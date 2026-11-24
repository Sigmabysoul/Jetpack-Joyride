package entities_test

import (
	"testing"

	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/entities"
)

func TestPlayerPhysicsAndBoundaries(t *testing.T) {
	player := entities.NewPlayer(config.PlayerX, 200.0)

	// Step 1: Upward thrust
	initialVy := player.VelocityY
	player.Update(true, 0.1) // thrusting upward for 0.1s
	if player.VelocityY <= initialVy {
		t.Errorf("expected vertical velocity to increase with thrust, got %f", player.VelocityY)
	}

	// Step 2: Gravity drop
	player.Update(false, 0.1) // gravity pulls downward for 0.1s while still in mid-air
	if player.VelocityY >= 0 {
		t.Errorf("expected negative velocity while falling, got %f", player.VelocityY)
	}

	// Step 3: Ground clamping
	for i := 0; i < 50; i++ {
		player.Update(false, 0.1)
	}
	if player.Y < config.GroundY {
		t.Errorf("player fell below ground: Y=%f, GroundY=%f", player.Y, config.GroundY)
	}
	if player.VelocityY != 0 {
		t.Errorf("expected velocity to reset to 0 when grounded, got %f", player.VelocityY)
	}

	// Step 4: Ceiling clamping
	for i := 0; i < 50; i++ {
		player.Update(true, 0.1)
	}
	if player.Y > config.CeilingY {
		t.Errorf("player clipped through ceiling: Y=%f, CeilingY=%f", player.Y, config.CeilingY)
	}
	if player.VelocityY != 0 {
		t.Errorf("expected velocity to reset to 0 at ceiling, got %f", player.VelocityY)
	}

	// Step 5: Toggle Gravity
	player.ToggleGravity()
	if !player.IsGravityInverted {
		t.Errorf("expected gravity to be inverted")
	}

	// Step 6: Bounds polygon
	bounds := player.Bounds()
	if len(bounds.Vertices) != 4 {
		t.Errorf("expected 4 vertices for bounding polygon, got %d", len(bounds.Vertices))
	}

	// Step 7: Shield
	player.SetShield(true)
	if !player.HasShield {
		t.Errorf("expected player to have active shield")
	}
}
