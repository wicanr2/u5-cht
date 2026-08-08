#!/usr/bin/env python3
"""把 FM Towns 的 15 首 `.EUP` 用**原版音色**(`FM_BANK.FMB`)離線渲染成 ogg。

格式的推導與證據見 `docs/formats/12-eup-and-fmb.md`,整條鏈路見 `docs/audio-pipeline.md`。

  音序     `.EUP`         音符(聲道 / 小節內 tick / 音高 / 力度)、速度、小節數
  音色     `FM_BANK.FMB`  128 筆 4-op 參數(DT/MUL・TL・KS/AR・DR・SR・SL/RR・FB/ALG)
  音量     `U5_BGM.TBL`   每首六個 FM 聲道的起始音量

★ 分清兩件事:**合成器(這支程式)是我們寫的,音色(資料)是原版的。**
  寫合成器不違反 `CLAUDE.md §3.0`;自己編一組音色參數才違反。

⚠ 這**不是** YM2612 的精確模擬。精確模擬需要原版的包絡速率表,而那張表在
  `TBIOS.BIN` / 晶片裡(`docs/re/89`)。近似的部分逐項標在 `ENV_*` 與 `render_note`。
  音高、時值、音色參數、聲道音量**全部來自原版資料**;近似的只有
  「這些參數怎麼變成波形」。

用法(全程 docker,見 tools/render_music.sh):
  eup2ogg.py <U5_E 目錄> <輸出目錄> [--rate 44100] [--only M1.EUP]
"""
import argparse
import math
import os
import sys
import wave

import numpy as np

# ── 格式常數(全部有出處,見 docs/formats/12)────────────────────────────────
SIG = b"EUPHONY "
REC = 6                 # 一筆記錄 6 byte
HDR_AFTER_SIG = 16      # 簽章 + 8 byte 小檔頭
TEMPO_OFF = 15          # 速度欄相對簽章的位移
TICKS_PER_BAR = 384     # 0xF2 的時間欄固定值 = 96 ppqn × 4/4
TICKS_PER_QUARTER = 96  # eupplayer 的 60e6/(96*tempo)
BAR_END = 0xF2
FM_CHANNELS = 6
VOICE_SIZE = 48
VOICE_COUNT = 128
BANK_HEADER = 8

# ── 近似的部分(逐項標明)────────────────────────────────────────────────────
#
# ⚠ 以下三個常數**不是**從原版讀出來的,是為了讓包絡的時間尺度落在
#   「聽起來像 FM 音源」的範圍而選的。原版的速率表在晶片 / BIOS 裡。
#   改這三個值會改變音色的攻擊與衰減感,但不會改變音高與時值。
ENV_DB_FLOOR = 96.0     # 包絡的動態範圍(dB),YM2612 的衰減暫存器是 10 bit
ENV_RATE_BASE = 1.0     # 速率 r 的 dB/s = ENV_RATE_BASE * 2**(r/2)
ENV_MIN_RATE = 1        # r == 0 視為「永不衰減」

# TL 是 0.75 dB 一階(這一條是 YM2612 規格,不是近似)
TL_DB_STEP = 0.75
# SL 是 3 dB 一階;SL == 15 視為全衰減
SL_DB_STEP = 3.0

# MUL == 0 代表 ×0.5(YM2612 規格)
MUL_HALF = 0.5

# ⚠⚠ **DT(detune)不實作** —— 這是量測出來的決定,不是懶。
#
# 第一版用「每階 6 音分」的比例近似,結果**整首歌的音高系統性偏高 1%**:
# 頻譜峰值 132.0 Hz 對 C3 的 130.8、66.0 對 65.4。查出來是 M8 有五個聲道的
# **載波** DT = 1 或 3(+6 ~ +18 音分),而全庫**有 25 個音色的載波 DT = 7**
# ⇒ 那些會偏 +42 音分 = +2.5%,明顯走音。
#
# 真實的 YM2612 DT1 是查一張**依 key code 索引的頻率偏移表**(單位是 Hz 不是比例),
# 低音時偏移很小。那張表在晶片裡,我們沒有一手來源(`docs/re/89`)。
# ⇒ 與其用一個發明的量級把音高弄歪,不如**不做**:音高因此完全正確,
# 只是少了 DT 帶來的細微加厚感。⬜ 記在 `docs/audio-pipeline.md`。
DT_ENABLED = False

