package entities

import (
	"math/rand"

	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/math2d"

	"github.com/hajimehoshi/ebiten/v2"
)

type PowerupType int

const (
	PowerupShield PowerupType = iota
	PowerupGravitySwap
	PowerupMagnet
	PowerupCoinBoost
)

func (p PowerupType) String() string {
	switch p {
	case PowerupShield:
		return "Shield"
	case PowerupGravitySwap:
		return "Gravity Swap"
	case PowerupMagnet:
		return "Coin Magnet"
	case PowerupCoinBoost:
		return "Coin Boost"
	default:
		return "Unknown"
	}
}

// PowerupToken represents the mystery item box floating across the screen.
type PowerupToken struct {
	X         float64
	Y         float64
	Collected bool
	Size      float64
}

// NewPowerupToken creates a powerup item at (x, y).
func NewPowerupToken(x, y float64) *PowerupToken {
	return &PowerupToken{
		X:         x,
		Y:         y,
		Collected: false,
		Size:      45.0,
	}
}

// Update moves the powerup leftward.
func (p *PowerupToken) Update(dt float64) {
	p.X -= config.BackgroundScrollSpeed * 1.2 * dt
}

// IsOffscreen returns true if the token is past the screen.
func (p *PowerupToken) IsOffscreen() bool {
	return p.X < -p.Size*2
}

// Activate chooses a random powerup when collected.
func (p *PowerupToken) Activate() PowerupType {
	p.Collected = true
	types := []PowerupType{PowerupShield, PowerupGravitySwap, PowerupMagnet, PowerupCoinBoost}
	return types[rand.Intn(len(types))]
}

// Bounds returns the collision polygon for the token.
func (p *PowerupToken) Bounds() math2d.Polygon {
	return math2d.NewRectangle(math2d.Vector{X: p.X, Y: p.Y}, p.Size, p.Size)
}

// Draw renders the mystery powerup box sprite.
func (p *PowerupToken) Draw(screen *ebiten.Image, sprite *ebiten.Image) {
	if p.Collected || sprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	w := float64(sprite.Bounds().Dx())
	scale := p.Size / w

	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-p.Size/2, -p.Size/2)

	screenY := float64(config.ScreenHeight) - p.Y
	op.GeoM.Translate(p.X, screenY)

	screen.DrawImage(sprite, op)
}
