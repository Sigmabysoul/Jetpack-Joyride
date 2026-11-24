# Jetpack Joyride: Go + Ebitengine Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite the Jetpack Joyride clone from C/SDL2 into an idiomatic, modular Go codebase using Ebitengine v2 with full feature parity and instant Web (WASM) and Desktop execution.

**Architecture:** The application is structured into decoupled packages: 2D mathematics with SAT polygon collisions (`pkg/math2d`), embedded asset caching (`pkg/assets`), cross-platform persistence (`pkg/storage`), physical game entities and spawners (`pkg/entities`), state-machine scenes (`pkg/scenes`), and an Ebitengine coordinator (`pkg/game`) with desktop (`cmd/game`) and web (`web/`) entry points.

**Tech Stack:** Go 1.27+, Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2`), WebAssembly (`GOOS=js GOARCH=wasm`), Standard library `embed.FS`.

**Spec:** [`docs/superpowers/specs/2026-11-24-jetpack-joyride-go-rewrite-design.md`](file:///home/me/Work/games/docs/superpowers/specs/2026-11-24-jetpack-joyride-go-rewrite-design.md)

## Global Constraints
- Target virtual resolution: 1000 x 500 pixels.
- Zero external C toolchain dependencies (pure Go + WebAssembly).
- All assets embedded directly into the Go binary (`//go:embed`).
- Consistent 60 TPS gameplay physics.

## Review Focus
1. Player boundary clamping: Player must never clip through ceiling ($Y=450$) or ground ($Y=50$).
2. Separating Axis Theorem (SAT): Non-axis-aligned rotated zappers must register exact polygon collisions without false positives.
3. Powerup state isolation: Gravity swap must invert thrust behavior; shield must safely absorb fatal hits and deactivate without crashing.
4. Rocket evasion window: Rocket warning must lock position 0.75s before launch so player can evade.
5. WASM localStorage persistence: Achievements and high scores must gracefully persist in browser without failing in non-browser environments.

---

### Task 1: Module Setup & 2D Math Library (`pkg/math2d`)

**Files:**
- Create: `go.mod`
- Create: `pkg/math2d/vector.go`
- Create: `pkg/math2d/polygon.go`
- Create: `pkg/math2d/collision.go`
- Test: `pkg/math2d/math2d_test.go`

**Interfaces:**
- Produces:
  - `type Vector struct { X, Y float64 }`
  - `(v Vector) Add(other Vector) Vector`
  - `(v Vector) Sub(other Vector) Vector`
  - `(v Vector) Mul(s float64) Vector`
  - `(v Vector) Dot(other Vector) float64`
  - `(v Vector) Length() float64`
  - `(v Vector) Unit() Vector`
  - `(v Vector) Rotate(angleRad float64) Vector`
  - `type Polygon struct { Vertices []Vector }`
  - `NewRectangle(center Vector, width, height float64) Polygon`
  - `(p Polygon) Translate(v Vector) Polygon`
  - `(p Polygon) Rotate(center Vector, angleRad float64) Polygon`
  - `CheckSATCollision(p1, p2 Polygon) bool`

- [ ] **Step 1: Initialize go.mod and write failing math tests**
Write tests covering vector arithmetic, polygon transformations, and SAT polygon intersection (overlapping rectangles, disjoint rectangles, and rotated polygons).
Run: `go test ./pkg/math2d -v`
Expected: FAIL (types not defined).

- [ ] **Step 2: Implement Vector math and Polygon utilities in `pkg/math2d/vector.go` and `polygon.go`**
Implement 2D vector operations and polygon translation/rotation.

- [ ] **Step 3: Implement Separating Axis Theorem in `pkg/math2d/collision.go`**
Project all vertices of both polygons onto perpendicular edge normals; verify overlap on all axes.

- [ ] **Step 4: Run tests to verify they pass**
Run: `go test ./pkg/math2d -v`
Expected: PASS.

- [ ] **Step 5: Commit**
```bash
git add go.mod pkg/math2d
git commit -m "feat(math2d): implement 2D vector and SAT polygon collision library"
```

---

