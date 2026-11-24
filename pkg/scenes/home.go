package scenes

import (
	"jetpack-joyride/pkg/assets"
	"jetpack-joyride/pkg/config"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// HomeScene is the title screen.
type HomeScene struct {
	PlayBtnX     float64
	PlayBtnY     float64
	PlayBtnW     float64
	PlayBtnH     float64
	SettingsBtnX float64
	SettingsBtnY float64
	SettingsBtnW float64
	SettingsBtnH float64
}

// NewHomeScene creates and positions the home menu buttons.
func NewHomeScene() *HomeScene {
	return &HomeScene{
		PlayBtnX:     400,
		PlayBtnY:     320,
		PlayBtnW:     200,
		PlayBtnH:     55,
		SettingsBtnX: 400,
		SettingsBtnY: 395,
		SettingsBtnW: 200,
		SettingsBtnH: 55,
	}
}

func (h *HomeScene) Update() (SceneID, error) {
	// Space or Enter also starts the game
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return SceneGameplay, nil
	}

	if IsClicked(h.PlayBtnX, h.PlayBtnY, h.PlayBtnW, h.PlayBtnH) {
		return SceneGameplay, nil
	}

	if IsClicked(h.SettingsBtnX, h.SettingsBtnY, h.SettingsBtnW, h.SettingsBtnH) {
		return SceneSettings, nil
	}

	return SceneNone, nil
}

func (h *HomeScene) Draw(screen *ebiten.Image) {
	bg := assets.GetImage("jetpack_joyride_home.jpg")
	if bg != nil {
		op := &ebiten.DrawImageOptions{}
		bw, bh := float64(bg.Bounds().Dx()), float64(bg.Bounds().Dy())
		op.GeoM.Scale(float64(config.ScreenWidth)/bw, float64(config.ScreenHeight)/bh)
		screen.DrawImage(bg, op)
	}

	btnSprite := assets.GetImage("button.png")
	btnFont := assets.GetFont("New Athletic M54.ttf", 26)

	DrawButton(screen, btnSprite, h.PlayBtnX, h.PlayBtnY, h.PlayBtnW, h.PlayBtnH, "PLAY", btnFont)
	DrawButton(screen, btnSprite, h.SettingsBtnX, h.SettingsBtnY, h.SettingsBtnW, h.SettingsBtnH, "SETTINGS", btnFont)
}
