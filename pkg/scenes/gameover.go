package scenes

import (
	"encoding/json"
	"fmt"
	"image/color"

	"jetpack-joyride/pkg/assets"
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/storage"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type SaveData struct {
	BestScore int `json:"best_score"`
	BestCoins int `json:"best_coins"`
}

// GameOverScene presents the score summary, high scores, and restart options.
type GameOverScene struct {
	FinalScore int
	FinalCoins int
	BestScore  int
	BestCoins  int
	RetryBtnX  float64
	RetryBtnY  float64
	RetryBtnW  float64
	RetryBtnH  float64
	HomeBtnX   float64
	HomeBtnY   float64
	HomeBtnW   float64
	HomeBtnH   float64
}

// NewGameOverScene creates the summary and persists any new high score.
func NewGameOverScene(score, coins int) *GameOverScene {
	store := storage.GetStorage()
	bestScore := score
	bestCoins := coins

	data, err := store.Load("stats")
	if err == nil {
		var saved SaveData
		if json.Unmarshal(data, &saved) == nil {
			if saved.BestScore > bestScore {
				bestScore = saved.BestScore
			}
			if saved.BestCoins > bestCoins {
				bestCoins = saved.BestCoins
			}
		}
	}

	// Save back updated stats
	updated, _ := json.Marshal(SaveData{BestScore: bestScore, BestCoins: bestCoins})
	_ = store.Save("stats", updated)

	return &GameOverScene{
		FinalScore: score,
		FinalCoins: coins,
		BestScore:  bestScore,
		BestCoins:  bestCoins,
		RetryBtnX:  400,
		RetryBtnY:  330,
		RetryBtnW:  200,
		RetryBtnH:  50,
		HomeBtnX:   400,
		HomeBtnY:   395,
		HomeBtnW:   200,
		HomeBtnH:   50,
	}
}

func (g *GameOverScene) Update() (SceneID, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return SceneGameplay, nil
	}

	if IsClicked(g.RetryBtnX, g.RetryBtnY, g.RetryBtnW, g.RetryBtnH) {
		return SceneGameplay, nil
	}

	if IsClicked(g.HomeBtnX, g.HomeBtnY, g.HomeBtnW, g.HomeBtnH) {
		return SceneHome, nil
	}

	return SceneNone, nil
}

func (g *GameOverScene) Draw(screen *ebiten.Image) {
	bg := assets.GetImage("jetpack_joyride_game_over.jpg")
	if bg != nil {
		op := &ebiten.DrawImageOptions{}
		bw, bh := float64(bg.Bounds().Dx()), float64(bg.Bounds().Dy())
		op.GeoM.Scale(float64(config.ScreenWidth)/bw, float64(config.ScreenHeight)/bh)
		screen.DrawImage(bg, op)
	}

	titleFont := assets.GetFont("New Athletic M54.ttf", 40)
	statFont := assets.GetFont("New Athletic M54.ttf", 24)
	btnFont := assets.GetFont("New Athletic M54.ttf", 24)
	btnSprite := assets.GetImage("button.png")

	if titleFont != nil {
		text.Draw(screen, "GAME OVER", titleFont, 410, 80, color.RGBA{255, 60, 60, 255})
	}

	if statFont != nil {
		distStr := fmt.Sprintf("DISTANCE: %d M  (BEST: %d M)", g.FinalScore, g.BestScore)
		coinStr := fmt.Sprintf("COINS: %d  (BEST: %d)", g.FinalCoins, g.BestCoins)

		text.Draw(screen, distStr, statFont, 360, 220, color.White)
		text.Draw(screen, coinStr, statFont, 400, 260, color.RGBA{255, 215, 0, 255})
	}

	DrawButton(screen, btnSprite, g.RetryBtnX, g.RetryBtnY, g.RetryBtnW, g.RetryBtnH, "RETRY", btnFont)
	DrawButton(screen, btnSprite, g.HomeBtnX, g.HomeBtnY, g.HomeBtnW, g.HomeBtnH, "MENU", btnFont)
}