### Task 2: Game Configuration & Cross-Platform Storage (`pkg/config`, `pkg/storage`)

**Files:**
- Create: `pkg/config/constants.go`
- Create: `pkg/storage/storage.go`
- Create: `pkg/storage/storage_wasm.go`
- Create: `pkg/storage/storage_native.go`
- Test: `pkg/storage/storage_test.go`

**Interfaces:**
- Produces:
  - `pkg/config`: `ScreenWidth = 1000`, `ScreenHeight = 500`, `Gravity = 900.0`, `JetpackThrust = 1500.0`, `MaxVelocity = 400.0`, `GroundY = 50.0`, `CeilingY = 450.0`, obstacle timings and speeds.
  - `type Storage interface { Save(key string, data []byte) error; Load(key string) ([]byte, error) }`
  - `GetStorage() Storage`

- [ ] **Step 1: Write failing storage tests**
Test saving and loading arbitrary JSON keys using the memory/desktop implementation.
Run: `go test ./pkg/storage -v`
Expected: FAIL.

- [ ] **Step 2: Implement `pkg/config/constants.go`**
Define screen sizes, physics constants, entity dimensions, colors, and difficulty spawn parameters.

- [ ] **Step 3: Implement `pkg/storage` with WASM and Desktop build tags**
Use `syscall/js` with `localStorage` for `//go:build js && wasm`, and file storage in `os.UserConfigDir` for native builds.

- [ ] **Step 4: Run storage tests to verify PASS**
Run: `go test ./pkg/storage -v`
Expected: PASS.

- [ ] **Step 5: Commit**
```bash
git add pkg/config pkg/storage
git commit -m "feat(config,storage): add constants and cross-platform storage system"
```

---

### Task 3: Embedded Asset & Font Loader (`pkg/assets`)

**Files:**
- Create: `pkg/assets/assets.go`
- Test: `pkg/assets/assets_test.go`

**Interfaces:**
- Consumes: Assets directory (`assets/*.png`, `assets/*.jpg`, `assets/*.ttf`).
- Produces:
  - `LoadAssets() error`
  - `GetImage(name string) *ebiten.Image`
  - `GetFont(name string, size float64) font.Face`

- [ ] **Step 1: Write failing asset loading test**
Write a test that calls `LoadAssets()` and checks that all critical sprites (`Barry.png`, `BackdropMain.png`, `Coin.png`, `Zapper1.png`, `laser_active.png`, `missle.png`, `warning.png`, `powerup.png`) and fonts load as non-nil.
Run: `go test ./pkg/assets -v`
Expected: FAIL.

- [ ] **Step 2: Implement `pkg/assets/assets.go` using `//go:embed`**
Embed all assets into the Go package, decode PNG/JPEG images into `*ebiten.Image`, and parse TTF files into OpenType font faces.

- [ ] **Step 3: Run tests to verify PASS**
Run: `go test ./pkg/assets -v`
Expected: PASS.

- [ ] **Step 4: Commit**
```bash
git add pkg/assets
git commit -m "feat(assets): implement embedded asset and font loader"
```

---

### Task 4: Player Physics & Flight Mechanics (`pkg/entities/player.go`)

**Files:**
- Create: `pkg/entities/player.go`
- Test: `pkg/entities/player_test.go`

**Interfaces:**
- Consumes: `pkg/math2d`, `pkg/config`.
- Produces:
  - `type Player struct`
  - `NewPlayer(x, y float64) *Player`
  - `(p *Player) Update(thrusting bool, dt float64)`
  - `(p *Player) ToggleGravity()`
  - `(p *Player) Bounds() math2d.Polygon`
  - `(p *Player) SetShield(active bool)`

- [ ] **Step 1: Write failing player flight & boundary tests**
Test:
1. When `thrusting = true`, vertical velocity increases upward.
2. When `thrusting = false`, vertical velocity decreases downward due to gravity.
3. Player Y position is clamped within `GroundY` and `CeilingY`.
4. `ToggleGravity()` inverts the direction of gravitational pull and ceiling/floor bounds.
Run: `go test ./pkg/entities -run TestPlayer -v`
Expected: FAIL.

