# Gravity Sandbox

2D N-body gravity simulation built with Go and [Ebitengine](https://ebitengine.org/). Launch particles with a slingshot, pin gravity anchors, and watch orbits form.

## Build & Run

```sh
make run           # build and launch
make build         # compile to bin/gravity
make build-wasm    # compile to WebAssembly
```

## Controls

| Input | Action |
|-------|--------|
| **Left-click + drag** on empty space | Slingshot aim — release to launch |
| **Left-click + drag** on particle | Drag it around |
| **Right-click** on particle | Select it |
| **Right-click** on empty space | Deselect |
| **Delete / Backspace** | Remove selected particle |
| **Space** | Pin/unpin selected particle (fixed gravity anchor) |
| **`[` / `]`** | Decrease / increase brush size |
| **Scroll wheel** | Zoom in/out |
| **Middle-click drag** | Pan camera |
| **Home** | Reset camera |
| **P** | Pause / unpause |
| **`+` / `-`** | Speed up / slow down simulation |
| **F** | Toggle friction (drag force on all particles) |
| **M** | Toggle merge mode (colliding particles merge) |
| **G** | Toggle gravity field heatmap |

## Physics

- **Velocity Verlet** integration for stable, energy-conserving orbits
- **Softened gravity** prevents singularities when particles are close
- **Collision separation** — overlapping particles are pushed apart before impulse
- **Restitution** — configurable bounciness (0 = inelastic, 1 = elastic, default 0.8)
- **Merge mode** — colliding particles combine mass and conserve momentum
- **Friction** — optional velocity drag on both axes
