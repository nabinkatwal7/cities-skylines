package zoning

import (
	"math"

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

// ZoneDepthForRoad returns how many grid cells may be zoned away from a road of this type.
func ZoneDepthForRoad(rt road.RoadType) int {
	switch road.HierarchyForRoad(rt) {
	case road.HierarchyHighway:
		return 8
	case road.HierarchyArterial:
		return 6
	case road.HierarchyCollector:
		return 4
	default:
		return 2
	}
}

type ZoneCell struct {
	Type    ZoneType
	Density float32
	LotID   int
}

type ZoneLot struct {
	ID     int
	X, Z   int
	Width  int
	Height int
	Type   ZoneType
	Cells  int
}

type ZoneManager struct {
	Cells        [][]ZoneCell
	width, height int
	roads        *road.RoadManager
	buildability *terrain.BuildabilityChecker
	services     ServiceCoverage
	demand       DemandProvider
	buildings    BuildingCatalog
	lots         []ZoneLot
	lotsDirty    bool
	districts    []District
	cellDistrict [][]uint8
	dirtyLand    [][2]int
}

func NewZoneManager(w, h int, roads *road.RoadManager, bc *terrain.BuildabilityChecker) *ZoneManager {
	cells := make([][]ZoneCell, h)
	for z := range cells {
		cells[z] = make([]ZoneCell, w)
		for x := range cells[z] {
			cells[z][x].LotID = -1
		}
	}
	return &ZoneManager{
		Cells:        cells,
		width:        w,
		height:       h,
		roads:        roads,
		buildability: bc,
		lotsDirty:    true,
	}
}

func (zm *ZoneManager) Init() {
	zm.initDistricts()
}

func (zm *ZoneManager) Width() int  { return zm.width }
func (zm *ZoneManager) Height() int { return zm.height }

func (zm *ZoneManager) markLandDirty(x, z int) {
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return
	}
	zm.dirtyLand = append(zm.dirtyLand, [2]int{x, z})
}

func (zm *ZoneManager) DrainDirtyLand() [][2]int {
	d := zm.dirtyLand
	zm.dirtyLand = nil
	return d
}

func (zm *ZoneManager) ExportZoneTypes() []uint8 {
	out := make([]uint8, zm.width*zm.height)
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			out[z*zm.width+x] = uint8(zm.Cells[z][x].Type)
		}
	}
	return out
}

func (zm *ZoneManager) ImportZoneTypes(flat []uint8) {
	if len(flat) != zm.width*zm.height {
		return
	}
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			zm.Cells[z][x].Type = ZoneType(flat[z*zm.width+x])
		}
	}
	zm.lotsDirty = true
}

func (zm *ZoneManager) cellSize() float32 {
	return terrain.WorldSize / float32(zm.width)
}

func (zm *ZoneManager) CellX(worldX float32) int {
	return int((worldX + terrain.WorldSize/2) / terrain.WorldSize * float32(zm.width-1))
}

func (zm *ZoneManager) CellZ(worldZ float32) int {
	return int((worldZ + terrain.WorldSize/2) / terrain.WorldSize * float32(zm.height-1))
}

func (zm *ZoneManager) CellCenter(x, z int) (float32, float32) {
	cs := zm.cellSize()
	return float32(x)*cs - terrain.WorldSize/2 + cs*0.5,
		float32(z)*cs - terrain.WorldSize/2 + cs*0.5
}

func (zm *ZoneManager) CanZone(worldX, worldZ float32) bool {
	if zm.buildability != nil {
		info := zm.buildability.GetBuildability(worldX, worldZ)
		if info.Score < 0.3 || info.IsUnderwater {
			return false
		}
	}
	return zm.roadConnected(worldX, worldZ)
}

func (zm *ZoneManager) RoadConnected(worldX, worldZ float32) bool {
	return zm.roadConnected(worldX, worldZ)
}

func (zm *ZoneManager) roadConnected(worldX, worldZ float32) bool {
	if zm.roads == nil {
		return false
	}
	rt, dist, ok := zm.roads.NearestRoad(worldX, worldZ)
	if !ok {
		return false
	}
	cs := zm.cellSize()
	if dist < cs*0.3 {
		return false
	}
	depth := int(math.Ceil(float64(dist / cs)))
	return depth >= 1 && depth <= ZoneDepthForRoad(rt)
}

