package zoning

import (
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ZoneType uint8

const (
	ZoneNone ZoneType = iota
	ZoneResidentialLow
	ZoneResidentialHigh
	ZoneCommercialLow
	ZoneCommercialHigh
	ZoneIndustrial
	ZoneOffice
)

type ZoneCategory uint8

const (
	CategoryResidential ZoneCategory = iota
	CategoryCommercial
	CategoryIndustrial
	CategoryOffice
)

// ZoneTraits holds per-type simulation weights used by demand, economy, and growth.
type ZoneTraits struct {
	Category          ZoneCategory
	Density           float32 // low=0.5, high/generic=1.0
	Housing           bool
	PopGrowth         bool
	TaxIncome         bool
	Shopping          bool
	Entertainment     bool
	Employment        bool
	Goods             bool
	Freight           bool
	Pollution         float32 // 0=none, 1=heavy
	HighEducationJobs bool
	RequiresEducation bool
}

func ZoneCategoryOf(zt ZoneType) ZoneCategory {
	switch zt {
	case ZoneResidentialLow, ZoneResidentialHigh:
		return CategoryResidential
	case ZoneCommercialLow, ZoneCommercialHigh:
		return CategoryCommercial
	case ZoneIndustrial:
		return CategoryIndustrial
	case ZoneOffice:
		return CategoryOffice
	default:
		return CategoryResidential
	}
}

func ZoneTraitsOf(zt ZoneType) ZoneTraits {
	switch zt {
	case ZoneResidentialLow:
		return ZoneTraits{CategoryResidential, 0.5, true, true, true, false, false, false, false, false, 0, false, false}
	case ZoneResidentialHigh:
		return ZoneTraits{CategoryResidential, 1.0, true, true, true, false, false, false, false, false, 0, false, false}
	case ZoneCommercialLow:
		return ZoneTraits{CategoryCommercial, 0.5, false, false, false, true, true, true, false, false, 0.1, false, false}
	case ZoneCommercialHigh:
		return ZoneTraits{CategoryCommercial, 1.0, false, false, false, true, true, true, false, false, 0.15, false, false}
	case ZoneIndustrial:
		return ZoneTraits{CategoryIndustrial, 1.0, false, false, false, false, false, true, true, true, 0.8, false, false}
	case ZoneOffice:
		return ZoneTraits{CategoryOffice, 1.0, false, false, true, false, false, true, false, false, 0.05, true, true}
	default:
		return ZoneTraits{}
	}
}

type ZoneCell struct {
	Type    ZoneType
	Density float32
}

type ZoneManager struct {
	Cells       [][]ZoneCell
	width, height int
	roads       *road.RoadManager
	buildability *terrain.BuildabilityChecker
	MaxZoneDepth int
}

func NewZoneManager(w, h int, roads *road.RoadManager, bc *terrain.BuildabilityChecker) *ZoneManager {
	cells := make([][]ZoneCell, h)
	for z := range cells {
		cells[z] = make([]ZoneCell, w)
	}
	return &ZoneManager{
		Cells:       cells,
		width:       w,
		height:      h,
		roads:       roads,
		buildability: bc,
		MaxZoneDepth: 4,
	}
}

func (zm *ZoneManager) CellX(worldX float32) int {
	return int((worldX + terrain.WorldSize/2) / terrain.WorldSize * float32(zm.width-1))
}

func (zm *ZoneManager) CellZ(worldZ float32) int {
	return int((worldZ + terrain.WorldSize/2) / terrain.WorldSize * float32(zm.height-1))
}

func (zm *ZoneManager) CanZone(worldX, worldZ float32) bool {
	cellSize := terrain.WorldSize / float32(zm.width)
	for dz := -zm.MaxZoneDepth; dz <= zm.MaxZoneDepth; dz++ {
		for dx := -zm.MaxZoneDepth; dx <= zm.MaxZoneDepth; dx++ {
			checkX := worldX + float32(dx)*cellSize
			checkZ := worldZ + float32(dz)*cellSize
			if zm.roads != nil && zm.roads.HasNearbyRoad(checkX, checkZ, cellSize*1.5) {
				if zm.buildability != nil {
					info := zm.buildability.GetBuildability(worldX, worldZ)
					if info.Score < 0.3 || info.IsUnderwater {
						return false
					}
				}
				return true
			}
		}
	}
	return false
}

func (zm *ZoneManager) SetZone(worldX, worldZ float32, zt ZoneType) {
	x := zm.CellX(worldX)
	z := zm.CellZ(worldZ)
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return
	}
	if !zm.CanZone(worldX, worldZ) {
		return
	}
	zm.Cells[z][x].Type = zt
}

func (zm *ZoneManager) RemoveZone(worldX, worldZ float32) {
	x := zm.CellX(worldX)
	z := zm.CellZ(worldZ)
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return
	}
	zm.Cells[z][x].Type = ZoneNone
	zm.Cells[z][x].Density = 0
}

func (zm *ZoneManager) CellTypeAt(worldX, worldZ float32) ZoneType {
	x := zm.CellX(worldX)
	z := zm.CellZ(worldZ)
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return ZoneNone
	}
	return zm.Cells[z][x].Type
}

func ZoneTypeName(zt ZoneType) string {
	switch zt {
	case ZoneResidentialLow:
		return "Res Low"
	case ZoneResidentialHigh:
		return "Res High"
	case ZoneCommercialLow:
		return "Com Low"
	case ZoneCommercialHigh:
		return "Com High"
	case ZoneIndustrial:
		return "Industrial"
	case ZoneOffice:
		return "Office"
	default:
		return "None"
	}
}

func ZoneColor(zt ZoneType) rl.Color {
	switch zt {
	case ZoneResidentialLow:
		return rl.NewColor(60, 200, 60, 120)
	case ZoneResidentialHigh:
		return rl.NewColor(40, 160, 40, 140)
	case ZoneCommercialLow:
		return rl.NewColor(60, 60, 220, 120)
	case ZoneCommercialHigh:
		return rl.NewColor(40, 40, 180, 140)
	case ZoneIndustrial:
		return rl.NewColor(220, 200, 40, 120)
	case ZoneOffice:
		return rl.NewColor(180, 180, 200, 120)
	default:
		return rl.Color{}
	}
}

func (zm *ZoneManager) Draw(h *terrain.Heightmap) {
	cellSize := terrain.WorldSize / float32(zm.width)
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			cell := &zm.Cells[z][x]
			if cell.Type == ZoneNone {
				continue
			}
			col := ZoneColor(cell.Type)
			cx := float32(x)*cellSize - terrain.WorldSize/2 + cellSize*0.5
			cz := float32(z)*cellSize - terrain.WorldSize/2 + cellSize*0.5
			hy := h.WorldHeight(cx, cz) + 0.2
			rl.DrawCube(rl.NewVector3(cx, hy, cz), cellSize*0.9, 0.1, cellSize*0.9, col)
		}
	}
}
