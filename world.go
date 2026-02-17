package main

type World struct {
	objects                   []*Object
	bounceOnScreenCollision   bool
	bounceOnParticleCollision bool
}

func newWorld() *World {
	return &World{
		objects:                   make([]*Object, 0),
		bounceOnScreenCollision:   true,
		bounceOnParticleCollision: true,
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

func (w *World) StepPhysics() {
	for i, o := range w.objects {
		if o.pinned {
			continue
		}

		o.UpdateVelocity(o.CalculateGravitationalForce(w.objects))
		o.UpdatePosition()

		if w.bounceOnParticleCollision {
			o.BounceOnObjectCollision(w.objects[i+1:])
		}

		if w.bounceOnScreenCollision {
			o.BounceOnScreenCollision()
		}
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
