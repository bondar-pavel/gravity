package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const screenWidth = 320
const screenHeight = 240

type Map struct {
	pix  []byte
	time int
}

func newMap() *Map {
	return &Map{
		pix: make([]byte, screenWidth*screenHeight),
	}
}

func (m *Map) SetObject(x, y int, size int, value byte) {
	for i := safeSub(x, size, screenWidth); i < safeAdd(x, size, screenWidth); i++ {
		for j := safeSub(y, size, screenHeight); j < safeAdd(y, size, screenHeight); j++ {
			m.pix[j*screenWidth+i] = value
		}
	}
}

func (m *Map) Update() {
	m.time++
	if m.time >= screenHeight {
		m.time = 0

	}

	m.SetObject(m.time, m.time, 20, 250)
}

func (m *Map) Draw(pixels []byte) {
	for i, v := range m.pix {
		pixels[4*i] = v   // R
		pixels[4*i+1] = v // G
		pixels[4*i+2] = v // B
		pixels[4*i+3] = v // ?
	}

}

func safeSub(a, b, limit int) int {
	if a < b {
		return 0
	}
	result := a - b
	if result > limit {
		return limit
	}
	return result
}

func safeAdd(a, b, limit int) int {
	if a+b > limit {
		return limit
	}
	return a + b
}

// Game implements ebiten.Game interface.
type Game struct {
	Map    *Map
	pixels []byte
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	/*
		for i, v := range g.Map.pixels {
			if v < 250 {
				g.Map.pixels[i] = v + 1
			}
		}
	*/

	//screen.WritePixels(g.pixels)
	g.Map.Update()

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	if g.pixels == nil {
		g.pixels = make([]byte, screenWidth*screenHeight*4)
	}

	g.Map.Draw(g.pixels)

	screen.WritePixels(g.pixels)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	game := &Game{Map: newMap()}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Gravity game")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
