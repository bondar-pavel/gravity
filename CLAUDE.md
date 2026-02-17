# Gravity Simulation

## Overview

2D gravity simulation and particle sandbox built with Go and [Ebitengine](https://ebitengine.org/). Particles attract each other via Newtonian gravity, collide elastically, and bounce off screen edges. Users click to create particles or drag existing ones.

## Tech Stack

- **Language:** Go 1.21
- **Engine:** Ebitengine v2.6.3 (2D game engine — handles window, input, rendering loop at 60 FPS)
- **Build targets:** Native binary, WebAssembly

## Project Structure

```
main.go          — All source code (~334 lines): physics, rendering, input, game loop
Makefile         — Build targets: build, build-wasm, run
go.mod / go.sum  — Dependencies
```

## Build & Run

```sh
make run           # Build and launch native binary
make build         # Compile to bin/gravity
make build-wasm    # Compile to bin/gravity.wasm
```

## Architecture

Single-file architecture with three main types:

| Type | Role |
|------|------|
| `Object` | Particle with position, velocity, mass, radius |
| `Map` | World state — object list, pixel buffer, feature flags |
| `Game` | Ebitengine interface — bridges Map to the engine's Update/Draw loop |

### Game Loop (60 FPS)

```
Game.Update() → Map.Update() → handle mouse input
                              → Map.ObjectsToPixels():
                                  for each object:
                                    1. Calculate gravitational forces from all others
                                    2. Update velocity (apply forces)
                                    3. Update position (Euler integration)
                                    4. Check object-object collisions
                                    5. Check screen-edge collisions
                                    6. Rasterize circle to pixel buffer

Game.Draw()   → Map.Draw()   → Convert grayscale buffer to RGBA
                              → screen.WritePixels()
```

### Physics Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `gravity` | 0.0001 | Constant downward acceleration |
| `graviationalConstant` | 0.005 | Mutual attraction strength |
| `friction` | 0.01 | Velocity damping (vertical only) |
| `screenBounceEfficiency` | 0.5 | Energy retained on wall bounce |
| `screenWidth` / `screenHeight` | 1600 / 1200 | Logical resolution |

### Physics Model

- **Gravitation:** Inverse-square law between all pairs. Force ∝ `G * (m_other / m_self) * displacement / distance²`
- **Integration:** Simple Euler (`position += velocity`, `velocity += force`)
- **Collisions:** Impulse-based elastic collision along the normal vector; mass-weighted
- **Boundaries:** Velocity reversal with 50% energy loss at screen edges
- **Friction:** Applied only to vertical velocity as `|friction * velocityY|`

### Rendering

- Grayscale pixel buffer (`[]byte`, one byte per pixel)
- Filled circles rasterized via distance check within bounding box
- Optional sub-pixel anti-aliasing (`shadeHalfCoveredPixels` flag, currently off)
- Grayscale → RGBA conversion for Ebitengine display

### Controls

- **Left-click empty space** → Create particle (radius 10)
- **Left-click + drag on particle** → Move it (teleport to cursor)

### Feature Flags (in `newMap()`)

| Flag | Default | Effect |
|------|---------|--------|
| `bounceOnScreenCollision` | true | Particles bounce off screen edges |
| `bounceOnParticleCollision` | true | Particles bounce off each other |
| `shadeHalfCoveredPixels` | false | Sub-pixel edge anti-aliasing |

## Known Issues & Quirks

1. **Friction only vertical** — `UpdateVelocityGravitational()` applies friction only to `velocityY`, not `velocityX`
2. **Particle spawning on hold** — Holding left-click continuously spawns particles every frame (no debounce)
3. **Drag resets velocity** — Dragging teleports the object without resetting velocity, so it launches away on release
4. **Collision bounce frames commented out** — Lines 113-114 are commented, meaning overlapping objects can trigger multiple collision responses per frame
5. **No object deletion** — Once created, objects persist forever
6. **Alpha channel set to grayscale value** — `pixels[4*i+3] = v` makes dark particles transparent instead of using `0xFF`
7. **Earth gravity always on** — `UpdateVelocityGravitational()` is not called in the current loop (only mutual gravitation is applied via `CalculateGraviationalForce`)

---

## Proposal: Gamified Gravity Sandbox

### Vision

Transform the current simulation into an engaging gravity sandbox game with objectives, interactive controls, and polished physics. Think "Angry Birds meets Universe Sandbox" — a playground where gravity is both the toy and the challenge.

### Phase 1: Controls & Interaction Overhaul

**Slingshot Launch (high priority)**
- Click and drag on empty space to define a velocity vector (rubber-band visual)
- Release to spawn a particle with that initial velocity
- Show a dotted trajectory preview while aiming

**Particle Size Control**
- Scroll wheel or `[`/`]` keys to adjust the radius of the next particle before placing it
- Show a ghost preview circle at cursor position
- Larger particles = more mass = stronger gravitational pull

**Object Selection & Deletion**
- Right-click to select a particle (highlight it)
- `Delete`/`Backspace` to remove the selected particle
- `Space` to pin/unpin a particle (make it immovable — acts as a gravity anchor)

**Camera Controls**
- Scroll to zoom in/out (adjust the logical-to-window mapping)
- Middle-click drag to pan the viewport
- `Home` key to reset view

**Time Control**
- `+`/`-` or `↑`/`↓` to speed up / slow down simulation
- `P` to pause/unpause
- While paused, allow placing and aiming particles, then unpause to watch them fly

### Phase 2: Physics Improvements

**Fix Euler Integration → Velocity Verlet**
- Current Euler integration (`v += a; x += v`) accumulates energy errors over time
- Switch to Velocity Verlet for stable orbits:
  ```
  x(t+dt) = x(t) + v(t)*dt + 0.5*a(t)*dt²
  v(t+dt) = v(t) + 0.5*(a(t) + a(t+dt))*dt
  ```

**Softened Gravity**
- Add a softening parameter ε to prevent singularity when particles are very close:
  `force = G * m1 * m2 / (distance² + ε²)`
- Prevents infinite forces and numeric explosions at close range

**Fix Friction Model**
- Apply friction to both axes as a drag force: `velocity *= (1 - friction)`
- Make it optional (toggle on/off) since space has no friction

**Proper Collision Separation**
- On collision, separate overlapping objects along the collision normal before applying impulse
- Re-enable `bouncedFrames` with a 1-frame cooldown
- Add coefficient of restitution (0 = perfectly inelastic, 1 = perfectly elastic) as a tunable parameter

**Merge on Collision (optional mode)**
- When two particles collide, optionally merge them:
  - Combined mass = m1 + m2
  - Conserve momentum: new velocity = (m1*v1 + m2*v2) / (m1 + m2)
  - New radius = sqrt(r1² + r2²) (area-preserving)

**Gravitational Field Toggle**
- On-demand overlay showing the gravitational field strength as a heatmap
- Color gradient from blue (weak) → red (strong)
- Helps players visualize orbits and plan shots

### Phase 3: Gamification

**Orbit Challenge Mode**
- Goal: launch a small particle into a stable orbit around a fixed "planet"
- Score based on how many complete orbits before it escapes or crashes
- Difficulty tiers: single planet → binary star → triple star chaos

**Target Practice Mode**
- Place target zones on the screen
- Launch particles that must pass through all targets using gravity assists
- Par system (minimum number of launches) like golf
- Star rating: 3 stars for par, 2 for par+1, 1 for par+2

**Sandbox Achievements**
- "First Contact" — two particles collide for the first time
- "Solar System" — get 5+ particles orbiting a central mass
- "Black Hole" — create a particle with mass > 10000 (via merging)
- "Butterfly Effect" — a tiny particle causes a chain reaction of 5+ collisions
- "Stable Orbit" — maintain a particle in orbit for 60 seconds

**Level System**
- Predefined initial configurations (particle positions, masses, fixed anchors)
- JSON-based level format for easy creation:
  ```json
  {
    "name": "Binary Star",
    "objects": [
      {"x": 400, "y": 600, "radius": 40, "pinned": true},
      {"x": 1200, "y": 600, "radius": 40, "pinned": true}
    ],
    "targets": [
      {"x": 800, "y": 300, "radius": 30}
    ],
    "launches": 3
  }
  ```

### Phase 4: Visual Polish

**Particle Trails**
- Fade-out trail behind each particle showing its trajectory
- Trail length proportional to velocity
- Different colors per particle for visual tracking

**Color by Property**
- Color particles by mass (small=blue, large=red) or velocity (slow=cool, fast=warm)
- Glow effect on fast-moving particles

**HUD Overlay**
- Particle count, simulation speed, current mode
- Selected particle info: mass, velocity, position
- FPS counter

**Background Star Field**
- Subtle parallax star background for depth
- Optional grid lines for spatial reference

### Implementation Priority

| Priority | Feature | Effort |
|----------|---------|--------|
| 1 | Slingshot launch + trajectory preview | Medium |
| 2 | Pause/play + time control | Small |
| 3 | Velocity Verlet integration | Small |
| 4 | Softened gravity + fix friction | Small |
| 5 | Particle size control + ghost preview | Small |
| 6 | Object deletion + pinning | Small |
| 7 | Particle trails | Medium |
| 8 | Color by property + HUD | Medium |
| 9 | Collision separation + merge mode | Medium |
| 10 | Orbit challenge mode | Large |
| 11 | Target practice mode + levels | Large |
| 12 | Gravitational field heatmap | Medium |
| 13 | Camera pan/zoom | Medium |
| 14 | Achievements system | Medium |
