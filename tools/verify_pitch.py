#!/usr/bin/env python3
"""驗收合成器的音高映射:**單音隔離**,不做全曲頻譜比對。

★ 為什麼要有這條:曲長對得上公式只證明**時間軸**對,不證明**音高**對。
  本專案真的踩過「整首高一個八度」而檢查回報 ✓(它對到了錯的音)。

⚠⚠ **前三版都用全曲頻譜比對,全部失敗**,值得記下來:

    第一版 「該頻率 ±20 音分有峰嗎」+ 鄰域基準  → 2/8。失敗原因是取「全域前 25 峰」
            當候選,短音符根本不在裡面 —— 量測錯,不是合成錯。
    第二版 換成全頻譜中位當基準              → 15/15 全綠,**但反對照也全綠**
            ⇒ 毫無鑑別力(FM 的泛音在 2f 本來就有能量)。
    第三版 樣板比對(比較移調 0/±5/±7/±12)  → 加泛音前 5 首紅、加泛音後 8 首紅,
            而反對照反而有一首綠。**三次迭代都在換哪些曲子紅,沒有收斂。**

⇒ `rulebook/41`:特例越補越多就該重想架構。全曲頻譜比對**是錯的工具** ——
  六個聲道的複音會互相干擾,而我要驗的只有「音高映射」這一件事。
  改成**渲染單一音符**:沒有複音、沒有泛音歧義,基頻一目了然。

用法:
  verify_pitch.py            # 正常驗收
  verify_pitch.py --control  # 反對照:故意把頻率 ×2,檢查驗證器會紅
"""
import os
import sys

import numpy as np

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from eup2ogg import ALGORITHMS, MUL_HALF, Voice, render_note

RATE = 44100
DUR_S = 1.5
CENTS_MAX = 10.0   # 單音的基頻允許偏差


def make_voice(alg, carrier_mul, ar=31, tl=0):
    """造一個最單純的測試音色:載波全開、AR 最快、不衰減。

    ⚠ 這是**測試夾具**,不是原版音色 —— 目的是隔離「音高映射」這一個變數。
    原版音色照樣走同一條 `render_note`,只是它們的包絡與調變會讓基頻不那麼乾淨。
    """
    raw = bytearray(48)
    raw[0:8] = b"TEST\x00\x00\x00\x00"
    for n in range(4):
        raw[8 + n] = carrier_mul if n in ALGORITHMS[alg]["carriers"] else 1
        raw[12 + n] = tl if n in ALGORITHMS[alg]["carriers"] else 127  # 調變器全靜音
        raw[16 + n] = ar
        raw[20 + n] = 0        # DR 0 = 不衰減
        raw[24 + n] = 0
        raw[28 + n] = 0x0F     # SL 0、RR 15
    raw[32] = alg              # FB 0
    raw[33] = 0xC0
    return Voice(bytes(raw))


def fundamental(w, rate):
    """量最強的譜線頻率(拋物線內插到子 bin 精度)。"""
    S = np.abs(np.fft.rfft(w * np.hanning(len(w))))
    f = np.fft.rfftfreq(len(w), 1 / rate)
    i = int(np.argmax(S))
    if 0 < i < len(S) - 1:
        a, b, c = S[i - 1], S[i], S[i + 1]
        d = a - 2 * b + c
        if d != 0:
            i = i + 0.5 * (a - c) / d
    return float(np.interp(i, np.arange(len(f)), f))


def main():
    control = "--control" in sys.argv
    factor = 2.0 if control else 1.0
    if control:
        print("== 反對照:期望頻率 ×2(假裝錯一個八度)—— 驗證器應該全紅")

    cases = []
    # 八種演算法 × 三種載波 MUL × 五個音高
    for alg in range(8):
        for mul in (1, 2, 4):
            for midi in (45, 57, 69, 76, 81):
                cases.append((alg, mul, midi))

    bad = 0
    worst = 0.0
    for alg, mul, midi in cases:
        v = make_voice(alg, mul)
        n_on = int(DUR_S * RATE)
        w = render_note(v, midi, 0x40, n_on, 0, RATE)
        if not np.any(np.abs(w) > 1e-6):
            print(f"  ✗ ALG={alg} MUL={mul} 音高={midi}:完全沒有輸出")
            bad += 1
            continue
        m = float(np.max(np.abs(w)))
        got = fundamental(w[: n_on], RATE)
        want = 440.0 * 2 ** ((midi - 69) / 12.0) * (MUL_HALF if mul == 0 else mul) * factor
        cents = 1200 * np.log2(got / want)
        worst = max(worst, abs(cents))
        if abs(cents) > CENTS_MAX:
            bad += 1
            if bad <= 6:
                print(f"  ✗ ALG={alg} MUL={mul} 音高={midi}:{got:.2f} Hz,"
                      f"預期 {want:.2f}({cents:+.1f} 音分,峰值 {m:.3f})")

    print(f"\n{len(cases) - bad}/{len(cases)} 個單音的基頻在 ±{CENTS_MAX} 音分內"
          f"(最大偏差 {worst:.1f} 音分)")
    if control:
        print("反對照:" + ("✓ 如預期變紅" if bad else "✗✗ 反對照竟然通過 —— 驗證器沒有鑑別力"))
        return 0 if bad else 1
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
