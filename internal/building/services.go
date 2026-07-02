package building

import "github.com/katwate/js-skylines/internal/zoning"

type Consumption struct {
	Electricity float32
	Water       float32
	Garbage     float32
	Healthcare  float32
	Education   float32
	Goods       float32
	Workers     float32
	Freight     float32
	Internet    float32 // ponytail: reserved for future DLC
}

type BuildingAI struct {
	Occupancy float32
	Resources float32
	Services  float32
	Income    float32
	Expenses  float32
	Happiness float32
	Growth    float32
	Pollution float32
}

func (m *Manager) tickServices(b *Building) {
	if b.State != StateOccupied || m.services == nil || m.demand == nil {
		return
	}
	wx, wz := b.WorldX, b.WorldZ
	f := m.demand.Factors
	scale := float32(b.Level) * 0.2
	cat := zoning.ZoneCategoryOf(b.Type)

	c := &b.Consumption
	c.Electricity = scale
	c.Water = scale * 0.8

	switch cat {
	case zoning.CategoryResidential:
		c.Garbage = float32(b.Occupancy.Residents) * 0.05
		c.Healthcare = float32(b.Occupancy.Residents) * 0.03
		c.Education = float32(b.Occupancy.Residents) * 0.02
	case zoning.CategoryCommercial:
		c.Goods = float32(b.Business.Customers) * 0.04
		c.Workers = float32(b.Business.Employees) * 0.1
	case zoning.CategoryIndustrial:
		c.Workers = float32(b.Business.Employees) * 0.12
		c.Freight = b.Business.Freight * 0.1
	case zoning.CategoryOffice:
		c.Workers = float32(b.Business.Employees) * 0.15
		c.Education = float32(b.Business.Employees) * 0.08
		c.Internet = scale * 0.5
	}

	_ = wx
	_ = wz
	_ = f
	b.ServiceOK = m.services.HasElectricity(wx, wz) && m.services.HasWater(wx, wz)
	if cat == zoning.CategoryResidential {
		b.ServiceOK = b.ServiceOK && f.ServiceScore > 0.3
	}
	if cat == zoning.CategoryCommercial && f.GoodsAvailability < 0.15 {
		b.ServiceOK = false
	}
	if cat == zoning.CategoryOffice && m.demand.Education < 0.15 {
		b.ServiceOK = false
	}
}
