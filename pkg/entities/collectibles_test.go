package entities_test

import (
	"testing"

	"jetpack-joyride/pkg/entities"
	"jetpack-joyride/pkg/math2d"
)

func TestCollectibles(t *testing.T) {
	// 1. Regular Coin movement
	coin := entities.NewCoin(500.0, 200.0)
	initialX := coin.X
	coin.Update(math2d.Vector{X: 100, Y: 200}, false, 0.1)
	if coin.X >= initialX {
		t.Errorf("coin should move left normally, got X=%f", coin.X)
	}

	// 2. Magnetic attraction
	playerPos := math2d.Vector{X: 100, Y: 400}
	distBefore := math2d.Vector{X: coin.X, Y: coin.Y}.Sub(playerPos).Length()
	coin.Update(playerPos, true, 0.2) // magnet active
	distAfter := math2d.Vector{X: coin.X, Y: coin.Y}.Sub(playerPos).Length()
	if distAfter >= distBefore {
		t.Errorf("coin should be drawn towards player with magnet active, before=%f, after=%f", distBefore, distAfter)
	}

	// 3. Coin Cluster
	cluster := entities.NewCoinCluster(1000.0, 200.0, 3, 4) // 3 rows, 4 cols
	if len(cluster.Coins) != 12 {
		t.Errorf("expected 12 coins in cluster, got %d", len(cluster.Coins))
	}

	// 4. Powerup Token
	token := entities.NewPowerupToken(800.0, 250.0)
	initialTokenX := token.X
	token.Update(0.1)
	if token.X >= initialTokenX {
		t.Errorf("powerup token should move left, got X=%f", token.X)
	}
	effect := token.Activate()
	if effect < entities.PowerupShield || effect > entities.PowerupCoinBoost {
		t.Errorf("expected valid powerup effect, got %v", effect)
	}
}
