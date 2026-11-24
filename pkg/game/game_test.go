package game_test

import (
	"testing"

	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/game"
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
