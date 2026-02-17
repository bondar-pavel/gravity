package main

import "github.com/hajimehoshi/ebiten/v2"

type InputState struct {
	// Slingshot aiming
	aiming    bool
	aimStartX float64
	aimStartY float64

	// Object dragging
	dragging bool
	dragObj  *Object

	// Selection
	selectedObj *Object

	// Particle size
	nextRadius int

	// Time
	paused   bool
	simSpeed float64

	// Camera panning
	panning   bool
	panStartX float64
	panStartY float64
	camStartX float64
	camStartY float64

	// Visualization
	showField bool

	// Debounce tracking
	prevKeys       map[ebiten.Key]bool
	prevRightClick bool
}

func newInputState() *InputState {
	return &InputState{
		nextRadius: 10,
		simSpeed:   1.0,
		prevKeys:   make(map[ebiten.Key]bool),
	}
}

// justPressed returns true on the frame a key transitions from up to down.
func (s *InputState) justPressed(key ebiten.Key) bool {
	pressed := ebiten.IsKeyPressed(key)
	was := s.prevKeys[key]
	s.prevKeys[key] = pressed
	return pressed && !was
}

func (s *InputState) Update(world *World, cam *Camera, challenge *Challenge) {
	// Challenge mode toggle
	if s.justPressed(ebiten.KeyO) {
		if challenge.active {
			challenge.Exit(world)
			s.aiming = false
			s.dragging = false
			return
		}
		challenge.Enter(world)
		s.aiming = false
		s.dragging = false
		s.selectedObj = nil
		return
	}

	if challenge.active {
		s.handleTimeControl()
		s.handleCamera(cam)
		s.handleChallengeInput(world, cam, challenge)
		return
	}

	// Normal sandbox mode
	s.handleTimeControl()
	s.handleSizeControl()
	s.handleCamera(cam)
	s.handleSelection(world, cam)
	s.handleMouse(world, cam)
	s.handleToggles(world)
}

func (s *InputState) handleChallengeInput(world *World, cam *Camera, ch *Challenge) {
	// Escape exits challenge
	if s.justPressed(ebiten.KeyEscape) {
		ch.Exit(world)
		s.aiming = false
		return
	}

	// Level cycling
	if s.justPressed(ebiten.KeyArrowLeft) {
		ch.ChangeLevel(-1, world)
	}
	if s.justPressed(ebiten.KeyArrowRight) {
		ch.ChangeLevel(1, world)
	}

	// After crash/escape: click to retry
	if ch.state == ChallengeCrashed || ch.state == ChallengeEscaped {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !s.aiming {
			ch.RetryLevel(world)
		}
		s.aiming = false
		s.dragging = false
		return
	}

	// Slingshot aiming (only when in aiming state)
	if ch.state != ChallengeAiming {
		return
	}

	if s.panning {
		return
	}

	cx, cy := ebiten.CursorPosition()
	wx, wy := cam.ScreenToWorld(float64(cx), float64(cy))

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !s.aiming {
			s.aiming = true
			s.aimStartX = wx
			s.aimStartY = wy
		}
	} else {
		if s.aiming {
			dx := wx - s.aimStartX
			dy := wy - s.aimStartY
			launchScale := 0.05
			ch.LaunchOrbiter(world, s.aimStartX, s.aimStartY, -dx*launchScale, -dy*launchScale)
		}
		s.aiming = false
	}
}

func (s *InputState) handleToggles(world *World) {
	if s.justPressed(ebiten.KeyG) {
		s.showField = !s.showField
	}
	if s.justPressed(ebiten.KeyM) {
		world.mergeOnCollision = !world.mergeOnCollision
	}
	if s.justPressed(ebiten.KeyF) {
		world.frictionEnabled = !world.frictionEnabled
	}
}

