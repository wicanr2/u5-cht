package game

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 存檔
//
// 寫回的格式與原版完全相同(`sub_284CC` 與讀取 `sub_27D24` 逐欄位對稱),
// 所以存出來的 `SAVED.GAM` / `SAVED.OOL` 拿回原版 DOS 版也讀得起來。
//
// **[HARD] 絕不寫進 cwd。** 存檔一律放 `os.UserConfigDir()` 底下 ——
// u1-cht 踩過這個坑:遊戲裝在唯讀目錄時存不進去,而且玩家的原版資料目錄
// 不該被引擎污染。

// SaveDirName 是設定目錄底下的資料夾名。
const SaveDirName = "u5cht"

// 存檔的兩個檔名,與原版一致。
const (
	SaveGameFile    = "SAVED.GAM"
	SaveObjectsFile = "SAVED.OOL"
)

// SaveDir 回傳存檔目錄,必要時建好。
func SaveDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("找不到使用者設定目錄:%w", err)
	}
	dir := filepath.Join(base, SaveDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// ExportSave 把目前的遊戲狀態寫回一份 Save。
//
// base 是**讀進來的那份存檔**,當作底稿 —— 引擎還沒實作的欄位(魔法、
// 任務旗標、地牢狀態…)原樣保留。沒有底稿就不能存,因為把未解欄位清成 0
// 會讓存檔在原版裡壞掉。
func (s *State) ExportSave(base *u5data.Save) (*u5data.Save, error) {
	if base == nil || len(base.Raw) != u5data.SaveFileSize {
		return nil, fmt.Errorf("沒有可用的底稿存檔 —— 引擎只解出部分欄位," +
			"必須以讀進來的原版存檔為底才存得出完整的檔")
	}
	out := *base
	out.Raw = append([]byte(nil), base.Raw...)

	copy(out.Roster[:], s.Roster)
	out.PartySize = s.PartySize
	out.Inventory = s.Inventory
	out.Year, out.Month, out.Day = s.Clock.Year, s.Clock.Month, s.Clock.Day
	out.Hour, out.Minute = s.Clock.Hour, s.Clock.Minute
	out.Karma = s.Karma
	out.Transport = s.Transport
	out.Location = s.Location
	out.Floor = s.Floor
	out.X, out.Y = s.X, s.Y
	out.ShrineQuest = s.ShrineQuestActive
	out.CodexLearned = s.ShrineQuestGiven
	out.ShrineFlag = s.ShrineFlag
	out.DungeonSeal = s.DungeonSeal
	out.ShadowlordAt = s.ShadowlordAt
	out.ShadowlordHere = s.ShadowlordHere

	// 在場景裡存檔時,X/Y 是**場景座標** —— 原版共用同一對變數,讀檔時
	// 靠 Location 判斷該怎麼解讀,所以照原樣寫回就好(見 state.go 的 LoadFrom)。
	return &out, nil
}

// Save 是玩家按下存檔鍵時走的路徑:用讀進來的那份當底稿寫回設定目錄。
func (s *State) Save() {
	dir, err := s.WriteSave(s.BaseSave)
	if err != nil {
		s.Log("存檔失敗:" + err.Error())
		return
	}
	s.Log("已存檔於 " + dir + "。")
}

// WriteSave 把遊戲狀態存到設定目錄。回傳實際寫到哪。
func (s *State) WriteSave(base *u5data.Save) (string, error) {
	sv, err := s.ExportSave(base)
	if err != nil {
		return "", err
	}
	blob, err := sv.Encode()
	if err != nil {
		return "", err
	}
	dir, err := SaveDir()
	if err != nil {
		return "", err
	}
	// 先寫暫存檔再 rename —— 中途失敗不會留下半份壞檔把舊存檔蓋掉。
	if err := writeAtomic(filepath.Join(dir, SaveGameFile), blob); err != nil {
		return "", err
	}
	objs := u5data.EncodeSaveObjects(s.Objects, s.UnderObjects)
	if err := writeAtomic(filepath.Join(dir, SaveObjectsFile), objs); err != nil {
		return "", err
	}
	return dir, nil
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// FindSave 依序找存檔:`explicit` 指定的 → 設定目錄的進度 → 遊戲資料目錄的原版存檔。
//
// 回傳的 base 一定是讀得起來的完整存檔,可以直接餵給 LoadFrom 與 ExportSave。
// 同時回傳同一份進度的物件表(`SAVED.OOL`);沒有的話兩者都是 nil,
// 呼叫端要退回 `BRIT.OOL` / `UNDER.OOL`。
func FindSave(gameData, explicit string) (sv *u5data.Save, from string, err error) {
	try := func(p string) (*u5data.Save, bool) {
		if p == "" {
			return nil, false
		}
		v, e := u5data.LoadSave(p)
		return v, e == nil
	}
	if v, ok := try(explicit); ok {
		return v, explicit, nil
	}
	if dir, e := SaveDir(); e == nil {
		p := filepath.Join(dir, SaveGameFile)
		if v, ok := try(p); ok {
			return v, p, nil
		}
	}
	for _, name := range []string{"SAVED.GAM", "INIT.GAM"} {
		p := filepath.Join(gameData, name)
		if v, ok := try(p); ok {
			return v, p, nil
		}
	}
	return nil, "", fmt.Errorf("在 %s 與設定目錄都找不到可讀的存檔", gameData)
}

// FindSaveObjects 找與進度同一份的物件表(`SAVED.OOL`)。
//
// 找不到就回 nil —— 呼叫端該退回世界預設的 `BRIT.OOL` / `UNDER.OOL`。
func FindSaveObjects() (surface, under *u5data.ObjectSet) {
	dir, err := SaveDir()
	if err != nil {
		return nil, nil
	}
	s, u, err := u5data.LoadSaveObjects(filepath.Join(dir, SaveObjectsFile))
	if err != nil {
		return nil, nil
	}
	return s, u
}
