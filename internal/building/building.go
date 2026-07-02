package building

import (
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MaxLevel           = 5
	baseConstructSec   = 10.0
	upgradeSecPerLevel = 20.0
	abandonGraceSec    = 8.0
	demolishGraceSec   = 15.0
	stageFoundationEnd = 0.33
	stageFrameworkEnd  = 0.66
)

type State uint8

const (
	StateConstructing State = iota
	StateOccupied
	StateAbandoned
)

type ConstructStage uint8

const (
	StageFoundation ConstructStage = iota
	StageFramework
	StageCompleted
)

type Building struct {
	ID       uint32
	LotID    int
	Type     zoning.ZoneType
	CellX    int
	CellZ    int
	Width    int
	Height   int
	WorldX   float32
	WorldZ   float32
	Level    int
	State    State
	Stage    ConstructStage
	Progress float32
	BuildTime float32
	UpgradeProgress float32
	AbandonTimer    float32
	DemolishTimer   float32
	LandValue       float32
	Household       Household
	Business        Business
	Occupancy       Occupancy
	Consumption     Consumption
	AI              BuildingAI
	ServiceOK       bool
}

type Manager struct {
	Buildings []Building
	byLot     map[int]int
	landValue [][]float32
	nextID    uint32

	zones    *zoning.ZoneManager
	demand   *zoning.DemandEngine
	services zoning.ServiceCoverage
	width    int
	height   int

	Stats        Statistics
	recentSpawns int
	lvTick       int
	aiCursor     int
}

func NewManager(zm *zoning.ZoneManager, demand *zoning.DemandEngine, services zoning.ServiceCoverage) *Manager {
	h := len(zm.Cells)
	w := len(zm.Cells[0])
	lv := make([][]float32, h)
	for z := range lv {
		lv[z] = make([]float32, w)
	}
	return &Manager{
		byLot:     make(map[int]int),
		landValue: lv,
		zones:     zm,
		demand:    demand,
		services:  services,
		width:     w,
		height:    h,
	}
}

func (m *Manager) Update(dt float64) {
	if m == nil || m.zones == nil || dt <= 0 {
		return
	}
	m.updateLandValueIncremental()
	for i := 0; i < len(m.Buildings); {
		if m.tickBuilding(i, dt) {
			continue
		}
		i++
	}
	m.tickAIBatch()
	m.trySpawn()
	m.refreshStats(dt)
}

func (m *Manager) trySpawn() {
	for _, lot := range m.zones.Lots() {
		if _, used := m.byLot[lot.ID]; used {
			continue
		}
		if !m.zones.CanDevelop(&lot) {
			continue
		}
		m.startConstruction(&lot)
		return
	}
}

func (m *Manager) startConstruction(lot *zoning.ZoneLot) {
	wx, wz := m.zones.CellCenter(lot.X+lot.Width/2, lot.Z+lot.Height/2)
	cells := float32(lot.Width * lot.Height)
	m.nextID++
	b := Building{
		ID:        m.nextID,
		LotID:     lot.ID,
		Type:      lot.Type,
		CellX:     lot.X,
		CellZ:     lot.Z,
		Width:     lot.Width,
		Height:    lot.Height,
		WorldX:    wx,
		WorldZ:    wz,
		Level:     1,
		State:     StateConstructing,
		Stage:     StageFoundation,
		BuildTime: baseConstructSec + cells*4,
	}
	idx := len(m.Buildings)
	m.Buildings = append(m.Buildings, b)
	m.byLot[lot.ID] = idx
	m.recentSpawns++
}

func (m *Manager) tickBuilding(idx int, dt float64) bool {
	b := &m.Buildings[idx]
	b.LandValue = m.landValueAt(b.CellX+b.Width/2, b.CellZ+b.Height/2)

	switch b.State {
	case StateConstructing:
		m.tickConstruction(b, dt)
	case StateOccupied:
		m.tickServices(b)
		m.applyPolicies(b)
		if m.shouldAbandon(b) {
			b.AbandonTimer += float32(dt)
			if b.AbandonTimer >= abandonGraceSec {
				b.State = StateAbandoned
				b.AbandonTimer = 0
			}
		} else {
			b.AbandonTimer = 0
			m.tickUpgrade(b, dt)
			m.tickOccupancy(b, dt)
		}
	case StateAbandoned:
		if !m.shouldAbandon(b) {
			b.State = StateOccupied
			b.DemolishTimer = 0
		} else {
			b.DemolishTimer += float32(dt)
			if b.DemolishTimer >= demolishGraceSec {
				m.removeAt(idx)
				return true
			}
		}
	}
	return false
}

func (m *Manager) tickConstruction(b *Building, dt float64) {
	b.Progress += float32(dt) / b.BuildTime
	switch {
	case b.Progress >= 1:
		b.Progress = 1
		b.Stage = StageCompleted
		b.State = StateOccupied
		m.initOccupancy(b)
	case b.Progress >= stageFrameworkEnd:
		b.Stage = StageCompleted
	case b.Progress >= stageFoundationEnd:
		b.Stage = StageFramework
	default:
		b.Stage = StageFoundation
	}
}

func (m *Manager) removeAt(idx int) {
	lotID := m.Buildings[idx].LotID
	delete(m.byLot, lotID)
	last := len(m.Buildings) - 1
	if idx != last {
		m.Buildings[idx] = m.Buildings[last]
		m.byLot[m.Buildings[idx].LotID] = idx
	}
	m.Buildings = m.Buildings[:last]
}

func (m *Manager) tickUpgrade(b *Building, dt float64) {
	maxLv := m.maxLevelFor(b)
	if b.Level >= maxLv || m.demand == nil {
		return
	}
	if !m.upgradeReady(b) {
		b.UpgradeProgress = 0
		return
	}
	b.UpgradeProgress += float32(dt)
	need := upgradeSecPerLevel * float32(b.Level)
	if b.UpgradeProgress >= need {
		b.Level++
		b.UpgradeProgress = 0
		m.initOccupancy(b)
	}
}

