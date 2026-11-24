package scenes

import (
	"fmt"
	"image/color"

	"jetpack-joyride/pkg/assets"
	"jetpack-joyride/pkg/config"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// SettingsScene allows configuring difficulty and viewing achievements.
type SettingsScene struct {
	difficulty   config.Difficulty
	DiffBtnX     float64
	DiffBtnY     float64
	DiffBtnW     float64
	DiffBtnH     float64
	BackBtnX     float64
	BackBtnY     float64
	BackBtnW     float64
	BackBtnH     float64
	Achievements []string
}

func NewSettingsScene(initialDiff config.Difficulty) *SettingsScene {
	return &SettingsScene{
		difficulty:   initialDiff,
		DiffBtnX:     350,
		DiffBtnY:     160,
		DiffBtnW:     300,
		DiffBtnH:     50,
		BackBtnX:     400,
		BackBtnY:     410,
		BackBtnW:     200,
		BackBtnH:     50,
		Achievements: []string{
			"Novice Flyer: Fly past 100 meters",
			"Coin Collector: Collect 50 coins",
			"Zapper Dodger: Dodge 10 zappers",
			"Laser Evader: Survive active laser beams",
			"Rocket Runner: Evade tracking missile",
		},
	}
}

func (s *SettingsScene) GetDifficulty() config.Difficulty {
	return s.difficulty
}

func (s *SettingsScene) CycleDifficulty() {
	switch s.difficulty {
	case config.DifficultyEasy:
		s.difficulty = config.DifficultyMedium
	case config.DifficultyMedium:
		s.difficulty = config.DifficultyHard
	case config.DifficultyHard:
		s.difficulty = config.DifficultyEasy
	}
}

func (s *SettingsScene) Update() (SceneID, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || IsClicked(s.BackBtnX, s.BackBtnY, s.BackBtnW, s.BackBtnH) {
		return SceneHome, nil
	}

	if IsClicked(s.DiffBtnX, s.DiffBtnY, s.DiffBtnW, s.DiffBtnH) {
		s.CycleDifficulty()
	}

	return SceneNone, nil
}

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	bg := assets.GetImage("jetpack_joyride_home.jpg")
	if bg != nil {
		op := &ebiten.DrawImageOptions{}
		bw, bh := float64(bg.Bounds().Dx()), float64(bg.Bounds().Dy())
		op.GeoM.Scale(float64(config.ScreenWidth)/bw, float64(config.ScreenHeight)/bh)
		op.ColorScale.Scale(0.3, 0.3, 0.3, 1.0) // Darken background for readability
		screen.DrawImage(bg, op)
	}

	titleFont := assets.GetFont("New Athletic M54.ttf", 36)
	labelFont := assets.GetFont("Roboto-Regular.ttf", 18)
	btnFont := assets.GetFont("New Athletic M54.ttf", 24)
	btnSprite := assets.GetImage("button.png")

	// Title
	if titleFont != nil {
		text.Draw(screen, "SETTINGS & ACHIEVEMENTS", titleFont, 290, 70, color.White)
	}

	// Difficulty Button
	diffLabel := fmt.Sprintf("DIFFICULTY: %s", s.difficulty.String())
	DrawButton(screen, btnSprite, s.DiffBtnX, s.DiffBtnY, s.DiffBtnW, s.DiffBtnH, diffLabel, btnFont)

	// Achievements Header & List
	if labelFont != nil {
		text.Draw(screen, "ACHIEVEMENTS:", labelFont, 350, 245, color.RGBA{255, 215, 0, 255})
		unlockedMap := GetUnlockedAchievements()
		for i, ach := range AllAchievements {
			statusIcon := "[ ]"
			col := color.RGBA{200, 200, 200, 255}
			if unlockedMap[ach.ID] {
				statusIcon = "[✓]"
				col = color.RGBA{100, 255, 100, 255}
			}
			line := fmt.Sprintf("%s %s: %s", statusIcon, ach.Title, ach.Description)
			text.Draw(screen, line, labelFont, 280, 275+i*24, col)
		}
	}

	// Back Button
	DrawButton(screen, btnSprite, s.BackBtnX, s.BackBtnY, s.BackBtnW, s.BackBtnH, "BACK", btnFont)
}
