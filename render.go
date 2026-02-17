package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Renderer handles all drawing operations.
type Renderer struct {
	pixels   []byte        // RGBA pixel buffer for screen
	hudImage *ebiten.Image // reusable off-screen image for scaled HUD text
}

func newRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Draw(screen *ebiten.Image, world *World, cam *Camera, input *InputState) {
	if r.pixels == nil {
		r.pixels = make([]byte, screenWidth*screenHeight*4)
	}

	// Clear to black
	for i := range r.pixels {
		r.pixels[i] = 0
	}

	// Draw gravity field heatmap (before objects so they render on top)
	if input.showField {
		r.drawGravityField(world, cam)
	}

	// Draw objects
	for _, o := range world.objects {
		r.drawObject(o, cam, o == input.selectedObj)
	}

	// Draw ghost preview at cursor
	if !input.aiming && !input.dragging {
		r.drawGhostCircle(input, cam)
	}

	// Draw slingshot aiming visuals
	if input.aiming {
		r.drawSlingshot(input, cam, world)
	}

	screen.WritePixels(r.pixels)

	// HUD on top (uses ebiten text rendering, not pixel buffer)
	r.drawHUD(screen, world, input)
}

func (r *Renderer) drawObject(o *Object, cam *Camera, selected bool) {
	sx, sy := cam.WorldToScreen(o.x, o.y)
	sr := cam.WorldRadius(o.radius)

	// Draw selection ring
	if selected {
		r.drawCircleOutline(sx, sy, sr+3, [3]byte{255, 255, 0})
	}

	// Draw pinned indicator (outer ring)
	if o.pinned {
		r.drawCircleOutline(sx, sy, sr+2, [3]byte{255, 100, 100})
	}

	// Draw filled circle
	r.drawFilledCircle(sx, sy, sr, o.color)
}

func (r *Renderer) drawFilledCircle(cx, cy float64, radius int, color [3]byte) {
	minX := clampInt(int(cx)-radius, 0, screenWidth)
	maxX := clampInt(int(cx)+radius, 0, screenWidth)
	minY := clampInt(int(cy)-radius, 0, screenHeight)
	maxY := clampInt(int(cy)+radius, 0, screenHeight)
	r2 := float64(radius * radius)

	for i := minX; i < maxX; i++ {
		for j := minY; j < maxY; j++ {
			dx := float64(i) - cx
			dy := float64(j) - cy
			if dx*dx+dy*dy < r2 {
				idx := (j*screenWidth + i) * 4
				r.pixels[idx] = color[0]
				r.pixels[idx+1] = color[1]
				r.pixels[idx+2] = color[2]
				r.pixels[idx+3] = 0xFF
			}
		}
	}
}

func (r *Renderer) drawCircleOutline(cx, cy float64, radius int, color [3]byte) {
	minX := clampInt(int(cx)-radius-1, 0, screenWidth)
	maxX := clampInt(int(cx)+radius+1, 0, screenWidth)
	minY := clampInt(int(cy)-radius-1, 0, screenHeight)
	maxY := clampInt(int(cy)+radius+1, 0, screenHeight)
	r2outer := float64(radius * radius)
	inner := radius - 1
	if inner < 0 {
		inner = 0
	}
	r2inner := float64(inner * inner)

	for i := minX; i < maxX; i++ {
		for j := minY; j < maxY; j++ {
			dx := float64(i) - cx
			dy := float64(j) - cy
			d := dx*dx + dy*dy
			if d >= r2inner && d < r2outer {
				idx := (j*screenWidth + i) * 4
				r.pixels[idx] = color[0]
				r.pixels[idx+1] = color[1]
				r.pixels[idx+2] = color[2]
				r.pixels[idx+3] = 0xFF
			}
		}
	}
}

func (r *Renderer) drawGhostCircle(input *InputState, cam *Camera) {
	wx, wy := input.cursorWorld(cam)
	sx, sy := cam.WorldToScreen(wx, wy)
	sr := cam.WorldRadius(input.nextRadius)

	// Draw faint outline
	r.drawCircleOutline(sx, sy, sr, [3]byte{80, 80, 80})
}

