package terrain

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameTool int

const (
	ToolPointer GameTool = iota
	ToolRoad
	ToolZone
	ToolPark
	ToolParking
	ToolTransport
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
	Selected       GameTool
	ZoneType       ZoneType
	RoadType       RoadType
	ParkMode       bool
	ParkingGarage  bool
	BusDepotMode   bool
	TramDepotMode  bool
	MetroDepotMode bool
	CargoMode      bool
	TransportType  TransportType
	Money          float32
	Population    int32
	ResDemand     int
	ComDemand     int
	IndDemand     int
	TimeStr       string
	HelpText      string
	MouseWorldX   float32
	MouseWorldZ   float32
	MouseOnGround bool
	BuildingInfo  string
}

func NewGameUI() *GameUI {
	return &GameUI{
		Selected: ToolPointer,
		ZoneType: ZoneResidentialLow,
		RoadType: RoadTwoLane,
	}
}

var ToolbarItems = []ToolbarItem{
	{ToolPointer, "Pointer", rl.KeyOne, rl.NewColor(200, 200, 200, 255), nil, 0},
	{ToolRoad, "Roads", rl.KeyTwo, rl.NewColor(180, 160, 120, 255), []string{"2-Lane", "1-Way", "4-Lane", "Gravel", "Highway", "6-Lane", "Avenue", "Bus Rd", "Tram Rd", "Bike Rd", "Tree Rd", "Asym Rd", "Pedestrian", "Quay"}, 0},
	{ToolZone, "Zones", rl.KeyThree, rl.NewColor(100, 200, 100, 255), []string{"Res Low", "Res High", "Com Low", "Com High", "Industrial", "Office"}, 0},
	{ToolPark, "Parks", rl.KeyFour, rl.NewColor(80, 200, 80, 255), nil, 0},
	{ToolParking, "Parking", rl.KeyFive, rl.NewColor(100, 100, 200, 255), []string{"Lot", "Garage", "Bus Depot", "Tram Depot", "Metro Depot"}, 0},
	{ToolTransport, "Transport", rl.KeySix, rl.NewColor(50, 150, 200, 255), []string{"Bus", "Tram", "Metro", "Train", "Ferry", "Monorail", "Cable Car", "Taxi", "Air", "Ship", "Walk", "Bicycle", "Car", "Blimp", "Cargo Stn"}, 0},
	{ToolRemove, "Remove", rl.KeySeven, rl.NewColor(200, 80, 80, 255), nil, 0},
	{ToolUpgrade, "Upgrade", rl.KeyEight, rl.NewColor(200, 200, 80, 255), nil, 0},
}

const (
	ToolbarY     = 660
	ToolbarH     = 60
	ToolbarBtnW  = 90
	ToolbarBtnH  = 48
	ToolbarPad   = 6
	OptionsBarH  = 40
	TopBarH      = 40
)

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
		ui.RoadType = RoadType(item.OptIndex)
	}
	if ui.Selected == ToolZone && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[2]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ui.ZoneType = ZoneType(item.OptIndex + 1)
	}
	if ui.Selected == ToolParking && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[4]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ui.ParkingGarage = item.OptIndex == 1
		ui.BusDepotMode = item.OptIndex == 2
		ui.TramDepotMode = item.OptIndex == 3
		ui.MetroDepotMode = item.OptIndex == 4
	}
	if ui.Selected == ToolTransport && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[5]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		if item.OptIndex >= int(TransportModeCount) {
			ui.CargoMode = true
		} else {
			ui.CargoMode = false
			ui.TransportType = TransportType(item.OptIndex)
		}
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
	tw := int32(rl.MeasureText(label, 14))
	rl.DrawText(label, x+(w-tw)/2, y+(h-14)/2, 14, textCol)
}

func (ui *GameUI) Draw() {
	ui.drawTopBar()
	ui.drawToolbar()
	ui.drawOptions()
	ui.drawHelpText()
}

