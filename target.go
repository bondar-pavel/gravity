package main

import "math"

type TargetState int

const (
	TargetAiming   TargetState = iota // waiting for player to launch
	TargetFlying                      // projectile in flight
	TargetComplete                    // all targets hit
)

type TargetZone struct {
	X, Y   float64
	Radius float64
	Hit    bool
}

type TargetLevel struct {
	Name    string
	Objects []LevelObject
	Targets []TargetZone
	Par     int
}

type TargetPractice struct {
	active       bool
	state        TargetState
	currentLevel int
	levels       []TargetLevel

	// Current attempt
	projectile *Object
	targets    []TargetZone // mutable copy for current attempt
	launches   int
	bestStars  []int // best star rating per level

	// Result display
	resultTimer int

	// Saved sandbox state
	savedObjects []*Object
	savedMerge   bool
}

func newTargetPractice() *TargetPractice {
	levels := []TargetLevel{
		{
			Name: "Straight Shot",
			Objects: []LevelObject{
				{800, 600, 30, true},
			},
			Targets: []TargetZone{
				{800, 300, 40, false},
			},
			Par: 1,
		},
		{
			Name: "Gravity Sling",
			Objects: []LevelObject{
				{800, 600, 40, true},
			},
			Targets: []TargetZone{
				{400, 300, 35, false},
				{1200, 300, 35, false},
			},
			Par: 2,
		},
		{
			Name: "Thread the Needle",
			Objects: []LevelObject{
				{600, 600, 25, true},
				{1000, 600, 25, true},
			},
			Targets: []TargetZone{
				{800, 400, 30, false},
				{800, 800, 30, false},
				{500, 300, 30, false},
			},
			Par: 2,
		},
		{
			Name: "Around the World",
			Objects: []LevelObject{
				{800, 600, 35, true},
			},
			Targets: []TargetZone{
				{800, 300, 30, false},
				{1100, 600, 30, false},
				{800, 900, 30, false},
				{500, 600, 30, false},
			},
			Par: 2,
		},
	}

	return &TargetPractice{
		levels:    levels,
		bestStars: make([]int, len(levels)),
	}
}

func (tp *TargetPractice) Enter(world *World) {
	tp.savedObjects = make([]*Object, len(world.objects))
	copy(tp.savedObjects, world.objects)
	tp.savedMerge = world.mergeOnCollision

	tp.active = true
	tp.projectile = nil
	world.mergeOnCollision = false
	tp.loadLevel(world)
}

func (tp *TargetPractice) Exit(world *World) {
	tp.active = false
	if tp.projectile != nil {
		world.RemoveObject(tp.projectile)
		tp.projectile = nil
	}

	world.objects = tp.savedObjects
	world.mergeOnCollision = tp.savedMerge
	tp.savedObjects = nil
}

func (tp *TargetPractice) loadLevel(world *World) {
	level := tp.levels[tp.currentLevel]
	world.objects = world.objects[:0]

	for _, lo := range level.Objects {
		obj := world.AddObject(lo.X, lo.Y, lo.Radius)
		obj.pinned = lo.Pinned
	}

	// Copy targets fresh
	tp.targets = make([]TargetZone, len(level.Targets))
	copy(tp.targets, level.Targets)

	if tp.projectile != nil {
		world.RemoveObject(tp.projectile)
	}
	tp.projectile = nil
	tp.state = TargetAiming
	tp.launches = 0
	tp.resultTimer = 0
}

func (tp *TargetPractice) ChangeLevel(delta int, world *World) {
	if tp.state == TargetFlying {
		return
	}
	tp.currentLevel += delta
	if tp.currentLevel < 0 {
		tp.currentLevel = len(tp.levels) - 1
	}
	if tp.currentLevel >= len(tp.levels) {
		tp.currentLevel = 0
	}
	tp.loadLevel(world)
}

func (tp *TargetPractice) LaunchProjectile(world *World, x, y, vx, vy float64) {
	if tp.state != TargetAiming {
		return
	}

	// Remove previous projectile if any
	if tp.projectile != nil {
		world.RemoveObject(tp.projectile)
	}

	obj := world.AddObject(x, y, 5)
	obj.velocityX = vx
	obj.velocityY = vy
	obj.color = [3]byte{100, 255, 200} // bright cyan-green

	tp.projectile = obj
	tp.state = TargetFlying
	tp.launches++
}

func (tp *TargetPractice) Update(world *World) {
	if !tp.active {
		return
	}

	switch tp.state {
	case TargetFlying:
		tp.trackProjectile(world)
	case TargetComplete:
		tp.resultTimer++
	}
}

func (tp *TargetPractice) trackProjectile(world *World) {
	if tp.projectile == nil {
		tp.state = TargetAiming
		return
	}

	// Check target hits
	allHit := true
	for i := range tp.targets {
		t := &tp.targets[i]
		if t.Hit {
			continue
		}
		dx := tp.projectile.x - t.X
		dy := tp.projectile.y - t.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < t.Radius {
			t.Hit = true
		}
		if !t.Hit {
			allHit = false
		}
	}

	if allHit {
		tp.completeLevel(world)
		return
	}

	// Check crash into planet
	for _, o := range world.objects {
		if o == tp.projectile || !o.pinned {
			continue
		}
		dx := tp.projectile.x - o.x
		dy := tp.projectile.y - o.y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < float64(tp.projectile.radius+o.radius) {
			tp.removeProjectile(world)
			return
		}
	}

	// Check escape (far from screen center)
	cx := float64(screenWidth) / 2
	cy := float64(screenHeight) / 2
	dx := tp.projectile.x - cx
	dy := tp.projectile.y - cy
	if dx*dx+dy*dy > cullDistance*cullDistance {
		tp.removeProjectile(world)
	}
}

func (tp *TargetPractice) removeProjectile(world *World) {
	if tp.projectile != nil {
		world.RemoveObject(tp.projectile)
		tp.projectile = nil
	}
	tp.state = TargetAiming
}

func (tp *TargetPractice) completeLevel(world *World) {
	tp.state = TargetComplete
	tp.resultTimer = 0

	stars := tp.StarRating()
	if stars > tp.bestStars[tp.currentLevel] {
		tp.bestStars[tp.currentLevel] = stars
	}

	if tp.projectile != nil {
		world.RemoveObject(tp.projectile)
		tp.projectile = nil
	}
}

func (tp *TargetPractice) StarRating() int {
	par := tp.levels[tp.currentLevel].Par
	diff := tp.launches - par
	if diff <= 0 {
		return 3
	}
	if diff == 1 {
		return 2
	}
	if diff == 2 {
		return 1
	}
	return 0
}

func (tp *TargetPractice) RetryLevel(world *World) {
	tp.loadLevel(world)
}

func (tp *TargetPractice) CurrentLevel() TargetLevel {
	return tp.levels[tp.currentLevel]
}

func (tp *TargetPractice) HitsCount() int {
	n := 0
	for _, t := range tp.targets {
		if t.Hit {
			n++
		}
	}
	return n
}
