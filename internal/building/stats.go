package building

import "github.com/katwate/js-skylines/internal/zoning"

type Statistics struct {
	Population      int
	VacantHomes     int
	VacantJobs      int
	DemandRes       float32
	DemandCom       float32
	DemandInd       float32
	DemandOffice    float32
	AvgLevel        float32
	AvgLandValue    float32
	DevelopmentRate float32
	Abandonment     int
}

func (m *Manager) refreshStats(dt float64) {
	if m == nil {
		return
	}
	s := &m.Stats
	s.Population = 0
	s.VacantHomes = 0
	s.VacantJobs = 0
	s.Abandonment = 0
	var levelSum float32
	var lvSum float32
	var n int

	for i := range m.Buildings {
		b := &m.Buildings[i]
		if b.State == StateAbandoned {
			s.Abandonment++
		}
		if b.State != StateOccupied {
			continue
		}
		n++
		levelSum += float32(b.Level)
		lvSum += b.LandValue
		switch zoning.ZoneCategoryOf(b.Type) {
		case zoning.CategoryResidential:
			s.Population += b.Occupancy.Residents
			s.VacantHomes += b.Occupancy.Vacancies
		case zoning.CategoryCommercial, zoning.CategoryIndustrial:
			cap := levelCapacity(b) / 4
			s.VacantJobs += cap - b.Occupancy.Workers
		case zoning.CategoryOffice:
			cap := levelCapacity(b) / 8
			s.VacantJobs += cap - b.Occupancy.Employees
		}
	}
	if n > 0 {
		s.AvgLevel = levelSum / float32(n)
		s.AvgLandValue = lvSum / float32(n)
	}
	if m.demand != nil {
		s.DemandRes = m.demand.Residential
		s.DemandCom = m.demand.Commercial
		s.DemandInd = m.demand.Industrial
		s.DemandOffice = m.demand.Office
	}
	if dt > 0 {
		s.DevelopmentRate = float32(m.recentSpawns) / float32(dt)
		m.recentSpawns = 0
	}
}