func (ui *GameUI) drawTopBar() {
	rl.DrawRectangle(0, 0, 1280, TopBarH, rl.NewColor(0, 0, 0, 180))

	moneyCol := rl.NewColor(100, 220, 100, 220)
	rl.DrawText(fmt.Sprintf("$%.0f", ui.Money), 10, 10, 20, moneyCol)

	popStr := fmt.Sprintf("Pop: %d", ui.Population)
	rl.DrawText(popStr, 160, 10, 20, rl.White)

	rl.DrawText(ui.TimeStr, 350, 10, 16, rl.Gray)

	bx := int32(480)
	bw := int32(80)
	by := int32(8)
	bar := func(label string, val int, yOff int32, col rl.Color) {
		rl.DrawText(label, bx, by+yOff, 12, rl.Gray)
		w := int32(val * 8)
		if w < 0 {
			w = 0
		}
		if w > bw {
			w = bw
		}
		rl.DrawRectangle(bx+20, by+yOff, w, 10, col)
		rl.DrawRectangleLines(bx+20, by+yOff, bw, 10, rl.NewColor(80, 80, 80, 200))
	}
	bar("R", ui.ResDemand, 0, rl.NewColor(100, 200, 100, 220))
	bar("C", ui.ComDemand, 14, rl.NewColor(100, 150, 255, 220))
	bar("I", ui.IndDemand, 28, rl.NewColor(255, 200, 80, 220))

	if ui.MouseOnGround {
		coordStr := fmt.Sprintf("(%.1f, %.1f)", ui.MouseWorldX, ui.MouseWorldZ)
		rl.DrawText(coordStr, 1280-180, 10, 14, rl.Gray)
	}
	if ui.BuildingInfo != "" {
		rl.DrawRectangle(10, 50, 400, 24, rl.NewColor(0, 0, 0, 180))
		rl.DrawText(ui.BuildingInfo, 16, 53, 16, rl.White)
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
	if ui.Selected == ToolPointer || ui.Selected == ToolPark || ui.Selected == ToolRemove || ui.Selected == ToolUpgrade {
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
				ui.RoadType = RoadType(oi)
			}
		}
	case ToolZone:
		item := &ToolbarItems[2]
		optW := int32(70)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (1280 - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			col := ZoneColor(ZoneType(oi + 1))
			col.A = 200
			uiBtn(bx, by, optW, OptionsBarH-8, opt, col, rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				ui.ZoneType = ZoneType(oi + 1)
			}
		}
	case ToolParking:
		item := &ToolbarItems[4]
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
				ui.ParkingGarage = oi == 1
				ui.BusDepotMode = oi == 2
				ui.TramDepotMode = oi == 3
				ui.MetroDepotMode = oi == 4
			}
		}
	case ToolTransport:
		item := &ToolbarItems[5]
		optW := int32(75)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (1280 - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			col := rl.NewColor(200, 150, 50, 200)
			if oi < int(TransportModeCount) {
				col = TransportStopColor(TransportType(oi))
				col.A = 200
			}
			uiBtn(bx, by, optW, OptionsBarH-8, opt, col, rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				if oi >= int(TransportModeCount) {
					ui.CargoMode = true
				} else {
					ui.CargoMode = false
					ui.TransportType = TransportType(oi)
				}
			}
		}
	}
}

func (ui *GameUI) drawHelpText() {
	helpY := int32(ToolbarY + ToolbarH + 5)
	switch ui.Selected {
	case ToolPointer:
		rl.DrawText("Click on buildings for info | 1-6 to select tools | WASD=pan | R-drag=orbit | Scroll=zoom", 10, helpY, 14, rl.White)
	case ToolRoad:
		rl.DrawText(fmt.Sprintf("L-click to place | R=cycle type | PgUp/PgDn=elevation | Esc=deselect | Current: %s", ToolbarItems[1].Options[ToolbarItems[1].OptIndex]), 10, helpY, 14, rl.White)
	case ToolZone:
		rl.DrawText(fmt.Sprintf("L-click to paint zones | R=cycle type | Esc=deselect | Current: %s", ToolbarItems[2].Options[ToolbarItems[2].OptIndex]), 10, helpY, 14, rl.White)
	case ToolPark:
		rl.DrawText("L-click to place park ($500) | Esc=deselect", 10, helpY, 14, rl.White)
	case ToolParking:
		rl.DrawText(fmt.Sprintf("L-click to place %s | R=cycle type | Esc=deselect", ToolbarItems[4].Options[ToolbarItems[4].OptIndex]), 10, helpY, 14, rl.White)
		switch ToolbarItems[4].OptIndex {
		case 0:
			rl.DrawText("Surface Parking Lot ($1000)", 10, helpY+18, 14, rl.White)
		case 1:
			rl.DrawText("Parking Garage ($3000)", 10, helpY+18, 14, rl.White)
		case 2:
			rl.DrawText("Bus Depot ($5000) — spawns buses for bus lines", 10, helpY+18, 14, rl.White)
		case 3:
			rl.DrawText("Tram Depot ($5000) — spawns trams for tram lines", 10, helpY+18, 14, rl.White)
		case 4:
			rl.DrawText("Metro Depot ($5000) — spawns metro trains for metro lines", 10, helpY+18, 14, rl.White)
		}
	case ToolTransport:
		if ui.CargoMode {
			rl.DrawText("L-click to place Cargo Station | R=cycle type | Esc=deselect", 10, helpY, 14, rl.White)
		} else {
			rl.DrawText(fmt.Sprintf("L-click to place %s stop | R=cycle type | Esc=deselect | Current: %s", ToolbarItems[5].Options[ToolbarItems[5].OptIndex], ToolbarItems[5].Options[ToolbarItems[5].OptIndex]), 10, helpY, 14, rl.White)
		}
	case ToolRemove:
		rl.DrawText("L-click on stop to remove, near line to remove route, or road segment | Esc=deselect", 10, helpY, 14, rl.White)
	case ToolUpgrade:
		rl.DrawText(fmt.Sprintf("L-click on road to upgrade to %s | R=change target type | Esc=deselect", ToolbarItems[1].Options[ToolbarItems[1].OptIndex]), 10, helpY, 14, rl.White)
	}
}
