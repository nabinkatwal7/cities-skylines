package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// InputDevice tracks the active input source (24.22).
type InputDevice int

const (
	DeviceKeyboard InputDevice = iota
	DeviceGamepad
)

// GamepadInput maps controller axes to camera and UI hints (24.22).
type GamepadInput struct {
	device InputDevice
	radial bool // ponytail: radial menu shell for future expansion
}

func NewGamepadInput() *GamepadInput { return &GamepadInput{device: DeviceKeyboard} }

func (g *GamepadInput) Device() InputDevice { return g.device }

func (g *GamepadInput) RadialOpen() bool { return g.radial }

func (g *GamepadInput) Update(cam *CameraController, a11y Accessibility) {
	if !rl.IsGamepadAvailable(0) {
		g.device = DeviceKeyboard
		return
	}
	g.device = DeviceGamepad
	if cam == nil {
		return
	}
	dt := float32(rl.GetFrameTime())
	lx := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftX)
	lz := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftY)
	if lx*lx+lz*lz > 0.01 {
		cam.goalTarget.X += lx * 90 * dt
		cam.goalTarget.Z += lz * 90 * dt
	}
	rx := rl.GetGamepadAxisMovement(0, rl.GamepadAxisRightX)
	ry := rl.GetGamepadAxisMovement(0, rl.GamepadAxisRightY)
	if rx*rx+ry*ry > 0.01 {
		cam.goalYaw -= rx * 1.5 * dt
		cam.goalPitch -= ry * 1.5 * dt
		if cam.goalPitch > 1.4 {
			cam.goalPitch = 1.4
		}
		if cam.goalPitch < -0.3 {
			cam.goalPitch = -0.3
		}
	}
	if rl.IsGamepadButtonPressed(0, rl.GamepadButtonRightTrigger2) {
		cam.goalDist -= 40 * dt
	}
	if rl.IsGamepadButtonPressed(0, rl.GamepadButtonLeftTrigger2) {
		cam.goalDist += 40 * dt
	}
	if cam.goalDist < 5 {
		cam.goalDist = 5
	}
	if cam.goalDist > 500 {
		cam.goalDist = 500
	}
	if rl.IsGamepadButtonPressed(0, rl.GamepadButtonMiddleRight) && !a11y.ReducedMotion {
		g.radial = !g.radial
	}
}

func (g *GamepadInput) DrawHint(y int32, a11y Accessibility) {
	if g.device != DeviceGamepad {
		return
	}
	DrawUITextScaled(T("input.gamepad"), 8, y, 11, rl.Gray, a11y.UIScale)
	if g.radial {
		DrawUITextScaled("Radial Menu", 80, y, 11, rl.LightGray, a11y.UIScale)
	}
}