func (zm *ZoneManager) SetZoneCell(x, z int, zt ZoneType) {
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return
	}
	zm.Cells[z][x].Type = zt
	zm.markLotsDirty()
	zm.markLandDirty(x, z)
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
	zm.SetZoneCell(x, z, zt)
}

func (zm *ZoneManager) RemoveZone(worldX, worldZ float32) {
	x := zm.CellX(worldX)
	z := zm.CellZ(worldZ)
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return
	}
	zm.RemoveZoneCell(x, z)
}

func (zm *ZoneManager) RemoveZoneCell(x, z int) {
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return
	}
	zm.Cells[z][x].Type = ZoneNone
	zm.Cells[z][x].Density = 0
	zm.Cells[z][x].LotID = -1
	zm.markLotsDirty()
	zm.markLandDirty(x, z)
}

func (zm *ZoneManager) CellTypeAt(worldX, worldZ float32) ZoneType {
	x := zm.CellX(worldX)
	z := zm.CellZ(worldZ)
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return ZoneNone
	}
	return zm.Cells[z][x].Type
}

func (zm *ZoneManager) CategoryCounts() (residential, commercial, industrial, office int) {
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			switch zm.Cells[z][x].Type {
			case ZoneResidentialLow, ZoneResidentialHigh:
				residential++
			case ZoneCommercialLow, ZoneCommercialHigh:
				commercial++
			case ZoneIndustrial:
				industrial++
			case ZoneOffice:
				office++
			}
		}
	}
	return
}

func (zm *ZoneManager) markLotsDirty() {
	zm.lotsDirty = true
}

func (zm *ZoneManager) Lots() []ZoneLot {
	if zm.lotsDirty {
		zm.rebuildLots()
	}
	return zm.lots
}

func (zm *ZoneManager) LotAtCell(x, z int) *ZoneLot {
	if x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return nil
	}
	zm.Lots()
	id := zm.Cells[z][x].LotID
	if id < 0 || id >= len(zm.lots) {
		return nil
	}
	return &zm.lots[id]
}

func (zm *ZoneManager) LotAt(worldX, worldZ float32) *ZoneLot {
	return zm.LotAtCell(zm.CellX(worldX), zm.CellZ(worldZ))
}

func (zm *ZoneManager) rebuildLots() {
	zm.lots = zm.lots[:0]
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			zm.Cells[z][x].LotID = -1
		}
	}

	visited := make([][]bool, zm.height)
	for z := range visited {
		visited[z] = make([]bool, zm.width)
	}

	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			zt := zm.Cells[z][x].Type
			if zt == ZoneNone || visited[z][x] {
				continue
			}
			cells := zm.floodFill(x, z, zt, visited)
			minX, maxX := x, x
			minZ, maxZ := z, z
			for _, c := range cells {
				if c.x < minX {
					minX = c.x
				}
				if c.x > maxX {
					maxX = c.x
				}
				if c.z < minZ {
					minZ = c.z
				}
				if c.z > maxZ {
					maxZ = c.z
				}
			}
			w := maxX - minX + 1
			h := maxZ - minZ + 1
			if len(cells) != w*h {
				continue
			}
			id := len(zm.lots)
			zm.lots = append(zm.lots, ZoneLot{
				ID: id, X: minX, Z: minZ, Width: w, Height: h, Type: zt, Cells: len(cells),
			})
			for _, c := range cells {
				zm.Cells[c.z][c.x].LotID = id
			}
		}
	}
	zm.lotsDirty = false
}

type gridPoint struct{ x, z int }

func (zm *ZoneManager) floodFill(startX, startZ int, zt ZoneType, visited [][]bool) []gridPoint {
	var cells []gridPoint
	stack := []gridPoint{{startX, startZ}}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if p.x < 0 || p.x >= zm.width || p.z < 0 || p.z >= zm.height {
			continue
		}
		if visited[p.z][p.x] || zm.Cells[p.z][p.x].Type != zt {
			continue
		}
		visited[p.z][p.x] = true
		cells = append(cells, p)
		stack = append(stack,
			gridPoint{p.x + 1, p.z},
			gridPoint{p.x - 1, p.z},
			gridPoint{p.x, p.z + 1},
			gridPoint{p.x, p.z - 1},
		)
	}
	return cells
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
	cellSize := zm.cellSize()
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
