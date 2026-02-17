package main

import "math"

type ChallengeState int

const (
	ChallengeAiming   ChallengeState = iota // waiting for player to launch
	ChallengeOrbiting                       // particle in flight
	ChallengeCrashed                        // hit a planet
	ChallengeEscaped                        // left orbit zone
)

type LevelObject struct {
	X, Y   float64
	Radius int
	Pinned bool
}

type Level struct {
	Name            string
	Objects         []LevelObject
	OrbitZoneRadius float64
}

type Challenge struct {
	active       bool
	state        ChallengeState
	currentLevel int
	levels       []Level

	// Orbit tracking
	orbiter     *Object
	orbitCenter [2]float64 // center point for orbit zone (centroid of planets)
	prevAngle   float64
	totalAngle  float64
	orbitCount  int
	bestScores  []int
	newBest     bool // flash "NEW BEST" on result screen

	// Zone
	orbitZoneRadius float64

	// Result display timer (frames)
	resultTimer int

	// Saved sandbox state
	savedObjects []*Object
	savedMerge   bool
}

func newChallenge() *Challenge {
	levels := []Level{
		{
			Name: "Single Planet",
			Objects: []LevelObject{
				{800, 600, 40, true},
			},
			OrbitZoneRadius: 500,
		},
		{
			Name: "Binary Star",
			Objects: []LevelObject{
				{500, 600, 30, true},
				{1100, 600, 30, true},
			},
			OrbitZoneRadius: 600,
		},
		{
			Name: "Triple Chaos",
			Objects: []LevelObject{
				{800, 300, 25, true},
				{500, 800, 25, true},
				{1100, 800, 25, true},
			},
			OrbitZoneRadius: 700,
		},
		{
			Name: "Giant and Moon",
			Objects: []LevelObject{
				{800, 600, 50, true},
				{1000, 600, 12, true},
			},
			OrbitZoneRadius: 500,
		},
	}

	return &Challenge{
		levels:     levels,
		bestScores: make([]int, len(levels)),
	}
}

func (c *Challenge) Enter(world *World) {
	// Save sandbox state
	c.savedObjects = make([]*Object, len(world.objects))
	copy(c.savedObjects, world.objects)
	c.savedMerge = world.mergeOnCollision

	c.active = true
	c.state = ChallengeAiming
	c.orbiter = nil
	world.mergeOnCollision = true
	c.loadLevel(world)
}

func (c *Challenge) Exit(world *World) {
	c.active = false
	c.orbiter = nil

	// Restore sandbox
	world.objects = c.savedObjects
	world.mergeOnCollision = c.savedMerge
	c.savedObjects = nil
}

func (c *Challenge) loadLevel(world *World) {
	level := c.levels[c.currentLevel]
	world.objects = world.objects[:0]

	var cx, cy float64
	for _, lo := range level.Objects {
		obj := world.AddObject(lo.X, lo.Y, lo.Radius)
		obj.pinned = lo.Pinned
		cx += lo.X
		cy += lo.Y
	}

	// Centroid for orbit zone center
	n := float64(len(level.Objects))
	c.orbitCenter = [2]float64{cx / n, cy / n}
	c.orbitZoneRadius = level.OrbitZoneRadius

	c.orbiter = nil
	c.state = ChallengeAiming
	c.totalAngle = 0
	c.orbitCount = 0
	c.resultTimer = 0
}

func (c *Challenge) ChangeLevel(delta int, world *World) {
	if c.state != ChallengeAiming {
		return
	}
	c.currentLevel += delta
	if c.currentLevel < 0 {
		c.currentLevel = len(c.levels) - 1
	}
	if c.currentLevel >= len(c.levels) {
		c.currentLevel = 0
	}
	c.loadLevel(world)
}

func (c *Challenge) LaunchOrbiter(world *World, x, y, vx, vy float64) {
	if c.state != ChallengeAiming {
		return
	}

	obj := world.AddObject(x, y, 5)
	obj.velocityX = vx
	obj.velocityY = vy
	obj.color = [3]byte{255, 255, 100} // bright yellow

	c.orbiter = obj
	c.state = ChallengeOrbiting
	c.totalAngle = 0
	c.orbitCount = 0
	c.newBest = false

	// Initialize angle tracking from orbit center
	dx := obj.x - c.orbitCenter[0]
	dy := obj.y - c.orbitCenter[1]
	c.prevAngle = math.Atan2(dy, dx)
}

// Update is called each physics tick while challenge is active.
func (c *Challenge) Update(world *World) {
	if !c.active {
		return
	}

	switch c.state {
	case ChallengeOrbiting:
		c.trackOrbit(world)
	case ChallengeCrashed, ChallengeEscaped:
		c.resultTimer++
	}
}

func (c *Challenge) trackOrbit(world *World) {
	if c.orbiter == nil {
		c.state = ChallengeAiming
		return
	}

	// Check crash: distance to any planet < sum of radii
	for _, o := range world.objects {
		if o == c.orbiter || !o.pinned {
			continue
		}
		dx := c.orbiter.x - o.x
		dy := c.orbiter.y - o.y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < float64(c.orbiter.radius+o.radius) {
			c.endRound(ChallengeCrashed, world)
			return
		}
	}

	// Check escape: distance from orbit center > zone radius
	dx := c.orbiter.x - c.orbitCenter[0]
	dy := c.orbiter.y - c.orbitCenter[1]
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > c.orbitZoneRadius {
		c.endRound(ChallengeEscaped, world)
		return
	}

	// Track angle
	currentAngle := math.Atan2(dy, dx)
	delta := currentAngle - c.prevAngle
	if delta > math.Pi {
		delta -= 2 * math.Pi
	} else if delta < -math.Pi {
		delta += 2 * math.Pi
	}
	c.totalAngle += delta
	c.orbitCount = int(math.Abs(c.totalAngle) / (2 * math.Pi))
	c.prevAngle = currentAngle
}

func (c *Challenge) endRound(state ChallengeState, world *World) {
	c.state = state
	c.resultTimer = 0

	if c.orbitCount > c.bestScores[c.currentLevel] {
		c.bestScores[c.currentLevel] = c.orbitCount
		c.newBest = true
	}

	// Remove orbiter
	if c.orbiter != nil {
		world.RemoveObject(c.orbiter)
		c.orbiter = nil
	}
}

func (c *Challenge) RetryLevel(world *World) {
	if c.state != ChallengeCrashed && c.state != ChallengeEscaped {
		return
	}
	c.loadLevel(world)
}

func (c *Challenge) CurrentLevel() Level {
	return c.levels[c.currentLevel]
}
