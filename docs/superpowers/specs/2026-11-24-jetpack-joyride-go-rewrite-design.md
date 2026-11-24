# Jetpack Joyride: Go + Ebitengine Modern Rewrite Design

## 1. Intent and the Why
The goal is to port and modernize the existing Caltech CS 3 Jetpack Joyride C codebase into an idiomatic, modular, and extensible Go implementation using **Ebitengine (v2)**. The user plans to significantly improve and expand this game in the future (e.g. adding new vehicles, enemy types, particle effects, sound effects, powerups, and custom game modes). 

The C codebase relied on raw manual memory management, brittle C-SDL2 bindings, and custom polygon wrappers compiled via a legacy Emscripten pipeline. Rewriting in Go with Ebitengine provides:
- **Instant Web (WASM) & Desktop execution:** One codebase builds seamlessly to WebAssembly for browsers (`GOOS=js GOARCH=wasm`) and native desktop (Linux/macOS/Windows) with zero C compiler or complex toolchain friction.
- **Superior developer ergonomics & extensibility:** Fast builds, garbage collection, and clean interface-based decoupling make adding new game mechanics and entities rapid and safe.
- **Self-contained asset packaging:** Embedding all textures and fonts via Go's standard `//go:embed` eliminates separate asset bundle syncing issues.

## 2. Goals, Non-Goals, and Anti-Goals

### Goals
- **100% Feature Parity with Original Game:**
  - Full player physics (vertical motion, gravity, jetpack thrust on Space/Left Click).
  - Continuous seamless scrolling background.
  - All four core scenes: **Home**, **Settings** (difficulty toggle + achievements), **GamePlay**, and **GameOver**.
  - All original obstacles with identical behaviors:
    - **Zappers:** Spawned in random orientations and vertical positions, moving left.
    - **Lasers:** Warning / inactive state with countdown, followed by full-screen lethal beam activation.
    - **Rockets:** Warning icon on the right following player's Y level, then launched horizontally.
  - Collectibles & Scoring: Coin clusters, distance counter, coin counter, and difficulty modifiers.
  - All 4 Powerups:
    - **Shield:** Protects against fatal collisions.
    - **Gravity Swap:** Inverts gravity direction upon Space/Click.
    - **Coin Magnet:** Attracts coins toward Barry.
    - **Extra Coins:** Multiplies or increases coin density.
  - Separating Axis Theorem (SAT) / Polygon bounding collision system for accurate hitboxes.
  - Web browser playability via `wasm_exec.js` and an HTML5 Canvas runner.
  - Native desktop playability (`go run main.go`).

### Non-Goals
- Adding brand new game mechanics or unrequested assets in this initial rewrite phase (new features will be built on top of this clean foundation in subsequent iterations).
- Implementing networking / multiplayer.

### Anti-Goals
- Complex framework bloat: We will not use heavyweight 3D engines or overly abstract entity component frameworks (ECS) when idiomatic Go interfaces and composition are cleaner and simpler to maintain.
- Hardcoded screen coordinates: Magic numbers will be centralized in a dedicated configuration package.

## 3. Constraints
- **Target Resolution:** Virtual resolution of 1000x500 (matching original `WINDOW_WIDTH` and `WINDOW_HEIGHT`), automatically scaled to window/browser canvas dimensions by Ebitengine with aspect ratio preservation.
- **Framerate:** 60 ticks/second (standard Ebitengine TPS).
- **Controls:** Spacebar, Up Arrow, or Mouse/Touch Left Click for jetpack thrust and gravity swap; Mouse click for menu buttons.
- **Browser Compatibility:** Pure WebAssembly + WebGL 1/2 supported by Chrome, Firefox, Safari, and Edge.

## 4. Approaches Considered
1. **Go with Ebitengine v2 (Selected):**
   - *Pros:* Ebitengine is the premier 2D game engine in Go. Out-of-the-box WebAssembly compilation (`GOOS=js GOARCH=wasm`), standard library asset embedding (`embed.FS`), fast iteration, zero C dependencies, native sound/font support, clean 2D drawing pipeline.
   - *Cons:* Slightly larger WASM binary (~5-10MB uncompressed) compared to hand-tuned C/WASM, but loads in milliseconds.
2. **Rust with Macroquad / Miniquad:**
   - *Pros:* Small WASM binary, zero-cost abstractions.
   - *Cons:* Slower iteration speed for game design changes; requires managing ownership/lifetimes across interconnected game scenes and entities; needs separate `wasm-bindgen` and toolchain setup.
3. **Keep C with Emscripten:**
   - *Pros:* Already written.
   - *Cons:* Manual memory management (`malloc`/`free`), pointer bugs, cumbersome to extend with new features, heavy Emscripten build requirements.

*Why Go + Ebitengine was chosen:* It provides the best balance of maintainability, rapid feature development for future upgrades, cross-platform deployment, and clean architecture.

## 5. System Architecture & Modular Design

