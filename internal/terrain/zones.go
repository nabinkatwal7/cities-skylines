package terrain

import (
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

type ZoneCell struct {
	Type  ZoneType
	Density float32
}

type ZoneManager struct {
	Cells [][]ZoneCell
	width, height int
}

func NewZoneManager(w, h int) *ZoneManager {
	cells := make([][]ZoneCell, h)
	for z := range cells {
		cells[z] = make([]ZoneCell, w)
	}
	return &ZoneManager{Cells: cells, width: w, height: h}
}

func (zm *ZoneManager) cellX(worldX float32) int {
	return int((worldX + WorldSize/2) / WorldSize * float32(zm.width-1))
}

func (zm *ZoneManager) cellZ(worldZ float32) int {
	return int((worldZ + WorldSize/2) / WorldSize * float32(zm.height-1))
}

const maxZoneDepth = 4

func (zm *ZoneManager) CanZone(worldX, worldZ float32, roads *RoadManager) bool {
	cellSize := WorldSize / float32(zm.width)
	for dz := -maxZoneDepth; dz <= maxZoneDepth; dz++ {
		for dx := -maxZoneDepth; dx <= maxZoneDepth; dx++ {
			checkX := worldX + float32(dx)*cellSize
			checkZ := worldZ + float32(dz)*cellSize
			if roads.HasNearbyRoad(checkX, checkZ, cellSize*1.5) {
				cx := zm.cellX(worldX)
				cz := zm.cellZ(worldZ)
				if abs(dx) <= maxZoneDepth && abs(dz) <= maxZoneDepth {
					_ = cx
					_ = cz
				}
				return true
			}
		}
	}
	return false
}

func (zm *ZoneManager) SetZone(worldX, worldZ float32, zt ZoneType, roads *RoadManager) {
	x := zm.cellX(worldX)
	z := zm.cellZ(worldZ)
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return
	}
	if !zm.CanZone(worldX, worldZ, roads) {
		return
	}
	zm.Cells[z][x].Type = zt
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func ZoneTypeName(zt ZoneType) string {
	switch zt {
	case ZoneResidentialLow:
		return "Residential Low"
	case ZoneResidentialHigh:
		return "Residential High"
	case ZoneCommercialLow:
		return "Commercial Low"
	case ZoneCommercialHigh:
		return "Commercial High"
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

func (zm *ZoneManager) CellTypeAt(worldX, worldZ float32) ZoneType {
	x := zm.cellX(worldX)
	z := zm.cellZ(worldZ)
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return ZoneNone
	}
	return zm.Cells[z][x].Type
}

func (zm *ZoneManager) Draw(h *Heightmap) {
	cellSize := WorldSize / float32(zm.width)
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			cell := &zm.Cells[z][x]
			if cell.Type == ZoneNone {
				continue
			}
			col := ZoneColor(cell.Type)
			cx := float32(x)*cellSize - WorldSize/2 + cellSize*0.5
			cz := float32(z)*cellSize - WorldSize/2 + cellSize*0.5
			hy := h.WorldHeight(cx, cz) + 0.2
			rl.DrawCube(rl.NewVector3(cx, hy, cz), cellSize*0.9, 0.1, cellSize*0.9, col)
		}
	}
}
