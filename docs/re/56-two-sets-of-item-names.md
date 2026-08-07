# 同一批東西有兩套名字 —— 而「特殊物品 158 項」是它的總長度

輸入檔:`gamedata/DATA.OVL`
IDA 位址:`.text:00040934`(FM Towns 的同一張指標表)、`sub_1A5E8`(U 指令)
日期:2026-08-08

---

## 0. `WORKLIST` 上那條標錯了

原本寫著:

> | 特殊物品 158 項 | `DATA.OVL` 0x04C3:`Magic Crpt / Skull Keys / Amulet / Crown / Sceptre` … | ⬜ |

0x04C3 是**字串**的位置(`strings` 掃出來的),不是表。而「158 項」也不是
「158 種特殊道具」—— 那是**整張指標表的長度**,裡面混著特殊道具、裝備的縮寫名、
角色狀態名等等。

照著那條去找「另外 136 種特殊道具」會找不到,因為它們不存在。

## 1. 找表的方法:反向查指標

字串位置好找(`strings` 就有),但**表**要反過來找:

```python
mc = DATA.OVL.find(b'Magic Crpt')          # 0x04C3
tgt = mc - 0x10                            # 指標表的 bias(既有的 ItemPointerBias)
[i for i in range(len(d)) if u16(d, i) == tgt]   # → [0x1946]
```

⇒ 指標表在 **0x1946**,而且全檔**只有一處**指向它 —— 沒有第二個候選要排除。

> 這比「從已知的表往後推位移」可靠:`ItemPointerTable` 在 0x1806、
> `CreaturePointerTable` 在 0x1866,兩者相鄰,很容易以為第三張表也緊接在後。
> 它不是(0x1946 與 0x1866 之間差 0xE0)。

## 2. ★ 兩張表指著同一批裝備,名字不同

| 表 | 筆數 | 例子 |
|---|---|---|
| 0x1806 | 48 | `Leather Helm` `Spiked Helm` `Small Shield` `Ring of Invisibility` |
| 0x1946 前 22 | 22 | `Magic Crpt` `Skull Keys` … `Wooden Box`(U 指令的特殊道具) |
| 0x1946 第 22 起 | 48 | `Leath Helm` `Spkd. Helm` `Sm. Shield`(**同一批裝備的縮寫名**) |

48 件裝備裡有 **28 件**的短名與長名不同(測試會印出來)。

⇒ 0x1946 不是「另一份裝備表」,是**背包 / 道具面板用的那一份** ——
U5 的面板很窄,長名字塞不進去。

★ 這也解釋了為什麼「特殊物品」與「裝備」會在同一張表裡:從 U 指令的角度
它們是**同一個清單**,索引連續。

## 3. 索引就是 U 指令的 case 編號 − 16

`docs/re/44` 當初用手抄出前 22 筆(`jumptable 0001A6DD case 16` 對上第 0 筆)。
現在整張表從玩家自己的檔案讀出來,兩者逐筆相同 —— 手抄那份是對的,
而現在它有了可重跑的來源。

驗證用三個錨:第 0 筆 `Magic Crpt`、第 21 筆 `Wooden Box`、第 22 筆 `Leath Helm`。
三個一起中,表就沒有滑動的餘地(`rulebook/62`)。

## 4. `(0`..`(7` 不是損毀

第 5..12 筆叫 `(0` `(1` … `(7`。看起來像資料壞了,其實是原版留的**佔位名** ——
它們在跳表裡走 default,根本用不到。

⚠ **不要試著「修好」它們**,也不要拿去查譯名(會得到一串問號)。
清單直接跳過:`SpecialItemPlaceholder()`。

## 5. 順手清掉兩條過期記載

| 項目 | 實際狀況 |
|---|---|
| **風向(航海)** | `game/wind.go` 早就實作:五種風、延遲表、**逆風完全動不了**、側風要兩三拍走一格(`CanSail`) |
| **魔法飾品** | 就是裝備 42..47(`Ring of Invisibility / Protection / Regeneration`、`Amulet/Turning`、`Spiked Collar`、`Ankh`),`ItemNames` 早就讀得到 |

## 6. 引擎對應

| 原版 | 引擎 |
|---|---|
| 0x1946 指標表 | `u5data.SpecialItemTable`(`Names[22]` + `EquipNames[48]`) |
| 索引 → case 編號 | `NameForUseCode(code)`(`code − 16`) |
| `(0`..`(7` | `SpecialItemPlaceholder` |
| 清單的中文 | `i18n.Name`(key 用原版的**縮寫**) |
