package assets_test

import (
	"testing"

	"jetpack-joyride/pkg/assets"
)

func TestLoadAssets(t *testing.T) {
	err := assets.LoadAssets()
	if err != nil {
		t.Fatalf("failed to load assets: %v", err)
	}

	requiredImages := []string{
		"Barry.png",
		"BackdropMain.png",
		"Coin.png",
		"Zapper1.png",
		"laser_active.png",
		"laser_noneactive.png",
		"missle.png",
		"warning.png",
		"powerup.png",
		"force_shield.png",
		"gravity_swap.png",
		"magnet.png",
		"jetpack_joyride_home.jpg",
		"jetpack_joyride_game_over.jpg",
		"button.png",
	}

	for _, name := range requiredImages {
		img := assets.GetImage(name)
		if img == nil {
			t.Errorf("expected image %s to be loaded, got nil", name)
		}
	}

	// Fonts
	f1 := assets.GetFont("Roboto-Regular.ttf", 24)
	if f1 == nil {
		t.Errorf("expected Roboto-Regular.ttf to be loaded, got nil")
	}

	f2 := assets.GetFont("New Athletic M54.ttf", 36)
	if f2 == nil {
		t.Errorf("expected New Athletic M54.ttf to be loaded, got nil")
	}
}
