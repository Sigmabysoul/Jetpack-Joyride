package main

import (
	"log"

	"jetpack-joyride/pkg/assets"
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	if err := assets.LoadAssets(); err != nil {
		log.Fatalf("failed to load embedded assets: %v", err)
	}

	ebiten.SetWindowTitle("Jetpack Joyride")
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := game.NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatalf("game exited with error: %v", err)
	}
}
