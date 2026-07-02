package building

import "github.com/katwate/js-skylines/internal/zoning"

func (m *Manager) updateLandValue() {
	if m.zones == nil {
		return
	}
	f := zoning.CityFactors{}
	edu := float32(0)
	if m.demand != nil {
		f = m.demand.Factors
		edu = m.demand.Education
	}

	for z := 0; z < m.height; z++ {
		for x := 0; x < m.width; x++ {
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
