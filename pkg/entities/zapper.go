package entities

import (
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/math2d"

	"github.com/hajimehoshi/ebiten/v2"
)

// Zapper represents a lethal high-voltage obstacle.
type Zapper struct {
	X      float64
	Y      float64
	Angle  float64
	Width  float64
	Height float64
}

// NewZapper creates a new zapper at (x, y) with a rotation angle in radians.
func NewZapper(x, y, angle float64) *Zapper {
	return &Zapper{
		X:      x,
		Y:      y,
		Angle:  angle,
		Width:  config.ZapperWidth,
		Height: config.ZapperHeight,
	}
}

// Update moves the zapper leftward.
func (z *Zapper) Update(dt float64) {
	z.X -= config.ZapperSpeed * dt
}

// IsOffscreen returns true if the zapper has scrolled completely off the left edge.
func (z *Zapper) IsOffscreen() bool {
	return z.X < -config.ZapperWidth*2
}

// Bounds returns the rotated collision polygon for the zapper.
func (z *Zapper) Bounds() math2d.Polygon {
	rect := math2d.NewRectangle(math2d.Vector{X: z.X, Y: z.Y}, z.Width*0.7, z.Height*0.9)
	return rect.Rotate(math2d.Vector{X: z.X, Y: z.Y}, z.Angle)
}

// Draw renders the zapper sprite rotated at its center.
func (z *Zapper) Draw(screen *ebiten.Image, sprite *ebiten.Image) {
	if sprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	w, h := float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy())
	scaleX := z.Width / w
	scaleY := z.Height / h

	// Center pivot, rotate, scale, translate
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Rotate(-z.Angle) // Invert angle for screen coordinates

	screenY := float64(config.ScreenHeight) - z.Y
	op.GeoM.Translate(z.X, screenY)

	screen.DrawImage(sprite, op)
}