func (r *Renderer) drawSlingshot(input *InputState, cam *Camera, world *World) {
	cx, cy := input.cursorWorld(cam)
	startSX, startSY := cam.WorldToScreen(input.aimStartX, input.aimStartY)
	endSX, endSY := cam.WorldToScreen(cx, cy)

	// Draw rubber band line from start to cursor
	r.drawLine(startSX, startSY, endSX, endSY, [3]byte{255, 100, 100})

	// Draw ghost at launch point
	sr := cam.WorldRadius(input.nextRadius)
	r.drawCircleOutline(startSX, startSY, sr, [3]byte{150, 150, 150})

	// Draw trajectory preview
	dx := cx - input.aimStartX
	dy := cy - input.aimStartY
	launchScale := 0.05
	vx := -dx * launchScale
	vy := -dy * launchScale

	r.drawTrajectory(input.aimStartX, input.aimStartY, vx, vy, input.nextRadius, world, cam)
}

func (r *Renderer) drawTrajectory(startX, startY, vx, vy float64, radius int, world *World, cam *Camera) {
	px, py := startX, startY
	svx, svy := vx, vy
	mass := float64(radius * radius)
	softSq := softeningParameter * softeningParameter

	for step := 0; step < 200; step++ {
		var fx, fy float64
		for _, o := range world.objects {
			dx := o.x - px
			dy := o.y - py
			distSq := dx*dx + dy*dy + softSq
			sizeAdj := o.mass / mass
			fx += gravitationalConstant * sizeAdj * dx / distSq
			fy += gravitationalConstant * sizeAdj * dy / distSq
		}

		svx += fx
		svy += fy
		px += svx
		py += svy

		if step%3 == 0 {
			sx, sy := cam.WorldToScreen(px, py)
			si := int(sx)
			sj := int(sy)
			if si >= 0 && si < screenWidth && sj >= 0 && sj < screenHeight {
				idx := (sj*screenWidth + si) * 4
				brightness := byte(200 - step)
				if step > 200 {
					brightness = 50
				}
				r.pixels[idx] = brightness
				r.pixels[idx+1] = brightness
				r.pixels[idx+2] = brightness
				r.pixels[idx+3] = 0xFF
			}
		}
	}
}

const fieldGridSize = 8 // render every 8th pixel

func (r *Renderer) drawGravityField(world *World, cam *Camera) {
	if len(world.objects) == 0 {
		return
	}
	softSq := softeningParameter * softeningParameter

	for sy := 0; sy < screenHeight; sy += fieldGridSize {
		for sx := 0; sx < screenWidth; sx += fieldGridSize {
			wx, wy := cam.ScreenToWorld(float64(sx+fieldGridSize/2), float64(sy+fieldGridSize/2))

			var field float64
			for _, o := range world.objects {
				dx := o.x - wx
				dy := o.y - wy
				distSq := dx*dx + dy*dy + softSq
				field += o.mass / distSq
			}
			field *= gravitationalConstant

			// Log scale mapping
			intensity := math.Log1p(field * 5000)
			if intensity > 4.0 {
				intensity = 4.0
			}

			cr, cg, cb := fieldColor(intensity / 4.0)
			if cr == 0 && cg == 0 && cb == 0 {
				continue
			}

			// Fill the grid cell
			maxX := sx + fieldGridSize
			if maxX > screenWidth {
				maxX = screenWidth
			}
			maxY := sy + fieldGridSize
			if maxY > screenHeight {
				maxY = screenHeight
			}
			for i := sx; i < maxX; i++ {
				for j := sy; j < maxY; j++ {
					idx := (j*screenWidth + i) * 4
					r.pixels[idx] = cr
					r.pixels[idx+1] = cg
					r.pixels[idx+2] = cb
					r.pixels[idx+3] = 0xFF
				}
			}
		}
	}
}

