package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Cities: Skylines inspired palette.
var (
	csBarTop      = rl.NewColor(22, 32, 42, 235)
	csBarBottom   = rl.NewColor(16, 22, 30, 245)
	csBarLine     = rl.NewColor(72, 152, 198, 180)
	csPanelBg     = rl.NewColor(24, 34, 44, 240)
	csPanelBorder = rl.NewColor(90, 140, 175, 200)
	csSelectGlow  = rl.NewColor(80, 190, 240, 255)
	csSelectFill  = rl.NewColor(35, 75, 100, 220)
	csText        = rl.NewColor(236, 240, 241, 255)
	csTextDim     = rl.NewColor(170, 185, 195, 255)
	csTextShadow  = rl.NewColor(0, 0, 0, 160)
	csMoney       = rl.NewColor(88, 214, 141, 255)
	csMoneyNeg    = rl.NewColor(235, 120, 110, 255)
	csPop         = rl.NewColor(130, 200, 255, 255)
	csHappy       = rl.NewColor(190, 220, 120, 255)
	csHappyLow    = rl.NewColor(235, 160, 100, 255)
	csInputBg     = rl.NewColor(12, 18, 24, 230)
	csBtnIdle     = rl.NewColor(38, 52, 64, 240)
	csBtnHover    = rl.NewColor(48, 68, 84, 250)
)

func drawBarTop() {
	rl.DrawRectangle(0, 0, ScreenW, TopBarH, csBarTop)
	rl.DrawRectangle(0, TopBarH-2, ScreenW, 2, csBarLine)
}

func drawBarBottom(y int32, h int32) {
	rl.DrawRectangle(0, y, ScreenW, h, csBarBottom)
	rl.DrawRectangle(0, y, ScreenW, 2, csBarLine)
}

func drawPanel(x, y, w, h int32) {
	rl.DrawRectangle(x, y, w, h, csPanelBg)
	rl.DrawRectangleLines(x, y, w, h, csPanelBorder)
	rl.DrawRectangle(x, y, w, 3, csBarLine)
}

func drawLabel(text string, x, y, size int32, col rl.Color) {
	DrawUIText(text, x+1, y+1, size, csTextShadow)
	DrawUIText(text, x, y, size, col)
}

func drawLabelScaled(text string, x, y, size int32, col rl.Color, scale float32) {
	DrawUITextScaled(text, x+1, y+1, size, csTextShadow, scale)
	DrawUITextScaled(text, x, y, size, col, scale)
}

func categoryIcon(cat ToolbarCategory) string {
	switch cat {
	case CatRoads:
		return "RD"
	case CatZoning:
		return "ZN"
	case CatDistricts:
		return "DT"
	case CatElectricity:
		return "EL"
	case CatWater:
		return "H2O"
	case CatGarbage:
		return "GB"
	case CatHealthcare:
		return "HP"
	case CatFireRescue:
		return "FR"
	case CatPolice:
		return "PD"
	case CatEducation:
		return "ED"
	case CatPublicTransport:
		return "PT"
	case CatLandscaping:
		return "LS"
	case CatParks:
		return "PK"
	case CatEconomy:
		return "EC"
	case CatPolicies:
		return "PO"
	case CatStatistics:
		return "ST"
	case CatOptions:
		return "OP"
	default:
		return "?"
	}
}

func toolIcon(tool GameTool) string {
	switch tool {
	case ToolPointer:
		return "Cu"
	case ToolInspect:
		return "In"
	case ToolMeasure:
		return "Ms"
	case ToolRemove:
		return "Bd"
	case ToolUpgrade:
		return "Up"
	default:
		return "·"
	}
}

func toolLabel(tool GameTool) string {
	switch tool {
	case ToolPointer:
		return "Cursor"
	case ToolInspect:
		return "Inspect"
	case ToolMeasure:
		return "Measure"
	case ToolRemove:
		return "Bulldoze"
	case ToolUpgrade:
		return "Upgrade"
	default:
		return "Tool"
	}
}

// csToolBtn draws a CS-style bottom-toolbar utility button.
func csToolBtn(x, y, w int32, tool GameTool, accent rl.Color, selected bool) {
	iconSz := int32(44)
	ix := x + (w-iconSz)/2
	iy := y + 2
	drawCategoryIcon(ix, iy, iconSz, toolIcon(tool), accent, selected)
	tw := MeasureUIText(toolLabel(tool), FontSm)
	drawLabel(toolLabel(tool), x+(w-tw)/2, y+iconSz+6, FontSm, csTextDim)
}

// csCategoryBtn draws icon + label like the CS build menu strip.
func csCategoryBtn(x, y, w int32, cat ToolbarCategory, def CategoryDef, selected bool) {
	iconSz := int32(44)
	ix := x + (w-iconSz)/2
	iy := y + 2
	drawCategoryIcon(ix, iy, iconSz, categoryIcon(cat), def.Color, selected)
	label := def.Label
	if len(label) > 10 {
		label = label[:9] + "…"
	}
	tw := MeasureUIText(label, FontSm)
	drawLabel(label, x+(w-tw)/2, y+iconSz+6, FontSm, csText)
}

func drawCategoryIcon(x, y, size int32, glyph string, accent rl.Color, selected bool) {
	if selected {
		rl.DrawRectangle(x-3, y-3, size+6, size+6, csSelectGlow)
	}
	rl.DrawRectangle(x, y, size, size, accent)
	rl.DrawRectangleLines(x, y, size, size, rl.NewColor(255, 255, 255, 80))
	if selected {
		rl.DrawRectangle(x, y, size, size, csSelectFill)
	}
	tw := MeasureUIText(glyph, FontMd)
	drawLabel(glyph, x+(size-tw)/2, y+(size-FontMd)/2, FontMd, csText)
}

func csOptionBtn(x, y, w, h int32, label string, fill rl.Color, selected bool) {
	if selected {
		rl.DrawRectangle(x-2, y-2, w+4, h+4, csSelectGlow)
	}
	rl.DrawRectangle(x, y, w, h, fill)
	if selected {
		rl.DrawRectangle(x, y, w, h, csSelectFill)
	}
	rl.DrawRectangleLines(x, y, w, h, csPanelBorder)
	tw := MeasureUIText(label, FontMd)
	drawLabel(label, x+(w-tw)/2, y+(h-FontMd)/2, FontMd, csText)
}

func csAssetBtn(x, y, w, h int32, name string, fav bool, selected bool) {
	fill := csBtnIdle
	if fav {
		fill = rl.NewColor(55, 50, 35, 240)
	}
	if selected {
		rl.DrawRectangle(x-2, y-2, w+4, h+4, csSelectGlow)
	}
	rl.DrawRectangle(x, y, w, h, fill)
	if selected {
		rl.DrawRectangle(x, y, w, h, csSelectFill)
	}
	rl.DrawRectangleLines(x, y, w, h, csPanelBorder)
	short := name
	if len(short) > 11 {
		short = short[:10] + "…"
	}
	tw := MeasureUIText(short, FontMd)
	drawLabel(short, x+(w-tw)/2, y+(h-FontMd)/2, FontMd, csText)
}

func csInputField(x, y, w, h int32) {
	rl.DrawRectangle(x, y, w, h, csInputBg)
	rl.DrawRectangleLines(x, y, w, h, csPanelBorder)
}
