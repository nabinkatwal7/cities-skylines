package ui

import (
	"math"

	"github.com/katwate/js-skylines/internal/sim"
)

const roadGrid = 4.0

// SnapRoadXZ aligns placement to the visible 4m build grid.
func SnapRoadXZ(x, z float32) (float32, float32) {
	return float32(math.Round(float64(x/roadGrid))) * roadGrid,
		float32(math.Round(float64(z/roadGrid))) * roadGrid
}

// SnapZoneXZ aligns placement to the center of the zone cell under the cursor.
func SnapZoneXZ(sm *sim.SimulationManager, x, z float32) (float32, float32) {
	if sm == nil || sm.Zones == nil {
		return x, z
	}
	cx := sm.Zones.CellX(x)
	cz := sm.Zones.CellZ(z)
	return sm.Zones.CellCenter(cx, cz)
}

// SnapPlacement returns snapped preview coordinates for the active tool.
func SnapPlacement(sm *sim.SimulationManager, tool GameTool, mode ToolMode, x, z float32) (float32, float32) {
	switch {
	case tool == ToolRoad || (mode == ModePlace && tool == ToolRoad):
		return SnapRoadXZ(x, z)
	case tool == ToolZone || mode == ModePaint:
		return SnapZoneXZ(sm, x, z)
	case tool == ToolTransport:
		return SnapRoadXZ(x, z)
	default:
		return x, z
	}
}