# YM2612 八種演算法的載波(輸出被聽到的運算子)與調變連線。
# conn[i] = 餵給運算子 i 的來源清單(運算子索引),空 = 只有自己的相位。
ALGORITHMS = {
    0: dict(conn={1: [0], 2: [1], 3: [2]}, carriers=[3]),
    1: dict(conn={2: [0, 1], 3: [2]}, carriers=[3]),
    2: dict(conn={2: [1], 3: [0, 2]}, carriers=[3]),
    3: dict(conn={1: [0], 3: [1, 2]}, carriers=[3]),
    4: dict(conn={1: [0], 3: [2]}, carriers=[1, 3]),
    5: dict(conn={1: [0], 2: [0], 3: [0]}, carriers=[1, 2, 3]),
    6: dict(conn={1: [0]}, carriers=[1, 2, 3]),
    7: dict(conn={}, carriers=[0, 1, 2, 3]),
}


def db_to_gain(db):
    return 10.0 ** (-db / 20.0)


class Voice:
    """一筆 FM 音色。位移的兩個獨立來源見 docs/formats/12 §3。"""

    def __init__(self, raw):
        self.name = raw[:8].rstrip(b"\x00 ").decode("latin1", "replace")
        self.op = []
        for n in range(4):
            self.op.append(dict(
                dt=(raw[8 + n] >> 4) & 7, mul=raw[8 + n] & 0x0F,
                tl=raw[12 + n] & 0x7F,
                ks=(raw[16 + n] >> 6) & 3, ar=raw[16 + n] & 0x1F,
                dr=raw[20 + n] & 0x1F,
                sr=raw[24 + n] & 0x1F,
                sl=(raw[28 + n] >> 4) & 0x0F, rr=raw[28 + n] & 0x0F,
            ))
        self.alg = raw[32] & 7
        self.fb = (raw[32] >> 3) & 7
        self.pan = raw[33]


def load_bank(path):
    raw = open(path, "rb").read()
    want = BANK_HEADER + VOICE_COUNT * VOICE_SIZE
    if len(raw) != want:
        sys.exit(f"{path}:大小 {len(raw)},預期 {want}")
    return [Voice(raw[BANK_HEADER + i * VOICE_SIZE:][:VOICE_SIZE]) for i in range(VOICE_COUNT)]


def load_bgm_table(path):
    """回傳 [(檔名, [六個聲道音量])]。⚠ 檔尾有 DOS 的 0x1A。"""
    text = open(path, "rb").read().split(b"\x1a")[0].decode("latin1")
    out = []
    for line in text.splitlines():
        f = line.split()
        if not f:
            continue
        if len(f) != 1 + FM_CHANNELS:
            sys.exit(f"{path}:一行有 {len(f)} 欄,預期 {1 + FM_CHANNELS}")
        out.append((f[0], [int(x) for x in f[1:]]))
    return out


def parse_eup(path):
    """回傳 (notes, programs, tempo, bars)。notes 的元素是 (聲道, 絕對tick, 音高, 力度)。"""
    d = open(path, "rb").read()
    sig = d.find(SIG)
    if sig < 0:
        sys.exit(f"{path}:找不到簽章")
    tempo = d[sig + TEMPO_OFF]
    notes, programs, bars = [], {}, 0
    i = sig + HDR_AFTER_SIG
    while i + REC <= len(d):
        r = d[i:i + REC]
        if r[0] == 0xFF:
            break
        hi = r[0] >> 4
        if hi == 0x9:
            # ★ 音符佔 12 byte:0x9n 前半 + 0x8n 後半(docs/formats/12 §2.1)
            if i + 2 * REC > len(d) or (d[i + REC] >> 4) != 0x8:
                sys.exit(f"{path}:位移 0x{i:X} 的音符後半不是 0x8n")
            notes.append((r[0] & 0x0F, bars * TICKS_PER_BAR + r[2] + 0x80 * r[3], r[4], r[5]))
            i += 2 * REC
            continue
        if hi == 0xC:
            programs[r[0] & 0x0F] = r[4]
        elif r[0] == BAR_END:
            bars += 1
        i += REC
    return notes, programs, tempo, bars


def env_rate_db_per_sec(r):
    if r <= 0:
        return 0.0
    return ENV_RATE_BASE * (2.0 ** (r / 2.0))


