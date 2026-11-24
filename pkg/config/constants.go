package config

// Difficulty levels matching original game
type Difficulty int

const (
	DifficultyEasy Difficulty = iota
	DifficultyMedium
	DifficultyHard
)

func (d Difficulty) String() string {
	switch d {
	case DifficultyEasy:
		return "Easy"
	case DifficultyMedium:
		return "Medium"
	case DifficultyHard:
		return "Hard"
	default:
		return "Unknown"
	}
}

const (
	// Screen Dimensions
	ScreenWidth  = 1000
	ScreenHeight = 500

	// Player Physics
	PlayerX       = 100.0 // Barry's fixed horizontal position
	PlayerWidth   = 60.0
	PlayerHeight  = 70.0
	GroundY       = 50.0  // Distance from bottom
	CeilingY      = 450.0 // Distance from bottom
	Gravity       = 900.0 // px/s^2 downward
	JetpackThrust = 1500.0 // px/s^2 upward
	MaxVelocity   = 400.0 // px/s terminal velocity

	// Fallback Kinematic Snap (matching original C code if toggled)
	UseKinematicSnap  = false
	KinematicSnapSpeed = 200.0

	// World & Background
	BackgroundScrollSpeed = 150.0 // px/s

	// Zappers
	ZapperWidth  = 46.0
	ZapperHeight = 109.0
	ZapperSpeed  = 200.0 // moves left at 200 px/s

	// Lasers
	LaserHeight         = 42.14
	LaserWarningDuration = 3.0 // seconds warning before active
	LaserActiveDuration  = 3.0 // seconds beam stays lethal

	// Rockets
	RocketWidth         = 80.0
	RocketHeight        = 40.0
	RocketSpeed         = 600.0 // px/s across screen
	RocketWarningTime   = 2.5   // seconds warning tracks before lock
	RocketLockTime      = 0.75  // seconds warning locks position before launch

	// Coins
	CoinSize  = 28.0
	CoinSpeed = 200.0

	// Powerup durations (seconds)
	ShieldDuration      = 15.0
	GravitySwapDuration = 12.0
	MagnetDuration      = 10.0
	CoinBoostDuration   = 10.0
)
