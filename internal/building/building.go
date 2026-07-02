package building

import (
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MaxLevel            = 5
	baseConstructSec    = 10.0
	upgradeSecPerLevel  = 20.0
	abandonGraceSec     = 8.0
)

type State uint8

const (
	StateConstructing State = iota
	StateOccupied
	StateAbandoned
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
	Progress float32
	BuildTime float32
	UpgradeProgress float32
	AbandonTimer    float32
	LandValue       float32
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
	m.updateLandValue()
	for i := range m.Buildings {
		m.tickBuilding(&m.Buildings[i], dt)
	}
	m.trySpawn()
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
		BuildTime: baseConstructSec + cells*4,
	}
	idx := len(m.Buildings)
	m.Buildings = append(m.Buildings, b)
	m.byLot[lot.ID] = idx
}

func (m *Manager) tickBuilding(b *Building, dt float64) {
	b.LandValue = m.landValueAt(b.CellX+b.Width/2, b.CellZ+b.Height/2)

	switch b.State {
	case StateConstructing:
		b.Progress += float32(dt) / b.BuildTime
		if b.Progress >= 1 {
			b.Progress = 1
			b.State = StateOccupied
		}
	case StateOccupied:
		if m.shouldAbandon(b) {
			b.AbandonTimer += float32(dt)
			if b.AbandonTimer >= abandonGraceSec {
				b.State = StateAbandoned
				b.AbandonTimer = 0
			}
		} else {
			b.AbandonTimer = 0
			m.tickUpgrade(b, dt)
		}
	case StateAbandoned:
		if !m.shouldAbandon(b) {
			b.State = StateOccupied
		}
	}
}

func (m *Manager) tickUpgrade(b *Building, dt float64) {
	if b.Level >= MaxLevel || m.demand == nil {
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
		if !m.services.HasElectricity(wx, wz) || !m.services.HasWater(wx, wz) {
			return true
		}
		if f.Pollution > 0.65 || f.Crime > 0.6 || f.ResidentialTax > 0.16 {
			return true
		}
	case zoning.CategoryCommercial:
		if f.ShoppingDemand < 0.15 || f.GoodsAvailability < 0.2 || f.WorkerShortage > 0.7 {
			return true
		}
	case zoning.CategoryIndustrial:
		if f.WorkerShortage > 0.75 || f.FreightCongestion > 0.7 || f.ResourceShortage > 0.65 {
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
		n += levelCapacity(b) / 2
	}
	return n
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
		hgt := 1.5 + float32(b.Level)*1.2
		if b.State == StateConstructing {
			hgt *= b.Progress
		}
		hy := h.WorldHeight(b.WorldX, b.WorldZ) + hgt*0.5
		w := cs * float32(b.Width) * 0.85
		d := cs * float32(b.Height) * 0.85
		col := buildingColor(b)
		rl.DrawCube(rl.NewVector3(b.WorldX, hy, b.WorldZ), w, hgt, d, col)
	}
}

func buildingColor(b *Building) rl.Color {
	if b.State == StateAbandoned {
		return rl.NewColor(90, 90, 90, 200)
	}
	if b.State == StateConstructing {
		return rl.NewColor(160, 140, 100, 180)
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
