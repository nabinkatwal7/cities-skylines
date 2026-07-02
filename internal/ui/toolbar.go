package ui

import (
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToolbarCategory int

const (
	CatRoads ToolbarCategory = iota
	CatZoning
	CatDistricts
	CatElectricity
	CatWater
	CatGarbage
	CatHealthcare
	CatFireRescue
	CatPolice
	CatEducation
	CatPublicTransport
	CatLandscaping
	CatParks
	CatEconomy
	CatPolicies
	CatStatistics
	CatOptions
	catCount
)

type CategoryDef struct {
	Cat   ToolbarCategory
	Label string
	Color rl.Color
	Key   int32 // 0 = no hotkey
}

var AllCategories = []CategoryDef{
	{CatRoads, "Roads", rl.NewColor(180, 160, 120, 255), rl.KeyTwo},
	{CatZoning, "Zones", rl.NewColor(120, 200, 120, 255), rl.KeyThree},
	{CatDistricts, "Districts", rl.NewColor(160, 140, 200, 255), 0},
	{CatElectricity, "Power", rl.NewColor(255, 220, 80, 255), 0},
	{CatWater, "Water", rl.NewColor(80, 160, 255, 255), 0},
	{CatGarbage, "Garbage", rl.NewColor(120, 100, 80, 255), 0},
	{CatHealthcare, "Health", rl.NewColor(255, 120, 120, 255), 0},
	{CatFireRescue, "Fire", rl.NewColor(255, 100, 50, 255), 0},
	{CatPolice, "Police", rl.NewColor(80, 120, 220, 255), 0},
	{CatEducation, "Education", rl.NewColor(100, 180, 255, 255), 0},
	{CatPublicTransport, "Transit", rl.NewColor(50, 150, 200, 255), rl.KeyFour},
	{CatLandscaping, "Landscape", rl.NewColor(100, 160, 80, 255), 0},
	{CatParks, "Parks", rl.NewColor(80, 200, 100, 255), 0},
	{CatEconomy, "Economy", rl.NewColor(200, 180, 60, 255), 0},
	{CatPolicies, "Policies", rl.NewColor(180, 120, 200, 255), 0},
	{CatStatistics, "Stats", rl.NewColor(160, 160, 200, 255), 0},
	{CatOptions, "Options", rl.NewColor(140, 140, 140, 255), 0},
}

// MainToolbar is the primary build-category selector (24.3).
type MainToolbar struct {
	Selected ToolbarCategory
	Scroll   int
	Unlocks  *UnlockRegistry
}

func NewMainToolbar(unlocks *UnlockRegistry) *MainToolbar {
	if unlocks == nil {
		unlocks = NewUnlockRegistry()
	}
	return &MainToolbar{
		Selected: CatRoads,
		Unlocks:  unlocks,
	}
}

func (tb *MainToolbar) VisibleCategories() []CategoryDef {
	out := make([]CategoryDef, 0, len(AllCategories))
	for _, c := range AllCategories {
		if tb.Unlocks.Unlocked(c.Cat) {
			out = append(out, c)
		}
	}
	return out
}

func (tb *MainToolbar) Select(cat ToolbarCategory, ts *ToolSystem, menus *BuildMenus) {
	tb.Selected = cat
	ts.ApplyCategory(cat)
	if menus != nil {
		menus.OpenCategory(cat)
	}
}

func (tb *MainToolbar) HandleKeyboard(ts *ToolSystem, menus *BuildMenus) {
	if rl.IsKeyPressed(rl.KeyOne) {
		ts.Activate(ToolPointer)
	}
	if rl.IsKeyPressed(rl.KeyI) {
		ts.Activate(ToolInspect)
	}
	if rl.IsKeyPressed(rl.KeyM) {
		ts.Activate(ToolMeasure)
	}
	for _, c := range tb.VisibleCategories() {
		if c.Key != 0 && rl.IsKeyPressed(c.Key) {
			tb.Select(c.Cat, ts, menus)
		}
	}
	if rl.IsKeyPressed(rl.KeySix) {
		ts.Activate(ToolRemove)
	}
	if rl.IsKeyPressed(rl.KeySeven) {
		ts.Activate(ToolUpgrade)
	}
}

const utilBtnW int32 = 58

func (tb *MainToolbar) HandleClick(ts *ToolSystem, menus *BuildMenus) GameTool {
	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)

	if my >= ToolbarY && my < ToolbarY+ToolbarBtnH {
		utils := []GameTool{ToolPointer, ToolInspect, ToolMeasure, ToolRemove, ToolUpgrade}
		for i, tool := range utils {
			bx := int32(4 + i*(int(utilBtnW)+int(ToolbarPad)))
			if mx >= bx && mx < bx+utilBtnW {
				ts.Activate(tool)
				return ts.Selected
			}
		}
	}

	visible := tb.VisibleCategories()
	btnW := int32(ToolbarBtnW)
	totalW := len(visible)*int(btnW+ToolbarPad) - int(ToolbarPad)
	startX := (ScreenW - totalW) / 2
	for i, c := range visible {
		bx := int32(startX + i*int(btnW+ToolbarPad))
		by := int32(ToolbarY)
		if mx >= bx && mx < bx+btnW && my >= by && my < by+ToolbarBtnH {
			tb.Select(c.Cat, ts, menus)
			return ts.Selected
		}
	}
	return ts.Selected
}

