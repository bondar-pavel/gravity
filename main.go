package main

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const screenWidth = 800
const screenHeight = 600

const gravity = 0.0001
const friction = 0.01
const screenBounceEfficiency = 0.5

type Object struct {
	x, y                 float64
	radius               int
	mass                 float64
	velocityX, velocityY float64
	bouncedFrames        int
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

		sizeAdjustment := float64(obj.radius*obj.radius) / float64(o.radius*o.radius)

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

func (o *Object) BounceOnScreenCollision() {
	if o.x-float64(o.radius) < 0 || o.x+float64(o.radius) > screenWidth {
		o.velocityX = -o.velocityX * screenBounceEfficiency
	}
	if o.y-float64(o.radius) < 0 || o.y+float64(o.radius) > screenHeight {
		o.velocityY = -o.velocityY * screenBounceEfficiency
	}
}

func (o *Object) BounceOnObjectCollision(objects []*Object) {
	// skip processing if object is in bounced state
	if o.bouncedFrames > 0 {
		o.bouncedFrames--
		return
	}

	for _, obj := range objects {
		if obj == o {
			continue
		}
		if obj.bouncedFrames > 0 {
			continue
		}

		dx := obj.x - o.x
		dy := obj.y - o.y

		distanceSquared := dx*dx + dy*dy
		distance := math.Sqrt(distanceSquared)

		if distance < float64(o.radius+obj.radius) {
			normalX := dx / distance
			normalY := dy / distance

			myProjection := o.velocityX*normalX + o.velocityY*normalY
			objProjection := obj.velocityX*normalX + obj.velocityY*normalY

			impulse := 2 * (myProjection - objProjection) / (o.mass + obj.mass)

			o.velocityX -= impulse * obj.mass * normalX
			o.velocityY -= impulse * obj.mass * normalY

			obj.velocityX += impulse * o.mass * normalX
			obj.velocityY += impulse * o.mass * normalY

			// set bounced frames to prevent multiple collision detection within one frame
			o.bouncedFrames = 10
			obj.bouncedFrames = 10
		}
	}
}

type Map struct {
	objects                 []*Object
	pix                     []byte
	time                    int
	bounceOnScreenCollision bool
	shadeHalfCoveredPixels  bool
}

func newMap() *Map {
	return &Map{
		pix:                     make([]byte, screenWidth*screenHeight),
		objects:                 make([]*Object, 0),
		bounceOnScreenCollision: true,
		shadeHalfCoveredPixels:  false,
	}
}

func (m *Map) SetObject(x, y int, radius int, value byte) {
	m.objects = append(m.objects, &Object{
		x:         float64(x),
		y:         float64(y),
		radius:    radius,
		mass:      float64(radius * radius),
		velocityX: 0, // rand.Float64()*1 - 0.5,
		velocityY: 0,
	})
}

func (m *Map) FindObject(x, y int, radius int) *Object {
	minX := float64(safeSub(float64(x), radius, screenWidth))
	maxX := float64(safeAdd(float64(x), radius, screenWidth))
	minY := float64(safeSub(float64(y), radius, screenHeight))
	maxY := float64(safeAdd(float64(y), radius, screenHeight))

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
		o.BounceOnObjectCollision(m.objects)

		o.UpdateVelocity(o.CalculateGraviationalForce(m.objects))
		o.UpdatePosition()

		if m.bounceOnScreenCollision {
			o.BounceOnScreenCollision()
		}

		if m.shadeHalfCoveredPixels {
			m.ShadeHalfCoveredPixels(o, m.pix)
		}

		// draw filled in circle
		for i := safeSub(o.x+1, o.radius, screenWidth); i < safeAdd(o.x+1, o.radius, screenWidth); i++ {
			for j := safeSub(o.y+1, o.radius, screenHeight); j < safeAdd(o.y+1, o.radius, screenHeight); j++ {
				dx := float64(i) - o.x
				dy := float64(j) - o.y
				if dx*dx+dy*dy < float64(o.radius*o.radius) {
					m.pix[j*screenWidth+i] = 255
				}
			}
		}

	}
}

// ShadeHalfCoveredPixels shades half covered pixels
func (m *Map) ShadeHalfCoveredPixels(o *Object, pix []byte) {
	xShade := o.x - float64(int(o.x))
	yShade := o.y - float64(int(o.y))

	xStart := safeSub(o.x, o.radius, screenWidth)
	yStart := safeSub(o.y, o.radius, screenHeight)

	xFinish := safeAdd(o.x, o.radius, screenWidth)
	yFinish := safeAdd(o.y, o.radius, screenHeight)

	for i := safeSub(o.x+1, o.radius, screenWidth); i < safeAdd(o.x, o.radius, screenWidth); i++ {
		if m.pix[yStart*screenWidth+i] < 250 {
			m.pix[yStart*screenWidth+i] = 255 - byte(255*yShade)
		}
		if m.pix[yFinish*screenWidth+i] < 250 {
			m.pix[yFinish*screenWidth+i] = byte(255 * yShade)
		}
	}

	for j := safeSub(o.y, o.radius, screenHeight); j < safeAdd(o.y, o.radius, screenHeight); j++ {
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

	m.SetObject(200, 250, 30, 255)
	m.SetObject(140, 100, 8, 255)

	m.SetObject(220, 110, 8, 255)

	game := &Game{Map: m}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Gravity game")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
