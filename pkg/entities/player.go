package entities

import (
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/math2d"

	"github.com/hajimehoshi/ebiten/v2"
)

// Player represents Barry, the jetpack user.
type Player struct {
	X                 float64
	Y                 float64
	VelocityY         float64
	Width             float64
	Height            float64
	IsGrounded        bool
	IsGravityInverted bool
	HasShield         bool
}

// NewPlayer creates a new player initialized at (x, y).
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:                 x,
		Y:                 y,
		VelocityY:         0,
		Width:             config.PlayerWidth,
		Height:            config.PlayerHeight,
		IsGrounded:        false,
		IsGravityInverted: false,
		HasShield:         false,
	}
}

// Update advances player physics by dt seconds based on thrust input.
func (p *Player) Update(thrusting bool, dt float64) {
	if config.UseKinematicSnap {
		if !p.IsGravityInverted {
			if thrusting {
				p.VelocityY = config.KinematicSnapSpeed
			} else {
				p.VelocityY = -config.KinematicSnapSpeed
			}
		} else {
			if thrusting {
				p.VelocityY = -config.KinematicSnapSpeed
			} else {
				p.VelocityY = config.KinematicSnapSpeed
			}
		}
	} else {
		// Normal gravity direction is downward (-Y)
		gravDir := -1.0
		thrustDir := 1.0
		if p.IsGravityInverted {
			gravDir = 1.0
			thrustDir = -1.0
		}

		accelY := config.Gravity * gravDir
		if thrusting {
			accelY += config.JetpackThrust * thrustDir
		}

		p.VelocityY += accelY * dt

		// Clamp terminal velocity
		if p.VelocityY > config.MaxVelocity {
			p.VelocityY = config.MaxVelocity
		} else if p.VelocityY < -config.MaxVelocity {
			p.VelocityY = -config.MaxVelocity
		}
	}

	p.Y += p.VelocityY * dt
	p.IsGrounded = false

	// Boundary clamping
	if !p.IsGravityInverted {
		if p.Y <= config.GroundY {
			p.Y = config.GroundY
			p.VelocityY = 0
			p.IsGrounded = true
		} else if p.Y >= config.CeilingY {
			p.Y = config.CeilingY
			p.VelocityY = 0
		}
	} else {
		if p.Y >= config.CeilingY {
			p.Y = config.CeilingY
			p.VelocityY = 0
			p.IsGrounded = true
		} else if p.Y <= config.GroundY {
			p.Y = config.GroundY
			p.VelocityY = 0
		}
	}
}

// ToggleGravity flips the gravitational direction for the gravity swap powerup.
func (p *Player) ToggleGravity() {
	p.IsGravityInverted = !p.IsGravityInverted
}

// SetShield enables or disables the energy shield.
func (p *Player) SetShield(active bool) {
	p.HasShield = active
}

// Bounds returns the collision polygon for Barry.
func (p *Player) Bounds() math2d.Polygon {
	// Slightly tighter hitbox for fairer gameplay (e.g. 75% size of sprite)
	hitboxW := p.Width * 0.75
	hitboxH := p.Height * 0.75
	return math2d.NewRectangle(math2d.Vector{X: p.X, Y: p.Y}, hitboxW, hitboxH)
}

// Draw renders Barry (and optional shield overlay) onto the screen.
func (p *Player) Draw(screen *ebiten.Image, barrySprite, shieldSprite *ebiten.Image) {
	if barrySprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}

	// Center sprite on (p.X, p.Y)
	w, h := float64(barrySprite.Bounds().Dx()), float64(barrySprite.Bounds().Dy())
	scaleX := p.Width / w
	scaleY := p.Height / h

	if p.IsGravityInverted {
		// Flip vertically when running on ceiling
		op.GeoM.Scale(scaleX, -scaleY)
		op.GeoM.Translate(-p.Width/2, p.Height/2)
	} else {
		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(-p.Width/2, -p.Height/2)
	}

	// In Ebitengine, Y=0 is top, so convert game bottom-relative Y to screen Y
	screenY := float64(config.ScreenHeight) - p.Y
	op.GeoM.Translate(p.X, screenY)

	screen.DrawImage(barrySprite, op)

	// Draw shield bubble if active
	if p.HasShield && shieldSprite != nil {
		sOp := &ebiten.DrawImageOptions{}
		sw, sh := float64(shieldSprite.Bounds().Dx()), float64(shieldSprite.Bounds().Dy())
		shieldScale := (p.Height * 1.4) / sh
		sOp.GeoM.Scale(shieldScale, shieldScale)
		sOp.GeoM.Translate(-sw*shieldScale/2, -sh*shieldScale/2)
		sOp.GeoM.Translate(p.X, screenY)
		screen.DrawImage(shieldSprite, sOp)
	}
}
