package ui

// ToolMode is the abstract interaction mode (24.5). Only one may be active.
type ToolMode int

const (
	ModePlace ToolMode = iota
	ModeBulldoze
	ModeUpgrade
	ModeMove
	ModeReplace
	ModePaint
	ModeMeasure
	ModeInspect
)

func (m ToolMode) String() string {
	switch m {
	case ModePlace:
		return "Place"
	case ModeBulldoze:
		return "Bulldoze"
	case ModeUpgrade:
		return "Upgrade"
	case ModeMove:
		return "Move"
	case ModeReplace:
		return "Replace"
	case ModePaint:
		return "Paint"
	case ModeMeasure:
		return "Measure"
	case ModeInspect:
		return "Inspect"
	default:
		return "Unknown"
	}
}

func modeForTool(t GameTool) ToolMode {
	switch t {
	case ToolRemove:
		return ModeBulldoze
	case ToolUpgrade:
		return ModeUpgrade
	case ToolMove:
		return ModeMove
	case ToolReplace:
		return ModeReplace
	case ToolZone:
		return ModePaint
	case ToolMeasure:
		return ModeMeasure
	case ToolInspect:
		return ModeInspect
	case ToolRoad, ToolParking, ToolTransport:
		return ModePlace
	default:
		return ModeInspect
	}
}
