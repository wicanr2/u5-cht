package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSceneMapsRejectsWrongSize(t *testing.T) {
	if _, err := ParseSceneMaps(make([]byte, 1024)); err == nil {
		t.Error("大小不對的場景檔應該被拒絕")
	}
}

func TestParseSceneMapsSplitsCorrectly(t *testing.T) {
	raw := make([]byte, SceneFileSize)
	for i := range raw {
		raw[i] = byte(i % 251)
	}
	scenes, err := ParseSceneMaps(raw)
	if err != nil {
		t.Fatalf("解析失敗: %v", err)
	}
	if len(scenes) != ScenesPerFile {
		t.Fatalf("解出 %d 張,預期 %d", len(scenes), ScenesPerFile)
	}
	// 第 2 張的起點應該對上第 1024 個 byte
	if scenes[1].At(0, 0) != raw[SceneTiles] {
		t.Error("第二張場景的起點對不上")
	}
	if scenes[0].At(1, 0) != raw[1] || scenes[0].At(0, 1) != raw[SceneSide] {
		t.Error("列寬不是 32 —— 但反編譯的 sub_86C 用的是 byte_3F789[32*dy+dx]")
	}
	// 邊界
	if scenes[0].At(-1, 0) != 0 || scenes[0].At(SceneSide, 0) != 0 {
		t.Error("超出範圍應回 0")
	}
}

// TestLoadSceneMapsFromGameData 對四個場景檔實測。
func TestLoadSceneMapsFromGameData(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過")
	}
	for _, name := range []string{"TOWNE.DAT", "CASTLE.DAT", "KEEP.DAT", "DWELLING.DAT"} {
		scenes, err := LoadSceneMaps(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(scenes) != ScenesPerFile {
			t.Errorf("%s 解出 %d 張,預期 %d", name, len(scenes), ScenesPerFile)
		}
		// 每張地圖都該有內容,而且用到的 tile 種類要夠多
		// (切錯的話會是大片同值或雜訊)
		for i := range scenes {
			seen := map[byte]bool{}
			for _, v := range scenes[i].Tiles {
				seen[v] = true
			}
			if len(seen) < 5 {
				t.Errorf("%s 第 %d 張只用了 %d 種 tile,可能切錯", name, i, len(seen))
			}
		}
		t.Logf("%s:%d 張 %d×%d", name, len(scenes), SceneSide, SceneSide)
	}
}
