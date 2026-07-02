package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// InputAction identifies a configurable binding (24.21).
type InputAction int

const (
	ActionPause InputAction = iota
	ActionSpeed1
	ActionSpeed2
	ActionSpeed3
	ActionUndo
	ActionScreenshot
	ActionCamReset
	ActionCamForward
	ActionCamBack
	ActionCamLeft
	ActionCamRight
	ActionCamUp
	ActionCamDown
	ActionInfoCycle
	ActionInfoClear
	ActionAdvisorToggle
	ActionSearchToggle
	ActionStatisticsToggle
)

const bindingCount = int(ActionStatisticsToggle) + 1

// KeyBindings stores configurable shortcuts (24.21).
type KeyBindings struct {
	keys [bindingCount]int32
}

func DefaultBindings() *KeyBindings {
	b := &KeyBindings{}
	b.keys[ActionPause] = rl.KeySpace
	b.keys[ActionSpeed1] = rl.KeyF1
	b.keys[ActionSpeed2] = rl.KeyF2
	b.keys[ActionSpeed3] = rl.KeyF3
	b.keys[ActionUndo] = rl.KeyZ
	b.keys[ActionScreenshot] = rl.KeyF12
	b.keys[ActionCamReset] = rl.KeyHome
	b.keys[ActionCamForward] = rl.KeyW
	b.keys[ActionCamBack] = rl.KeyS
	b.keys[ActionCamLeft] = rl.KeyA
	b.keys[ActionCamRight] = rl.KeyD
	b.keys[ActionCamUp] = rl.KeyE
	b.keys[ActionCamDown] = rl.KeyQ
	b.keys[ActionInfoCycle] = rl.KeyF4
	b.keys[ActionInfoClear] = rl.KeyF5
	b.keys[ActionAdvisorToggle] = rl.KeyF6
	b.keys[ActionSearchToggle] = rl.KeySlash
	b.keys[ActionStatisticsToggle] = rl.KeyF7
	return b
}

func (b *KeyBindings) Get(action InputAction) int32 {
	if b == nil || int(action) < 0 || int(action) >= bindingCount {
		return 0
	}
	return b.keys[action]
}

func (b *KeyBindings) Set(action InputAction, key int32) {
	if b == nil || int(action) < 0 || int(action) >= bindingCount {
		return
	}
	b.keys[action] = key
}

func (b *KeyBindings) Pressed(action InputAction) bool {
	k := b.Get(action)
	if k == 0 {
		return false
	}
	if action == ActionUndo {
		return rl.IsKeyPressed(k) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl))
	}
	return rl.IsKeyPressed(k)
}

func (b *KeyBindings) Down(action InputAction) bool {
	k := b.Get(action)
	return k != 0 && rl.IsKeyDown(k)
}

func actionLabel(action InputAction) string {
	switch action {
	case ActionPause:
		return T("bind.pause")
	case ActionSpeed1:
		return T("bind.speed1")
	case ActionSpeed2:
		return T("bind.speed2")
	case ActionSpeed3:
		return T("bind.speed3")
	case ActionUndo:
		return T("bind.undo")
	case ActionScreenshot:
		return T("bind.screenshot")
	case ActionCamReset:
		return T("bind.cam_reset")
	case ActionInfoCycle:
		return T("bind.info_cycle")
	case ActionInfoClear:
		return T("bind.info_clear")
	case ActionAdvisorToggle:
		return T("bind.advisors")
	case ActionSearchToggle:
		return T("bind.search")
	case ActionStatisticsToggle:
		return T("bind.statistics")
	default:
		return "?"
	}
}

func keyName(key int32) string {
	switch key {
	case rl.KeySpace:
		return "Space"
	case rl.KeySlash:
		return "/"
	case rl.KeyHome:
		return "Home"
	case rl.KeyF1:
		return "F1"
	case rl.KeyF2:
		return "F2"
	case rl.KeyF3:
		return "F3"
	case rl.KeyF4:
		return "F4"
	case rl.KeyF5:
		return "F5"
	case rl.KeyF6:
		return "F6"
	case rl.KeyF7:
		return "F7"
	case rl.KeyF12:
		return "F12"
	case rl.KeyZ:
		return "Ctrl+Z"
	default:
		if key >= rl.KeyA && key <= rl.KeyZ {
			return string(rune('A' + (key - rl.KeyA)))
		}
		return "?"
	}
}
