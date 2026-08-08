#!/usr/bin/env python3
"""XMI(AIL / Miles Sound System)→ 標準 MIDI。

推導見 `docs/formats/13`。XMI 與 SMF 差在三處,少處理任何一處都會得到
「音全部黏在一起」或「速度快十倍」的結果:

  1. **delta 時間是累加的裸位元組**,不是 SMF 的 VLQ:
     連續的 `0x00..0x7F` 各自貢獻自己的值,加起來才是這一步的等待;
     碰到 `>= 0x80` 就是狀態位元組,delta 結束。
     ⚠ 這**不是** VLQ —— VLQ 用最高位元當續接旗標,XMI 用「還是不是 < 0x80」。
  2. **沒有 note-off**:note-on 之後接 note、velocity,再接一個
     **SMF 風格的 VLQ 時長**。要自己排一個 note-off 到 (現在 + 時長)。
  3. **時基固定**:XMI 的一個 tick 是 1/120 秒(AIL 的硬體節拍)。
     輸出用 ppqn=60 + tempo 500,000 µs/四分音符 ⇒ 60 ticks / 0.5 s = 120 ticks/s。
     ⚠ 曲子裡的 `FF 51`(set tempo)meta 事件**照抄**——AIL 的驅動也是照著跑的。

用法:
  xmi2mid.py <in.XMI> <out.mid>
  xmi2mid.py --check <in.XMI>     只解析並印統計(不寫檔)
"""
import struct
import sys

# XMI 的一個 tick 是 1/120 秒 ⇒ 這一組讓輸出的 SMF 有同樣的實際速度。
PPQN = 60
DEFAULT_TEMPO = 500000  # µs / 四分音符 = 120 BPM


def chunks(raw: bytes, pos: int, end: int):
    """走一層 IFF chunk。回傳 (id, 內容起點, 內容長度, 下一個 chunk 的位置)。"""
    while pos + 8 <= end:
        cid = raw[pos : pos + 4]
        n = struct.unpack_from(">I", raw, pos + 4)[0]
        body = pos + 8
        nxt = body + n + (n & 1)  # IFF 的 chunk 補齊到偶數
        yield cid, body, n
        pos = nxt


def find_evnt(raw: bytes) -> tuple[bytes, bytes]:
    """從 XMI 容器裡取出 (TIMB, EVNT)。

    結構:FORM XDIR { INFO } ・ CAT_ XMID { FORM XMID { TIMB, [RBRN], EVNT } … }
    ⚠ 只取**第一首** —— 本專案的 15 個檔各只有一個 EVNT(已驗證)。
    多首的檔案會在這裡明確報錯,不會安靜只轉第一首。
    """
    timb = b""
    evnts = []

    def walk(pos: int, end: int) -> None:
        nonlocal timb
        for cid, body, n in chunks(raw, pos, end):
            if cid in (b"FORM", b"CAT "):
                # 前四個位元組是型別(XDIR / XMID),之後是子 chunk
                walk(body + 4, body + n)
            elif cid == b"TIMB":
                timb = raw[body : body + n]
            elif cid == b"EVNT":
                evnts.append(raw[body : body + n])

    walk(0, len(raw))
    if not evnts:
        raise ValueError("找不到 EVNT chunk —— 這不是 XMI?")
    if len(evnts) > 1:
        raise ValueError(f"這個檔有 {len(evnts)} 首序列,本工具只處理單首")
    return timb, evnts[0]


def read_vlq(buf: bytes, i: int) -> tuple[int, int]:
    """讀一個 SMF 風格的 VLQ(note-on 的時長用的是這種)。"""
    v = 0
    while i < len(buf):
        b = buf[i]
        i += 1
        v = (v << 7) | (b & 0x7F)
        if not b & 0x80:
            break
    return v, i


def write_vlq(v: int) -> bytes:
    out = bytearray([v & 0x7F])
    v >>= 7
    while v:
        out.append(0x80 | (v & 0x7F))
        v >>= 7
    return bytes(reversed(out))


