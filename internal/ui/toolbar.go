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
	{CatRoads, "Roads", rl.NewColor(145, 125, 95, 255), rl.KeyTwo},
	{CatZoning, "Zoning", rl.NewColor(95, 175, 110, 255), rl.KeyThree},
	{CatDistricts, "Districts", rl.NewColor(130, 115, 175, 255), 0},
	{CatElectricity, "Power", rl.NewColor(235, 200, 75, 255), 0},
	{CatWater, "Water", rl.NewColor(70, 145, 220, 255), 0},
	{CatGarbage, "Garbage", rl.NewColor(110, 95, 75, 255), 0},
	{CatHealthcare, "Health", rl.NewColor(235, 105, 105, 255), 0},
	{CatFireRescue, "Fire", rl.NewColor(235, 95, 55, 255), 0},
	{CatPolice, "Police", rl.NewColor(75, 115, 210, 255), 0},
	{CatEducation, "Education", rl.NewColor(95, 165, 235, 255), 0},
	{CatPublicTransport, "Transport", rl.NewColor(55, 140, 185, 255), rl.KeyFour},
	{CatLandscaping, "Landscape", rl.NewColor(95, 150, 80, 255), 0},
	{CatParks, "Parks", rl.NewColor(75, 185, 95, 255), 0},
	{CatEconomy, "Economy", rl.NewColor(190, 170, 55, 255), 0},
	{CatPolicies, "Policies", rl.NewColor(170, 110, 190, 255), 0},
	{CatStatistics, "Stats", rl.NewColor(140, 150, 190, 255), 0},
	{CatOptions, "Options", rl.NewColor(120, 125, 130, 255), 0},
}

// MainToolbar is the primary build-category selector (24.3).
type MainToolbar struct {
	Selected  ToolbarCategory
	catScroll int32
	optScroll int32
	Unlocks   *UnlockRegistry
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
	tb.optScroll = 0
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

func (tb *MainToolbar) utilStripX() int32 { return 8 }

func (tb *MainToolbar) toolbarBtnY() int32 { return ToolbarY + 4 }

func (tb *MainToolbar) categoryTotalW(count int) int32 {
	if count <= 0 {
		return 0
	}
	return int32(count)*int32(ToolbarBtnW+ToolbarPad) - ToolbarPad
}

func (tb *MainToolbar) clampCatScroll() {
	visible := tb.VisibleCategories()
	max := tb.categoryTotalW(len(visible)) - CategoryStripW()
	if max < 0 {
		max = 0
	}
	if tb.catScroll < 0 {
		tb.catScroll = 0
	}
	if tb.catScroll > max {
		tb.catScroll = max
	}
}

func (tb *MainToolbar) clampOptScroll(count int) {
	total := int32(count)*int32(OptBtnW+ToolbarPad) - ToolbarPad
	max := total - OptionsStripW()
	if max < 0 {
		max = 0
	}
	if tb.optScroll < 0 {
		tb.optScroll = 0
	}
	if tb.optScroll > max {
		tb.optScroll = max
	}
}

func (tb *MainToolbar) ensureOptVisible(idx int, count int) {
	if idx < 0 || idx >= count {
		return
	}
	bx := int32(idx) * int32(OptBtnW+ToolbarPad)
	areaW := OptionsStripW()
	if bx < tb.optScroll {
		tb.optScroll = bx
	}
	if bx+OptBtnW > tb.optScroll+areaW {
		tb.optScroll = bx + OptBtnW - areaW
	}
	tb.clampOptScroll(count)
}

func (tb *MainToolbar) HandleWheel(ts *ToolSystem, menus *BuildMenus, delta float32) {
	if delta == 0 {
		return
	}
	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)
	step := int32(delta * 56)
	if step == 0 {
		if delta > 0 {
			step = 56
		} else {
			step = -56
		}
	}

	if menus != nil && menus.Visible() {
		y := menus.chromeY(ts)
		if my >= y && my < y+BuildMenuH {
			return
		}
	}

	if ts.HasOptionsBar() {
		y := tb.optionsY()
		if my >= y && my < ToolbarY && mx >= 0 && mx < ScreenW {
			tb.optScroll -= step
			tb.clampOptScroll(tb.activeOptionCount(ts))
			return
		}
	}
	if my >= ToolbarY && my < ToolbarY+ToolbarH {
		tb.catScroll -= step
		tb.clampCatScroll()
	}
}

