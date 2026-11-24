package entities

import (
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/math2d"

	"github.com/hajimehoshi/ebiten/v2"
)

type LaserState int

const (
	LaserStateWarning LaserState = iota
	LaserStateActive
	LaserStateFinished
)

// Laser represents a horizontal beam that warms up before firing across the screen.
type Laser struct {
	Y     float64
	Timer float64
	State LaserState
}

// NewLaser creates a new laser positioned at vertical coordinate y.
func NewLaser(y float64) *Laser {
	return &Laser{
		Y:     y,
		Timer: 0,
		State: LaserStateWarning,
	}
}

// Update advances the laser countdown and active phase.
func (l *Laser) Update(dt float64) {
	l.Timer += dt

	if l.State == LaserStateWarning {
		if l.Timer >= config.LaserWarningDuration {
			l.State = LaserStateActive
			l.Timer = 0
		}
	} else if l.State == LaserStateActive {
		if l.Timer >= config.LaserActiveDuration {
			l.State = LaserStateFinished
		}
	}
}

// IsActive returns true if the laser is currently lethal.
func (l *Laser) IsActive() bool {
	return l.State == LaserStateActive
}

// IsFinished returns true if the laser sequence is complete.
func (l *Laser) IsFinished() bool {
	return l.State == LaserStateFinished
}

// Bounds returns the lethal collision rectangle covering the screen width when active.
func (l *Laser) Bounds() math2d.Polygon {
	if !l.IsActive() {
		return math2d.Polygon{}
	}
	center := math2d.Vector{X: config.ScreenWidth / 2.0, Y: l.Y}
	return math2d.NewRectangle(center, config.ScreenWidth, config.LaserHeight*0.7)
}

// Draw renders either the inactive warning line or the active laser beam.
func (l *Laser) Draw(screen *ebiten.Image, activeSprite, inactiveSprite *ebiten.Image) {
	if l.IsFinished() {
		return
	}

	sprite := inactiveSprite
	if l.IsActive() {
		sprite = activeSprite
	}
	if sprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	w, h := float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy())
	scaleX := float64(config.ScreenWidth) / w
	scaleY := config.LaserHeight / h

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(0, -config.LaserHeight/2)

	screenY := float64(config.ScreenHeight) - l.Y
	op.GeoM.Translate(0, screenY)

	screen.DrawImage(sprite, op)
}
