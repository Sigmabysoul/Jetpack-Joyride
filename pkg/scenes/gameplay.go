package scenes

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"jetpack-joyride/pkg/assets"
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/entities"
	"jetpack-joyride/pkg/math2d"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// GameplayScene is the core game loop.
type GameplayScene struct {
	difficulty   config.Difficulty
	Player       *entities.Player
	bgScrollX    float64
	distance     float64
	coins        int
	gameOver     bool

	// Obstacles
	zappers []*entities.Zapper
	lasers  []*entities.Laser
	rockets []*entities.Rocket

	// Collectibles & Powerups
	coinsList      []*entities.Coin
	powerupTokens  []*entities.PowerupToken
	activePowerup  entities.PowerupType
	powerupTimer   float64
	isPowerupOn    bool

	// Timers for spawning
	zapperSpawnTimer float64
	laserSpawnTimer  float64
	rocketSpawnTimer float64
	coinSpawnTimer   float64
	powerupTimerNext float64
}

// NewGameplayScene initializes a fresh game run.
func NewGameplayScene(diff config.Difficulty) *GameplayScene {
	return &GameplayScene{
		difficulty:       diff,
		Player:           entities.NewPlayer(config.PlayerX, 150.0),
		bgScrollX:        0,
		distance:         0,
		coins:            0,
		gameOver:         false,
		zappers:          make([]*entities.Zapper, 0),
		lasers:           make([]*entities.Laser, 0),
		rockets:          make([]*entities.Rocket, 0),
		coinsList:        make([]*entities.Coin, 0),
		powerupTokens:    make([]*entities.PowerupToken, 0),
		activePowerup:    entities.PowerupShield,
		powerupTimer:     0,
		isPowerupOn:      false,
		zapperSpawnTimer: 2.0,
		laserSpawnTimer:  10.0,
		rocketSpawnTimer: 5.0,
		coinSpawnTimer:   1.0,
		powerupTimerNext: 12.0,
	}
}

func (g *GameplayScene) GetDistance() int {
	return int(g.distance)
}

func (g *GameplayScene) GetCoins() int {
	return g.coins
}

