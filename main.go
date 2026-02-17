package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const screenWidth = 1600
const screenHeight = 1200

const gravitationalConstant = 0.005
const screenBounceEfficiency = 0.5
const softeningParameter = 10.0

// Game implements ebiten.Game interface.
type Game struct {
	world    *World
	camera   *Camera
	input    *InputState
	renderer *Renderer
}

// Update proceeds the game state.
func (g *Game) Update() error {
	g.input.Update(g.world, g.camera)

	if !g.input.paused {
		// Run multiple physics sub-steps for higher sim speeds
		steps := int(g.input.simSpeed * 2)
		if steps < 1 {
			steps = 1
		}
		for i := 0; i < steps; i++ {
			g.world.StepPhysics()
		}
	}
	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen, g.world, g.camera, g.input)
}

// Layout returns the logical screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	world := newWorld()

	world.AddObject(200, 250, 30)
	world.AddObject(140, 100, 8)
	world.AddObject(220, 110, 8)

	game := &Game{
		world:    world,
		camera:   newCamera(),
		input:    newInputState(),
		renderer: newRenderer(),
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Gravity Sandbox")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