func (tb *MainToolbar) activeOptionCount(ts *ToolSystem) int {
	switch ts.Selected {
	case ToolRoad:
		return len(roadOptions().Options)
	case ToolParking:
		return len(parkingOptions().Options)
	case ToolTransport:
		return len(transportOptions().Options)
	case ToolZone:
		return len(zoneOptions().Options)
	default:
		return 0
	}
}

func (tb *MainToolbar) HandleClick(ts *ToolSystem, menus *BuildMenus) GameTool {
	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)

	if my >= ToolbarY && my < ToolbarY+ToolbarH {
		utils := []GameTool{ToolPointer, ToolInspect, ToolMeasure, ToolRemove, ToolUpgrade}
		by := tb.toolbarBtnY()
		for i, tool := range utils {
			bx := tb.utilStripX() + int32(i)*(UtilBtnW+ToolbarPad)
			if mx >= bx && mx < bx+UtilBtnW && my >= by && my < by+ToolbarBtnH {
				ts.Activate(tool)
				return ts.Selected
			}
		}
		if idx, ok := tb.categoryHit(mx, my); ok {
			visible := tb.VisibleCategories()
			tb.Select(visible[idx].Cat, ts, menus)
			return ts.Selected
		}
	}
	return ts.Selected
}

func (tb *MainToolbar) categoryHit(mx, my int32) (int, bool) {
	if my < tb.toolbarBtnY() || my >= tb.toolbarBtnY()+ToolbarBtnH {
		return -1, false
	}
	x0 := CategoryStripX0()
	areaW := CategoryStripW()
	if mx < x0 || mx >= x0+areaW {
		return -1, false
	}
	relX := mx - x0 + tb.catScroll
	idx := int(relX / int32(ToolbarBtnW+ToolbarPad))
	visible := tb.VisibleCategories()
	if idx < 0 || idx >= len(visible) {
		return -1, false
	}
	cellX := int32(idx) * int32(ToolbarBtnW+ToolbarPad)
	if relX < cellX || relX >= cellX+ToolbarBtnW {
		return -1, false
	}
	return idx, true
}

func (tb *MainToolbar) Draw(ts *ToolSystem) {
	drawBarBottom(ToolbarY, ToolbarH)

	utils := []struct {
		tool GameTool
		col  rl.Color
	}{
		{ToolPointer, rl.NewColor(140, 145, 150, 255)},
		{ToolInspect, rl.NewColor(100, 155, 200, 255)},
		{ToolMeasure, rl.NewColor(170, 150, 100, 255)},
		{ToolRemove, rl.NewColor(190, 75, 75, 255)},
		{ToolUpgrade, rl.NewColor(190, 185, 75, 255)},
	}
	by := tb.toolbarBtnY()
	for i, u := range utils {
		bx := tb.utilStripX() + int32(i)*(UtilBtnW+ToolbarPad)
		csToolBtn(bx, by, UtilBtnW, ToolbarBtnH, u.tool, u.col, ts.Selected == u.tool)
	}

	sepX := CategoryStripX0() - 6
	rl.DrawRectangle(sepX, ToolbarY+10, 2, ToolbarH-20, csBarLine)

	x0 := CategoryStripX0()
	areaW := CategoryStripW()
	btnW := ToolbarBtnW
	visible := tb.VisibleCategories()
	tb.clampCatScroll()

	rl.BeginScissorMode(x0, ToolbarY, areaW, ToolbarH)
	for i, c := range visible {
		bx := x0 + int32(i)*(btnW+ToolbarPad) - tb.catScroll
		sel := (tb.Selected == c.Cat && ts.Mode == ModePlace) || (ts.Mode == ModePaint && c.Cat == CatZoning)
		csCategoryBtn(bx, by, btnW, ToolbarBtnH, c.Cat, c, sel)
	}
	rl.EndScissorMode()

	total := tb.categoryTotalW(len(visible))
	drawScrollFade(x0, ToolbarY, areaW, ToolbarH, tb.catScroll > 0, total > areaW+tb.catScroll)
}

