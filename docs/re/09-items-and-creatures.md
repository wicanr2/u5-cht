# 裝備表與生物名表

> 來源:DOS `DATA.OVL` 的兩張指標表(0x1806 / 0x1866),欄位語意以角色紀錄與
> `sub_B98` 的範圍判斷交叉驗證。

## 裝備:48 件,順序由指標表決定

`DATA.OVL` 裡的裝備名字**分成兩段**放:

- **0x0052 起**:Leather Helm、Spiked Helm、Small Shield…(28 筆)
- **0x175C 起**:Chain Coif、Iron Helm、Ring Mail、Dagger、Sling…(20 筆)

只讀其中一段會缺一半,而且缺的正是最常見的那些(長劍、十字弓、鎖甲頭巾)。

真正的順序在 **0x1806 的指標表**:48 個 u16,依裝備編號排好,**交錯**指向兩段字串。

⚠ 指標要 **加 0x10** 才是檔案位移:`0x0042 + 0x10 = 0x0052`("Leather Helm")、
`0x174C + 0x10 = 0x175C`("Chain Coif")。這 0x10 是 DOS overlay 的載入偏移。
忘了加會讀到字串中間,解出 `" Helm"`、`"elm"` 這種殘片 —— 看起來像編碼問題,
其實是位移問題。

| 編號 | 分類 |
|---|---|
| 0–3 | 頭盔(Leather / Chain Coif / Iron / Spiked) |
| 4–8 | 盾(Small / Large / Spiked / Magic / Jewel) |
| 9–15 | 護甲(Cloth / Leather / Ring Mail / Scale / Chain / Plate / Mystic) |
| 16–41 | 武器(Dagger … Mystic Sword) |
| 42–47 | 戒指與護符(Ring of Invisibility … Ankh) |

## 角色紀錄的裝備欄位

| 位移 | 欄位 |
|---|---|
| 0x17 | 留在旅店的天數(每天 +1,上限 25;0xFF = 沒在計數)|
| 0x19 | 頭盔 |
| 0x1A | 護甲 |
| 0x1B | 武器(右手) |
| 0x1C | 副手(盾或第二把武器) |
| 0x1D | 戒指 |
| 0x1E | 護符 |
| 0x1F | 留在旅店的標記(入隊時設 0) |

0x17 與 0x1F 的依據是位址:日期進位那段對 16 名角色跑 `byte_3DDCB[i*32]`
而 `0x3DDCB − 0x3DDB4 = 0x17`;入隊時寫 `byte_3DDD3[i*32]` 而差是 `0x1F`。

裝備欄位靠**橫向對照**確認 —— 每個人的配備都符合他的定位:

```
聖者      Chain Coif   Chain Mail      Long Sword    —              Ankh
Shamino   Leather Helm Ring Mail       Short Sword   Small Shield
Iolo      Leather Helm Leather Armour  Main Gauche   Short Sword     ← 雙手各一把
Mariah    —            Cloth Armour    Dagger        —               ← 法師
Geoffrey  Spiked Helm  Scale Mail      Mace          Spiked Shield   Spiked Collar
```

## 生物名:索引 =(編號 − 64)/ 4

緊接在裝備指標表後面(0x1866)是另一張 48 個 u16 的指標表:
Mage / Bard / Fighter / Avatar / Villager / Merchant / … / Guard / Blackthorn /
Lord British / Sea Horse / … / Shadow Lord。

人物圖以四張(四個朝向)為一組,所以生物編號是 4 的倍數;減 64 是因為表從編號 64 開始。

```
64 Mage   68 Bard   80 Villager   84 Merchant   104 Child   112 Guard
```

這個公式與 `sub_B98`(平民被攻擊會嚇到)的範圍 `0x40 <= 編號 < 0x74` 互相印證:
換算成索引正好是 **0..11**,**剛好排除**衛兵(12)、遊民(13)、黑刺(14)、
不列顛王(15)—— 這幾種本來就不該被嚇跑。

⚠ 全遊戲 325 個 NPC 裡有 **29 個查不到名字**:編號 < 64 的物件與動物
(tile 落在 256–319 那一列),以及兩隻編號不是 4 的倍數的怪物。
原版資料就這樣,不是解析漏掉。

## 驗收

- 48 件裝備全數有名字,兩段字串的交錯點(編號 0 與 1、9 與 11)都對得上
- 五名初始角色的裝備逐欄比對相符
- 296/325 個 NPC 查得到生物名,查不到的全部落在已知的兩種例外裡
