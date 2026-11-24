package game

import (
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/scenes"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game implements the ebiten.Game interface and coordinates screen scenes.
type Game struct {
	activeScene scenes.Scene
	difficulty  config.Difficulty
}

// NewGame initializes the game with the home scene.
func NewGame() *Game {
	return &Game{
		activeScene: scenes.NewHomeScene(),
		difficulty:  config.DifficultyEasy,
	}
}

// Update coordinates scene transitions and advances the active scene.
func (g *Game) Update() error {
	nextSceneID, err := g.activeScene.Update()
	if err != nil {
		return err
	}

	// Persist difficulty selection whenever transitioning away from Settings
	if settings, ok := g.activeScene.(*scenes.SettingsScene); ok {
		g.difficulty = settings.GetDifficulty()
	}

	switch nextSceneID {
	case scenes.SceneHome:
		g.activeScene = scenes.NewHomeScene()

	case scenes.SceneSettings:
		g.activeScene = scenes.NewSettingsScene(g.difficulty)

	case scenes.SceneGameplay:
		g.activeScene = scenes.NewGameplayScene(g.difficulty)

	case scenes.SceneGameOver:
		score := 0
		coins := 0
		if gp, ok := g.activeScene.(*scenes.GameplayScene); ok {
			score = gp.GetDistance()
			coins = gp.GetCoins()
		}
		g.activeScene = scenes.NewGameOverScene(score, coins)
	}

	return nil
}

// Draw delegates rendering to the active scene.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.activeScene != nil {
		g.activeScene.Draw(screen)
	}
}

// Layout maintains a fixed virtual resolution of 1000x500 scaled to window dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

// Test helpers
func (g *Game) SetSceneForTest(s scenes.Scene) {
	g.activeScene = s
}

func (g *Game) GetActiveSceneForTest() scenes.Scene {
	return g.activeScene
}

func (g *Game) NavigateToSceneForTest(target scenes.SceneID) {
	if settings, ok := g.activeScene.(*scenes.SettingsScene); ok {
		g.difficulty = settings.GetDifficulty()
	}

	switch target {
	case scenes.SceneHome:
		g.activeScene = scenes.NewHomeScene()
	case scenes.SceneGameplay:
		g.activeScene = scenes.NewGameplayScene(g.difficulty)
	case scenes.SceneSettings:
		g.activeScene = scenes.NewSettingsScene(g.difficulty)
	}
}