func (m *Manager) upgradeReady(b *Building) bool {
	if m.demand == nil {
		return false
	}
	f := m.demand.Factors
	score := b.LandValue*0.35 +
		m.demand.Education*0.15 +
		f.Happiness*0.12 +
		f.ServiceScore*0.1 +
		coverageStub(f)*0.18 +
		f.ShoppingDemand*0.05 +
		float32(b.Level)*0.02
	threshold := 0.35 + float32(b.Level)*0.12
	return score >= threshold
}

func coverageStub(f zoning.CityFactors) float32 {
	return (f.ServiceScore + (1 - f.Crime)) * 0.5
}

func (m *Manager) shouldAbandon(b *Building) bool {
	if m.services == nil || m.demand == nil {
		return true
	}
	wx, wz := b.WorldX, b.WorldZ
	f := m.demand.Factors
	cat := zoning.ZoneCategoryOf(b.Type)

	switch cat {
	case zoning.CategoryResidential:
		if !b.ServiceOK || !m.services.HasElectricity(wx, wz) || !m.services.HasWater(wx, wz) {
			return true
		}
		if f.Pollution > 0.65 || f.Crime > 0.6 || f.ResidentialTax > 0.16 {
			return true
		}
	case zoning.CategoryCommercial:
		if b.Business.Profitability < -0.2 || f.ShoppingDemand < 0.15 || f.GoodsAvailability < 0.2 || f.WorkerShortage > 0.7 {
			return true
		}
	case zoning.CategoryIndustrial:
		if b.Business.Profitability < -0.2 || f.WorkerShortage > 0.75 || f.FreightCongestion > 0.7 || f.ResourceShortage > 0.65 {
			return true
		}
	case zoning.CategoryOffice:
		if m.demand.Education < 0.2 || f.EducatedWorkers < 0.25 {
			return true
		}
	}
	return false
}

func (m *Manager) landValueAt(cx, cz int) float32 {
	if cx < 0 || cx >= m.width || cz < 0 || cz >= m.height {
		return 0
	}
	return m.landValue[cz][cx]
}

func (m *Manager) Residents() int {
	n := 0
	for i := range m.Buildings {
		b := &m.Buildings[i]
		if b.State != StateOccupied {
			continue
		}
		n += b.Occupancy.Residents
	}
	return n
}

// NearestAt returns the building index within maxDist of world x/z, or -1.
func (m *Manager) NearestAt(x, z float32, maxDist float32) int {
	if m == nil {
		return -1
	}
	maxD := maxDist * maxDist
	best := -1
	for i := range m.Buildings {
		b := &m.Buildings[i]
		dx := b.WorldX - x
		dz := b.WorldZ - z
		if d := dx*dx + dz*dz; d < maxD {
			maxD = d
			best = i
		}
	}
	return best
}

// LandValueGrid returns the cached land-value grid and its size (read-only).
func (m *Manager) LandValueGrid() ([][]float32, int, int) {
	if m == nil {
		return nil, 0, 0
	}
	return m.landValue, m.width, m.height
}

// BuildingAt returns a building by index or nil.
func (m *Manager) BuildingAt(idx int) *Building {
	if m == nil || idx < 0 || idx >= len(m.Buildings) {
		return nil
	}
	return &m.Buildings[idx]
}

func levelCapacity(b *Building) int {
	base := 4
	if b.Type == zoning.ZoneResidentialHigh || b.Type == zoning.ZoneCommercialHigh {
		base = 8
	}
	return base * b.Level * b.Width * b.Height
}

func (m *Manager) Draw(h *terrain.Heightmap) {
	if m == nil {
		return
	}
	cs := terrain.WorldSize / float32(m.width)
	for i := range m.Buildings {
		b := &m.Buildings[i]
		fullHgt := 1.5 + float32(b.Level)*1.2
		hgt := fullHgt
		if b.State == StateConstructing {
			hgt = constructionHeight(b, fullHgt)
		}
		hy := h.WorldHeight(b.WorldX, b.WorldZ) + hgt*0.5
		w := cs * float32(b.Width) * 0.85
		d := cs * float32(b.Height) * 0.85
		col := buildingColor(b)
		rl.DrawCube(rl.NewVector3(b.WorldX, hy, b.WorldZ), w, hgt, d, col)
	}
}

func constructionHeight(b *Building, full float32) float32 {
	switch b.Stage {
	case StageFoundation:
		return full * 0.12
	case StageFramework:
		return full * 0.45
	default:
		return full * clampf(b.Progress, 0.5, 1)
	}
}

func buildingColor(b *Building) rl.Color {
	if b.State == StateAbandoned {
		return rl.NewColor(90, 90, 90, 200)
	}
	if b.State == StateConstructing {
		switch b.Stage {
		case StageFoundation:
			return rl.NewColor(120, 110, 90, 180)
		case StageFramework:
			return rl.NewColor(150, 140, 110, 190)
		default:
			return rl.NewColor(180, 170, 140, 200)
		}
	}
	switch zoning.ZoneCategoryOf(b.Type) {
	case zoning.CategoryResidential:
		return rl.NewColor(80, 200, 80, 220)
	case zoning.CategoryCommercial:
		return rl.NewColor(80, 80, 220, 220)
	case zoning.CategoryIndustrial:
		return rl.NewColor(220, 200, 60, 220)
	case zoning.CategoryOffice:
		return rl.NewColor(200, 200, 220, 220)
	default:
		return rl.NewColor(180, 180, 180, 200)
	}
}

func clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