def parse(evnt: bytes):
    """把 EVNT 解成 (絕對 tick, MIDI 位元組) 的清單。"""
    events: list[tuple[int, int, bytes]] = []  # (tick, 次序, 位元組)
    order = 0
    t = 0
    i = 0
    running = 0
    while i < len(evnt):
        b = evnt[i]
        # ── ① delta:連續的裸位元組累加
        if b < 0x80:
            while i < len(evnt) and evnt[i] < 0x80:
                t += evnt[i]
                i += 1
            continue
        # ── ② 狀態位元組
        if b == 0xFF:  # meta
            i += 1
            mtype = evnt[i]
            i += 1
            n, i = read_vlq(evnt, i)
            data = evnt[i : i + n]
            i += n
            events.append((t, order, bytes([0xFF, mtype]) + write_vlq(n) + data))
            order += 1
            if mtype == 0x2F:  # end of track
                break
            continue
        if b in (0xF0, 0xF7):  # sysex
            i += 1
            n, i = read_vlq(evnt, i)
            events.append((t, order, bytes([b]) + write_vlq(n) + evnt[i : i + n]))
            order += 1
            i += n
            continue
        status = b
        i += 1
        running = status
        kind = status & 0xF0
        if kind == 0x90:  # ★ note-on:多讀一個 VLQ 時長,並排一個 note-off
            note, vel = evnt[i], evnt[i + 1]
            i += 2
            dur, i = read_vlq(evnt, i)
            events.append((t, order, bytes([status, note, vel])))
            order += 1
            # ⚠ note-off 用 velocity 0 的 note-on 表示 —— 與原版驅動等價,
            # 而且省一個狀態位元組(SMF 的 running status 更容易接上)。
            events.append((t + dur, order, bytes([status, note, 0])))
            order += 1
            continue
        # 其餘:依 status 決定資料位元組數
        nbytes = 1 if kind in (0xC0, 0xD0) else 2
        events.append((t, order, bytes([status]) + evnt[i : i + nbytes]))
        order += 1
        i += nbytes
    _ = running
    events.sort(key=lambda e: (e[0], e[1]))
    return events


def build_mid(events) -> bytes:
    """把事件清單寫成單軌 SMF(format 0)。"""
    track = bytearray()
    # 開頭放一個 tempo,讓 ppqn=60 對上 1 tick = 1/120 秒。
    track += write_vlq(0) + bytes([0xFF, 0x51, 0x03]) + DEFAULT_TEMPO.to_bytes(3, "big")
    prev = 0
    saw_end = False
    for t, _order, data in events:
        if data[:2] == b"\xff\x2f":
            saw_end = True
            continue  # 收尾自己補,避免中途 end-of-track 把後面截掉
        track += write_vlq(t - prev) + data
        prev = t
    _ = saw_end
    track += write_vlq(0) + bytes([0xFF, 0x2F, 0x00])
    head = b"MThd" + struct.pack(">IHHH", 6, 0, 1, PPQN)
    return head + b"MTrk" + struct.pack(">I", len(track)) + bytes(track)


def main() -> int:
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    check = "--check" in sys.argv
    if not args or (not check and len(args) < 2):
        print(__doc__)
        return 2
    raw = open(args[0], "rb").read()
    timb, evnt = find_evnt(raw)
    events = parse(evnt)
    notes = sum(1 for _, _, d in events if d[0] & 0xF0 == 0x90 and d[2] != 0)
    last = events[-1][0] if events else 0
    secs = last / 120.0
    chans = sorted({d[0] & 0x0F for _, _, d in events if d[0] < 0xF0})
    progs = sorted({d[1] for _, _, d in events if d[0] & 0xF0 == 0xC0})
    print(
        f"{args[0]}: {notes} 音、{len(chans)} 聲道 {chans}、"
        f"音色 {progs}、TIMB {len(timb)} B、長度 {secs:.1f} 秒"
    )
    if check:
        return 0
    open(args[1], "wb").write(build_mid(events))
    print(f"  → {args[1]}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
