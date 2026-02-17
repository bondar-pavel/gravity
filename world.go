package main

type World struct {
	objects                   []*Object
	bounceOnScreenCollision   bool
	bounceOnParticleCollision bool
	mergeOnCollision          bool
	frictionEnabled           bool
	frictionCoeff             float64
	restitution               float64
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
}

func (w *World) handleCollisions() {
	var toRemove []*Object

	for i := 0; i < len(w.objects); i++ {
		o := w.objects[i]
		for j := i + 1; j < len(w.objects); j++ {
			obj := w.objects[j]
			shouldMerge := o.CollideWith(obj, w.restitution, w.mergeOnCollision)
			if shouldMerge {
				o.MergeFrom(obj)
				toRemove = append(toRemove, obj)
			}
		}
	}

	for _, obj := range toRemove {
		w.RemoveObject(obj)
	}
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
