package entities

import (
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/math2d"

	"github.com/hajimehoshi/ebiten/v2"
)

type RocketState int

const (
	RocketStateWarningTracking RocketState = iota
	RocketStateWarningLocked
	RocketStateLaunched
)

// Rocket represents an incoming missile tracking the player before launch.
type Rocket struct {
	X      float64
	Y      float64
	Timer  float64
	State  RocketState
	Width  float64
	Height float64
}

// NewRocket creates a new rocket initialized in the tracking warning phase.
func NewRocket() *Rocket {
	return &Rocket{
		X:      config.ScreenWidth - 50.0,
		Y:      config.ScreenHeight / 2.0,
		Timer:  0,
		State:  RocketStateWarningTracking,
		Width:  config.RocketWidth,
		Height: config.RocketHeight,
	}
}

// Update advances the rocket warning timer or moves the rocket if launched.
func (r *Rocket) Update(playerY float64, dt float64) {
	r.Timer += dt

	if r.State == RocketStateWarningTracking {
		diff := playerY - r.Y
		r.Y += diff * 5.0 * dt
		if r.Timer >= config.RocketWarningTime {
			r.State = RocketStateWarningLocked
		}
	}

	if r.State == RocketStateWarningLocked {
		if r.Timer >= config.RocketWarningTime+config.RocketLockTime {
			r.State = RocketStateLaunched
			r.X = config.ScreenWidth + 50.0
		}
	} else if r.State == RocketStateLaunched {
		r.X -= config.RocketSpeed * dt
	}
}

// IsLaunched returns true if the missile is in flight.
func (r *Rocket) IsLaunched() bool {
	return r.State == RocketStateLaunched
}

// IsOffscreen returns true if the missile has flown past the left edge.
func (r *Rocket) IsOffscreen() bool {
	return r.State == RocketStateLaunched && r.X < -r.Width*2
}

// Bounds returns the collision polygon for the flying rocket.
func (r *Rocket) Bounds() math2d.Polygon {
	if !r.IsLaunched() {
		return math2d.Polygon{}
	}
	return math2d.NewRectangle(math2d.Vector{X: r.X, Y: r.Y}, r.Width*0.8, r.Height*0.6)
}

// Draw renders either the warning exclamation icon or the missile sprite.
func (r *Rocket) Draw(screen *ebiten.Image, rocketSprite, warningSprite *ebiten.Image) {
	screenY := float64(config.ScreenHeight) - r.Y

	if r.State == RocketStateWarningTracking || r.State == RocketStateWarningLocked {
		if warningSprite == nil {
			return
		}
		// In locked state, flash every 0.1s
		if r.State == RocketStateWarningLocked && int(r.Timer*10)%2 == 0 {
			return
		}
		op := &ebiten.DrawImageOptions{}
		w, h := float64(warningSprite.Bounds().Dx()), float64(warningSprite.Bounds().Dy())
		scale := 40.0 / h
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(-w*scale/2, -40.0/2)
		op.GeoM.Translate(config.ScreenWidth-40.0, screenY)
		screen.DrawImage(warningSprite, op)
		return
	}

	if rocketSprite == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	w, h := float64(rocketSprite.Bounds().Dx()), float64(rocketSprite.Bounds().Dy())
	scaleX := r.Width / w
	scaleY := r.Height / h

	// Mirror horizontally so rocket faces left
	op.GeoM.Scale(-scaleX, scaleY)
	op.GeoM.Translate(r.Width/2, -r.Height/2)
	op.GeoM.Translate(r.X, screenY)

	screen.DrawImage(rocketSprite, op)
}
