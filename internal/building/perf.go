package building

import "github.com/katwate/js-skylines/internal/zoning"

const (
	landValueFullInterval = 30
	aiBatchSize           = 16
)

func (m *Manager) updateLandValueIncremental() {
	if m.zones == nil {
		return
	}
	dirty := m.zones.DrainDirtyLand()
	if len(dirty) > 0 {
		for _, c := range dirty {
			m.calcLandValueAt(c[0], c[1])
			for dz := -1; dz <= 1; dz++ {
				for dx := -1; dx <= 1; dx++ {
					m.calcLandValueAt(c[0]+dx, c[1]+dz)
				}
			}
		}
		return
	}
	m.lvTick++
	if m.lvTick%landValueFullInterval != 0 {
		return
	}
	for z := 0; z < m.height; z++ {
		for x := 0; x < m.width; x++ {
			m.calcLandValueAt(x, z)
		}
	}
}

func (m *Manager) calcLandValueAt(x, z int) {
	if x < 0 || x >= m.width || z < 0 || z >= m.height {
		return
	}
	f := zoning.CityFactors{}
	edu := float32(0)
	if m.demand != nil {
		f = m.demand.Factors
		edu = m.demand.Education
	}
	lv := float32(0.3)
	zt := m.zones.Cells[z][x].Type
	wx, wz := m.zones.CellCenter(x, z)
	if m.zones.RoadConnected(wx, wz) {
		lv += 0.08
	}
	lv += edu * 0.12
	lv += f.ServiceScore * 0.08
	lv += f.Happiness * 0.06
	lv += f.Tourism * 0.04
	if zt == zoning.ZoneIndustrial {
		lv -= 0.15
	}
	lv -= f.Pollution * 0.2
	lv -= f.Crime * 0.15
	lv -= f.FreightCongestion * 0.05
	m.landValue[z][x] = clampf(lv, 0, 1)
}

func (m *Manager) tickAIBatch() {
	n := len(m.Buildings)
	if n == 0 {
		return
	}
	for i := 0; i < aiBatchSize; i++ {
		idx := (m.aiCursor + i) % n
		b := &m.Buildings[idx]
		if b.State == StateOccupied {
			m.tickServices(b)
			m.applyPolicies(b)
			m.evaluateAI(b)
		}
	}
	m.aiCursor = (m.aiCursor + aiBatchSize) % n
}
