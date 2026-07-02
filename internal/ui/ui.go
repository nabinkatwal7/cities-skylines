package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameTool int

const (
	ToolPointer GameTool = iota
	ToolRoad
	ToolParking
	ToolTransport
	ToolZone
	ToolRemove
	ToolUpgrade
)

type ToolbarItem struct {
	Tool     GameTool
	Label    string
	Key      int32
	Color    rl.Color
	Options  []string
	OptIndex int
}

type GameUI struct {
	Selected          GameTool
	RoadType          road.RoadType
	ZoneType          int
	ParkingGarage     bool
	BusDepotMode      bool
	TramDepotMode     bool
	MetroDepotMode    bool
	FerryDepotMode    bool
	MonorailDepotMode bool
	CableCarDepotMode bool
	TaxiDepotMode     bool
	AirportMode       bool
	PortMode          bool
	CargoMode         bool
	TransportType     transport.TransportType
	Money             float32
	TimeStr           string
	HelpText          string
	MouseWorldX       float32
	MouseWorldZ       float32
	MouseOnGround     bool
}

func NewGameUI() *GameUI {
	return &GameUI{
		Selected: ToolPointer,
		RoadType: road.RoadTwoLane,
	}
}

var ToolbarItems = []ToolbarItem{
	{ToolPointer, "Pointer", rl.KeyOne, rl.NewColor(200, 200, 200, 255), nil, 0},
	{ToolRoad, "Roads", rl.KeyTwo, rl.NewColor(180, 160, 120, 255), []string{"2-Lane", "1-Way", "4-Lane", "Gravel", "Highway", "6-Lane", "Avenue", "Bus Rd", "Tram Rd", "Bike Rd", "Tree Rd", "Asym Rd", "Pedestrian", "Quay"}, 0},
	{ToolParking, "Parking", rl.KeyThree, rl.NewColor(100, 100, 200, 255), []string{"Lot", "Garage", "Bus Depot", "Tram Depot", "Metro Depot", "Ferry Depot", "Monorail Depot", "Cable Car Depot", "Taxi Depot", "Airport", "Port"}, 0},
	{ToolTransport, "Transport", rl.KeyFour, rl.NewColor(50, 150, 200, 255), []string{"Bus", "Tram", "Metro", "Train", "Ferry", "Monorail", "Cable Car", "Taxi", "Air", "Ship", "Walk", "Bicycle", "Car", "Blimp", "Cargo Stn"}, 0},
	{ToolZone, "Zones", rl.KeyFive, rl.NewColor(120, 200, 120, 255), []string{"Res Low", "Res High", "Com Low", "Com High", "Industrial", "Office"}, 0},
	{ToolRemove, "Remove", rl.KeySix, rl.NewColor(200, 80, 80, 255), nil, 0},
	{ToolUpgrade, "Upgrade", rl.KeySeven, rl.NewColor(200, 200, 80, 255), nil, 0},
}

const (
	ToolbarY    = 660
	ToolbarH    = 60
	ToolbarBtnW = 90
	ToolbarBtnH = 48
	ToolbarPad  = 6
	OptionsBarH = 40
	TopBarH     = 40
)

func (ui *GameUI) setParkingMode(oi int) {
	ui.ParkingGarage = oi == 1
	ui.BusDepotMode = oi == 2
	ui.TramDepotMode = oi == 3
	ui.MetroDepotMode = oi == 4
	ui.FerryDepotMode = oi == 5
	ui.MonorailDepotMode = oi == 6
	ui.CableCarDepotMode = oi == 7
	ui.TaxiDepotMode = oi == 8
	ui.AirportMode = oi == 9
	ui.PortMode = oi == 10
}

func (ui *GameUI) HasOptionsBar() bool {
	switch ui.Selected {
	case ToolRoad, ToolParking, ToolTransport, ToolZone:
		return true
	default:
		return false
	}
}

func (ui *GameUI) HandleInput() GameTool {
	for _, item := range ToolbarItems {
		if rl.IsKeyPressed(item.Key) {
			ui.Selected = item.Tool
			return item.Tool
		}
	}
	if ui.Selected == ToolRoad && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[1]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ui.RoadType = road.RoadType(item.OptIndex)
	}
	if ui.Selected == ToolParking && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[2]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ui.setParkingMode(item.OptIndex)
	}
	if ui.Selected == ToolTransport && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[3]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		if item.OptIndex >= int(transport.TransportModeCount) {
			ui.CargoMode = true
		} else {
			ui.CargoMode = false
			ui.TransportType = transport.TransportType(item.OptIndex)
		}
	}
	if ui.Selected == ToolZone && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[4]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ui.ZoneType = item.OptIndex
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.Selected = ToolPointer
	}
	return ui.Selected
}

func (ui *GameUI) handleToolbarClick() GameTool {
	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)
	totalW := len(ToolbarItems)*ToolbarBtnW + (len(ToolbarItems)-1)*ToolbarPad
	startX := (1280 - totalW) / 2
	for i, item := range ToolbarItems {
		bx := int32(startX + i*(ToolbarBtnW+ToolbarPad))
		by := int32(ToolbarY)
		if mx >= bx && mx < bx+ToolbarBtnW && my >= by && my < by+ToolbarBtnH {
			ui.Selected = item.Tool
			return item.Tool
		}
	}
	return ui.Selected
}

