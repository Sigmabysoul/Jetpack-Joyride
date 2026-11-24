package scenes_test

import (
	"testing"

	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/entities"
	"jetpack-joyride/pkg/scenes"
)

func TestSceneCreationAndState(t *testing.T) {
	// 1. Home scene
	home := scenes.NewHomeScene()
	if home == nil {
		t.Fatalf("expected NewHomeScene() to return non-nil")
	}

	// 2. Settings scene
	settings := scenes.NewSettingsScene(config.DifficultyMedium)
	if settings.GetDifficulty() != config.DifficultyMedium {
		t.Errorf("expected difficulty Medium, got %v", settings.GetDifficulty())
	}
	settings.CycleDifficulty()
	if settings.GetDifficulty() != config.DifficultyHard {
		t.Errorf("expected difficulty Hard after cycle, got %v", settings.GetDifficulty())
	}

	// 3. Gameplay scene
	gameplay := scenes.NewGameplayScene(config.DifficultyEasy)
	if gameplay == nil {
		t.Fatalf("expected NewGameplayScene() to return non-nil")
	}
	if gameplay.GetDistance() != 0 || gameplay.GetCoins() != 0 {
		t.Errorf("expected 0 initial distance and coins, got %d, %d", gameplay.GetDistance(), gameplay.GetCoins())
	}

	// 4. Game Over scene
	gameOver := scenes.NewGameOverScene(520, 42)
	if gameOver.FinalScore != 520 || gameOver.FinalCoins != 42 {
		t.Errorf("expected 520 distance and 42 coins, got %d, %d", gameOver.FinalScore, gameOver.FinalCoins)
	}
}

func TestShieldAbsorbsLaserWithoutImmediateDeath(t *testing.T) {
	gameplay := scenes.NewGameplayScene(config.DifficultyEasy)

	// Equip shield
	gameplay.ActivatePowerupForTest(entities.PowerupShield)

	// Position player at Y=250 and spawn an active laser at Y=250
	gameplay.Player.Y = 250.0
	laser := entities.NewLaser(250.0)
	laser.State = entities.LaserStateActive
	gameplay.AddLaserForTest(laser)

	// Update tick 1: Shield should absorb laser hit, deactivating shield
	nextScene, err := gameplay.Update()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextScene == scenes.SceneGameOver {
		t.Fatalf("player died on frame 1 with full shield")
	}
	if gameplay.IsPowerupActive() {
		t.Errorf("expected shield to be consumed after absorbing hit")
	}

	// Update tick 2: Player must NOT die on frame 2 (laser must be finished/disarmed)
	nextScene, err = gameplay.Update()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextScene == scenes.SceneGameOver {
		t.Fatalf("player died on frame 2 from laser that was supposed to be absorbed by shield")
	}
}

func TestPowerupStateIsolation(t *testing.T) {
	gameplay := scenes.NewGameplayScene(config.DifficultyEasy)

	// Equip Shield first
	gameplay.ActivatePowerupForTest(entities.PowerupShield)
	if !gameplay.Player.HasShield {
		t.Fatalf("expected player to have shield")
	}

	// Switch to Magnet: player shield must be cleared
	gameplay.ActivatePowerupForTest(entities.PowerupMagnet)
	if gameplay.Player.HasShield {
		t.Errorf("expected player shield visual to be cleared when switching to Magnet")
	}

	// Invert gravity with GravitySwap
	gameplay.ActivatePowerupForTest(entities.PowerupGravitySwap)
	gameplay.Player.ToggleGravity()
	if !gameplay.Player.IsGravityInverted {
		t.Fatalf("expected player to be inverted")
	}

	// Switch to CoinBoost: gravity inversion must be normalized
	gameplay.ActivatePowerupForTest(entities.PowerupCoinBoost)
	if gameplay.Player.IsGravityInverted {
		t.Errorf("expected player gravity inversion to be normalized when switching away from GravitySwap")
	}
}

func TestAchievementsTrackingAndPersistence(t *testing.T) {
	scenes.ResetAchievementsForTest()

	// Initial check: all locked
	unlocked := scenes.GetUnlockedAchievements()
	if len(unlocked) != 0 {
		t.Errorf("expected 0 unlocked achievements initially, got %d", len(unlocked))
	}

	gameplay := scenes.NewGameplayScene(config.DifficultyEasy)
	gameplay.SimulateProgressForTest(150.0, 60) // 150m, 60 coins

	gameplay.CheckAchievementsForTest()

	unlocked = scenes.GetUnlockedAchievements()
	if len(unlocked) < 2 {
		t.Errorf("expected at least 2 unlocked achievements (100m flyer and 50 coins), got %d", len(unlocked))
	}
}
