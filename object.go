package main

import "math"

type Object struct {
	x, y                 float64
	radius               int
	mass                 float64
	velocityX, velocityY float64
	bouncedFrames        int
	pinned               bool
	color                [3]byte // R, G, B
}

// CalculateGravitationalForce calculates resulting force of gravity for passed in objects
func (o *Object) CalculateGravitationalForce(objects []*Object) (float64, float64) {
	var forceX, forceY float64

	for _, obj := range objects {
		if obj == o {
			continue
		}
		dx := obj.x - o.x
		dy := obj.y - o.y
		distanceSquared := dx*dx + dy*dy

		sizeAdjustment := obj.mass / o.mass

		forceX += gravitationalConstant * sizeAdjustment * dx / distanceSquared
		forceY += gravitationalConstant * sizeAdjustment * dy / distanceSquared
	}

	return forceX, forceY
}

func (o *Object) UpdateVelocity(forceX, forceY float64) {
	o.velocityX += forceX
	o.velocityY += forceY
}

func (o *Object) UpdatePosition() {
	o.x += o.velocityX
	o.y += o.velocityY
}

func (o *Object) BounceOnScreenCollision() {
	if o.x-float64(o.radius) < 0 && o.velocityX < 0 || o.x+float64(o.radius) > screenWidth && o.velocityX > 0 {
		o.velocityX = -o.velocityX * screenBounceEfficiency
	}
	if o.y-float64(o.radius) < 0 && o.velocityY < 0 || o.y+float64(o.radius) > screenHeight && o.velocityY > 0 {
		o.velocityY = -o.velocityY * screenBounceEfficiency
	}
}

func (o *Object) BounceOnObjectCollision(objects []*Object) {
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

			o.bouncedFrames = 1
			obj.bouncedFrames = 1
		}
	}
}
