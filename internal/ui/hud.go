package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// HUD presents critical city information (24.2).
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

func (h *HUD) Draw(notices *Notifications, settings *PlayerSettings) {
	h.dirty = false

	rl.DrawRectangle(0, 0, ScreenW, TopBarH, rl.NewColor(0, 0, 0, 190))

	// Row 1: treasury + economy
	DrawUIText(fmt.Sprintf("$%.0f", h.view.Money), 8, 4, 18, rl.NewColor(100, 220, 100, 230))
	incomeCol := rl.NewColor(120, 200, 120, 200)
	if h.view.WeeklyIncome < 0 {
		incomeCol = rl.NewColor(220, 120, 120, 200)
	}
	DrawUIText(fmt.Sprintf("wk %+0.f", h.view.WeeklyIncome), 100, 8, 13, incomeCol)

	DrawUIText(fmt.Sprintf("Pop %d", h.view.Population), 200, 4, 15, rl.NewColor(180, 220, 255, 220))
	happyPct := int(h.view.Happiness * 100)
	happyCol := rl.NewColor(180, 220, 120, 220)
	if happyPct < 40 {
		happyCol = rl.NewColor(220, 140, 100, 220)
	}
	DrawUIText(fmt.Sprintf("Happy %d%%", happyPct), 290, 4, 15, happyCol)

	if h.view.Milestone != "" {
		DrawUIText(h.view.Milestone, 400, 4, 14, rl.NewColor(255, 220, 140, 210))
	}

	// Row 2: date, time, speed, coords
	if h.view.DateStr != "" {
		DrawUIText(h.view.DateStr, 8, 28, 13, rl.Gray)
	}
	DrawUIText(h.view.TimeStr, 120, 28, 13, rl.Gray)
	if h.view.SpeedStr != "" {
		DrawUIText(h.view.SpeedStr, 220, 28, 13, rl.NewColor(160, 200, 255, 200))
	}
	if h.view.MouseOnGround {
		coordStr := fmt.Sprintf("(%.1f, %.1f)", h.view.MouseWorldX, h.view.MouseWorldZ)
		DrawUIText(coordStr, ScreenW-160, 28, 13, rl.Gray)
	}

	if notices != nil {
		notices.DrawHUD(TopBarH + 2)
	}
}
