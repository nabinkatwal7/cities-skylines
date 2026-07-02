package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Panel is a drawable UI subsystem. Simulation logic never lives here.
type Panel interface {
	Draw()
}

// InputPanel optionally handles player input before it reaches the world.
type InputPanel interface {
	HandleInput() bool
	HandleClick(mx, my int32) bool
}

var (
	ScreenW  int32 = 1280
	ScreenH  int32 = 720
	ToolbarY int32 = ScreenH - ToolbarH
)

const (
	ToolbarH    int32 = 96
	ToolbarBtnW int32 = 54
	ToolbarBtnH int32 = 86
	ToolbarPad  int32 = 4
	UtilBtnW    int32 = 54
	UtilBtnCnt  int32 = 5
	OptionsBarH int32 = 48
	BuildMenuH  int32 = 148
	TopBarH     int32 = 56
	OptBtnW     int32 = 92
)

// SetScreenSize keeps layout in sync with the game window.
func SetScreenSize(w, h int32) {
	if w < 640 {
		w = 640
	}
	if h < 480 {
		h = 480
	}
	ScreenW, ScreenH = w, h
	ToolbarY = ScreenH - ToolbarH
}

func UtilStripWidth() int32 {
	return 8 + UtilBtnCnt*(UtilBtnW+ToolbarPad) + 10
}

func CategoryStripX0() int32 { return UtilStripWidth() }

func CategoryStripW() int32 {
	w := ScreenW - CategoryStripX0() - 8
	if w < 64 {
		return 64
	}
	return w
}

func OptionsStripW() int32 { return ScreenW - 16 }

// ChromeTopY is the Y of the lowest stacked chrome row above the main toolbar.
func ChromeTopY(hasOptionsBar, hasBuildMenu bool) int32 {
	y := ToolbarY
	if hasBuildMenu {
		y -= BuildMenuH
	}
	if hasOptionsBar {
		y -= OptionsBarH
	}
	return y
}

// ViewState carries read-only presentation data from the simulation layer.
type ViewState struct {
	Money         float32
	WeeklyIncome  float32
	Population    int
	Happiness     float32 // 0..1
	DateStr       string
	TimeStr       string
	SpeedStr      string
	Milestone     string
	MouseWorldX   float32
	MouseWorldZ   float32
	MouseOnGround bool
}

func uiBtn(x, y, w, h int32, label string, col, textCol rl.Color, selected bool) {
	csOptionBtn(x, y, w, h, label, col, selected)
}
