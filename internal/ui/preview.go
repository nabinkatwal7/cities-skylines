package ui

import (
	"fmt"
	"math"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// PlacementPreview holds read-only placement feedback (24.6).
type PlacementPreview struct {
	Valid      bool
	Cost       float32
	FootprintW float32
	FootprintH float32
	Elevation  int32
	TerrainOK  bool
	RoadOK     bool
	UtilityOK  bool
	Collision  bool
	Messages   []string
}

// WorldContext passes simulation read-only state into UI evaluation.
type WorldContext struct {
	Sim                 *sim.SimulationManager
	WorldX, WorldZ      float32
	PreviewX, PreviewZ  float32
	OnGround            bool
	RoadActive          bool
	RoadStartNode       uint32
	RoadElevation       int32
	TransportActive     bool
	TransportStartStop  uint32
	Tools               *ToolSystem
}

func zmCellSize(sm *sim.SimulationManager) float32 {
	if sm == nil || sm.Zones == nil {
		return 4
	}
	return terrain.WorldSize / float32(sm.Zones.Width())
}

func EvalPlacement(ctx WorldContext) PlacementPreview {
	p := PlacementPreview{Valid: true, TerrainOK: true, RoadOK: true, UtilityOK: true}
	if !ctx.OnGround || ctx.Sim == nil || ctx.Tools == nil {
		p.Valid = false
		p.Messages = append(p.Messages, "No terrain target")
		return p
	}
	ts := ctx.Tools
	sm := ctx.Sim
	px, pz := ctx.PreviewX, ctx.PreviewZ
	p.Elevation = ctx.RoadElevation
	_ = sm.Heightmap.WorldHeight(px, pz)

	switch ts.Mode {
	case ModePlace:
		switch ts.Selected {
		case ToolRoad:
			p.Cost = sm.RoadPlacementCost(ts.RoadType, ctx.RoadElevation)
			p.FootprintW, p.FootprintH = 4, 4
			if sm.Heightmap.IsUnderwater(px, pz) {
				p.TerrainOK = false
				p.Messages = append(p.Messages, "Underwater")
			}
			if ctx.RoadActive && !sm.Roads.ValidNodeIndex(ctx.RoadStartNode) {
				p.Valid = false
				p.Messages = append(p.Messages, "Road chain invalid")
			} else if ctx.RoadActive && sm.Roads.ValidNodeIndex(ctx.RoadStartNode) {
				sn := &sm.Roads.Nodes[ctx.RoadStartNode]
				if reason := sm.CanPlaceRoad(sn.X, sn.Z, px, pz, ts.RoadType, ctx.RoadElevation, math.MaxUint32); reason != "" {
					p.Valid = false
					p.Collision = true
					p.Messages = append(p.Messages, reason)
				}
			}
			if sm.Money < p.Cost {
				p.Valid = false
				p.Messages = append(p.Messages, "Insufficient funds")
			}
		case ToolZone:
			p.FootprintW = zmCellSize(sm)
			p.FootprintH = zmCellSize(sm)
			if sm.Zones != nil && !sm.Zones.CanZone(px, pz) {
				p.Valid = false
				p.RoadOK = false
				p.Messages = append(p.Messages, "Needs road frontage")
			}
		case ToolParking:
			p.Cost = parkingCost(ts)
			p.FootprintW, p.FootprintH = parkingFootprint(ts)
			if sm.Heightmap.IsUnderwater(px, pz) {
				p.TerrainOK = false
				p.Valid = false
				p.Messages = append(p.Messages, "Underwater")
			}
			if sm.Money < p.Cost {
				p.Valid = false
				p.Messages = append(p.Messages, "Insufficient funds")
			}
		case ToolTransport:
			p.Cost = 500
			if ts.CargoMode {
				p.Cost = 5000
				p.FootprintW, p.FootprintH = 5, 5
			} else {
				p.FootprintW, p.FootprintH = 2, 2
			}
			if sm.Heightmap.IsUnderwater(px, pz) {
				p.TerrainOK = false
				p.Valid = false
			}
			if sm.Money < p.Cost {
				p.Valid = false
				p.Messages = append(p.Messages, "Insufficient funds")
			}
		}
	case ModePaint:
		px, pz := ctx.PreviewX, ctx.PreviewZ
		p.FootprintW = zmCellSize(sm)
		p.FootprintH = zmCellSize(sm)
		if sm.Zones != nil && !sm.Zones.CanZone(px, pz) {
			p.Valid = false
			p.RoadOK = false
			p.Messages = append(p.Messages, "Cannot paint zone here")
		}
	case ModeBulldoze, ModeUpgrade, ModeInspect:
		// selection feedback handled separately
	default:
	}

	if sm.Services != nil && ts.Mode == ModePlace && ts.Selected != ToolRoad {
		p.UtilityOK = true
	}

	if p.TerrainOK && p.RoadOK && !p.Collision && (p.Cost == 0 || sm.Money >= p.Cost) {
		if len(p.Messages) == 0 {
			p.Valid = true
		}
	} else if len(p.Messages) > 0 {
		p.Valid = false
	}
	return p
}

func parkingCost(ts *ToolSystem) float32 {
	switch {
	case ts.AirportMode:
		return 10000
	case ts.PortMode:
		return 8000
	case ts.BusDepotMode, ts.TramDepotMode, ts.MetroDepotMode, ts.FerryDepotMode,
		ts.MonorailDepotMode, ts.CableCarDepotMode, ts.TaxiDepotMode:
		return 5000
	case ts.ParkingGarage:
		return 3000
	default:
		return 1000
	}
}

func parkingFootprint(ts *ToolSystem) (float32, float32) {
	switch {
	case ts.AirportMode:
		return 12, 8
	case ts.PortMode:
		return 8, 10
	case ts.BusDepotMode, ts.TramDepotMode, ts.MetroDepotMode, ts.TaxiDepotMode:
		return 6, 4
	default:
		return 20, 15
	}
}

func previewColor(valid bool) rl.Color {
	if valid {
		return rl.NewColor(80, 220, 80, 140)
	}
	return rl.NewColor(220, 80, 80, 140)
}

func previewWire(valid bool) rl.Color {
	if valid {
		return rl.NewColor(100, 255, 100, 220)
	}
	return rl.NewColor(255, 100, 100, 220)
}

// DrawPreview3D renders footprint and guides in world space (24.6).
func DrawPreview3D(ctx WorldContext, p PlacementPreview) {
	if !ctx.OnGround || ctx.Sim == nil || ctx.Tools == nil {
		return
	}
	sm := ctx.Sim
	ts := ctx.Tools
	px, pz := ctx.PreviewX, ctx.PreviewZ
	wx, wz := ctx.WorldX, ctx.WorldZ
	h := sm.Heightmap.WorldHeight(px, pz)
	col := previewColor(p.Valid)
	wire := previewWire(p.Valid)

	switch ts.Mode {
	case ModePlace, ModePaint:
		switch ts.Selected {
		case ToolRoad:
			rl.DrawSphere(rl.NewVector3(px, h+0.5, pz), 0.8, col)
			if ctx.RoadActive && sm.Roads.ValidNodeIndex(ctx.RoadStartNode) {
				sn := &sm.Roads.Nodes[ctx.RoadStartNode]
				sh := sm.Heightmap.WorldHeight(sn.X, sn.Z)
				lineCol := wire
				if !p.Valid {
					lineCol = rl.Red
				}
				rl.DrawLine3D(rl.NewVector3(sn.X, sh+0.2, sn.Z), rl.NewVector3(px, h+0.2, pz), lineCol)
			}
		case ToolZone:
			drawFootprint(px, h, pz, zmCellSize(sm), 0.2, zmCellSize(sm), col, wire)
		case ToolParking:
			drawFootprint(wx, sm.Heightmap.WorldHeight(wx, wz), wz, p.FootprintW, 0.5, p.FootprintH, col, wire)
		case ToolTransport:
			if ts.CargoMode {
				drawFootprint(wx, h, wz, 5, 2, 5, col, wire)
			} else {
				stopCol := transport.TransportStopColor(ts.TransportType)
				if !p.Valid {
					stopCol = rl.Red
				}
				if ctx.TransportActive {
					if sn := sm.Transport.StopByID(ctx.TransportStartStop); sn != nil {
						sh := sm.Heightmap.WorldHeight(sn.X, sn.Z)
						rl.DrawLine3D(rl.NewVector3(sn.X, sh+0.5, sn.Z), rl.NewVector3(px, h+0.5, pz), stopCol)
					}
				}
				rl.DrawSphere(rl.NewVector3(px, h+0.5, pz), 0.6, stopCol)
			}
		}
	case ModeBulldoze:
		drawBulldozePreview(sm, wx, wz)
	case ModeUpgrade:
		drawUpgradePreview(sm, wx, wz)
	case ModeMeasure:
		drawMeasurePreview(ctx)
	case ModeInspect:
		rl.DrawSphere(rl.NewVector3(wx, sm.Heightmap.WorldHeight(wx, wz)+0.4, wz), 0.5, rl.SkyBlue)
	}
}

func drawFootprint(x, h, z, w, height, d float32, fill, wire rl.Color) {
	rl.DrawCube(rl.NewVector3(x, h+height*0.5, z), w, height, d, fill)
	rl.DrawCubeWires(rl.NewVector3(x, h+height*0.5, z), w, height, d, wire)
}

func drawBulldozePreview(sm *sim.SimulationManager, x, z float32) {
	if sm.Transport != nil {
		if stop := sm.Transport.NearestStop(x, z, 8); stop != nil {
			sh := sm.Heightmap.WorldHeight(stop.X, stop.Z) + 0.5
			rl.DrawCubeWires(rl.NewVector3(stop.X, sh, stop.Z), 2, 2, 2, rl.Red)
			return
		}
	}
	if idx := sm.Roads.NearestSegment(x, z); idx >= 0 {
		seg := sm.Roads.Segments[idx]
		na := &sm.Roads.Nodes[seg.NodeA]
		nb := &sm.Roads.Nodes[seg.NodeB]
		ha := sm.Heightmap.WorldHeight(na.X, na.Z) + 0.5
		hb := sm.Heightmap.WorldHeight(nb.X, nb.Z) + 0.5
		rl.DrawLine3D(rl.NewVector3(na.X, ha, na.Z), rl.NewVector3(nb.X, hb, nb.Z), rl.Red)
	}
}

func drawUpgradePreview(sm *sim.SimulationManager, x, z float32) {
	if idx := sm.Roads.NearestSegment(x, z); idx >= 0 {
		seg := sm.Roads.Segments[idx]
		na := &sm.Roads.Nodes[seg.NodeA]
		nb := &sm.Roads.Nodes[seg.NodeB]
		ha := sm.Heightmap.WorldHeight(na.X, na.Z) + 0.5
		hb := sm.Heightmap.WorldHeight(nb.X, nb.Z) + 0.5
		rl.DrawLine3D(rl.NewVector3(na.X, ha, na.Z), rl.NewVector3(nb.X, hb, nb.Z), rl.Yellow)
	}
}

func drawMeasurePreview(ctx WorldContext) {
	if !ctx.Tools.measureASet {
		return
	}
	ax, az := ctx.Tools.measureAX, ctx.Tools.measureAZ
	sm := ctx.Sim
	ha := sm.Heightmap.WorldHeight(ax, az) + 0.5
	hb := sm.Heightmap.WorldHeight(ctx.WorldX, ctx.WorldZ) + 0.5
	rl.DrawLine3D(rl.NewVector3(ax, ha, az), rl.NewVector3(ctx.WorldX, hb, ctx.WorldZ), rl.Orange)
	rl.DrawSphere(rl.NewVector3(ax, ha, az), 0.4, rl.Orange)
}

// DrawPreviewHUD shows cost and warnings on screen (24.6).
func DrawPreviewHUD(p PlacementPreview, y int32) {
	if p.Cost > 0 {
		col := csMoney
		if !p.Valid {
			col = csMoneyNeg
		}
		drawLabel(fmt.Sprintf("Cost $%.0f", p.Cost), 12, y, FontMd, col)
		y += 20
	}
	if p.Elevation != 0 {
		drawLabel(fmt.Sprintf("Elevation %+d", p.Elevation), 12, y, FontSm, csTextDim)
		y += 18
	}
	for _, msg := range p.Messages {
		drawLabel(msg, 12, y, FontSm, rl.NewColor(255, 190, 110, 255))
		y += 18
	}
	flags := ""
	if !p.TerrainOK {
		flags += " terrain"
	}
	if !p.RoadOK {
		flags += " road"
	}
	if !p.UtilityOK {
		flags += " utility"
	}
	if p.Collision {
		flags += " collision"
	}
	if flags != "" {
		drawLabel("Conflict:"+flags, 12, y, FontSm, csMoneyNeg)
	}
}

func zoneTypeName(zt zoning.ZoneType) string {
	switch zt {
	case zoning.ZoneResidentialLow:
		return "Residential Low"
	case zoning.ZoneResidentialHigh:
		return "Residential High"
	case zoning.ZoneCommercialLow:
		return "Commercial Low"
	case zoning.ZoneCommercialHigh:
		return "Commercial High"
	case zoning.ZoneIndustrial:
		return "Industrial"
	case zoning.ZoneOffice:
		return "Office"
	default:
		return "Building"
	}
}
