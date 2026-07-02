package building

import "github.com/katwate/js-skylines/internal/zoning"

func (m *Manager) evaluateAI(b *Building) {
	if m.demand == nil {
		return
	}
	f := m.demand.Factors
	cap := float32(levelCapacity(b))
	if cap < 1 {
		cap = 1
	}

	ai := &b.AI
	switch zoning.ZoneCategoryOf(b.Type) {
	case zoning.CategoryResidential:
		ai.Occupancy = float32(b.Occupancy.Residents) / cap
		ai.Resources = f.HouseholdIncome
	case zoning.CategoryCommercial:
		ai.Occupancy = float32(b.Occupancy.Workers+b.Occupancy.Customers) / cap
		ai.Resources = f.GoodsAvailability
	case zoning.CategoryIndustrial:
		ai.Occupancy = float32(b.Occupancy.Workers) / cap
		ai.Resources = 1 - f.ResourceShortage
	case zoning.CategoryOffice:
		ai.Occupancy = float32(b.Occupancy.Employees) / cap
		ai.Resources = f.EducatedWorkers
	}

	if b.ServiceOK {
		ai.Services = f.ServiceScore
	} else {
		ai.Services = 0.2
	}
	ai.Income = b.Business.Profitability*0.5 + f.HouseholdIncome*0.3
	ai.Expenses = b.Consumption.Electricity + b.Consumption.Water + f.ResidentialTax*0.2
	ai.Happiness = b.Household.Happiness
	if ai.Happiness == 0 {
		ai.Happiness = f.Happiness
	}
	ai.Growth = m.demand.Value(zoning.ZoneCategoryOf(b.Type))
	ai.Pollution = f.Pollution
	if b.State == StateAbandoned {
		ai.Growth = -1
	}
}

func (m *Manager) applyPolicies(b *Building) {
	if m.zones == nil {
		return
	}
	p := m.zones.PoliciesAt(b.CellX, b.CellZ)
	if p&zoning.PolicyITCluster != 0 && zoning.ZoneCategoryOf(b.Type) == zoning.CategoryOffice {
		b.Business.Profitability += 0.15
	}
	if p&zoning.PolicyOrganicProduce != 0 && zoning.ZoneCategoryOf(b.Type) == zoning.CategoryCommercial {
		b.Business.Profitability += 0.08
	}
	if p&zoning.PolicySelfSufficientHousing != 0 && zoning.ZoneCategoryOf(b.Type) == zoning.CategoryResidential {
		b.Consumption.Electricity *= 0.7
		b.Consumption.Water *= 0.7
	}
}

func (m *Manager) maxLevelFor(b *Building) int {
	maxLv := MaxLevel
	if m.zones != nil && m.zones.PoliciesAt(b.CellX, b.CellZ)&zoning.PolicyHighRiseBan != 0 {
		maxLv = 3
	}
	return maxLv
}
