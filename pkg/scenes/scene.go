package scenes

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type SceneID int

const (
	SceneNone SceneID = iota
	SceneHome
	SceneSettings
	SceneGameplay
	SceneGameOver
)

// Scene defines the interface that all screen states implement.
type Scene interface {
	Update() (SceneID, error)
	Draw(screen *ebiten.Image)
}

// IsClicked returns true if the user clicked or tapped inside the specified bounding box this frame.
func IsClicked(x, y, w, h float64) bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		fx, fy := float64(cx), float64(cy)
		if fx >= x && fx <= x+w && fy >= y && fy <= y+h {
			return true
		}
	}
	// Also check touch taps for mobile/web touch screens
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range touchIDs {
		tx, ty := ebiten.TouchPosition(id)
		fx, fy := float64(tx), float64(ty)
		if fx >= x && fx <= x+w && fy >= y && fy <= y+h {
			return true
		}
	}
	return false
}

// DrawButton renders a stylized button box with text label.
func DrawButton(screen *ebiten.Image, buttonSprite *ebiten.Image, x, y, w, h float64, label string, fontFace font.Face) {
	if buttonSprite != nil {
		op := &ebiten.DrawImageOptions{}
		bw, bh := float64(buttonSprite.Bounds().Dx()), float64(buttonSprite.Bounds().Dy())
		op.GeoM.Scale(w/bw, h/bh)
		op.GeoM.Translate(x, y)
		screen.DrawImage(buttonSprite, op)
	}

	if fontFace != nil && label != "" {
		bounds := text.BoundString(fontFace, label)
		textX := int(x + (w-float64(bounds.Dx()))/2)
		textY := int(y + (h+float64(bounds.Dy()))/2)
		text.Draw(screen, label, fontFace, textX, textY, color.White)
	}
}