func (g *GameplayScene) Update() (SceneID, error) {
	if g.gameOver {
		return SceneGameOver, nil
	}

	dt := 1.0 / 60.0 // 60 FPS tick

	// 1. Player controls
	thrusting := ebiten.IsKeyPressed(ebiten.KeySpace) ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Gravity Swap tap behavior
	if g.isPowerupOn && g.activePowerup == entities.PowerupGravitySwap {
		thrusting = false // Disable continuous thrust so it does not fight tap-to-swap
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.Player.ToggleGravity()
		}
	}

	g.Player.Update(thrusting, dt)

	// 2. Scroll Background & distance
	g.bgScrollX += config.BackgroundScrollSpeed * dt
	g.distance += dt * 35.0

	// 3. Powerup countdown
	if g.isPowerupOn {
		g.powerupTimer -= dt
		if g.powerupTimer <= 0 {
			g.isPowerupOn = false
			g.Player.SetShield(false)
			if g.Player.IsGravityInverted {
				g.Player.ToggleGravity()
			}
		}
	}

	// 4. Update and spawn Zappers
	g.zapperSpawnTimer -= dt
	if g.zapperSpawnTimer <= 0 {
		g.spawnZapper()
		g.zapperSpawnTimer = g.randomZapperInterval()
	}
	for i := len(g.zappers) - 1; i >= 0; i-- {
		z := g.zappers[i]
		z.Update(dt)
		if z.IsOffscreen() {
			g.zappers = append(g.zappers[:i], g.zappers[i+1:]...)
		}
	}

	// 5. Update and spawn Lasers
	g.laserSpawnTimer -= dt
	if g.laserSpawnTimer <= 0 {
		g.spawnLaser()
		g.laserSpawnTimer = g.randomLaserInterval()
	}
	for i := len(g.lasers) - 1; i >= 0; i-- {
		l := g.lasers[i]
		l.Update(dt)
		if l.IsFinished() {
			g.lasers = append(g.lasers[:i], g.lasers[i+1:]...)
		}
	}

	// 6. Update and spawn Rockets
	g.rocketSpawnTimer -= dt
	if g.rocketSpawnTimer <= 0 {
		g.rockets = append(g.rockets, entities.NewRocket())
		g.rocketSpawnTimer = g.randomRocketInterval()
	}
	for i := len(g.rockets) - 1; i >= 0; i-- {
		r := g.rockets[i]
		r.Update(g.Player.Y, dt)
		if r.IsOffscreen() {
			g.rockets = append(g.rockets[:i], g.rockets[i+1:]...)
		}
	}

	// 7. Update and spawn Coins
	g.coinSpawnTimer -= dt
	if g.coinSpawnTimer <= 0 {
		g.spawnCoins()
		g.coinSpawnTimer = 4.0
	}
	magnetActive := g.isPowerupOn && g.activePowerup == entities.PowerupMagnet
	playerPos := math2d.Vector{X: g.Player.X, Y: g.Player.Y}
	for i := len(g.coinsList) - 1; i >= 0; i-- {
		c := g.coinsList[i]
		c.Update(playerPos, magnetActive, dt)
		if c.Collected || c.IsOffscreen() {
			g.coinsList = append(g.coinsList[:i], g.coinsList[i+1:]...)
		}
	}

	// 8. Update and spawn Powerup Tokens
	g.powerupTimerNext -= dt
	if g.powerupTimerNext <= 0 {
		tokenY := 120.0 + rand.Float64()*260.0
		g.powerupTokens = append(g.powerupTokens, entities.NewPowerupToken(config.ScreenWidth+50, tokenY))
		g.powerupTimerNext = 20.0 + rand.Float64()*10.0
	}
	for i := len(g.powerupTokens) - 1; i >= 0; i-- {
		pt := g.powerupTokens[i]
		pt.Update(dt)
		if pt.Collected || pt.IsOffscreen() {
			g.powerupTokens = append(g.powerupTokens[:i], g.powerupTokens[i+1:]...)
		}
	}

	// 9. Check Collisions
	playerBounds := g.Player.Bounds()

	// Check Coins
	for _, c := range g.coinsList {
		if !c.Collected && math2d.CheckSATCollision(playerBounds, c.Bounds()) {
			c.Collected = true
			if g.isPowerupOn && g.activePowerup == entities.PowerupCoinBoost {
				g.coins += 2
			} else {
				g.coins++
			}
		}
	}

	// Check Powerup Tokens
	for _, pt := range g.powerupTokens {
		if !pt.Collected && math2d.CheckSATCollision(playerBounds, pt.Bounds()) {
			effect := pt.Activate()
			g.activatePowerup(effect)
		}
	}

	// Check Lethal Obstacles
	hitFatal := false

	// Zappers
	for i, z := range g.zappers {
		if math2d.CheckSATCollision(playerBounds, z.Bounds()) {
			if g.handlePlayerHit() {
				// Destroy zapper on shield absorb
				g.zappers = append(g.zappers[:i], g.zappers[i+1:]...)
				break
			} else {
				hitFatal = true
				break
			}
		}
	}

	// Lasers
	if !hitFatal {
		for _, l := range g.lasers {
			if l.IsActive() && math2d.CheckSATCollision(playerBounds, l.Bounds()) {
				if g.handlePlayerHit() {
					l.State = entities.LaserStateFinished // Disarm beam so it doesn't kill player on next frame
				} else {
					hitFatal = true
					break
				}
			}
		}
	}

	// Rockets
	if !hitFatal {
		for i, r := range g.rockets {
			if r.IsLaunched() && math2d.CheckSATCollision(playerBounds, r.Bounds()) {
				if g.handlePlayerHit() {
					g.rockets = append(g.rockets[:i], g.rockets[i+1:]...)
					break
				} else {
					hitFatal = true
					break
				}
			}
		}
	}

	if hitFatal {
		g.gameOver = true
		return SceneGameOver, nil
	}

	CheckAndUnlockAchievements(g.distance, g.coins, g.isPowerupOn && g.activePowerup == entities.PowerupShield, g.isPowerupOn && g.activePowerup == entities.PowerupGravitySwap)

	return SceneNone, nil
}