func (s *InputState) handleTimeControl() {
	if s.justPressed(ebiten.KeyP) {
		s.paused = !s.paused
	}
	if s.justPressed(ebiten.KeyEqual) || s.justPressed(ebiten.KeyKPAdd) {
		s.simSpeed *= 1.5
		if s.simSpeed > 4.0 {
			s.simSpeed = 4.0
		}
	}
	if s.justPressed(ebiten.KeyMinus) || s.justPressed(ebiten.KeyKPSubtract) {
		s.simSpeed /= 1.5
		if s.simSpeed < 0.25 {
			s.simSpeed = 0.25
		}
	}
}

func (s *InputState) handleSizeControl() {
	if s.justPressed(ebiten.KeyBracketRight) {
		s.nextRadius += 3
		if s.nextRadius > 60 {
			s.nextRadius = 60
		}
	}
	if s.justPressed(ebiten.KeyBracketLeft) {
		s.nextRadius -= 3
		if s.nextRadius < 3 {
			s.nextRadius = 3
		}
	}
}

func (s *InputState) handleCamera(cam *Camera) {
	_, dy := ebiten.Wheel()
	if dy > 0 {
		cam.ZoomAt(1.1)
	} else if dy < 0 {
		cam.ZoomAt(1.0 / 1.1)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		cx, cy := ebiten.CursorPosition()
		if !s.panning {
			s.panning = true
			s.panStartX = float64(cx)
			s.panStartY = float64(cy)
			s.camStartX = cam.x
			s.camStartY = cam.y
		} else {
			dx := (float64(cx) - s.panStartX) / cam.zoom
			dy := (float64(cy) - s.panStartY) / cam.zoom
			cam.x = s.camStartX - dx
			cam.y = s.camStartY - dy
		}
	} else {
		s.panning = false
	}

	if s.justPressed(ebiten.KeyHome) {
		cam.Reset()
	}
}

func (s *InputState) handleSelection(world *World, cam *Camera) {
	rightDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	justRightClicked := rightDown && !s.prevRightClick
	s.prevRightClick = rightDown

	if justRightClicked {
		cx, cy := ebiten.CursorPosition()
		wx, wy := cam.ScreenToWorld(float64(cx), float64(cy))
		obj := world.FindObject(wx, wy, 15)
		s.selectedObj = obj
	}

	if s.selectedObj != nil {
		if s.justPressed(ebiten.KeyDelete) || s.justPressed(ebiten.KeyBackspace) {
			world.RemoveObject(s.selectedObj)
			s.selectedObj = nil
		}
		if s.justPressed(ebiten.KeySpace) {
			s.selectedObj.pinned = !s.selectedObj.pinned
		}
	}
}

func (s *InputState) handleMouse(world *World, cam *Camera) {
	if s.panning {
		return
	}

	cx, cy := ebiten.CursorPosition()
	wx, wy := cam.ScreenToWorld(float64(cx), float64(cy))

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !s.aiming && !s.dragging {
			obj := world.FindObject(wx, wy, 15)
			if obj != nil {
				s.dragging = true
				s.dragObj = obj
			} else {
				s.aiming = true
				s.aimStartX = wx
				s.aimStartY = wy
			}
		}

		if s.dragging && s.dragObj != nil {
			s.dragObj.x = wx
			s.dragObj.y = wy
			s.dragObj.velocityX = 0
			s.dragObj.velocityY = 0
		}
	} else {
		if s.aiming {
			dx := wx - s.aimStartX
			dy := wy - s.aimStartY
			launchScale := 0.05
			obj := world.AddObject(s.aimStartX, s.aimStartY, s.nextRadius)
			obj.velocityX = -dx * launchScale
			obj.velocityY = -dy * launchScale
		}

		s.aiming = false
		s.dragging = false
		s.dragObj = nil
	}
}

func (s *InputState) cursorWorld(cam *Camera) (float64, float64) {
	cx, cy := ebiten.CursorPosition()
	return cam.ScreenToWorld(float64(cx), float64(cy))
}
