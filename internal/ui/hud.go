package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/core"

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

func (h *HUD) Draw(notices *Notifications, settings *PlayerSettings) {
	h.dirty = false
	_ = settings
	drawBarTop()

	// Left cluster — money & weekly change (CS style)
	drawLabel(fmt.Sprintf("$%s", formatMoney(h.view.Money)), 14, 8, FontXl, csMoney)
	incomeCol := csMoney
	if h.view.WeeklyIncome < 0 {
		incomeCol = csMoneyNeg
	}
	drawLabel(fmt.Sprintf("%+0.f /wk", h.view.WeeklyIncome), 14, 34, FontSm, incomeCol)

	// Population & happiness
	px := int32(200)
	drawLabel(fmt.Sprintf("%d", h.view.Population), px, 8, FontLg, csPop)
	drawLabel("Population", px, 32, FontXs, csTextDim)

	hx := int32(310)
	happyPct := int(h.view.Happiness * 100)
	happyCol := csHappy
	if happyPct < 40 {
		happyCol = csHappyLow
	}
	drawLabel(fmt.Sprintf("%d%%", happyPct), hx, 8, FontLg, happyCol)
	drawLabel("Happiness", hx, 32, FontXs, csTextDim)

	// Center — date & time
	if h.view.DateStr != "" {
		dw := MeasureUIText(h.view.DateStr, FontMd)
		drawLabel(h.view.DateStr, ScreenW/2-dw/2-40, 10, FontMd, csText)
	}
	if h.view.TimeStr != "" {
		tw := MeasureUIText(h.view.TimeStr, FontLg)
		drawLabel(h.view.TimeStr, ScreenW/2-tw/2+20, 8, FontLg, csText)
	}

	// Milestone badge
	if h.view.Milestone != "" {
		mw := MeasureUIText(h.view.Milestone, FontSm)
		bx := ScreenW/2 - mw/2
		rl.DrawRectangle(bx-8, 30, mw+16, 16, rl.NewColor(50, 90, 120, 200))
		drawLabel(h.view.Milestone, bx, 30, FontSm, csTextDim)
	}

	// Right — speed & coords
	if h.view.SpeedStr != "" {
		drawLabel(h.view.SpeedStr, ScreenW-120, 10, FontMd, csBarLine)
	}
	if h.view.MouseOnGround {
		coordStr := fmt.Sprintf("%.0f, %.0f", h.view.MouseWorldX, h.view.MouseWorldZ)
		cw := MeasureUIText(coordStr, FontSm)
		drawLabel(coordStr, ScreenW-cw-12, 32, FontSm, csTextDim)
	}

	if notices != nil {
		notices.DrawHUD(TopBarH + 4)
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
