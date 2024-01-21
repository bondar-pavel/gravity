package main

import (
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const screenWidth = 320
const screenHeight = 240

const gravity = 0.0001
const friction = 0.01

type Object struct {
	x, y                 float64
	size                 int
	velocityX, velocityY float64
}

// CalculateGraviationalForce calculates resulting force of gravity for passed in objects
func (o *Object) CalculateGraviationalForce(objects []*Object) (float64, float64) {
	var forceX, forceY float64

	for _, obj := range objects {
		if obj == o {
			continue
		}
		dx := obj.x - o.x
		dy := obj.y - o.y
		distance := dx*dx + dy*dy

		sizeAdjustment := float64(obj.size) / float64(o.size)

		forceX += sizeAdjustment * dx / distance
		forceY += sizeAdjustment * dy / distance
	}

	return forceX, forceY
}

func (o *Object) UpdateVelocity(forceX, forceY float64) {
	o.velocityX += forceX
	o.velocityY += forceY
}

func (o *Object) UpdateVelocityGravitational() {
	o.velocityY += gravity

	slowDown := friction * o.velocityY
	if slowDown < 0 {
		slowDown = -slowDown
	}
	o.velocityY -= slowDown
}

func (o *Object) UpdatePosition() {
	o.x += o.velocityX
	o.y += o.velocityY
}

func (o *Object) BounceOnCollision() {
	if o.x-float64(o.size) < 0 || o.x+float64(o.size) > screenWidth {
		o.velocityX = -o.velocityX
	}
	if o.y-float64(o.size) < 0 || o.y+float64(o.size) > screenHeight {
		o.velocityY = -o.velocityY
	}
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
		velocityX: rand.Float64()*1 - 0.5,
		velocityY: 0.5,
	})
}

func (m *Map) FindObject(x, y int, size int) *Object {
	minX := float64(safeSub(float64(x), size, screenWidth))
	maxX := float64(safeAdd(float64(x), size, screenWidth))
	minY := float64(safeSub(float64(y), size, screenHeight))
	maxY := float64(safeAdd(float64(y), size, screenHeight))

	for _, o := range m.objects {
		if o.x < minX || o.x > maxX || o.y < minY || o.y > maxY {
			continue
		}
		return o
	}
	return nil
}

func (m *Map) ObjectsToPixels() {
	m.pix = make([]byte, screenWidth*screenHeight)

	for _, o := range m.objects {
		o.UpdateVelocity(o.CalculateGraviationalForce(m.objects))
		o.UpdatePosition()
		o.BounceOnCollision()

		/*
			// shade half covered pixels
			xShade := o.x - float64(int(o.x))
			yShade := o.y - float64(int(o.y))

			xStart := safeSub(o.x, o.size, screenWidth)
			yStart := safeSub(o.y, o.size, screenHeight)

			xFinish := safeAdd(o.x, o.size, screenWidth)
			yFinish := safeAdd(o.y, o.size, screenHeight)

			for i := safeSub(o.x+1, o.size, screenWidth); i < safeAdd(o.x, o.size, screenWidth); i++ {
				if m.pix[yStart*screenWidth+i] < 250 {
					m.pix[yStart*screenWidth+i] = 255 - byte(255*yShade)
				}
				if m.pix[yFinish*screenWidth+i] < 250 {
					m.pix[yFinish*screenWidth+i] = byte(255 * yShade)
				}
			}

			for j := safeSub(o.y, o.size, screenHeight); j < safeAdd(o.y, o.size, screenHeight); j++ {
				if m.pix[j*screenWidth+xStart] < 250 {
					v := 255 - byte(255*xShade)
					if m.pix[j*screenWidth+xStart] > 0 {
						v = m.pix[j*screenWidth+xStart]/2 + v/2
					}
					m.pix[j*screenWidth+xStart] = v
				}
				if m.pix[j*screenWidth+xFinish] < 250 {
					v := byte(255 * xShade)
					if m.pix[j*screenWidth+xFinish] > 0 {
						v = m.pix[j*screenWidth+xFinish]/2 + v/2
					}
					m.pix[j*screenWidth+xFinish] = v
				}
			}
		*/

		// fill the rest
		for i := safeSub(o.x+1, o.size, screenWidth); i < safeAdd(o.x, o.size, screenWidth); i++ {
			for j := safeSub(o.y+1, o.size, screenHeight); j < safeAdd(o.y, o.size, screenHeight); j++ {
				m.pix[j*screenWidth+i] = 255
			}
		}

	}
}

func (m *Map) Update() {
	m.time++
	if m.time >= screenHeight {
		m.time = 0
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		obj := m.FindObject(x, y, 20)
		if obj != nil {
			obj.x = float64(x)
			obj.y = float64(y)
		} else {
			m.SetObject(x, y, 10, 255)
		}
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
	if m+b >= limit {
		return limit - 1
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
func (g *Game) Layout(outsideWidth, outsideHeight int) (sWidth, sHeight int) {
	return screenWidth, screenHeight
}

func main() {
	m := newMap()

	m.SetObject(50, 50, 20, 255)
	m.SetObject(140, 100, 10, 255)

	m.SetObject(220, 110, 8, 255)

	game := &Game{Map: m}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Gravity game")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
