package ui

import (
	"math"

	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	roadGrid      = 4.0
	roadAngleSnap = math.Pi / 12.0 // 15° like Cities: Skylines
	nodeSnapDist  = 4.0 // matches sim.RoadProximityDist
)

type SnapContext struct {
	Sim           *sim.SimulationManager
	Tool          GameTool
	Mode          ToolMode
	RoadActive bool
	RoadStartX float32
	RoadStartZ float32
	PreviewX      float32
	PreviewZ      float32
}

func SnapRoadXZ(x, z float32) (float32, float32) {
	return float32(math.Round(float64(x/roadGrid))) * roadGrid,
		float32(math.Round(float64(z/roadGrid))) * roadGrid
}

func SnapZoneXZ(sm *sim.SimulationManager, x, z float32) (float32, float32) {
	if sm == nil || sm.Zones == nil {
		return x, z
	}
	cx := sm.Zones.CellX(x)
	cz := sm.Zones.CellZ(z)
	return sm.Zones.CellCenter(cx, cz)
}

func SnapPlacement(ctx SnapContext, x, z float32) (float32, float32) {
	switch {
	case ctx.Tool == ToolRoad || (ctx.Mode == ModePlace && ctx.Tool == ToolRoad):
		return snapRoad(ctx, x, z)
	case ctx.Tool == ToolZone || ctx.Mode == ModePaint:
		return SnapZoneXZ(ctx.Sim, x, z)
	case ctx.Tool == ToolTransport:
		return SnapRoadXZ(x, z)
	default:
		return x, z
	}
}

func snapRoad(ctx SnapContext, x, z float32) (float32, float32) {
	sx, sz := SnapRoadXZ(x, z)
	sm := ctx.Sim
	if sm == nil || sm.Roads == nil {
		return sx, sz
	}
	if ctx.RoadActive {
		if idx, ok := sm.Roads.NearestNode(sx, sz); ok {
			n := &sm.Roads.Nodes[idx]
			dx := n.X - sx
			dz := n.Z - sz
			if dx*dx+dz*dz < nodeSnapDist*nodeSnapDist {
				return n.X, n.Z
			}
		}
		return snapRoadFromNode(ctx.RoadStartX, ctx.RoadStartZ, sx, sz)
	}
	return sx, sz
}

func snapRoadFromNode(ax, az, x, z float32) (float32, float32) {
	dx := x - ax
	dz := z - az
	dist := float32(math.Hypot(float64(dx), float64(dz)))
	if dist < roadGrid*0.5 {
		return ax, az
	}
	angle := math.Atan2(float64(dz), float64(dx))
	angle = math.Round(angle/roadAngleSnap) * roadAngleSnap
	dist = float32(math.Round(float64(dist/roadGrid))) * roadGrid
	if dist < roadGrid {
		dist = roadGrid
	}
	return ax + float32(math.Cos(angle))*dist, az + float32(math.Sin(angle))*dist
}

func DrawBuildGrid(ctx SnapContext, camX, camZ float32) {
	switch {
	case ctx.Tool == ToolRoad:
		rl.DrawGrid(100, roadGrid)
		drawRoadNodeHints(ctx)
	case ctx.Tool == ToolZone || ctx.Mode == ModePaint:
		drawZonePlacementGrid(ctx, camX, camZ)
	}
}

func drawRoadNodeHints(ctx SnapContext) {
	sm := ctx.Sim
	if sm == nil || sm.Roads == nil {
		return
	}
	for i := range sm.Roads.Nodes {
		n := &sm.Roads.Nodes[i]
		h := sm.Heightmap.WorldHeight(n.X, n.Z)
		rl.DrawSphere(rl.NewVector3(n.X, h+0.15, n.Z), 0.35, rl.NewColor(255, 255, 255, 90))
	}
}