// handlePlayerHit absorbs the fatal hit if shield is active. Returns true if hit was absorbed.
func (g *GameplayScene) handlePlayerHit() bool {
	if g.isPowerupOn && g.activePowerup == entities.PowerupShield {
		g.isPowerupOn = false
		g.Player.SetShield(false)
		return true
	}
	return false
}

func (g *GameplayScene) activatePowerup(pType entities.PowerupType) {
	// Clean up any lingering previous powerup states
	g.Player.SetShield(false)
	if g.Player.IsGravityInverted {
		g.Player.ToggleGravity()
	}

	g.isPowerupOn = true
	g.activePowerup = pType

	switch pType {
	case entities.PowerupShield:
		g.powerupTimer = config.ShieldDuration
		g.Player.SetShield(true)
	case entities.PowerupGravitySwap:
		g.powerupTimer = config.GravitySwapDuration
	case entities.PowerupMagnet:
		g.powerupTimer = config.MagnetDuration
	case entities.PowerupCoinBoost:
		g.powerupTimer = config.CoinBoostDuration
	}
}

// Test helpers
func (g *GameplayScene) ActivatePowerupForTest(pType entities.PowerupType) {
	g.activatePowerup(pType)
}

func (g *GameplayScene) AddLaserForTest(l *entities.Laser) {
	g.lasers = append(g.lasers, l)
}

func (g *GameplayScene) IsPowerupActive() bool {
	return g.isPowerupOn
}

func (g *GameplayScene) GetDifficulty() config.Difficulty {
	return g.difficulty
}

func (g *GameplayScene) SimulateProgressForTest(dist float64, coins int) {
	g.distance = dist
	g.coins = coins
}

func (g *GameplayScene) CheckAchievementsForTest() {
	CheckAndUnlockAchievements(g.distance, g.coins, g.isPowerupOn && g.activePowerup == entities.PowerupShield, g.isPowerupOn && g.activePowerup == entities.PowerupGravitySwap)
}

func (g *GameplayScene) spawnZapper() {
	y := 100.0 + rand.Float64()*300.0
	angles := []float64{0.0, math.Pi / 6.0, -math.Pi / 6.0, math.Pi / 3.0}
	angle := angles[rand.Intn(len(angles))]
	g.zappers = append(g.zappers, entities.NewZapper(config.ScreenWidth+50, y, angle))
}

func (g *GameplayScene) spawnLaser() {
	y := 80.0 + rand.Float64()*340.0
	g.lasers = append(g.lasers, entities.NewLaser(y))
}

func (g *GameplayScene) spawnCoins() {
	patterns := [][]int{
		{2, 5}, // 2 rows, 5 cols
		{3, 4}, // 3 rows, 4 cols
		{1, 8}, // 1 long line
	}
	pat := patterns[rand.Intn(len(patterns))]
	y := 120.0 + rand.Float64()*260.0
	cluster := entities.NewCoinCluster(config.ScreenWidth+50, y, pat[0], pat[1])
	g.coinsList = append(g.coinsList, cluster.Coins...)
}

func (g *GameplayScene) randomZapperInterval() float64 {
	switch g.difficulty {
	case config.DifficultyEasy:
		return 3.0 + rand.Float64()*2.0
	case config.DifficultyMedium:
		return 2.0 + rand.Float64()*1.5
	default:
		return 1.2 + rand.Float64()*1.0
	}
}

