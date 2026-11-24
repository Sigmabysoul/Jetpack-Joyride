package game_test

import (
	"testing"

	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/game"
	"jetpack-joyride/pkg/scenes"
)

func TestGameCoordinatorLayoutAndLifecycle(t *testing.T) {
	g := game.NewGame()
	if g == nil {
		t.Fatalf("expected NewGame() to return non-nil")
	}

	w, h := g.Layout(1920, 1080)
	if w != config.ScreenWidth || h != config.ScreenHeight {
		t.Errorf("expected virtual layout (%d, %d), got (%d, %d)", config.ScreenWidth, config.ScreenHeight, w, h)
	}

	// Update tick should run without error
	err := g.Update()
	if err != nil {
		t.Errorf("expected Update() to succeed, got %v", err)
	}
}

func TestDifficultyPersistenceAcrossNavigation(t *testing.T) {
	g := game.NewGame()

	// Switch to Settings scene
	g.SetSceneForTest(scenes.NewSettingsScene(config.DifficultyEasy))

	// Cycle difficulty in Settings to Hard
	settings := g.GetActiveSceneForTest().(*scenes.SettingsScene)
	settings.CycleDifficulty() // Medium
	settings.CycleDifficulty() // Hard

	// Simulate navigating back to Home
	g.NavigateToSceneForTest(scenes.SceneHome)

	// Now start GamePlay from Home
	g.NavigateToSceneForTest(scenes.SceneGameplay)

	// Verify the difficulty of GameplayScene is Hard, NOT Easy!
	gp := g.GetActiveSceneForTest().(*scenes.GameplayScene)
	if gp.GetDifficulty() != config.DifficultyHard {
		t.Errorf("expected difficulty to persist as Hard, got %v", gp.GetDifficulty())
	}
}
