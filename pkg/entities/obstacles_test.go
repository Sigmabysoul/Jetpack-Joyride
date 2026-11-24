package entities_test

import (
	"testing"

	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/entities"
)

func TestObstacles(t *testing.T) {
	// 1. Zapper
	zapper := entities.NewZapper(1000.0, 200.0, 0.0)
	initialX := zapper.X
	zapper.Update(0.1)
	if zapper.X >= initialX {
		t.Errorf("expected zapper to move left, got X=%f", zapper.X)
	}

	zapper.X = -100.0
	if !zapper.IsOffscreen() {
		t.Errorf("expected zapper at X=-100 to be offscreen")
	}

	// 2. Laser
	laser := entities.NewLaser(250.0)
	if laser.IsActive() {
		t.Errorf("laser should start inactive (warning state)")
	}

	// Advance through warning duration (3.0s)
	laser.Update(config.LaserWarningDuration + 0.1)
	if !laser.IsActive() {
		t.Errorf("laser should be active after warning duration")
	}

	// Advance through active duration (3.0s)
	laser.Update(config.LaserActiveDuration + 0.1)
	if !laser.IsFinished() {
		t.Errorf("laser should be finished after active duration")
	}

	// 3. Rocket
	rocket := entities.NewRocket()
	if rocket.IsLaunched() {
		t.Errorf("rocket should start in warning state")
	}

	// Warning should track player Y
	targetY := 350.0
	rocket.Update(targetY, 0.5)
	if rocket.Y == 0 {
		t.Errorf("rocket should have tracked toward %f, got %f", targetY, rocket.Y)
	}

	// Advance past total warning time (WarningTime + LockTime)
	rocket.Update(targetY, config.RocketWarningTime+config.RocketLockTime+0.1)
	if !rocket.IsLaunched() {
		t.Errorf("rocket should be launched after warning duration")
	}

	// Launched rocket should fly leftward
	rocketX := rocket.X
	rocket.Update(targetY, 0.1)
	if rocket.X >= rocketX {
		t.Errorf("launched rocket should move left, initial=%f, current=%f", rocketX, rocket.X)
	}
}
