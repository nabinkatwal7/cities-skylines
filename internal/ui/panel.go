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

const (
	ScreenW     int32 = 1280
	ScreenH     int32 = 720
	ToolbarH    int32 = 82
	ToolbarY    int32 = ScreenH - ToolbarH
	ToolbarBtnW int32 = 72
	ToolbarBtnH int32 = 74
	ToolbarPad  int32 = 4
	UtilBtnW    int32 = 64
	OptionsBarH int32 = 52
	BuildMenuH  int32 = 156
	TopBarH     int32 = 48
)

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