func (g *GameplayScene) randomLaserInterval() float64 {
	switch g.difficulty {
	case config.DifficultyEasy:
		return 18.0 + rand.Float64()*5.0
	case config.DifficultyMedium:
		return 12.0 + rand.Float64()*4.0
	default:
		return 8.0 + rand.Float64()*3.0
	}
}

func (g *GameplayScene) randomRocketInterval() float64 {
	switch g.difficulty {
	case config.DifficultyEasy:
		return 8.0 + rand.Float64()*4.0
	case config.DifficultyMedium:
		return 5.0 + rand.Float64()*3.0
	default:
		return 3.5 + rand.Float64()*2.0
	}
}

func (g *GameplayScene) Draw(screen *ebiten.Image) {
	// 1. Draw seamlessly wrapping background
	bg := assets.GetImage("BackdropMain.png")
	if bg != nil {
		bw, bh := float64(bg.Bounds().Dx()), float64(bg.Bounds().Dy())
		scale := float64(config.ScreenHeight) / bh
		scaledW := bw * scale

		offset := math.Mod(g.bgScrollX, scaledW)

		op1 := &ebiten.DrawImageOptions{}
		op1.GeoM.Scale(scale, scale)
		op1.GeoM.Translate(-offset, 0)
		screen.DrawImage(bg, op1)

		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Scale(scale, scale)
		op2.GeoM.Translate(-offset+scaledW, 0)
		screen.DrawImage(bg, op2)
	}

	// 2. Draw Obstacles
	zapperSprite := assets.GetImage("Zapper1.png")
	for _, z := range g.zappers {
		z.Draw(screen, zapperSprite)
	}

	laserActiveSprite := assets.GetImage("laser_active.png")
	laserInactiveSprite := assets.GetImage("laser_noneactive.png")
	for _, l := range g.lasers {
		l.Draw(screen, laserActiveSprite, laserInactiveSprite)
	}

	rocketSprite := assets.GetImage("missle.png")
	warningSprite := assets.GetImage("warning.png")
	for _, r := range g.rockets {
		r.Draw(screen, rocketSprite, warningSprite)
	}

	// 3. Draw Collectibles & Powerup Tokens
	coinSprite := assets.GetImage("Coin.png")
	for _, c := range g.coinsList {
		c.Draw(screen, coinSprite)
	}

	tokenSprite := assets.GetImage("powerup.png")
	for _, pt := range g.powerupTokens {
		pt.Draw(screen, tokenSprite)
	}

	// 4. Draw Player
	barrySprite := assets.GetImage("barry.png")
	shieldSprite := assets.GetImage("force_shield.png")
	g.Player.Draw(screen, barrySprite, shieldSprite)

	// 5. Draw HUD
	hudFont := assets.GetFont("New Athletic M54.ttf", 26)
	labelFont := assets.GetFont("Roboto-Regular.ttf", 16)

	if hudFont != nil {
		// Distance in meters
		distText := fmt.Sprintf("%05d M", int(g.distance))
		text.Draw(screen, distText, hudFont, 30, 45, color.White)

		// Coins collected
		coinText := fmt.Sprintf("%d", g.coins)
		text.Draw(screen, coinText, hudFont, 65, 85, color.RGBA{255, 215, 0, 255})
		if coinSprite != nil {
			cOp := &ebiten.DrawImageOptions{}
			cOp.GeoM.Scale(24.0/float64(coinSprite.Bounds().Dx()), 24.0/float64(coinSprite.Bounds().Dy()))
			cOp.GeoM.Translate(32, 65)
			screen.DrawImage(coinSprite, cOp)
		}

		// Active Powerup banner
		if g.isPowerupOn && labelFont != nil {
			bannerText := fmt.Sprintf("POWERUP: %s (%.1fs)", g.activePowerup.String(), g.powerupTimer)
			text.Draw(screen, bannerText, labelFont, 30, 115, color.RGBA{0, 255, 255, 255})
		}
	}
}
