package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const screenWidth = 320
const screenHeight = 240

type Object struct {
	x, y                 float64
	size                 int
	velocityX, velocityY float64
}

func (o *Object) UpdatePosition() {
	o.x += o.velocityX
	o.y += o.velocityY
}

type Map struct {
	objects []*Object
	pix     []byte
	time    int
}

func newMap() *Map {
	return &Map{
		pix:     make([]byte, screenWidth*screenHeight),
		objects: make([]*Object, 0),
	}
}

func (m *Map) SetObject(x, y int, size int, value byte) {
	m.objects = append(m.objects, &Object{
		x:         float64(x),
		y:         float64(y),
		size:      size,
		velocityX: 0,
		velocityY: 0.5,
	})
}

func (m *Map) ObjectsToPixels() {
	m.pix = make([]byte, screenWidth*screenHeight)

	for _, o := range m.objects {
		o.UpdatePosition()

		for i := safeSub(o.x, o.size, screenWidth); i < safeAdd(o.x, o.size, screenWidth); i++ {
			for j := safeSub(o.y, o.size, screenHeight); j < safeAdd(o.y, o.size, screenHeight); j++ {
				m.pix[j*screenWidth+i] = 250
			}
		}
	}
}

func (m *Map) Update() {
	m.time++
	if m.time >= screenHeight {
		m.time = 0
	}

	m.ObjectsToPixels()
}

func (m *Map) Draw(pixels []byte) {
	for i, v := range m.pix {
		pixels[4*i] = v   // R
		pixels[4*i+1] = v // G
		pixels[4*i+2] = v // B
		pixels[4*i+3] = v // ?
	}

}

func safeSub(a float64, b, limit int) int {
	m := int(a)
	if m < b {
		return 0
	}
	result := m - b
	if result > limit {
		return limit
	}
	return result
}

func safeAdd(a float64, b, limit int) int {
	m := int(a)
	if m+b > limit {
		return limit
	}
	return m + b
}

// Game implements ebiten.Game interface.
type Game struct {
	Map    *Map
	pixels []byte
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
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
	m := newMap()

	m.SetObject(50, 50, 20, 250)
	m.SetObject(140, 100, 10, 250)

	game := &Game{Map: m}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Gravity game")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
