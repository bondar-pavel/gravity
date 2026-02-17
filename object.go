package main

import "math"

type Object struct {
	x, y                 float64
	radius               int
	mass                 float64
	velocityX, velocityY float64
	ax, ay               float64 // acceleration (stored for Verlet integration)
	bouncedFrames        int
	pinned               bool
	color                [3]byte
}

// CalculateAcceleration returns gravitational acceleration from all other objects (with softening).
func (o *Object) CalculateAcceleration(objects []*Object) (float64, float64) {
	var ax, ay float64
	softSq := softeningParameter * softeningParameter

	for _, obj := range objects {
		if obj == o {
			continue
		}
		dx := obj.x - o.x
		dy := obj.y - o.y
		distSq := dx*dx + dy*dy + softSq

		sizeAdj := obj.mass / o.mass
		ax += gravitationalConstant * sizeAdj * dx / distSq
		ay += gravitationalConstant * sizeAdj * dy / distSq
	}

	return ax, ay
}

// UpdatePositionVerlet performs the position half of Velocity Verlet: x += v*dt + 0.5*a*dt²
func (o *Object) UpdatePositionVerlet() {
	o.x += o.velocityX + 0.5*o.ax
	o.y += o.velocityY + 0.5*o.ay
}

// UpdateVelocityVerlet performs the velocity half: v += 0.5*(a_old + a_new)*dt
func (o *Object) UpdateVelocityVerlet(newAX, newAY float64) {
	o.velocityX += 0.5 * (o.ax + newAX)
	o.velocityY += 0.5 * (o.ay + newAY)
	o.ax = newAX
	o.ay = newAY
}

func (o *Object) BounceOnScreenCollision() {
	if o.x-float64(o.radius) < 0 && o.velocityX < 0 || o.x+float64(o.radius) > screenWidth && o.velocityX > 0 {
		o.velocityX = -o.velocityX * screenBounceEfficiency
	}
	if o.y-float64(o.radius) < 0 && o.velocityY < 0 || o.y+float64(o.radius) > screenHeight && o.velocityY > 0 {
		o.velocityY = -o.velocityY * screenBounceEfficiency
	}
}

// CollideWith checks collision with another object, separates overlap, and applies impulse.
// Returns true if a merge should happen (caller handles removal).
func (o *Object) CollideWith(obj *Object, restitution float64, merge bool) bool {
	dx := obj.x - o.x
	dy := obj.y - o.y
	distSq := dx*dx + dy*dy
	distance := math.Sqrt(distSq)
	minDist := float64(o.radius + obj.radius)

	if distance >= minDist {
		return false
	}
	if distance < 0.001 {
		distance = 0.001
	}

	normalX := dx / distance
	normalY := dy / distance

	// Separate overlapping objects
	overlap := minDist - distance
	totalMass := o.mass + obj.mass

	if o.pinned {
		obj.x += normalX * overlap
		obj.y += normalY * overlap
	} else if obj.pinned {
		o.x -= normalX * overlap
		o.y -= normalY * overlap
	} else {
		o.x -= normalX * overlap * (obj.mass / totalMass)
		o.y -= normalY * overlap * (obj.mass / totalMass)
		obj.x += normalX * overlap * (o.mass / totalMass)
		obj.y += normalY * overlap * (o.mass / totalMass)
	}

	if merge && !o.pinned && !obj.pinned {
		return true
	}

	// Impulse-based collision with restitution
	myProj := o.velocityX*normalX + o.velocityY*normalY
	objProj := obj.velocityX*normalX + obj.velocityY*normalY

	if o.pinned {
		// Only obj bounces
		obj.velocityX += -(1 + restitution) * (objProj - myProj) * normalX
		obj.velocityY += -(1 + restitution) * (objProj - myProj) * normalY
	} else if obj.pinned {
		// Only o bounces
		o.velocityX += -(1 + restitution) * (myProj - objProj) * normalX
		o.velocityY += -(1 + restitution) * (myProj - objProj) * normalY
	} else {
		impulse := (1 + restitution) * (myProj - objProj) / totalMass
		o.velocityX -= impulse * obj.mass * normalX
		o.velocityY -= impulse * obj.mass * normalY
		obj.velocityX += impulse * o.mass * normalX
		obj.velocityY += impulse * o.mass * normalY
	}

	return false
}

// MergeFrom absorbs another object: conserves momentum, area-preserving radius.
func (o *Object) MergeFrom(obj *Object) {
	newMass := o.mass + obj.mass
	o.velocityX = (o.mass*o.velocityX + obj.mass*obj.velocityX) / newMass
	o.velocityY = (o.mass*o.velocityY + obj.mass*obj.velocityY) / newMass
	o.radius = int(math.Sqrt(float64(o.radius*o.radius + obj.radius*obj.radius)))
	if o.radius < 1 {
		o.radius = 1
	}
	o.mass = float64(o.radius * o.radius)
}
