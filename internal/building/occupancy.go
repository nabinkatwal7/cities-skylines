package building

import (
	"github.com/katwate/js-skylines/internal/zoning"
)

type Household struct {
	Members   int
	Wealth    float32
	Education float32
	Vehicles  int
	Happiness float32
}

type Business struct {
	Employees     int
	Production    float32
	Storage       float32
	Freight       float32
	Profitability float32
	Customers     int
}

type Occupancy struct {
	Residents  int
	Vacancies  int
	Workers    int
	Customers  int
	Production float32
	Employees  int
}

func (m *Manager) initOccupancy(b *Building) {
	cap := levelCapacity(b)
	cat := zoning.ZoneCategoryOf(b.Type)
	switch cat {
	case zoning.CategoryResidential:
		b.Household.Members = cap / 4
		if b.Household.Members < 1 {
			b.Household.Members = 1
		}
		b.Occupancy.Residents = b.Household.Members
		b.Occupancy.Vacancies = cap - b.Occupancy.Residents
	case zoning.CategoryCommercial:
		b.Business.Employees = cap / 6
		b.Business.Customers = cap / 2
		b.Occupancy.Workers = b.Business.Employees
		b.Occupancy.Customers = b.Business.Customers
	case zoning.CategoryIndustrial:
		b.Business.Employees = cap / 5
		b.Occupancy.Workers = b.Business.Employees
		b.Occupancy.Production = b.Business.Production
	case zoning.CategoryOffice:
		b.Business.Employees = cap / 8
		b.Occupancy.Employees = b.Business.Employees
	}
}

func (m *Manager) tickOccupancy(b *Building, dt float64) {
	if b.State != StateOccupied || m.demand == nil {
		return
	}
	f := m.demand.Factors
	cap := levelCapacity(b)
	cat := zoning.ZoneCategoryOf(b.Type)

	switch cat {
	case zoning.CategoryResidential:
		m.tickHousehold(b, cap, f, dt)
	case zoning.CategoryCommercial, zoning.CategoryIndustrial, zoning.CategoryOffice:
		m.tickBusiness(b, cap, f, dt)
	}
}

func (m *Manager) tickHousehold(b *Building, cap int, f zoning.CityFactors, dt float64) {
	hh := &b.Household
	target := cap / 3
	if target < 1 {
		target = 1
	}
	if hh.Members < target && m.demand.HasDemand(zoning.CategoryResidential) {
		hh.Members++
	}
	hh.Wealth = f.HouseholdIncome
	hh.Education = m.demand.Education
	hh.Happiness = clampf(f.Happiness*0.6+b.LandValue*0.25+(1-f.Pollution)*0.15, 0, 1)
	hh.Vehicles = hh.Members / 2
	b.Occupancy.Residents = hh.Members
	if cap > hh.Members {
		b.Occupancy.Vacancies = cap - hh.Members
	} else {
		b.Occupancy.Vacancies = 0
	}
	_ = dt
}

func (m *Manager) tickBusiness(b *Building, cap int, f zoning.CityFactors, dt float64) {
	biz := &b.Business
	cat := zoning.ZoneCategoryOf(b.Type)
	workerCap := cap / 4
	if workerCap < 1 {
		workerCap = 1
	}

	switch cat {
	case zoning.CategoryCommercial:
		biz.Employees = scaleTo(workerCap, f.ShoppingDemand, f.WorkerShortage)
		biz.Customers = int(float32(cap/2) * f.ShoppingDemand)
		biz.Production = 0
		biz.Storage = float32(biz.Customers) * 0.1
		biz.Freight = 0
		biz.Profitability = evaluateProfit(f.ShoppingDemand, f.GoodsAvailability, f.WorkerShortage, f.ResidentialTax)
		b.Occupancy.Workers = biz.Employees
		b.Occupancy.Customers = biz.Customers
	case zoning.CategoryIndustrial:
		biz.Employees = scaleTo(workerCap, 1-f.WorkerShortage, f.WorkerShortage)
		biz.Production = float32(biz.Employees) * f.GoodsShortage * 0.5
		biz.Storage = biz.Production * 0.3
		biz.Freight = biz.Production * (1 - f.FreightCongestion)
		biz.Customers = 0
		biz.Profitability = evaluateProfit(f.GoodsShortage, f.ExportOpportunity, f.WorkerShortage, f.IndustrialTax)
		b.Occupancy.Workers = biz.Employees
		b.Occupancy.Production = biz.Production
	case zoning.CategoryOffice:
		biz.Employees = scaleTo(workerCap, f.EducatedWorkers, 1-f.EducatedWorkers)
		biz.Production = float32(biz.Employees) * m.demand.Education * 0.2
		biz.Storage = 0
		biz.Freight = 0
		biz.Customers = 0
		biz.Profitability = evaluateProfit(f.EducatedWorkers, f.HighTechEconomy, 0, f.IndustrialTax)
		b.Occupancy.Employees = biz.Employees
	}
	_ = dt
}

func scaleTo(cap int, demand, shortage float32) int {
	n := int(float32(cap) * demand * (1 - shortage*0.5))
	if n < 0 {
		return 0
	}
	if n > cap {
		return cap
	}
	return n
}

func evaluateProfit(positive, secondary, penalty, tax float32) float32 {
	return clampf(positive*0.5+secondary*0.3-penalty*0.35-(tax-0.1)*0.5, -1, 1)
}
