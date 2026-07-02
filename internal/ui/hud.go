package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// HUD presents critical city information (24.2) — CS top bar layout.
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

func (h *HUD) timeControlRect() (x, y, w, btnH int32) {
	btnW := int32(34)
	btnH = 30
	w = 4*btnW + 3*4
	x = ScreenW - w - 10
	y = (TopBarH - btnH) / 2
	return x, y, w, btnH
}

func (h *HUD) HandleClick(mx, my int32, sm *sim.SimulationManager) bool {
	if sm == nil || my >= TopBarH {
		return false
	}
	rx, ry, _, btnH := h.timeControlRect()
	btnW := int32(34)
	if mx < rx || mx >= rx+4*btnW+12 {
		return false
	}
	if my < ry || my >= ry+btnH {
		return false
	}
	idx := (mx - rx) / (btnW + 4)
	switch idx {
	case 0:
		sm.TogglePause()
	case 1:
		sm.SetSpeed(1)
	case 2:
		sm.SetSpeed(2)
	case 3:
		sm.SetSpeed(3)
	default:
		return false
	}
	return true
}

func (h *HUD) Draw(notices *Notifications, settings *PlayerSettings, sm *sim.SimulationManager) {
	h.dirty = false
	_ = settings
	_ = sm
	drawBarTop()

	drawLabel(fmt.Sprintf("$%s", formatMoney(h.view.Money)), 14, 10, FontXl, csMoney)
	incomeCol := csMoney
	if h.view.WeeklyIncome < 0 {
		incomeCol = csMoneyNeg
	}
	drawLabel(fmt.Sprintf("%+0.f /wk", h.view.WeeklyIncome), 14, 38, FontSm, incomeCol)

	px := int32(210)
	drawLabel(fmt.Sprintf("%d", h.view.Population), px, 10, FontLg, csPop)
	drawLabel("Population", px, 34, FontXs, csTextDim)

	hx := int32(330)
	happyPct := int(h.view.Happiness * 100)
	happyCol := csHappy
	if happyPct < 40 {
		happyCol = csHappyLow
	}
	drawLabel(fmt.Sprintf("%d%%", happyPct), hx, 10, FontLg, happyCol)
	drawLabel("Happiness", hx, 34, FontXs, csTextDim)

	if h.view.DateStr != "" {
		dw := MeasureUIText(h.view.DateStr, FontMd)
		drawLabel(h.view.DateStr, ScreenW/2-dw/2-50, 12, FontMd, csTextDim)
	}
	if h.view.TimeStr != "" {
		tw := MeasureUIText(h.view.TimeStr, FontLg)
		drawLabel(h.view.TimeStr, ScreenW/2-tw/2, 8, FontLg, csText)
	}

	if h.view.Milestone != "" {
		mw := MeasureUIText(h.view.Milestone, FontSm)
		bx := ScreenW/2 - mw/2
		rl.DrawRectangle(bx-8, 36, mw+16, 16, rl.NewColor(50, 90, 120, 200))
		drawLabel(h.view.Milestone, bx, 36, FontSm, csTextDim)
	}

	h.drawTimeControls(sm)

	if h.view.MouseOnGround {
		coordStr := fmt.Sprintf("%.0f, %.0f", h.view.MouseWorldX, h.view.MouseWorldZ)
		cw := MeasureUIText(coordStr, FontXs)
		drawLabel(coordStr, ScreenW-cw-12, TopBarH-18, FontXs, csTextDim)
	}

	if notices != nil {
		notices.DrawHUD(TopBarH + 4)
	}
}

func (h *HUD) drawTimeControls(sm *sim.SimulationManager) {
	rx, ry, _, btnH := h.timeControlRect()
	btnW := int32(34)
	labels := []string{"||", ">", ">>", ">>>"}
	paused := sm != nil && sm.IsPaused()
	speed := float64(1)
	if sm != nil && sm.Time != nil {
		speed = sm.Time.Speed
	}
	for i, label := range labels {
		bx := rx + int32(i)*(btnW+4)
		active := (i == 0 && paused) || (i == 1 && !paused && speed <= 1) || (i == 2 && !paused && speed > 1 && speed <= 2) || (i == 3 && !paused && speed > 2)
		fill := csBtnIdle
		if active {
			fill = csSelectFill
		}
		csOptionBtn(bx, ry, btnW, btnH, label, fill, active)
	}
}

func formatMoney(v float32) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", v/1_000_000)
	}
	if v >= 10_000 {
		return fmt.Sprintf("%.0fK", v/1_000)
	}
	return fmt.Sprintf("%.0f", v)
}
