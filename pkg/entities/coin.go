package entities

import (
	"jetpack-joyride/pkg/config"
	"jetpack-joyride/pkg/math2d"

	"github.com/hajimehoshi/ebiten/v2"
)

// Coin represents a collectible golden coin.
type Coin struct {
	X         float64
	Y         float64
	Collected bool
}

// NewCoin creates a single coin at (x, y).
func NewCoin(x, y float64) *Coin {
	return &Coin{
		X:         x,
		Y:         y,
		Collected: false,
	}
}

// Update moves the coin leftward and applies magnetic attraction towards player if magnet is active.
func (c *Coin) Update(playerPos math2d.Vector, magnetActive bool, dt float64) {
	if c.Collected {
		return
	}

	// Base leftward scroll
	c.X -= config.CoinSpeed * dt

	if magnetActive {
		coinPos := math2d.Vector{X: c.X, Y: c.Y}
		diff := playerPos.Sub(coinPos)
		dist := diff.Length()

		// Magnet attraction radius of 400 pixels
		if dist > 0 && dist < 400.0 {
			dir := diff.Unit()
			speed := 500.0 // Attraction speed
			c.X += dir.X * speed * dt
			c.Y += dir.Y * speed * dt
		}
	}
}

// IsOffscreen returns true if the coin has scrolled past the left edge.
func (c *Coin) IsOffscreen() bool {
	return c.X < -config.CoinSize*2
}

// Bounds returns the collision polygon for the coin.
func (c *Coin) Bounds() math2d.Polygon {
	return math2d.NewRectangle(math2d.Vector{X: c.X, Y: c.Y}, config.CoinSize, config.CoinSize)
}

// Draw renders the coin sprite.
func (c *Coin) Draw(screen *ebiten.Image, sprite *ebiten.Image) {
	if c.Collected || sprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	w := float64(sprite.Bounds().Dx())
	scale := config.CoinSize / w

	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-config.CoinSize/2, -config.CoinSize/2)

	screenY := float64(config.ScreenHeight) - c.Y
	op.GeoM.Translate(c.X, screenY)

	screen.DrawImage(sprite, op)
}

// CoinCluster groups multiple coins spawned in a grid pattern.
type CoinCluster struct {
	Coins []*Coin
}

// NewCoinCluster spawns a rectangular grid of coins with given rows and columns.
func NewCoinCluster(startX, startY float64, rows, cols int) *CoinCluster {
	coins := make([]*Coin, 0, rows*cols)
	spacing := config.CoinSize * 1.3

	for r := 0; r < rows; r++ {
		for col := 0; col < cols; col++ {
			x := startX + float64(col)*spacing
			y := startY + float64(r)*spacing
			coins = append(coins, NewCoin(x, y))
		}
	}

	return &CoinCluster{Coins: coins}
}