```
jetpack-go/
├── assets/                  # Original image and font assets (embedded into binary)
├── cmd/
│   └── game/
│       └── main.go          # Entry point for desktop and web
├── web/
│   ├── index.html           # WebAssembly HTML runner with responsive canvas
│   └── wasm_exec.js         # Go WebAssembly runtime bridge
├── pkg/
│   ├── config/              # Central game constants (gravity, speeds, dimensions, timing)
│   ├── math2d/              # Vector2D, Polygon math, SAT collision detection
│   ├── storage/             # Cross-platform persistent storage (localStorage for WASM, JSON for desktop)
│   ├── assets/              # Asset loader, embedded filesystem, image/font cache
│   ├── entities/            # Player, Zappers, Lasers, Rockets, Coins, Powerups
│   │   ├── player.go        # Barry state, jetpack physics, collision bounds
│   │   ├── zapper.go        # Zapper obstacle generator & movement
│   │   ├── laser.go         # Laser inactive/active state machine
│   │   ├── rocket.go        # Tracking warning & launch mechanics
│   │   ├── coin.go          # Coin formations & magnetic attraction
│   │   └── powerup.go       # Shield, GravitySwap, Magnet, CoinBoost
│   ├── scenes/              # Scene interface and scene implementations
│   │   ├── scene.go         # Scene interface (Update, Draw)
│   │   ├── home.go          # Main menu (Play, Settings)
│   │   ├── settings.go      # Difficulty selector, Achievements view
│   │   ├── gameplay.go      # Primary game loop, spawner, HUD, collision coordinator
│   │   └── gameover.go      # Score summary, Retry, Back to Menu
│   └── game/
│       └── game.go          # Implements ebiten.Game (Update, Draw, Layout)
├── Makefile                 # Builds desktop binary and web WASM bundle
└── go.mod
```

### Key Subsystems & Architectural Calls:
1. **Player Flight Dynamics:**
   - Provides a smooth, responsive Newtonian acceleration model (gravity $g = 900\text{ px/s}^2$, jetpack thrust impulse $1500\text{ px/s}^2$, terminal velocity clamped to $\pm 400\text{ px/s}$) to give fluid flight feel and enable future vehicle upgrades.
   - Includes a configuration switch in `pkg/config` (`UseKinematicSnap`) allowing instant fallback to the C codebase's rigid $\pm 200\text{ px/s}$ step velocities.
2. **Rocket Obstacle Behavior:**
   - Multi-phase state machine:
     - `WarningTracking`: Rocket warning indicator on the right edge follows Barry's vertical coordinate with slight damping.
     - `WarningLocked`: 0.75 seconds before launch, the indicator locks onto its final $Y$ position and flashes, providing an evasion window.
     - `Launched`: Rocket fires horizontally from right to left at high speed ($600\text{ px/s}$).
3. **Cross-Platform Storage (`pkg/storage`):**
   - Implements a unified interface for achievements, high scores, and settings:
     - On WASM (`GOOS=js`): interfaces with browser `localStorage` via `syscall/js`.
     - On Desktop: writes JSON to the user's configuration directory (`os.UserConfigDir`).
     - Graceful in-memory fallback if persistence fails.
4. **Math & Collision (`pkg/math2d`):**
   - `Vector`: 2D vector operations (Add, Sub, Mul, Dot, Length, Rotate, Unit).
   - `Polygon`: Convex polygon vertices, bounding box, centroid, translation.
   - `Collision (SAT)`: Separating Axis Theorem implementation to test arbitrary polygon collisions between Barry and obstacles (zappers, lasers, rockets, coins).
5. **Entity Management (`pkg/entities`):**
   - Clean structs for each entity with distinct update and render logic.
   - **Player (Barry):** Handles velocity, gravity acceleration, thrust impulse, ceiling/floor clamping, shield bubble rendering, and gravity flip state.
   - **Spawners:** Deterministic or pseudo-random spawning based on difficulty settings (Easy, Medium, Hard), matching the original timing intervals.
6. **Scene State Machine (`pkg/scenes`):**
   - Standard `Scene` interface: `Update() (SceneID, error)` and `Draw(screen *ebiten.Image)`.
   - Transitions between `Home` -> `Settings` -> `Gameplay` -> `GameOver`.
7. **Asset & Font Loader (`pkg/assets`):**
   - Uses `//go:embed ../../assets/*` so the game is completely self-contained in a single executable or WASM file.
   - Parses TTF fonts (`Roboto-Regular.ttf` and `New Athletic M54.ttf`) using `golang.org/x/image/font/opentype`.
8. **Web Runner (`web/`):**
   - Clean, modern HTML5 page styled to center the canvas, auto-focus keyboard inputs, and support full-screen responsive scaling.

## 6. What's Left to the Builder
- Precise bounding polygon coordinates for each sprite to guarantee pixel-accurate collision detection.
- Fine-tuning font rendering DPI/size in HUD elements for crisp display.
- Exact audio hooks structure for when the user wants to add sound effects and music in future expansions.
