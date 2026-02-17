# Gravity Simulation

## Overview

2D gravity simulation and gamified particle sandbox built with Go and [Ebitengine](https://ebitengine.org/). Features a solar system simulation, slingshot particle launching, merge physics with animations, two game modes (Orbit Challenge and Target Practice), and a full HUD with camera controls. Starts paused with a solar system model.

## Tech Stack

- **Language:** Go 1.21
- **Engine:** Ebitengine v2.6.3 (2D game engine — handles window, input, rendering loop at 60 FPS)
- **Build targets:** Native binary, WebAssembly

## Project Structure

```
main.go       — Game struct, constants, solar system setup, entry point
object.go     — Object struct, Velocity Verlet physics, collision, merge with angular momentum
world.go      — World state, physics loop, collision handling, ejecta particle system, culling
camera.go     — Camera with pan/zoom, world↔screen coordinate transforms
input.go      — Input handling for sandbox, challenge, and target modes, key debouncing
render.go     — All rendering: objects, HUD, slingshot, trajectories, gravity field, game mode visuals
challenge.go  — Orbit Challenge game mode: state machine, orbit tracking, level definitions
target.go     — Target Practice game mode: state machine, target zones, star scoring, level definitions
Makefile      — Build targets: build, build-wasm, run
go.mod/go.sum — Dependencies
```

## Build & Run

```sh
make run           # Build and launch native binary
make build         # Compile to bin/gravity
make build-wasm    # Compile to bin/gravity.wasm
```

## Architecture

### Core Types

| Type | File | Role |
|------|------|------|
| `Object` | object.go | Particle with position, velocity, mass, radius, rotation, merge animation state |
| `World` | world.go | Object list, ejecta particles, physics loop, feature flags, collision handling |
| `Camera` | camera.go | View center + zoom, world↔screen coordinate conversion |
| `InputState` | input.go | Mouse/keyboard state, mode-specific input handlers |
| `Renderer` | render.go | Pixel buffer rendering, HUD text, all visual effects |
| `Challenge` | challenge.go | Orbit Challenge game mode state machine |
| `TargetPractice` | target.go | Target Practice game mode state machine |
| `Game` | main.go | Ebitengine interface, wires everything together |

### Game Loop (60 FPS)

```
Game.Update() → InputState.Update()        — handle keyboard/mouse per active mode
              → if !paused:
                  for steps (simSpeed * 2):
                    World.StepPhysics()     — Verlet integration, gravity, collisions, culling
                    Challenge.Update()      — orbit tracking (if active)
                    TargetPractice.Update() — target hit detection (if active)

Game.Draw()   → Renderer.Draw()
                → clear pixel buffer
                → draw gravity field heatmap (if enabled)
                → draw all objects (with merge animation, rotation, selection highlight)
                → draw ejecta debris particles
                → draw mode-specific visuals (orbit zone / target zones / ghost preview / slingshot)
                → draw trajectory projections (if enabled)
                → screen.WritePixels()
                → draw HUD text overlay (scaled 2x for readability)
```

### Physics Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `gravitationalConstant` | 0.005 | Mutual attraction strength (G) |
| `softeningParameter` | 10.0 | Prevents singularity at close range: F ∝ 1/(r² + ε²) |
| `screenBounceEfficiency` | 0.5 | Energy retained on wall bounce (when enabled) |
| `cullDistance` | 5000 | Remove objects beyond this distance from screen center |
| `screenWidth` / `screenHeight` | 1600 / 1200 | Logical resolution |

### Physics Model

- **Gravitation:** Softened inverse-square law between all pairs. `a = G * m_other / (r² + ε²) * unit_vector`
- **Integration:** Velocity Verlet (3-phase: position update → recalculate accelerations → velocity update)
- **Collisions:** Impulse-based with overlap separation along normal; configurable restitution (default 0.8)
- **Merge:** When enabled, colliding unpinned objects merge conserving linear and angular momentum; area-preserving radius; triggers multi-phase visual animation (flash → shockwave → ejecta)
- **Friction:** Optional velocity damping `v *= (1 - frictionCoeff)` on both axes
- **Boundaries:** Screen bounce (disabled by default); distant object culling (always on)
- **Mass model:** `mass = radius²` by default; sun mass overridden to 10000 for stable orbits

### Rendering

- RGBA pixel buffer (`[]byte`, 4 bytes per pixel) written directly via `screen.WritePixels()`
- Objects: filled circles with rotation lines, selection highlight ring, merge animation (triple shockwave rings, white-hot flash with color cooling, glow halo, size oscillation)
- Ejecta: cosmetic debris particles with color gradient (white → yellow → orange → red) and drag
- Trajectories: forward-simulated dotted paths (200 steps), colored per object
- HUD: drawn to off-screen image at half resolution, scaled 2x onto screen via `DrawImageOptions`
- Gravity field: heatmap overlay sampled on 8px grid, blue (weak) → red (strong)

### Default Scene

Solar system model with approximate planetary positions for 2026-02-17 (J2000 mean orbital elements):
- **Sun:** radius 30, pinned, mass overridden to 10000
- **Mercury–Saturn:** 6 planets at computed orbital distances (80–720px) with circular orbit velocities
- Starts **paused** — press `P` to begin simulation

### Controls

**Sandbox Mode:**

| Key | Action |
|-----|--------|
| LMB (empty space) | Slingshot aim — drag and release to launch particle |
| LMB (on object) | Drag to reposition |
| RMB | Select object |
| `[` / `]` | Decrease / increase next particle radius |
| `Delete` / `Backspace` | Remove selected object |
| `Space` | Pin/unpin selected object |
| `P` | Pause / unpause |
| `+` / `-` | Speed up / slow down simulation (0.25x–4x) |
| Scroll wheel | Zoom in/out |
| Middle-click drag | Pan camera |
| `Home` | Reset camera |
| `F` | Toggle friction |
| `M` | Toggle merge-on-collision |
| `G` | Toggle gravitational field heatmap |
| `V` | Toggle trajectory projections for all objects |
| `O` | Enter Orbit Challenge mode |
| `T` | Enter Target Practice mode |

**Orbit Challenge Mode:**

| Key | Action |
|-----|--------|
| LMB | Slingshot launch orbiter |
| `←` / `→` | Change level (when aiming) |
| `P` / `+` / `-` | Time controls |
| `O` / `Esc` | Exit to sandbox |

**Target Practice Mode:**

| Key | Action |
|-----|--------|
| LMB | Slingshot launch projectile |
| `←` / `→` | Change level (when not flying) |
| LMB (after complete) | Retry level |
| `P` / `+` / `-` | Time controls |
| `T` / `Esc` | Exit to sandbox |

### Game Modes

**Orbit Challenge** (`O` key) — Launch a particle into orbit around fixed planets. Score = number of complete orbits before crash or escape. 4 levels: Single Planet, Binary Star, Triple Chaos, Giant and Moon. Best scores tracked per session. Merge mode enabled.

**Target Practice** (`T` key) — Launch particles through target zones using gravity assists. Golf-style par scoring (3/2/1/0 stars based on launches vs par). 4 levels: Straight Shot, Gravity Sling, Thread the Needle, Around the World. Multiple launches per attempt; targets persist across launches. Merge mode disabled.

Both modes save/restore the sandbox state on enter/exit and are mutually exclusive.

### World Feature Flags (in `newWorld()`)

| Flag | Default | Effect |
|------|---------|--------|
| `bounceOnScreenCollision` | false | Particles bounce off screen edges |
| `bounceOnParticleCollision` | true | Elastic collisions between particles |
| `mergeOnCollision` | true | Colliding unpinned particles merge |
| `frictionEnabled` | false | Velocity damping on both axes |
| `restitution` | 0.8 | Coefficient of restitution for collisions |
| `frictionCoeff` | 0.001 | Friction damping factor per tick |
