package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// NPC 鏡射進物件表(原版 `sub_1E74` 配置 / `sub_268` 釋放 / `sub_218` 記錄離場)
//
// 城鎮裡的東西分兩處放:`.OOL` 是大地圖的物件,而**寶箱、地上的物品、
// 甚至檀香木盒,全都在 `.NPC` 檔裡**,只是生物編號小於 0x40。
// 原版的處理是:每個「此刻在本層」的 NPC 都在物件表 `dword_3E46C` 佔一格,
// NPC 記錄的 +0xC(`word_3E77C`)記住配到哪一槽。
//
// 這條鏡射是 Get / Open 看得到它們的**唯一**原因 —— `sub_15A94` 掃的是物件表,
// 不是 NPC 表。少了它,城裡的寶箱開不了、檀香木盒撿不起來,
// 而症狀看起來會像「Get 壞了」。
//
// ⚠ **與原版的一處刻意差異**:原版連人也鏡射,而且畫面就是照物件表畫的;
// 本引擎的人另有一條算繪路徑(`VisibleNPCs`,帶朝向與逐格移動),
// 所以鏡射出來的槽在 `VisibleObjects` 會被跳過,免得同一格畫兩次。
// 鏡射本身照抄(人也配槽),因為 Get 的「掃到不可撿的就往下一槽」
// 這個行為要有那些槽在場才成立。

// syncNPCObjects 讓「這一層在場的 NPC」各佔一個場景物件槽。
//
// 每回合跑一次(NPC 會走動),換樓層時把不在本層的槽放掉。
func (s *State) syncNPCObjects() {
	if !s.InScene() || s.npcs == nil || s.Chamber != nil {
		return
	}
	if len(s.rtNPCs) != u5data.NPCsPerLocation {
		s.initRuntimeNPCs()
	}
	objs := s.currentObjects()
	if objs == nil {
		return
	}
	for i := 1; i < u5data.NPCsPerLocation; i++ {
		n := &s.npcs[i]
		rt := &s.rtNPCs[i]
		gone := !n.Present() || s.removed[s.Location<<8|i] || rt.Mode == ModeAbsent
		if gone || rt.Floor != s.Floor {
			s.releaseNPCObject(i)
			continue
		}
		slot := rt.ObjSlot
		if slot == 0 {
			slot = s.freeObjectSlot()
			if slot == 0 {
				continue // 三十二格滿了 —— 原版的配置函式同樣會配不到
			}
			rt.ObjSlot = slot
		}
		o := &objs.Objects[slot]
		o.Kind, o.Tile = rt.Creature, rt.Creature
		o.X, o.Y, o.Floor = rt.X, rt.Y, rt.Floor
		o.Raw[u5data.ObjKind], o.Raw[u5data.ObjTile] = rt.Creature, rt.Creature
		o.Raw[u5data.ObjX], o.Raw[u5data.ObjY] = byte(rt.X), byte(rt.Y)
		o.Raw[u5data.ObjFloor] = byte(rt.Floor)
		o.Raw[u5data.ObjQuality] = u5data.NPCObjectQuality(rt.Creature, s.Location, i)
		o.Raw[6], o.Raw[7] = 0, 0
	}
}

// releaseNPCObject 放掉某個 NPC 佔著的物件槽(原版 `sub_268` 的前半)。
func (s *State) releaseNPCObject(i int) {
	if i <= 0 || i >= len(s.rtNPCs) {
		return
	}
	rt := &s.rtNPCs[i]
	if rt.ObjSlot == 0 {
		return
	}
	if objs := s.currentObjects(); objs != nil && rt.ObjSlot < u5data.ObjectSlots {
		objs.Objects[rt.ObjSlot] = u5data.MapObject{}
	}
	rt.ObjSlot = 0
}

// freeObjectSlot 找一個空的物件槽(原版 `sub_2B57C`)。0 代表沒有。
func (s *State) freeObjectSlot() int {
	objs := s.currentObjects()
	if objs == nil {
		return 0
	}
	for i := 1; i < u5data.ObjectSlots; i++ {
		if !objs.Objects[i].Present() {
			return i
		}
	}
	return 0
}

// npcOfObject 回報某個物件槽是不是某個 NPC 的鏡射,是的話回傳槽號。
func (s *State) npcOfObject(slot int) (int, bool) {
	if slot <= 0 {
		return 0, false
	}
	for i := range s.rtNPCs {
		if s.rtNPCs[i].ObjSlot == slot {
			return i, true
		}
	}
	return 0, false
}

// takeNPCObject 是「地上那樣東西被撿走了」(原版 `sub_268` + `sub_218`)。
//
// 三件事一起做:放掉物件槽、這一次進場景不再出現、把位元寫進存檔的
// `dword_3E368[地點]`。少了最後一項,離開城堡再回來檀香木盒會躺在原地第二次。
func (s *State) takeNPCObject(i int) {
	if i <= 0 || i >= len(s.rtNPCs) {
		return
	}
	s.releaseNPCObject(i)
	s.markNPCRemoved(i)
	s.removeNPC(i)
}
