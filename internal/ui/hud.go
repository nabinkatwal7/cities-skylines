package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// HUD presents critical city information. Values refresh on simulation events.
type HUD struct {
	dirty bool
	view  ViewState
}

func NewHUD() *HUD {
	return &HUD{dirty: true}
}

func (h *HUD) Subscribe(bus *core.EventBus) []func() {
	if bus == nil {
		return nil
	}
	mark := func(any) { h.dirty = true }
	return []func(){
		bus.On(string(core.EventTimeMinute), mark),
		bus.On(string(core.EventTimeHour), mark),
		bus.On(string(core.EventTimeDay), mark),
		bus.On(string(core.EventTaxCollected), mark),
		bus.On(string(core.EventDayNightCycle), mark),
	}
}

func (h *HUD) MarkDirty() { h.dirty = true }

func (h *HUD) Sync(view ViewState) {
	h.view = view
}

func (h *HUD) Draw() {
	rl.DrawRectangle(0, 0, ScreenW, TopBarH, rl.NewColor(0, 0, 0, 180))
	DrawUIText(fmt.Sprintf("$%.0f", h.view.Money), 10, 10, 20, rl.NewColor(100, 220, 100, 220))
	DrawUIText(h.view.TimeStr, 160, 10, 16, rl.Gray)
	if h.view.MouseOnGround {
		coordStr := fmt.Sprintf("(%.1f, %.1f)", h.view.MouseWorldX, h.view.MouseWorldZ)
		DrawUIText(coordStr, ScreenW-180, 10, 14, rl.Gray)
	}
}