- [ ] **Step 2: Implement `Player` in `pkg/entities/player.go`**
Implement the physics updates, collision polygon generation, shield state, and bounds clamping.

- [ ] **Step 3: Run tests to verify PASS**
Run: `go test ./pkg/entities -run TestPlayer -v`
Expected: PASS.

- [ ] **Step 4: Commit**
```bash
git add pkg/entities/player.go pkg/entities/player_test.go
git commit -m "feat(entities): implement player physics, boundaries, and flight controls"
```

---

### Task 5: Obstacles Subsystem (`pkg/entities/zapper.go`, `laser.go`, `rocket.go`)

**Files:**
- Create: `pkg/entities/zapper.go`
- Create: `pkg/entities/laser.go`
- Create: `pkg/entities/rocket.go`
- Test: `pkg/entities/obstacles_test.go`

**Interfaces:**
- Consumes: `pkg/math2d`, `pkg/config`.
- Produces:
  - `type Zapper struct` with `Update(dt float64)`, `Bounds() math2d.Polygon`, `IsOffscreen() bool`
  - `type Laser struct` with `Update(dt float64)`, `IsActive() bool`, `Bounds() math2d.Polygon`
  - `type Rocket struct` with `Update(playerY float64, dt float64)`, `IsLaunched() bool`, `Bounds() math2d.Polygon`

- [ ] **Step 1: Write failing obstacles tests**
Test:
1. Zapper moves left at constant speed and reports offscreen when `X < -ZAPPER_WIDTH`.
2. Laser progresses from Inactive -> Active -> Off.
3. Rocket tracks player Y until lock-on timer (0.75s remaining), then launches horizontally at constant high speed.
Run: `go test ./pkg/entities -run TestObstacles -v`
Expected: FAIL.

- [ ] **Step 2: Implement Zapper, Laser, and Rocket entities**
Implement movement, timers, lock-on mechanics, and bounding polygons.

- [ ] **Step 3: Run tests to verify PASS**
Run: `go test ./pkg/entities -run TestObstacles -v`
Expected: PASS.

- [ ] **Step 4: Commit**
```bash
git add pkg/entities/zapper.go pkg/entities/laser.go pkg/entities/rocket.go pkg/entities/obstacles_test.go
git commit -m "feat(entities): implement zappers, lasers, and tracking rockets"
```

---

### Task 6: Collectibles & Powerups (`pkg/entities/coin.go`, `powerup.go`)

**Files:**
- Create: `pkg/entities/coin.go`
- Create: `pkg/entities/powerup.go`
- Test: `pkg/entities/collectibles_test.go`

**Interfaces:**
- Produces:
  - `type Coin struct`, `type CoinCluster struct`
  - `(c *Coin) Update(playerPos math2d.Vector, magnetActive bool, dt float64)`
  - `type PowerupType int` (Shield, GravitySwap, Magnet, CoinBoost)
  - `type PowerupToken struct`
  - `(p *PowerupToken) Update(dt float64)`

- [ ] **Step 1: Write failing collectibles tests**
Test:
1. Coins move leftward normally.
2. When `magnetActive = true`, coins accelerate toward player position.
3. Powerup tokens move leftward and generate a random powerup effect when collected.
Run: `go test ./pkg/entities -run TestCollectibles -v`
Expected: FAIL.

- [ ] **Step 2: Implement Coin, CoinCluster, and PowerupToken**
Implement clustering patterns, magnetic force attraction, and powerup activation timers.

- [ ] **Step 3: Run tests to verify PASS**
Run: `go test ./pkg/entities -run TestCollectibles -v`
Expected: PASS.

- [ ] **Step 4: Commit**
```bash
git add pkg/entities/coin.go pkg/entities/powerup.go pkg/entities/collectibles_test.go
git commit -m "feat(entities): implement coins, magnetic attraction, and powerups"
```

---

### Task 7: Scene State Machine & UI (`pkg/scenes/`)

**Files:**
- Create: `pkg/scenes/scene.go`
- Create: `pkg/scenes/home.go`
- Create: `pkg/scenes/settings.go`
- Create: `pkg/scenes/gameplay.go`
- Create: `pkg/scenes/gameover.go`
- Test: `pkg/scenes/scenes_test.go`