func (tb *MainToolbar) drawLegacyOptions(ts *ToolSystem) {
	if ts.Selected == ToolPointer || ts.Selected == ToolRemove || ts.Selected == ToolUpgrade {
		return
	}
	y := tb.optionsY()
	drawBarBottom(y, OptionsBarH)

	switch ts.Selected {
	case ToolRoad:
		item := roadOptions()
		tb.ensureOptVisible(item.OptIndex, len(item.Options))
		tb.drawOptionRow(item, y, -1, -1, nil)
	case ToolParking:
		item := parkingOptions()
		tb.ensureOptVisible(item.OptIndex, len(item.Options))
		tb.drawOptionRow(item, y, -1, -1, nil)
	case ToolTransport:
		item := transportOptions()
		tb.ensureOptVisible(item.OptIndex, len(item.Options))
		tb.drawOptionRow(item, y, -1, -1, nil)
	case ToolZone:
		item := zoneOptions()
		tb.ensureOptVisible(item.OptIndex, len(item.Options))
		tb.drawOptionRowColored(item, y, -1, -1, func(oi int) rl.Color {
			if oi < 6 {
				c := zoning.ZoneColor(zoning.ZoneType(oi + 1))
				c.A = 220
				return c
			}
			return csBtnIdle
		}, nil)
	}
}

func (tb *MainToolbar) optionsY() int32 {
	return ToolbarY - OptionsBarH
}

func (tb *MainToolbar) DrawOptions(ts *ToolSystem) {
	tb.drawLegacyOptions(ts)
}

func (tb *MainToolbar) drawOptionRow(item *ToolbarItem, y int32, mx, my int32, onSelect func(int)) {
	tb.drawOptionRowColored(item, y, mx, my, func(int) rl.Color {
		return csBtnIdle
	}, onSelect)
}

func (tb *MainToolbar) drawOptionRowColored(item *ToolbarItem, y int32, mx, my int32, colFn func(int) rl.Color, onSelect func(int)) {
	if item == nil {
		return
	}
	tb.clampOptScroll(len(item.Options))
	areaX := int32(8)
	areaW := OptionsStripW()
	optH := OptionsBarH - 12
	by := y + 6

	rl.BeginScissorMode(areaX, y, areaW, OptionsBarH)
	for oi, opt := range item.Options {
		bx := areaX + int32(oi)*int32(OptBtnW+ToolbarPad) - tb.optScroll
		sel := oi == item.OptIndex
		csOptionBtn(bx, by, OptBtnW, optH, opt, colFn(oi), sel)
		if onSelect != nil && mx >= bx && mx < bx+OptBtnW && my >= by && my < by+optH {
			onSelect(oi)
		}
	}
	rl.EndScissorMode()

	total := int32(len(item.Options))*int32(OptBtnW+ToolbarPad) - ToolbarPad
	drawScrollFade(areaX, y, areaW, OptionsBarH, tb.optScroll > 0, total > areaW+tb.optScroll)
}

func (tb *MainToolbar) HandleOptionsClick(ts *ToolSystem, mx, my int32) bool {
	if !ts.HasOptionsBar() {
		return false
	}
	y := tb.optionsY()
	if my < y || my >= ToolbarY {
		return false
	}
	handled := false
	switch ts.Selected {
	case ToolRoad:
		item := roadOptions()
		tb.drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			ts.RoadType = road.RoadTypeFromOptionIndex(oi)
			tb.ensureOptVisible(oi, len(item.Options))
			handled = true
		})
	case ToolParking:
		item := parkingOptions()
		tb.drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			ts.setParkingMode(oi)
			tb.ensureOptVisible(oi, len(item.Options))
			handled = true
		})
	case ToolTransport:
		item := transportOptions()
		tb.drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			if oi >= int(transport.TransportModeCount) {
				ts.CargoMode = true
			} else {
				ts.CargoMode = false
				ts.TransportType = transport.TransportType(oi)
			}
			tb.ensureOptVisible(oi, len(item.Options))
			handled = true
		})
	case ToolZone:
		item := zoneOptions()
		tb.drawOptionRowColored(item, y, mx, my, func(oi int) rl.Color {
			if oi < 6 {
				c := zoning.ZoneColor(zoning.ZoneType(oi + 1))
				c.A = 220
				return c
			}
			return csBtnIdle
		}, func(oi int) {
			item.OptIndex = oi
			ts.ZoneType = oi
			tb.ensureOptVisible(oi, len(item.Options))
			handled = true
		})
	}
	return handled
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
	hasMenu := menus != nil && menus.Visible()
	return ChromeTopY(ts.HasOptionsBar(), hasMenu)
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
