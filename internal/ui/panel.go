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
	ScreenW     = 1280
	ScreenH     = 720
	ToolbarY    = 660
	ToolbarH    = 60
	ToolbarBtnW = 72
	ToolbarBtnH = 48
	ToolbarPad  = 4
	OptionsBarH = 40
	BuildMenuH  = 128
	TopBarH     = 52
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
	if selected {
		rl.DrawRectangle(x-2, y-2, w+4, h+4, rl.NewColor(255, 255, 200, 200))
	}
	rl.DrawRectangle(x, y, w, h, col)
	rl.DrawRectangleLines(x, y, w, h, rl.NewColor(60, 60, 60, 200))
	tw := MeasureUIText(label, 14)
	DrawUIText(label, x+(w-tw)/2, y+(h-14)/2, 14, textCol)
}
