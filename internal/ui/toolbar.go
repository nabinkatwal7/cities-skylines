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
	{CatElectricity, "Electricity", rl.NewColor(235, 200, 75, 255), 0},
	{CatWater, "Water & Sewage", rl.NewColor(70, 145, 220, 255), 0},
	{CatGarbage, "Garbage", rl.NewColor(110, 95, 75, 255), 0},
	{CatHealthcare, "Healthcare", rl.NewColor(235, 105, 105, 255), 0},
	{CatFireRescue, "Fire Dept", rl.NewColor(235, 95, 55, 255), 0},
	{CatPolice, "Police", rl.NewColor(75, 115, 210, 255), 0},
	{CatEducation, "Education", rl.NewColor(95, 165, 235, 255), 0},
	{CatPublicTransport, "Transport", rl.NewColor(55, 140, 185, 255), rl.KeyFour},
	{CatLandscaping, "Landscaping", rl.NewColor(95, 150, 80, 255), 0},
	{CatParks, "Parks", rl.NewColor(75, 185, 95, 255), 0},
	{CatEconomy, "Economy", rl.NewColor(190, 170, 55, 255), 0},
	{CatPolicies, "Policies", rl.NewColor(170, 110, 190, 255), 0},
	{CatStatistics, "Statistics", rl.NewColor(140, 150, 190, 255), 0},
	{CatOptions, "Options", rl.NewColor(120, 125, 130, 255), 0},
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

func (tb *MainToolbar) utilStripX() int32 { return 8 }

func (tb *MainToolbar) categoryStripX(visible int, btnW int32) int32 {
	totalW := int32(visible)*int32(btnW+ToolbarPad) - ToolbarPad
	return (ScreenW - totalW) / 2
}

func (tb *MainToolbar) HandleClick(ts *ToolSystem, menus *BuildMenus) GameTool {
	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)

	if my >= ToolbarY && my < ToolbarY+ToolbarH {
		utils := []GameTool{ToolPointer, ToolInspect, ToolMeasure, ToolRemove, ToolUpgrade}
		for i, tool := range utils {
			bx := tb.utilStripX() + int32(i)*(UtilBtnW+ToolbarPad)
			if mx >= bx && mx < bx+UtilBtnW {
				ts.Activate(tool)
				return ts.Selected
			}
		}
	}

	visible := tb.VisibleCategories()
	btnW := int32(ToolbarBtnW)
	startX := tb.categoryStripX(len(visible), btnW)
	for i, c := range visible {
		bx := startX + int32(i)*(btnW+ToolbarPad)
		by := ToolbarY + 2
		if mx >= bx && mx < bx+btnW && my >= by && my < by+ToolbarBtnH {
			tb.Select(c.Cat, ts, menus)
			return ts.Selected
		}
	}
	return ts.Selected
}

func (tb *MainToolbar) Draw(ts *ToolSystem) {
	drawBarBottom(ToolbarY, ToolbarH)

	utils := []struct {
		tool  GameTool
		col   rl.Color
	}{
		{ToolPointer, rl.NewColor(140, 145, 150, 255)},
		{ToolInspect, rl.NewColor(100, 155, 200, 255)},
		{ToolMeasure, rl.NewColor(170, 150, 100, 255)},
		{ToolRemove, rl.NewColor(190, 75, 75, 255)},
		{ToolUpgrade, rl.NewColor(190, 185, 75, 255)},
	}
	for i, u := range utils {
		bx := tb.utilStripX() + int32(i)*(UtilBtnW+ToolbarPad)
		by := ToolbarY + 2
		csToolBtn(bx, by, UtilBtnW, u.tool, u.col, ts.Selected == u.tool)
	}

	visible := tb.VisibleCategories()
	btnW := int32(ToolbarBtnW)
	startX := tb.categoryStripX(len(visible), btnW)
	for i, c := range visible {
		bx := startX + int32(i)*(btnW+ToolbarPad)
		by := ToolbarY + 2
		sel := (tb.Selected == c.Cat && ts.Mode == ModePlace) || (ts.Mode == ModePaint && c.Cat == CatZoning)
		csCategoryBtn(bx, by, btnW, c.Cat, c, sel)
	}
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
		drawOptionRow(item, y, -1, -1, nil)
	case ToolParking:
		item := parkingOptions()
		drawOptionRow(item, y, -1, -1, nil)
	case ToolTransport:
		item := transportOptions()
		drawOptionRow(item, y, -1, -1, nil)
	case ToolZone:
		item := zoneOptions()
		drawOptionRowColored(item, y, -1, -1, func(oi int) rl.Color {
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

func drawOptionRow(item *ToolbarItem, y int32, mx, my int32, onSelect func(int)) {
	drawOptionRowColored(item, y, mx, my, func(int) rl.Color {
		return csBtnIdle
	}, onSelect)
}

func drawOptionRowColored(item *ToolbarItem, y int32, mx, my int32, colFn func(int) rl.Color, onSelect func(int)) {
	optW := int32(100)
	total := int32(len(item.Options)) * int32(optW+ToolbarPad)
	sx := (ScreenW - total) / 2
	for oi, opt := range item.Options {
		bx := sx + int32(oi)*int32(optW+ToolbarPad)
		by := y + 6
		sel := oi == item.OptIndex
		csOptionBtn(bx, by, optW, OptionsBarH-12, opt, colFn(oi), sel)
		if onSelect != nil && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-12 {
			onSelect(oi)
		}
	}
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
		drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			ts.RoadType = road.RoadType(oi)
			handled = true
		})
	case ToolParking:
		item := parkingOptions()
		drawOptionRow(item, y, mx, my, func(oi int) {
			item.OptIndex = oi
			ts.setParkingMode(oi)
			handled = true
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
			handled = true
		})
	case ToolZone:
		item := zoneOptions()
		drawOptionRowColored(item, y, mx, my, func(oi int) rl.Color {
			if oi < 6 {
				c := zoning.ZoneColor(zoning.ZoneType(oi + 1))
				c.A = 220
				return c
			}
			return csBtnIdle
		}, func(oi int) {
			item.OptIndex = oi
			ts.ZoneType = oi
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
