package main

import "math"

type World struct {
	objects                   []*Object
	ejecta                    []Ejecta
	bounceOnScreenCollision   bool
	bounceOnParticleCollision bool
	mergeOnCollision          bool
	frictionEnabled           bool
	frictionCoeff             float64
	restitution               float64
}

type Ejecta struct {
	x, y   float64
	vx, vy float64
	life   float64 // 1.0 → 0.0
	size   float64 // initial pixel radius
}

func newWorld() *World {
	return &World{
		objects:                   make([]*Object, 0),
		bounceOnScreenCollision:   false,
		bounceOnParticleCollision: true,
		frictionCoeff:             0.001,
		restitution:               0.8,
	}
}

func (w *World) AddObject(x, y float64, radius int) *Object {
	obj := &Object{
		x:      x,
		y:      y,
		radius: radius,
		mass:   float64(radius * radius),
		color:  defaultParticleColor(len(w.objects)),
	}
	w.objects = append(w.objects, obj)
	return obj
}

func (w *World) RemoveObject(obj *Object) {
	for i, o := range w.objects {
		if o == obj {
			w.objects = append(w.objects[:i], w.objects[i+1:]...)
			return
		}
	}
}

func (w *World) FindObject(wx, wy float64, radius int) *Object {
	r := float64(radius)
	for _, o := range w.objects {
		dx := o.x - wx
		dy := o.y - wy
		dist := dx*dx + dy*dy
		threshold := r + float64(o.radius)
		if dist < threshold*threshold {
			return o
		}
	}
	return nil
}

// StepPhysics runs one tick using Velocity Verlet integration.
func (w *World) StepPhysics() {
	// Phase 1: Update positions using current velocity and acceleration
	for _, o := range w.objects {
		if o.pinned {
			continue
		}
		o.UpdatePositionVerlet()
	}

	// Phase 2: Calculate new accelerations from updated positions
	type accel struct{ ax, ay float64 }
	newAccels := make([]accel, len(w.objects))
	for i, o := range w.objects {
		if o.pinned {
			continue
		}
		newAccels[i].ax, newAccels[i].ay = o.CalculateAcceleration(w.objects)
	}

	// Phase 3: Update velocities using average of old and new acceleration
	for i, o := range w.objects {
		if o.pinned {
			continue
		}
		o.UpdateVelocityVerlet(newAccels[i].ax, newAccels[i].ay)

		if w.frictionEnabled {
			o.velocityX *= (1 - w.frictionCoeff)
			o.velocityY *= (1 - w.frictionCoeff)
		}
	}

	// Collisions
	if w.bounceOnParticleCollision || w.mergeOnCollision {
		w.handleCollisions()
	}

	// Screen boundary
	if w.bounceOnScreenCollision {
		for _, o := range w.objects {
			if o.pinned {
				continue
			}
			o.BounceOnScreenCollision()
		}
	}

	// Remove objects that drifted far outside the observable area
	w.cullDistantObjects()

	// Rotation and merge animation
	for _, o := range w.objects {
		o.UpdateRotation()
	}

	// Update ejecta
	w.updateEjecta()
}

func (w *World) handleCollisions() {
	var toRemove []*Object

	for i := 0; i < len(w.objects); i++ {
		o := w.objects[i]
		for j := i + 1; j < len(w.objects); j++ {
			obj := w.objects[j]
			shouldMerge := o.CollideWith(obj, w.restitution, w.mergeOnCollision)
			if shouldMerge {
				speed := math.Sqrt(
					(o.velocityX-obj.velocityX)*(o.velocityX-obj.velocityX)+
						(o.velocityY-obj.velocityY)*(o.velocityY-obj.velocityY)) + 1.0
				mx := (o.x + obj.x) / 2
				my := (o.y + obj.y) / 2
				o.MergeFrom(obj)
				w.SpawnEjecta(mx, my, speed, 8+int(speed))
				toRemove = append(toRemove, obj)
			}
		}
	}

	for _, obj := range toRemove {
		w.RemoveObject(obj)
	}
}

func (w *World) SpawnEjecta(x, y, speed float64, count int) {
	if count > 16 {
		count = 16
	}
	for i := 0; i < count; i++ {
		angle := 2 * math.Pi * float64(i) / float64(count)
		// Vary speed slightly per particle
		s := speed * (0.5 + 0.8*float64((i*7+3)%10)/10.0)
		w.ejecta = append(w.ejecta, Ejecta{
			x:    x,
			y:    y,
			vx:   math.Cos(angle) * s,
			vy:   math.Sin(angle) * s,
			life: 1.0,
			size: 2.0 + float64(i%3),
		})
	}
}

func (w *World) updateEjecta() {
	n := 0
	for i := range w.ejecta {
		e := &w.ejecta[i]
		e.x += e.vx
		e.y += e.vy
		e.vx *= 0.97 // drag
		e.vy *= 0.97
		e.life -= 0.015
		if e.life > 0 {
			w.ejecta[n] = *e
			n++
		}
	}
	w.ejecta = w.ejecta[:n]
}

const cullDistance = 5000 // remove objects this far from screen center

func (w *World) cullDistantObjects() {
	cx := float64(screenWidth) / 2
	cy := float64(screenHeight) / 2
	var toRemove []*Object
	for _, o := range w.objects {
		if o.pinned {
			continue
		}
		dx := o.x - cx
		dy := o.y - cy
		if dx*dx+dy*dy > cullDistance*cullDistance {
			toRemove = append(toRemove, o)
		}
	}
	for _, o := range toRemove {
		w.RemoveObject(o)
	}
}

func defaultParticleColor(index int) [3]byte {
	colors := [][3]byte{
		{255, 255, 255}, // white
		{100, 180, 255}, // light blue
		{255, 130, 100}, // salmon
		{130, 255, 130}, // light green
		{255, 220, 100}, // yellow
		{200, 140, 255}, // purple
		{255, 160, 200}, // pink
		{100, 255, 220}, // cyan
	}
	return colors[index%len(colors)]
}
