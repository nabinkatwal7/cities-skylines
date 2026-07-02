package ui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CameraController handles pan, rotate, zoom, focus, follow, and reset (24.18).
type CameraController struct {
	Target     rl.Vector3
	Yaw        float32
	Pitch      float32
	Dist       float32
	goalTarget rl.Vector3
	goalYaw    float32
	goalPitch  float32
	goalDist   float32
	followX    float32
	followZ    float32
	following  bool
}

func NewCameraController() *CameraController {
	c := &CameraController{
		Yaw:   float32(-135 * math.Pi / 180),
		Pitch: float32(25 * math.Pi / 180),
		Dist:  140,
	}
	c.Reset()
	c.goalTarget = c.Target
	c.goalYaw = c.Yaw
	c.goalPitch = c.Pitch
	c.goalDist = c.Dist
	return c
}

func (c *CameraController) Reset() {
	c.Target = rl.NewVector3(0, 0, 0)
	c.goalTarget = c.Target
	c.Yaw = float32(-135 * math.Pi / 180)
	c.Pitch = float32(25 * math.Pi / 180)
	c.Dist = 140
	c.goalYaw = c.Yaw
	c.goalPitch = c.Pitch
	c.goalDist = c.Dist
	c.following = false
}

func (c *CameraController) FocusOn(x, z float32) {
	c.goalTarget.X = x
	c.goalTarget.Z = z
	c.following = false
}

func (c *CameraController) SetFollowTarget(x, z float32) {
	c.followX, c.followZ = x, z
	c.following = true
}

func (c *CameraController) ClearFollow() { c.following = false }

func (c *CameraController) Update(dt float32, bindings *KeyBindings, a11y Accessibility, camPos rl.Vector3) {
	if bindings == nil {
		bindings = DefaultBindings()
	}
	lerp := float32(0.08)
	if a11y.ReducedMotion {
		lerp = 1
	}

	wheel := rl.GetMouseWheelMove()
	c.goalDist -= wheel * 5
	if c.goalDist < 5 {
		c.goalDist = 5
	}
	if c.goalDist > 500 {
		c.goalDist = 500
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		delta := rl.GetMouseDelta()
		c.goalYaw -= delta.X * 0.005
		c.goalPitch -= delta.Y * 0.005
		if c.goalPitch > 1.4 {
			c.goalPitch = 1.4
		}
		if c.goalPitch < -0.3 {
			c.goalPitch = -0.3
		}
	}

	speed := float32(80) * dt
	if bindings.Down(ActionCamForward) {
		forward := rl.Vector3Subtract(c.Target, camPos)
		forward.Y = 0
		forward = rl.Vector3Normalize(forward)
		c.goalTarget = rl.Vector3Add(c.goalTarget, rl.Vector3Scale(forward, speed))
	}
	if bindings.Down(ActionCamBack) {
		forward := rl.Vector3Subtract(c.Target, camPos)
		forward.Y = 0
		forward = rl.Vector3Normalize(forward)
		c.goalTarget = rl.Vector3Add(c.goalTarget, rl.Vector3Scale(forward, -speed))
	}
	if bindings.Down(ActionCamLeft) {
		forward := rl.Vector3Subtract(c.Target, camPos)
		forward.Y = 0
		forward = rl.Vector3Normalize(forward)
		right := rl.Vector3CrossProduct(forward, rl.NewVector3(0, 1, 0))
		c.goalTarget = rl.Vector3Add(c.goalTarget, rl.Vector3Scale(right, -speed))
	}
	if bindings.Down(ActionCamRight) {
		forward := rl.Vector3Subtract(c.Target, camPos)
		forward.Y = 0
		forward = rl.Vector3Normalize(forward)
		right := rl.Vector3CrossProduct(forward, rl.NewVector3(0, 1, 0))
		c.goalTarget = rl.Vector3Add(c.goalTarget, rl.Vector3Scale(right, speed))
	}
	if bindings.Down(ActionCamDown) {
		c.goalTarget.Y -= speed
	}
	if bindings.Down(ActionCamUp) {
		c.goalTarget.Y += speed
	}

	if c.following {
		c.goalTarget.X = c.followX
		c.goalTarget.Z = c.followZ
	}

	c.Target.X += (c.goalTarget.X - c.Target.X) * lerp
	c.Target.Y += (c.goalTarget.Y - c.Target.Y) * lerp
	c.Target.Z += (c.goalTarget.Z - c.Target.Z) * lerp
	c.Yaw += (c.goalYaw - c.Yaw) * lerp
	c.Pitch += (c.goalPitch - c.Pitch) * lerp
	c.Dist += (c.goalDist - c.Dist) * lerp
}

func (c *CameraController) Camera3D() rl.Camera3D {
	pos := rl.NewVector3(
		c.Target.X+c.Dist*float32(math.Cos(float64(c.Pitch)))*float32(math.Sin(float64(c.Yaw))),
		c.Target.Y+c.Dist*float32(math.Sin(float64(c.Pitch))),
		c.Target.Z+c.Dist*float32(math.Cos(float64(c.Pitch)))*float32(math.Cos(float64(c.Yaw))),
	)
	return rl.Camera3D{
		Position:   pos,
		Target:     c.Target,
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       60,
		Projection: rl.CameraPerspective,
	}
}

func (c *CameraController) Position() rl.Vector3 { return c.Camera3D().Position }