func (tb *MainToolbar) Draw(ts *ToolSystem) {
	rl.DrawRectangle(0, ToolbarY, ScreenW, ToolbarH, rl.NewColor(0, 0, 0, 200))

	// Utility strip
	utils := []struct {
		tool  GameTool
		label string
		col   rl.Color
	}{
		{ToolPointer, "Ptr", rl.NewColor(200, 200, 200, 255)},
		{ToolInspect, "Insp", rl.NewColor(160, 200, 255, 255)},
		{ToolMeasure, "Meas", rl.NewColor(200, 180, 120, 255)},
		{ToolRemove, "Del", rl.NewColor(200, 80, 80, 255)},
		{ToolUpgrade, "Upg", rl.NewColor(200, 200, 80, 255)},
	}
	for i, u := range utils {
		bx := int32(4 + i*(int(utilBtnW)+int(ToolbarPad)))
		by := int32(ToolbarY + (ToolbarH-ToolbarBtnH)/2)
		sel := ts.Selected == u.tool
		textCol := rl.White
		if sel {
			textCol = rl.NewColor(255, 255, 200, 255)
		}
		uiBtn(bx, by, utilBtnW, ToolbarBtnH, u.label, u.col, textCol, sel)
	}

	visible := tb.VisibleCategories()
	btnW := int32(ToolbarBtnW)
	totalW := len(visible)*int(btnW+ToolbarPad) - int(ToolbarPad)
	startX := (ScreenW - totalW) / 2
	for i, c := range visible {
		bx := int32(startX + i*int(btnW+ToolbarPad))
		by := int32(ToolbarY + (ToolbarH-ToolbarBtnH)/2)
		sel := (tb.Selected == c.Cat && ts.Mode == ModePlace) || (ts.Mode == ModePaint && c.Cat == CatZoning)
		textCol := rl.White
		if sel {
			textCol = rl.NewColor(255, 255, 200, 255)
		}
		label := c.Label
		if len(label) > 8 {
			label = label[:7] + "…"
		}
		uiBtn(bx, by, btnW, ToolbarBtnH, label, c.Color, textCol, sel)
	}
}

