package scenes

import (
	"encoding/json"
	"sync"

	"jetpack-joyride/pkg/storage"
)

type Achievement struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

var AllAchievements = []Achievement{
	{ID: "novice_flyer", Title: "Novice Flyer", Description: "Fly past 100 meters"},
	{ID: "coin_collector", Title: "Coin Collector", Description: "Collect 50 coins"},
	{ID: "shield_master", Title: "Shield Master", Description: "Acquire an energy shield"},
	{ID: "gravity_defier", Title: "Gravity Defier", Description: "Activate gravity swap"},
	{ID: "survivor_500m", Title: "Iron Man", Description: "Fly past 500 meters in a single run"},
}

var (
	achMu    sync.RWMutex
	unlocked = make(map[string]bool)
	loaded   = false
)

func LoadAchievements() {
	achMu.Lock()
	defer achMu.Unlock()
	if loaded {
		return
	}
	store := storage.GetStorage()
	data, err := store.Load("achievements")
	if err == nil {
		_ = json.Unmarshal(data, &unlocked)
	}
	loaded = true
}

func GetUnlockedAchievements() map[string]bool {
	LoadAchievements()
	achMu.RLock()
	defer achMu.RUnlock()
	res := make(map[string]bool, len(unlocked))
	for k, v := range unlocked {
		if v {
			res[k] = true
		}
	}
	return res
}

func UnlockAchievement(id string) {
	LoadAchievements()
	achMu.Lock()
	defer achMu.Unlock()
	if !unlocked[id] {
		unlocked[id] = true
		data, _ := json.Marshal(unlocked)
		_ = storage.GetStorage().Save("achievements", data)
	}
}

func CheckAndUnlockAchievements(distance float64, coins int, hasShield, hasGravity bool) {
	if distance >= 100.0 {
		UnlockAchievement("novice_flyer")
	}
	if distance >= 500.0 {
		UnlockAchievement("survivor_500m")
	}
	if coins >= 50 {
		UnlockAchievement("coin_collector")
	}
	if hasShield {
		UnlockAchievement("shield_master")
	}
	if hasGravity {
		UnlockAchievement("gravity_defier")
	}
}

func ResetAchievementsForTest() {
	achMu.Lock()
	defer achMu.Unlock()
	unlocked = make(map[string]bool)
	loaded = true
	_ = storage.GetStorage().Save("achievements", []byte("{}"))
}