def op_envelope(op, n_on, n_off, rate):
    """算一個運算子的包絡(線性增益陣列),長度 n_on + n_off。

    ⚠ 近似:YM2612 的包絡在 dB 域以整數速率遞增衰減暫存器,速率表在晶片裡。
    這裡用「dB/s = 2**(r/2)」的連續近似,四個階段照 AR → DR → SR → RR。
    """
    total = n_on + n_off
    att = np.full(total, ENV_DB_FLOOR)  # 衰減量(dB),越大越安靜
    t = np.arange(total) / rate

    ar = env_rate_db_per_sec(op["ar"])
    dr = env_rate_db_per_sec(op["dr"])
    sr = env_rate_db_per_sec(op["sr"])
    rr = env_rate_db_per_sec(max(op["rr"] * 2, ENV_MIN_RATE))  # RR 是 4 bit ⇒ ×2 對齊 5 bit
    sl_db = ENV_DB_FLOOR if op["sl"] >= 15 else op["sl"] * SL_DB_STEP

    # 起音:從 FLOOR 降到 0
    if ar <= 0:
        att[:] = ENV_DB_FLOOR
        return db_to_gain(att)
    t_att = ENV_DB_FLOOR / ar
    n_att = min(int(t_att * rate) + 1, n_on)
    att[:n_att] = np.maximum(ENV_DB_FLOOR - ar * t[:n_att], 0.0)

    # 衰減到 SL
    idx = n_att
    if idx < n_on:
        tt = t[idx:n_on] - t[idx]
        d = dr * tt if dr > 0 else np.zeros_like(tt)
        att[idx:n_on] = np.minimum(d, sl_db)
        # 到達 SL 之後轉 SR(第二段衰減)
        if dr > 0:
            reached = np.searchsorted(d, sl_db)
            j = idx + reached
            if j < n_on:
                tt2 = t[j:n_on] - t[j]
                att[j:n_on] = np.minimum(sl_db + sr * tt2, ENV_DB_FLOOR)
        else:
            att[idx:n_on] = np.minimum(sl_db + sr * tt, ENV_DB_FLOOR)

    # 放音
    if n_off > 0:
        start = att[n_on - 1] if n_on > 0 else ENV_DB_FLOOR
        tt = t[n_on:] - t[n_on]
        att[n_on:] = np.minimum(start + rr * tt, ENV_DB_FLOOR)

    return db_to_gain(np.minimum(att, ENV_DB_FLOOR))


def render_note(voice, midi, vel, n_on, n_off, rate):
    """把一個音符渲染成單聲道波形(float32)。

    ⚠ 近似:回饋(FB)用「上一個樣本」的一階近似,不是晶片的兩段平均;
    LFO(AMS/PMS)完全沒做。這兩者都在 `+33` 那一格裡,而它的低六位是
    AMS/PMS —— ⬜ 未實作,已記在 docs/audio-pipeline.md。
    """
    total = n_on + n_off
    if total <= 0:
        return np.zeros(0, dtype=np.float32)
    f0 = 440.0 * 2.0 ** ((midi - 69) / 12.0)
    alg = ALGORITHMS[voice.alg]
    t = np.arange(total) / rate

    phase, envs = {}, {}
    for n in range(4):
        op = voice.op[n]
        mul = MUL_HALF if op["mul"] == 0 else float(op["mul"])
        # ⚠ DT 不套用(見 DT_ENABLED 的說明:發明的量級會把音高弄歪 1..2.5%)。
        phase[n] = 2.0 * math.pi * f0 * mul * t
        envs[n] = op_envelope(op, n_on, n_off, rate)

    # 力度:原版的力度絕大多數是 0x40。當成 TL 之外的線性縮放。
    vgain = max(vel, 1) / 127.0

    out = {}
    for n in range(4):
        src = alg["conn"].get(n, [])
        mod = np.zeros(total) if not src else sum(out[s] for s in src)
        if n == 0 and voice.fb > 0:
            # ⚠ 回饋的近似:**一次不動點迭代**(用未回饋的訊號當調變),
            # 不是晶片的「前兩個樣本平均」遞迴。
            #
            # 為什麼不做真遞迴:那是逐樣本的相依運算,Python 迴圈跑不完
            # (十五首合計 23 分鐘 × 44.1 kHz × 每音符四個運算子)。
            # 一次迭代抓到回饋「讓波形變亮」的主要效果,但不會產生真遞迴的
            # 混沌成分 ⇒ 回饋值大(FB 6-7)的音色會比原版乾淨。⬜ 記在文件裡。
            fb_scale = (2.0 ** voice.fb) / 64.0
            ph, ev = phase[0], envs[0]
            tl = db_to_gain(voice.op[0]["tl"] * TL_DB_STEP)
            out[0] = np.sin(ph + fb_scale * np.sin(ph)) * ev * tl
            continue
        tl = db_to_gain(voice.op[n]["tl"] * TL_DB_STEP)
        out[n] = np.sin(phase[n] + mod * math.pi) * envs[n] * tl

    mix = sum(out[c] for c in alg["carriers"]) / len(alg["carriers"])
    return (mix * vgain).astype(np.float32)


