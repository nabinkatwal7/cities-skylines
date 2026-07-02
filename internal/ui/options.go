package ui

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// OptionsPanel exposes bindings, accessibility, and locale (24.21, 24.23, 24.24).
type OptionsPanel struct {
	open   bool
	tab    int
	inited bool
}

func NewOptionsPanel() *OptionsPanel { return &OptionsPanel{} }

func (o *OptionsPanel) Open() bool { return o.open }

func (o *OptionsPanel) ensureInit() {
	if o.inited {
		return
	}
	o.inited = true
}

func (o *OptionsPanel) Draw(settings *PlayerSettings) {
	if !o.open || settings == nil {
		return
	}
	o.ensureInit()
	scale := settings.UIScale()
	w := ScaleSize(440, scale)
	h := ScaleSize(320, scale)
	x := int32((ScreenW - w) / 2)
	y := ScaleY(TopBarH+8, scale)
	bg := rl.NewColor(0, 0, 0, 220)
	if settings.A11y.HighContrast {
		bg = rl.Black
	}
	rl.DrawRectangle(x, y, w, h, bg)
	rl.DrawRectangleLines(x, y, w, h, rl.Gray)
	DrawUITextScaled(T("options.title"), x+10, y+8, 16, rl.White, scale)

	tabs := []string{T("options.accessibility"), T("options.bindings"), T("options.language")}
	for i, label := range tabs {
		bx := x + 10 + int32(i)*130
		sel := i == o.tab
		col := rl.NewColor(40, 40, 45, 200)
		if sel {
			col = rl.NewColor(60, 70, 90, 220)
		}
		uiBtn(bx, y+28, 120, 20, label, col, rl.White, sel)
	}
	ly := y + 56
	switch o.tab {
	case 0:
		o.drawA11y(x, ly, settings)
	case 1:
		o.drawBindings(x, ly, settings)
	case 2:
		o.drawLocale(x, ly, settings)
	}
}

func (o *OptionsPanel) drawA11y(x, ly int32, settings *PlayerSettings) {
	scale := settings.UIScale()
	a := &settings.A11y
	DrawUITextScaled(fmt.Sprintf("%s: %.0f%%", T("a11y.ui_scale"), a.UIScale*100), x+10, ly, 13, rl.LightGray, scale)
	DrawUITextScaled(toggleLabel(T("a11y.high_contrast"), a.HighContrast), x+10, ly+22, 13, rl.LightGray, scale)
	DrawUITextScaled(toggleLabel(T("a11y.subtitles"), a.Subtitles), x+10, ly+44, 13, rl.LightGray, scale)
	DrawUITextScaled(toggleLabel(T("a11y.reduced_motion"), a.ReducedMotion), x+10, ly+66, 13, rl.LightGray, scale)
	DrawUITextScaled(fmt.Sprintf("%s: %d", T("a11y.color_blind"), int(a.ColorBlindMode)), x+10, ly+88, 13, rl.LightGray, scale)
}

func (o *OptionsPanel) drawBindings(x, ly int32, settings *PlayerSettings) {
	scale := settings.UIScale()
	actions := []InputAction{
		ActionPause, ActionSpeed1, ActionSpeed2, ActionSpeed3, ActionUndo,
		ActionScreenshot, ActionCamReset, ActionSearchToggle, ActionStatisticsToggle,
	}
	for i, act := range actions {
		DrawUITextScaled(fmt.Sprintf("%s: %s", actionLabel(act), keyName(settings.Bindings.Get(act))),
			x+10, ly+int32(i*18), 12, rl.LightGray, scale)
	}
}

func (o *OptionsPanel) drawLocale(x, ly int32, settings *PlayerSettings) {
	scale := settings.UIScale()
	for i, loc := range AvailableLocales() {
		col := rl.LightGray
		if loc == settings.Locale {
			col = rl.SkyBlue
		}
		DrawUITextScaled(loc, x+10, ly+int32(i*22), 14, col, scale)
	}
}

