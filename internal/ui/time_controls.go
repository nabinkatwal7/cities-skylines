package ui

import (
	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TimeControls routes pause and speed keys to the simulation (24.19).
type TimeControls struct{}

func NewTimeControls() *TimeControls { return &TimeControls{} }

func (t *TimeControls) HandleInput(sm *sim.SimulationManager, bindings *KeyBindings) {
	if sm == nil || bindings == nil {
		return
	}
	if bindings.Pressed(ActionPause) {
		sm.TogglePause()
	}
	if bindings.Pressed(ActionSpeed1) {
		sm.SetSpeed(1)
	}
	if bindings.Pressed(ActionSpeed2) {
		sm.SetSpeed(2)
	}
	if bindings.Pressed(ActionSpeed3) {
		sm.SetSpeed(3)
	}
}

func (t *TimeControls) DrawHUD(x, y int32, view ViewState, a11y Accessibility) {
	if view.SpeedStr == "" {
		return
	}
	col := rl.Gray
	if a11y.HighContrast {
		col = rl.White
	}
	DrawUITextScaled(T("hud.speed")+": "+view.SpeedStr, x, y, 12, col, a11y.UIScale)
}