def render_song(path, bank, volumes, rate, release_s=0.25):
    notes, programs, tempo, bars = parse_eup(path)
    if tempo <= 0 or bars <= 0:
        sys.exit(f"{path}:速度 {tempo} / 小節 {bars} 不合理")
    tick_s = 60.0 / (TICKS_PER_QUARTER * tempo)
    dur_s = bars * TICKS_PER_BAR * tick_s
    n_total = int((dur_s + release_s) * rate) + 1
    buf = np.zeros(n_total, dtype=np.float32)

    # ★ 音長:**同聲道的下一個音符**必然結束前一個 —— YM2612 每個 FM 聲道
    # 同時只有一個音。這是硬體推導,不是猜的。
    # ⬜ 明確的 gate 欄位沒定案(後半 5 個位元組語意未定,docs/audio-pipeline §3.5):
    # 四個候選公式都給不出音樂性的值(43/91/92 這種),所以不選。
    by_ch = {c: [] for c in range(FM_CHANNELS)}
    for ch, tk, midi, vel in notes:
        if ch < FM_CHANNELS:
            by_ch[ch].append((tk, midi, vel))

    for ch, seq in by_ch.items():
        if not seq:
            continue
        v = programs.get(ch)
        if v is None or v >= len(bank):
            # ⬜ M4.EUP 的聲道 4 真的沒有 program change。沒有依據可循 ⇒ 跳過並說明,
            # 不要拿 0 號音色湊(那是自創)。
            print(f"    ⚠ 聲道 {ch} 有 {len(seq)} 個音符但沒選音色 ⇒ 跳過(見 audio-pipeline §3.5)")
            continue
        voice = bank[v]
        # 聲道音量(表裡是 0..127)
        cg = min(volumes[ch], 127) / 127.0
        for i, (tk, midi, vel) in enumerate(seq):
            end = seq[i + 1][0] if i + 1 < len(seq) else bars * TICKS_PER_BAR
            n_on = max(int((end - tk) * tick_s * rate), 1)
            n_off = int(release_s * rate)
            w = render_note(voice, midi, vel, n_on, n_off, rate)
            s0 = int(tk * tick_s * rate)
            s1 = min(s0 + len(w), n_total)
            if s1 > s0:
                buf[s0:s1] += w[:s1 - s0] * cg
    return buf, tempo, bars, dur_s, len(notes)


def write_wav(path, mono, rate):
    peak = float(np.max(np.abs(mono))) if mono.size else 0.0
    if peak > 0:
        mono = mono / peak * 0.89   # 留一點餘裕,避免 ogg 編碼削波
    pcm = (mono * 32767).astype("<i2")
    with wave.open(path, "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(rate)
        w.writeframes(pcm.tobytes())


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("u5e")
    ap.add_argument("out")
    ap.add_argument("--rate", type=int, default=44100)
    ap.add_argument("--only", default="")
    a = ap.parse_args()

    bank = load_bank(os.path.join(a.u5e, "FM_BANK.FMB"))
    table = load_bgm_table(os.path.join(a.u5e, "U5_BGM.TBL"))
    os.makedirs(a.out, exist_ok=True)
    named = sum(1 for v in bank if v.name)
    print(f"音色庫 {len(bank)} 筆({named} 筆有名字);曲目表 {len(table)} 首")

    total_err = 0
    for idx, (fn, vols) in enumerate(table):
        if a.only and fn != a.only:
            continue
        src = os.path.join(a.u5e, fn)
        buf, tempo, bars, dur_s, n_notes = render_song(src, bank, vols, a.rate)
        wav = os.path.join(a.out, os.path.splitext(fn)[0] + ".wav")
        write_wav(wav, buf, a.rate)
        got = len(buf) / a.rate
        # ★★ 硬驗收:渲染出的長度必須與公式算出來的秒數相符。
        if abs(got - dur_s) > 0.5:
            print(f"  ✗ {fn}:渲染 {got:.2f}s 與公式 {dur_s:.2f}s 差太多")
            total_err += 1
        print(f"  曲{idx:2d} {fn:<10} {tempo:3d}BPM {bars:3d}小節 {n_notes:4d}音符 "
              f"→ {os.path.basename(wav)} {dur_s:6.1f}s")
    return 1 if total_err else 0


if __name__ == "__main__":
    sys.exit(main())
