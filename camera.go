package main

type Camera struct {
	x, y float64 // world position of view center
	zoom float64 // 1.0 = default
}

func newCamera() *Camera {
	return &Camera{
		x:    screenWidth / 2,
		y:    screenHeight / 2,
		zoom: 1.0,
	}
}

// WorldToScreen converts world coordinates to screen pixel coordinates.
func (c *Camera) WorldToScreen(wx, wy float64) (float64, float64) {
	sx := (wx-c.x)*c.zoom + screenWidth/2
	sy := (wy-c.y)*c.zoom + screenHeight/2
	return sx, sy
}

// ScreenToWorld converts screen pixel coordinates to world coordinates.
func (c *Camera) ScreenToWorld(sx, sy float64) (float64, float64) {
	wx := (sx-screenWidth/2)/c.zoom + c.x
	wy := (sy-screenHeight/2)/c.zoom + c.y
	return wx, wy
}

// WorldRadius converts a world-space radius to screen pixels.
func (c *Camera) WorldRadius(r int) int {
	sr := float64(r) * c.zoom
	if sr < 1 {
		return 1
	}
	return int(sr)
}

func (c *Camera) Reset() {
	c.x = screenWidth / 2
	c.y = screenHeight / 2
	c.zoom = 1.0
}

func (c *Camera) ZoomAt(factor float64) {
	c.zoom *= factor
	if c.zoom < 0.25 {
		c.zoom = 0.25
	}
	if c.zoom > 4.0 {
		c.zoom = 4.0
	}
}