func toggleLabel(label string, on bool) string {
	if on {
		return label + ": ON"
	}
	return label + ": OFF"
}

func (o *OptionsPanel) HandleClick(mx, my int32, settings *PlayerSettings) bool {
	if !o.open || settings == nil {
		return false
	}
	scale := settings.UIScale()
	w := ScaleSize(440, scale)
	h := ScaleSize(320, scale)
	x := int32((ScreenW - w) / 2)
	y := ScaleY(TopBarH+8, scale)
	if mx < x || mx >= x+w || my < y || my >= y+h {
		return false
	}
	if my >= y+28 && my < y+48 {
		tab := int((mx - x - 10) / 130)
		if tab >= 0 && tab < 3 {
			o.tab = tab
		}
		return true
	}
	if o.tab == 0 {
		row := int((my - (y + 56)) / 22)
		switch row {
		case 1:
			settings.A11y.HighContrast = !settings.A11y.HighContrast
		case 2:
			settings.A11y.Subtitles = !settings.A11y.Subtitles
		case 3:
			settings.A11y.ReducedMotion = !settings.A11y.ReducedMotion
		case 4:
			settings.A11y.ColorBlindMode = (settings.A11y.ColorBlindMode + 1) % 4
		}
	}
	if o.tab == 2 {
		row := int((my - (y + 56)) / 22)
		locales := AvailableLocales()
		if row >= 0 && row < len(locales) {
			settings.Locale = locales[row]
			SetLocale(settings.Locale)
		}
	}
	return true
}

// GlobalShortcuts handles cross-cutting hotkeys (24.21).
type GlobalShortcuts struct {
	screenshot bool
	subtitle   string
	subtitleAt time.Time
}

func NewGlobalShortcuts() *GlobalShortcuts { return &GlobalShortcuts{} }

func (g *GlobalShortcuts) Handle(m *UIManager) {
	if m == nil || m.Settings == nil || m.Settings.Bindings == nil {
		return
	}
	b := m.Settings.Bindings
	if b.Pressed(ActionCamReset) {
		m.Camera.Reset()
	}
	if b.Pressed(ActionUndo) && m.lastSim != nil && m.lastSim.Undo != nil {
		if m.lastSim.Undo.Undo(m.lastSim) {
			g.flashSubtitle(T("bind.undo"))
		}
	}
	if b.Pressed(ActionScreenshot) {
		g.screenshot = true
	}
	if b.Pressed(ActionStatisticsToggle) {
		if m.Toolbar.Selected == CatStatistics {
			m.Toolbar.Selected = CatRoads
			m.Statistics.open = false
		} else {
			m.Toolbar.Select(CatStatistics, &m.ToolSystem, m.BuildMenus)
			m.Statistics.open = true
		}
	}
	if b.Pressed(ActionInfoCycle) {
		m.InfoViews.Cycle()
	}
	if b.Pressed(ActionInfoClear) {
		m.InfoViews.Clear()
	}
	if b.Pressed(ActionAdvisorToggle) {
		m.Advisors.Toggle()
	}
	if b.Pressed(ActionSearchToggle) {
		if m.Search.IsOpen() {
			m.Search.Close()
		} else {
			m.Search.Open()
		}
	}
}

func (g *GlobalShortcuts) ConsumeScreenshot() bool {
	if !g.screenshot {
		return false
	}
	g.screenshot = false
	return true
}

func (g *GlobalShortcuts) flashSubtitle(msg string) {
	g.subtitle = msg
	g.subtitleAt = time.Now()
}

func (g *GlobalShortcuts) Draw(settings *PlayerSettings) {
	if settings == nil || !settings.A11y.Subtitles || g.subtitle == "" {
		return
	}
	if time.Since(g.subtitleAt) > 3*time.Second {
		g.subtitle = ""
		return
	}
	DrawUITextScaled(T("subtitle.prefix")+": "+g.subtitle, 10, ScreenH-28, 14, rl.White, settings.UIScale())
}

func (g *GlobalShortcuts) ScreenReaderHint() string { return g.subtitle }
