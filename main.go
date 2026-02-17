package main

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const screenWidth = 1600
const screenHeight = 1200

const gravitationalConstant = 0.005
const screenBounceEfficiency = 0.5
const softeningParameter = 10.0

// Game implements ebiten.Game interface.
type Game struct {
	world     *World
	camera    *Camera
	input     *InputState
	renderer  *Renderer
	challenge *Challenge
	target    *TargetPractice
}

// Update proceeds the game state.
func (g *Game) Update() error {
	g.input.Update(g.world, g.camera, g.challenge, g.target)

	if !g.input.paused {
		steps := int(g.input.simSpeed * 2)
		if steps < 1 {
			steps = 1
		}
		for i := 0; i < steps; i++ {
			g.world.StepPhysics()
			g.challenge.Update(g.world)
			g.target.Update(g.world)
		}
	}
	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen, g.world, g.camera, g.input, g.challenge, g.target)
}

// Layout returns the logical screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// setupSolarSystem creates a scale model of the solar system with approximate
// planetary positions for 2026-02-17, computed from J2000 mean orbital elements.
func setupSolarSystem(world *World) {
	cx, cy := float64(screenWidth)/2, float64(screenHeight)/2

	// Sun (pinned at center, mass overridden for stable planetary orbits)
	sun := world.AddObject(cx, cy, 30)
	sun.pinned = true
	sun.color = [3]byte{255, 220, 50}
	sun.mass = 10000
	sunMass := sun.mass

	// Mean longitudes for 2026-02-17 (9545 days from J2000)
	// Computed: L = L0 + rate_per_day * 9545, then mod 360
	type body struct {
		angle    float64 // mean longitude in degrees
		distance float64 // pixels from sun
		radius   int
		color    [3]byte
	}

	planets := []body{
		{73, 80, 3, [3]byte{180, 160, 140}},   // Mercury
		{346, 130, 5, [3]byte{230, 200, 150}},  // Venus
		{148, 190, 5, [3]byte{100, 150, 255}},  // Earth
		{317, 260, 4, [3]byte{220, 100, 60}},   // Mars
		{108, 480, 14, [3]byte{200, 170, 130}},  // Jupiter
		{9, 720, 11, [3]byte{220, 200, 150}},    // Saturn
	}

	for _, p := range planets {
		rad := p.angle * math.Pi / 180
		px := cx + p.distance*math.Cos(rad)
		py := cy - p.distance*math.Sin(rad)

		// Circular orbit velocity: v = sqrt(G * M_sun / r)
		v := math.Sqrt(gravitationalConstant * sunMass / p.distance)
		// Counterclockwise orbit (screen Y-down): tangent = (-sin θ, -cos θ)
		vx := -v * math.Sin(rad)
		vy := -v * math.Cos(rad)

		obj := world.AddObject(px, py, p.radius)
		obj.velocityX = vx
		obj.velocityY = vy
		obj.color = p.color
	}
}

func main() {
	world := newWorld()
	setupSolarSystem(world)

	game := &Game{
		world:     world,
		camera:    newCamera(),
		input:     newInputState(),
		renderer:  newRenderer(),
		challenge: newChallenge(),
		target:    newTargetPractice(),
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Gravity Sandbox")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