func (tb *MainToolbar) drawLegacyOptions(ts *ToolSystem) {
	if ts.Selected == ToolPointer || ts.Selected == ToolRemove || ts.Selected == ToolUpgrade {
		return
	}
	y := tb.optionsY()
	rl.DrawRectangle(0, y, ScreenW, OptionsBarH, rl.NewColor(0, 0, 0, 160))

	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)

	switch ts.Selected {
	case ToolRoad:
		item := roadOptions()
		drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			ts.RoadType = road.RoadType(oi)
		})
	case ToolParking:
		item := parkingOptions()
		drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			ts.setParkingMode(oi)
		})
	case ToolTransport:
		item := transportOptions()
		drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			if oi >= int(transport.TransportModeCount) {
				ts.CargoMode = true
			} else {
				ts.CargoMode = false
				ts.TransportType = transport.TransportType(oi)
			}
		})
	case ToolZone:
		item := zoneOptions()
		drawOptionRowColored(item, y, mx, my, func(oi int) rl.Color {
			if oi < 6 {
				c := zoning.ZoneColor(zoning.ZoneType(oi + 1))
				c.A = 200
				return c
			}
			return rl.NewColor(40, 40, 40, 200)
		}, func(oi int) {
			item.OptIndex = oi
			ts.ZoneType = oi
		})
	}
}

func (tb *MainToolbar) optionsY() int32 {
	return ToolbarY - OptionsBarH
}

func (tb *MainToolbar) DrawOptions(ts *ToolSystem) {
	tb.drawLegacyOptions(ts)
}

func drawOptionRow(item *ToolbarItem, y int32, mx, my int32, onSelect func(int)) {
	drawOptionRowColored(item, y, mx, my, func(int) rl.Color {
		return rl.NewColor(40, 40, 40, 200)
	}, onSelect)
}

func drawOptionRowColored(item *ToolbarItem, y int32, mx, my int32, colFn func(int) rl.Color, onSelect func(int)) {
	optW := int32(80)
	total := len(item.Options) * int(optW+ToolbarPad)
	sx := (ScreenW - total) / 2
	for oi, opt := range item.Options {
		bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
		by := y + 4
		sel := oi == item.OptIndex
		uiBtn(bx, by, optW, OptionsBarH-8, opt, colFn(oi), rl.White, sel)
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
			onSelect(oi)
		}
	}
}

func roadOptions() *ToolbarItem {
	for i := range ToolbarItems {
		if ToolbarItems[i].Tool == ToolRoad {
			return &ToolbarItems[i]
		}
	}
	return nil
}

func parkingOptions() *ToolbarItem {
	for i := range ToolbarItems {
		if ToolbarItems[i].Tool == ToolParking {
			return &ToolbarItems[i]
		}
	}
	return nil
}

func transportOptions() *ToolbarItem {
	for i := range ToolbarItems {
		if ToolbarItems[i].Tool == ToolTransport {
			return &ToolbarItems[i]
		}
	}
	return nil
}

func zoneOptions() *ToolbarItem {
	for i := range ToolbarItems {
		if ToolbarItems[i].Tool == ToolZone {
			return &ToolbarItems[i]
		}
	}
	return nil
}

func (tb *MainToolbar) ChromeTopY(ts *ToolSystem, menus *BuildMenus) int32 {
	y := ToolbarY
	if menus != nil && menus.Visible() {
		y -= BuildMenuH
	}
	if ts.HasOptionsBar() {
		y -= OptionsBarH
	}
	return int32(y)
}

func categoryUsesLegacyOptions(cat ToolbarCategory) bool {
	switch cat {
	case CatRoads, CatZoning, CatPublicTransport:
		return true
	default:
		return false
	}
}

func (ts *ToolSystem) ApplyCategory(cat ToolbarCategory) {
	switch cat {
	case CatRoads:
		ts.Activate(ToolRoad)
	case CatZoning:
		ts.Activate(ToolZone)
	case CatPublicTransport:
		ts.Activate(ToolTransport)
	case CatStatistics:
		ts.Activate(ToolPointer)
	default:
		ts.Activate(ToolPointer)
	}
}

// UsesBuildMenu reports whether the category shows the asset browser instead of legacy options.
func UsesBuildMenu(cat ToolbarCategory) bool {
	return !categoryUsesLegacyOptions(cat) && cat != CatStatistics && cat != CatOptions
}
