package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 在城裡跟 NPC 打起來(原版 `sub_C74`)
//
//	sub_218(npc)       ★ 只有某些生物會被**永久**記成「這一格已經清掉了」
//	sub_2E58C(物件槽)   開打 —— 與撞上野外怪物走同一支
//	sub_268(npc)       從場上抹掉
//	sub_5C8(0)         重載地圖
//
// 引擎的場景 NPC 與大地圖物件是兩套資料,所以這裡合成一個等價的物件
// 餵給既有的 `BeginCombat` —— 戰鬥本身(選圖、排隊伍、生怪、回合)完全共用。

// beginNPCCombat 與場景裡的一個 NPC 開打。回傳有沒有真的打起來。
func (s *State) beginNPCCombat(i int) bool {
	if s.npcs == nil || i < 0 || i >= len(s.npcs) {
		return false
	}
	n := &s.npcs[i]
	rt := &s.rtNPCs[i]
	// 合成一個與這名 NPC 等價的物件:種類碼就是生物編號,位置就是他站的地方。
	//
	// ⚠ `Raw[ObjShipHull]` 留 0 —— 那個位元組的最高位元是「大版本生物」旗標,
	// 設起來會讓城鎮裡「只打得到眼前那一個」的規則失效(見 spawnEnemies)。
	obj := u5data.MapObject{Kind: n.Creature}
	obj.X, obj.Y, obj.Floor = rt.X, rt.Y, rt.Floor

	// ★ 順序照原版:先記帳、再開打、最後才把人抹掉。
	// 反過來的話,`sub_218` 要用的生物編號已經被清成 0 了。
	if u5data.NPCKillIsPermanent(n.Creature) {
		s.markNPCRemoved(i)
	}
	ok := s.beginCombatWith(&obj)
	s.removeNPC(i)
	return ok
}

// markNPCRemoved 把「這一格的第 i 個 NPC 已經清掉了」記進**存檔**。
//
// 與 `removeNPC` 不同:那一支只影響這一次進場景(離場再回來就復原),
// 這一支寫的是 `dword_3E368[地點]` 的位元,存檔帶得走。
func (s *State) markNPCRemoved(i int) {
	if s.Location < 1 || s.Location > len(s.RemovedNPC) {
		return
	}
	s.RemovedNPC[s.Location-1] |= 1 << uint(i)
}

// applyRemovedNPCs 重建這一次進場景時「誰不在」的名單。
//
// ⚠ **先清乾淨再套存檔的位元。** `removed` 這個 map 有兩種來源:
// 存檔裡的永久移除,以及這一次進場景中途被抹掉的人(打死的衛兵、
// 用碎片消滅的暗影君主)。原版的後者放在**每次進場景重新從 `.NPC` 載入**
// 的暫存表裡,離場就沒了 —— 不清的話,衛兵死一次就再也不會補上。
func (s *State) applyRemovedNPCs() {
	if s.Location < 1 || s.Location > len(s.RemovedNPC) {
		return
	}
	if s.removed == nil {
		s.removed = map[int]bool{}
	}
	for i := 0; i < u5data.NPCsPerLocation; i++ {
		delete(s.removed, s.Location<<8|i)
	}
	mask := s.RemovedNPC[s.Location-1]
	for i := 0; i < u5data.NPCsPerLocation; i++ {
		if mask&(1<<uint(i)) != 0 {
			s.removed[s.Location<<8|i] = true
		}
	}
}