**Interfaces:**
- Produces:
  - `type SceneID int` (`SceneHome`, `SceneSettings`, `SceneGameplay`, `SceneGameOver`)
  - `type Scene interface { Update() (SceneID, error); Draw(screen *ebiten.Image) }`
  - Constructors: `NewHomeScene()`, `NewSettingsScene()`, `NewGameplayScene(diff Difficulty)`, `NewGameOverScene(score, coins int)`

- [ ] **Step 1: Write failing scene transition tests**
Test button click coordinate checking and scene transition return values.
Run: `go test ./pkg/scenes -v`
Expected: FAIL.

- [ ] **Step 2: Implement Home, Settings, Gameplay, and GameOver scenes**
- `Home`: Displays backdrop, Play and Settings buttons.
- `Settings`: Allows toggling difficulty (Easy, Medium, Hard), displays achievements.
- `Gameplay`: Handles infinite background scrolling, entity spawning, HUD rendering (distance, coins, active powerup), and collision checking.
- `GameOver`: Displays final run stats and Retry/Home buttons.

- [ ] **Step 3: Run scene tests to verify PASS**
Run: `go test ./pkg/scenes -v`
Expected: PASS.

- [ ] **Step 4: Commit**
```bash
git add pkg/scenes
git commit -m "feat(scenes): implement home, settings, gameplay, and gameover scenes"
```

---

### Task 8: Ebitengine Game Coordinator & Desktop App (`pkg/game`, `cmd/game/main.go`)

**Files:**
- Create: `pkg/game/game.go`
- Create: `cmd/game/main.go`
- Test: `pkg/game/game_test.go`

**Interfaces:**
- Produces:
  - `type Game struct` implementing `ebiten.Game` (`Update`, `Draw`, `Layout`).
  - Desktop executable in `cmd/game`.

- [ ] **Step 1: Write test for Game coordinator**
Verify that `Game.Layout(w, h)` returns `(1000, 500)` and `Game.Update()` delegates to active scene.
Run: `go test ./pkg/game -v`
Expected: PASS.

- [ ] **Step 2: Implement `Game` and `main.go`**
Wire up window initialization (`ebiten.SetWindowTitle("Jetpack Joyride")`, `ebiten.SetWindowSize(1000, 500)`, `ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)`).

- [ ] **Step 3: Verify native build**
Run: `go build -o bin/jetpack-joyride ./cmd/game`
Expected: Success, executable generated.

- [ ] **Step 4: Commit**
```bash
git add pkg/game cmd/game
git commit -m "feat(game): add Ebitengine coordinator and desktop entrypoint"
```

---

### Task 9: WebAssembly Runner & Build Pipeline (`web/`, `Makefile`)

**Files:**
- Create: `web/index.html`
- Create: `web/wasm_exec.js`
- Modify: `Makefile`

**Interfaces:**
- Produces:
  - Web runner at `web/index.html` loading `web/game.wasm`.
  - Makefile targets: `build-go-wasm`, `build-go-desktop`, `serve-go`.

- [ ] **Step 1: Copy standard Go `wasm_exec.js` and craft `web/index.html`**
Use `$(go env GOROOT)/lib/wasm/wasm_exec.js` (or `/misc/wasm/wasm_exec.js`).
Create modern `web/index.html` with responsive centered canvas and viewport meta tags.

- [ ] **Step 2: Add Go build targets to `Makefile`**
Add:
```makefile
build-go-wasm:
	GOOS=js GOARCH=wasm go build -o web/game.wasm ./cmd/game

serve-go: build-go-wasm
	python3 -m http.server 8080 --directory web
```

- [ ] **Step 3: Compile WebAssembly bundle**
Run: `make build-go-wasm`
Expected: `web/game.wasm` created successfully.

- [ ] **Step 4: Run full test suite**
Run: `go test ./...`
Expected: ALL PASS.

- [ ] **Step 5: Commit**
```bash
git add web Makefile
git commit -m "feat(web): add WebAssembly runner, wasm_exec.js, and build pipeline"
```
