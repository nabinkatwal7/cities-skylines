package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/transport"

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

// ToolSystem tracks the active build tool and its mode options.
type ToolSystem struct {
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
	HelpText          string
}

func NewToolSystem() *ToolSystem {
	return &ToolSystem{
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

func (ts *ToolSystem) setParkingMode(oi int) {
	ts.ParkingGarage = oi == 1
	ts.BusDepotMode = oi == 2
	ts.TramDepotMode = oi == 3
	ts.MetroDepotMode = oi == 4
	ts.FerryDepotMode = oi == 5
	ts.MonorailDepotMode = oi == 6
	ts.CableCarDepotMode = oi == 7
	ts.TaxiDepotMode = oi == 8
	ts.AirportMode = oi == 9
	ts.PortMode = oi == 10
}

func (ts *ToolSystem) HasOptionsBar() bool {
	switch ts.Selected {
	case ToolRoad, ToolParking, ToolTransport, ToolZone:
		return true
	default:
		return false
	}
}

func (ts *ToolSystem) HandleKeyboard() GameTool {
	for _, item := range ToolbarItems {
		if rl.IsKeyPressed(item.Key) {
			ts.Selected = item.Tool
			return item.Tool
		}
	}
	if ts.Selected == ToolRoad && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[1]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ts.RoadType = road.RoadType(item.OptIndex)
	}
	if ts.Selected == ToolParking && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[2]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ts.setParkingMode(item.OptIndex)
	}
	if ts.Selected == ToolTransport && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[3]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		if item.OptIndex >= int(transport.TransportModeCount) {
			ts.CargoMode = true
		} else {
			ts.CargoMode = false
			ts.TransportType = transport.TransportType(item.OptIndex)
		}
	}
	if ts.Selected == ToolZone && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[4]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ts.ZoneType = item.OptIndex
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		ts.Selected = ToolPointer
	}
	return ts.Selected
}

func (ts *ToolSystem) DrawHelp() {
	helpY := int32(ToolbarY + ToolbarH + 5)
	switch ts.Selected {
	case ToolPointer:
		DrawUIText("1-7 tools | F1-F3 speed | Space pause | WASD pan | R-drag orbit | Scroll zoom", 10, helpY, 14, rl.White)
	case ToolRoad:
		DrawUIText(fmt.Sprintf("L-click place | R cycle type | PgUp/PgDn elevation | Current: %s", ToolbarItems[1].Options[ToolbarItems[1].OptIndex]), 10, helpY, 14, rl.White)
	case ToolParking:
		DrawUIText(fmt.Sprintf("L-click place %s | R cycle type", ToolbarItems[2].Options[ToolbarItems[2].OptIndex]), 10, helpY, 14, rl.White)
	case ToolTransport:
		if ts.CargoMode {
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
	if ts.HelpText != "" {
		DrawUIText(ts.HelpText, 10, helpY+18, 14, rl.NewColor(255, 180, 80, 255))
	}
}