// fieldColor maps a 0..1 intensity to a blue → cyan → green → yellow → red gradient.
func fieldColor(t float64) (byte, byte, byte) {
	if t < 0.01 {
		return 0, 0, 0
	}
	if t < 0.25 {
		s := t / 0.25
		return 0, byte(s * 80), byte(40 + s*80)
	}
	if t < 0.5 {
		s := (t - 0.25) / 0.25
		return 0, byte(80 + s*100), byte(120 - s*40)
	}
	if t < 0.75 {
		s := (t - 0.5) / 0.25
		return byte(s * 200), byte(180 + s*75), byte(80 - s*80)
	}
	s := (t - 0.75) / 0.25
	return byte(200 + s*55), byte(255 - s*155), 0
}

func (r *Renderer) drawLine(x0, y0, x1, y1 float64, color [3]byte) {
	dx := x1 - x0
	dy := y1 - y0
	length := math.Sqrt(dx*dx + dy*dy)
	if length < 1 {
		return
	}

	steps := int(length)
	for s := 0; s <= steps; s++ {
		t := float64(s) / length
		px := x0 + dx*t
		py := y0 + dy*t
		i := int(px)
		j := int(py)
		if i >= 0 && i < screenWidth && j >= 0 && j < screenHeight {
			idx := (j*screenWidth + i) * 4
			r.pixels[idx] = color[0]
			r.pixels[idx+1] = color[1]
			r.pixels[idx+2] = color[2]
			r.pixels[idx+3] = 0xFF
		}
	}
}

const hudScale = 2.0

func (r *Renderer) drawHUD(screen *ebiten.Image, world *World, input *InputState) {
	// Draw HUD text to a temporary image, then scale it up
	hudW := screenWidth / hudScale
	hudH := screenHeight / hudScale
	if r.hudImage == nil {
		r.hudImage = ebiten.NewImage(int(hudW), int(hudH))
	}
	r.hudImage.Clear()

	// Top-left: status
	speedStr := fmt.Sprintf("%.1fx", input.simSpeed)
	pauseStr := ""
	if input.paused {
		pauseStr = "  [PAUSED]"
	}
	fps := ebiten.ActualFPS()
	status := fmt.Sprintf("Particles: %d  Speed: %s%s  Brush: %d  FPS: %.0f",
		len(world.objects), speedStr, pauseStr, input.nextRadius, fps)
	ebitenutil.DebugPrintAt(r.hudImage, status, 8, 8)

	// Physics modes
	frictionStr := "OFF"
	if world.frictionEnabled {
		frictionStr = "ON"
	}
	mergeStr := "OFF"
	if world.mergeOnCollision {
		mergeStr = "ON"
	}
	fieldStr := "OFF"
	if input.showField {
		fieldStr = "ON"
	}
	modes := fmt.Sprintf("Friction: %s  Merge: %s  Restitution: %.1f  Field: %s",
		frictionStr, mergeStr, world.restitution, fieldStr)
	ebitenutil.DebugPrintAt(r.hudImage, modes, 8, 24)

	// Selected object info
	if input.selectedObj != nil {
		o := input.selectedObj
		vel := math.Sqrt(o.velocityX*o.velocityX + o.velocityY*o.velocityY)
		pinnedStr := ""
		if o.pinned {
			pinnedStr = " [PINNED]"
		}
		info := fmt.Sprintf("Selected: mass=%.0f vel=%.3f%s", o.mass, vel, pinnedStr)
		ebitenutil.DebugPrintAt(r.hudImage, info, 8, 40)
	}

	// Controls help (bottom)
	help := "LMB: aim  RMB: select  [/]: size  P: pause  +/-: speed  Scroll: zoom  Del: remove  Space: pin  F: friction  M: merge  G: field"
	ebitenutil.DebugPrintAt(r.hudImage, help, 8, int(hudH)-20)

	// Draw HUD scaled up onto the main screen
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(hudScale, hudScale)
	screen.DrawImage(r.hudImage, op)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
