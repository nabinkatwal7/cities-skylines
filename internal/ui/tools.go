package ui

import (
	"fmt"
	"math"

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
	ToolInspect
	ToolMeasure
	ToolMove
	ToolReplace
)

type ToolbarItem struct {
	Tool     GameTool
	Label    string
	Key      int32
	Color    rl.Color
	Options  []string
	OptIndex int
}

// ToolSystem tracks the active build tool and its mode options (24.5).
type ToolSystem struct {
	Mode              ToolMode
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
	measureASet       bool
	measureAX           float32
	measureAZ           float32
	measureDist         float32
}

func NewToolSystem() *ToolSystem {
	return &ToolSystem{
		Selected: ToolPointer,
		Mode:     ModeInspect,
		RoadType: road.RoadTwoLane,
	}
}

// Activate selects exactly one tool and its mode (24.5).
func (ts *ToolSystem) Activate(tool GameTool) {
	ts.Selected = tool
	ts.Mode = modeForTool(tool)
	if tool != ToolMeasure {
		ts.measureASet = false
	}
}

func (ts *ToolSystem) MeasureClick(x, z float32) {
	if !ts.measureASet {
		ts.measureAX, ts.measureAZ = x, z
		ts.measureASet = true
		ts.measureDist = 0
		return
	}
	dx := x - ts.measureAX
	dz := z - ts.measureAZ
	ts.measureDist = float32(math.Sqrt(float64(dx*dx + dz*dz)))
	ts.measureASet = false
}

var ToolbarItems = []ToolbarItem{
	{ToolPointer, "Pointer", rl.KeyOne, rl.NewColor(200, 200, 200, 255), nil, 0},
	{ToolRoad, "Roads", rl.KeyTwo, rl.NewColor(180, 160, 120, 255), road.RoadTypeOptionNames, 0},
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
	if ts.Selected == ToolRoad && rl.IsKeyPressed(rl.KeyR) {
		item := &ToolbarItems[1]
		item.OptIndex = (item.OptIndex + 1) % len(item.Options)
		ts.RoadType = road.RoadTypeFromOptionIndex(item.OptIndex)
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
		ts.Activate(ToolPointer)
	}
	return ts.Selected
}

func (ts *ToolSystem) DrawHelp(chromeTop int32) {
	helpY := chromeTop - 22
	if helpY < TopBarH+4 {
		helpY = TopBarH + 4
	}
	switch ts.Mode {
	case ModeInspect:
		drawLabel("Click to inspect  ·  I inspect  ·  M measure  ·  1 cursor  ·  6 bulldoze  ·  7 upgrade", 12, helpY, FontSm, csText)
	case ModeMeasure:
		hint := "Click two points to measure distance"
		if ts.measureDist > 0 {
			hint = fmt.Sprintf("Distance: %.1f m", ts.measureDist)
		} else if ts.measureASet {
			hint = "Click second point"
		}
		drawLabel(hint, 12, helpY, FontSm, csText)
	case ModeMove:
		drawLabel("Move tool: select object then click destination (coming soon)", 12, helpY, FontSm, csText)
	case ModeReplace:
		drawLabel("Replace tool: click asset to swap (coming soon)", 12, helpY, FontSm, csText)
	case ModeBulldoze:
		drawLabel("Click to bulldoze  ·  Esc cancel", 12, helpY, FontSm, csText)
	case ModeUpgrade:
		drawLabel(fmt.Sprintf("Click road to upgrade to %s  ·  R change type", ToolbarItems[1].Options[ToolbarItems[1].OptIndex]), 12, helpY, FontSm, csText)
	case ModePaint:
		drawLabel(fmt.Sprintf("Click to paint %s  ·  R cycle type", ToolbarItems[4].Options[ToolbarItems[4].OptIndex]), 12, helpY, FontSm, csText)
	case ModePlace:
		switch ts.Selected {
		case ToolRoad:
			drawLabel(fmt.Sprintf("Click to place  ·  R cycle  ·  PgUp/Dn elevation  ·  %s", ToolbarItems[1].Options[ToolbarItems[1].OptIndex]), 12, helpY, FontSm, csText)
		case ToolParking:
			drawLabel(fmt.Sprintf("Click to place %s  ·  R cycle type", ToolbarItems[2].Options[ToolbarItems[2].OptIndex]), 12, helpY, FontSm, csText)
		case ToolTransport:
			if ts.CargoMode {
				drawLabel("Click to place cargo station  ·  R cycle type", 12, helpY, FontSm, csText)
			} else {
				drawLabel(fmt.Sprintf("Click to place %s stop  ·  R cycle type", ToolbarItems[3].Options[ToolbarItems[3].OptIndex]), 12, helpY, FontSm, csText)
			}
		}
	}
	if ts.HelpText != "" {
		drawLabel(ts.HelpText, 12, helpY+20, FontSm, rl.NewColor(255, 190, 110, 255))
	}
}

// ApplyAsset configures the tool system from a build-menu selection.
func (ts *ToolSystem) ApplyAsset(a BuildAsset) {
	switch a.Category {
	case CatRoads:
		ts.Activate(ToolRoad)
		for i, name := range roadOptions().Options {
			if name == a.Name {
				roadOptions().OptIndex = i
				ts.RoadType = road.RoadTypeFromOptionIndex(i)
				return
			}
		}
	case CatZoning:
		ts.Activate(ToolZone)
		for i, name := range zoneOptions().Options {
			if name == a.Name {
				zoneOptions().OptIndex = i
				ts.ZoneType = i
				return
			}
		}
	case CatPublicTransport:
		for i, name := range transportOptions().Options {
			if name == a.Name {
				transportOptions().OptIndex = i
				if i >= int(transport.TransportModeCount) {
					ts.CargoMode = true
					ts.Activate(ToolTransport)
				} else {
					ts.CargoMode = false
					ts.TransportType = transport.TransportType(i)
					ts.Activate(ToolTransport)
				}
				return
			}
		}
		for i, name := range parkingOptions().Options {
			if name == a.Name {
				parkingOptions().OptIndex = i
				ts.setParkingMode(i)
				ts.Activate(ToolParking)
				return
			}
		}
	default:
		ts.HelpText = a.Name + ": placement coming soon"
	}
}
