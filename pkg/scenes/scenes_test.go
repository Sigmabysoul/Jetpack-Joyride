package scenes_test

import (
	"testing"

	"jetpack-joyride/pkg/config"
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