func (ui *GameUI) HandleClick() GameTool {
	return ui.handleToolbarClick()
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

func (ui *GameUI) Draw() {
	ui.drawTopBar()
	ui.drawToolbar()
	ui.drawOptions()
	ui.drawHelpText()
}

func (ui *GameUI) drawTopBar() {
	rl.DrawRectangle(0, 0, 1280, TopBarH, rl.NewColor(0, 0, 0, 180))
	DrawUIText(fmt.Sprintf("$%.0f", ui.Money), 10, 10, 20, rl.NewColor(100, 220, 100, 220))
	DrawUIText(ui.TimeStr, 160, 10, 16, rl.Gray)
	if ui.MouseOnGround {
		coordStr := fmt.Sprintf("(%.1f, %.1f)", ui.MouseWorldX, ui.MouseWorldZ)
		DrawUIText(coordStr, 1280-180, 10, 14, rl.Gray)
	}
}

func (ui *GameUI) drawToolbar() {
	rl.DrawRectangle(0, ToolbarY, 1280, ToolbarH, rl.NewColor(0, 0, 0, 200))
	totalW := len(ToolbarItems)*ToolbarBtnW + (len(ToolbarItems)-1)*ToolbarPad
	startX := (1280 - totalW) / 2
	for i, item := range ToolbarItems {
		bx := int32(startX + i*(ToolbarBtnW+ToolbarPad))
		by := int32(ToolbarY + (ToolbarH-ToolbarBtnH)/2)
		col := item.Color
		sel := ui.Selected == item.Tool
		textCol := rl.White
		if sel {
			textCol = rl.NewColor(255, 255, 200, 255)
		}
		uiBtn(bx, by, ToolbarBtnW, ToolbarBtnH, item.Label, col, textCol, sel)
	}
}

func (ui *GameUI) drawOptions() {
	if ui.Selected == ToolPointer || ui.Selected == ToolRemove || ui.Selected == ToolUpgrade {
		return
	}
	rl.DrawRectangle(0, ToolbarY-OptionsBarH, 1280, OptionsBarH, rl.NewColor(0, 0, 0, 160))

	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)

	switch ui.Selected {
	case ToolRoad:
		item := &ToolbarItems[1]
		optW := int32(80)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (1280 - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			uiBtn(bx, by, optW, OptionsBarH-8, opt, rl.NewColor(40, 40, 40, 200), rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				ui.RoadType = road.RoadType(oi)
			}
		}
	case ToolParking:
		item := &ToolbarItems[2]
		optW := int32(95)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (1280 - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			uiBtn(bx, by, optW, OptionsBarH-8, opt, rl.NewColor(40, 40, 40, 200), rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				ui.setParkingMode(oi)
			}
		}
	case ToolTransport:
		item := &ToolbarItems[3]
		optW := int32(75)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (1280 - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			col := rl.NewColor(200, 150, 50, 200)
			if oi < int(transport.TransportModeCount) {
				col = transport.TransportStopColor(transport.TransportType(oi))
				col.A = 200
			}
			uiBtn(bx, by, optW, OptionsBarH-8, opt, col, rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				if oi >= int(transport.TransportModeCount) {
					ui.CargoMode = true
				} else {
					ui.CargoMode = false
					ui.TransportType = transport.TransportType(oi)
				}
			}
		}
	case ToolZone:
		item := &ToolbarItems[4]
		optW := int32(80)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (1280 - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			col := rl.NewColor(40, 40, 40, 200)
			if oi < 6 {
				col = zoning.ZoneColor(zoning.ZoneType(oi+1))
				col.A = 200
			}
			uiBtn(bx, by, optW, OptionsBarH-8, opt, col, rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				ui.ZoneType = oi
			}
		}
	}
}

func (ui *GameUI) drawHelpText() {
	helpY := int32(ToolbarY + ToolbarH + 5)
	switch ui.Selected {
	case ToolPointer:
		DrawUIText("1-6 tools | F1-F3 speed | Space pause | WASD pan | R-drag orbit | Scroll zoom", 10, helpY, 14, rl.White)
	case ToolRoad:
		DrawUIText(fmt.Sprintf("L-click place | R cycle type | PgUp/PgDn elevation | Current: %s", ToolbarItems[1].Options[ToolbarItems[1].OptIndex]), 10, helpY, 14, rl.White)
	case ToolParking:
		DrawUIText(fmt.Sprintf("L-click place %s | R cycle type", ToolbarItems[2].Options[ToolbarItems[2].OptIndex]), 10, helpY, 14, rl.White)
	case ToolTransport:
		if ui.CargoMode {
			DrawUIText("L-click place Cargo Station | R cycle type", 10, helpY, 14, rl.White)
		} else {
			DrawUIText(fmt.Sprintf("L-click place %s stop | R cycle type", ToolbarItems[3].Options[ToolbarItems[3].OptIndex]), 10, helpY, 14, rl.White)
		}
	case ToolZone:
		DrawUIText(fmt.Sprintf("L-click paint %s | R cycle type", ToolbarItems[4].Options[ToolbarItems[4].OptIndex]), 10, helpY, 14, rl.White)
	case ToolRemove:
		DrawUIText("L-click stop, line, or road segment to remove | Esc deselect", 10, helpY, 14, rl.White)
	case ToolUpgrade:
		DrawUIText(fmt.Sprintf("L-click road to upgrade to %s | R change type", ToolbarItems[1].Options[ToolbarItems[1].OptIndex]), 10, helpY, 14, rl.White)
	}
}
