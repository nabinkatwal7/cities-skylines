package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/building"
	"github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/zoning"
)

func buildBuildingSelection(sm *sim.SimulationManager, idx int) Selection {
	b := sm.Buildings.BuildingAt(idx)
	if b == nil {
		return Selection{}
	}
	cat := zoning.ZoneCategoryOf(b.Type)
	occ := fmt.Sprintf("%d residents", b.Occupancy.Residents)
	workers := b.Occupancy.Workers + b.Occupancy.Employees
	visitors := b.Occupancy.Customers
	switch cat {
	case zoning.CategoryCommercial, zoning.CategoryIndustrial, zoning.CategoryOffice:
		occ = fmt.Sprintf("%d workers", workers)
		if visitors > 0 {
			occ = fmt.Sprintf("%d workers, %d visitors", workers, visitors)
		}
	}

	svcPct := int(b.AI.Services * 100)
	if svcPct == 0 && b.ServiceOK {
		svcPct = 85
	} else if svcPct == 0 && sm.Demand != nil {
		svcPct = int(sm.Demand.Factors.ServiceScore * 100)
	}

	lines := []string{
		fmt.Sprintf("Level: %d", b.Level),
		fmt.Sprintf("Occupancy: %s", occ),
		fmt.Sprintf("Workers: %d", workers),
		fmt.Sprintf("Visitors: %d", visitors),
		fmt.Sprintf("Maintenance: $%.0f/wk", buildingMaintenance(b)),
		fmt.Sprintf("Electricity: %.1f", b.Consumption.Electricity),
		fmt.Sprintf("Water: %.1f", b.Consumption.Water),
		fmt.Sprintf("Service coverage: %d%%", svcPct),
	}
	for _, w := range buildingWarnings(b, sm) {
		lines = append(lines, "⚠ "+w)
	}
	lines = append(lines,
		fmt.Sprintf("Profitability: %.0f%%", b.Business.Profitability*100),
		fmt.Sprintf("Land value: %.0f", b.LandValue*100),
		fmt.Sprintf("Growth score: %.0f%%", b.AI.Growth*100),
	)

	actions := []InspectorAction{{ID: "inspect", Label: "Inspect"}}
	if b.Level < building.MaxLevel && b.State == building.StateOccupied {
		actions = append(actions, InspectorAction{ID: "upgrade", Label: "Upgrade"})
	}
	if cat == zoning.CategoryResidential {
		actions = append(actions, InspectorAction{ID: "citizen", Label: "View Citizen"})
	}
	actions = append(actions, InspectorAction{ID: "bulldoze", Label: "Bulldoze"})

	return Selection{
		Kind:        InspBuilding,
		Title:       zoneTypeName(b.Type),
		Lines:       lines,
		Actions:     actions,
		buildingIdx: idx,
		followX:     b.WorldX,
		followZ:     b.WorldZ,
	}
}

func buildingMaintenance(b *building.Building) float32 {
	return float32(b.Level*b.Width*b.Height) * 12
}

func buildingWarnings(b *building.Building, sm *sim.SimulationManager) []string {
	var out []string
	if !b.ServiceOK {
		out = append(out, "Service shortage")
	}
	if b.State == building.StateAbandoned {
		out = append(out, "Abandoned")
	}
	if sm != nil && sm.Demand != nil {
		if sm.Demand.Factors.Pollution > 0.6 {
			out = append(out, "High pollution")
		}
		if sm.Demand.Factors.Crime > 0.5 {
			out = append(out, "High crime")
		}
	}
	return out
}

func buildCitizenSelection(sm *sim.SimulationManager, buildingIdx int) Selection {
	b := sm.Buildings.BuildingAt(buildingIdx)
	if b == nil || zoning.ZoneCategoryOf(b.Type) != zoning.CategoryResidential {
		return Selection{}
	}
	hh := &b.Household
	name := fmt.Sprintf("Citizen #%d", b.ID)
	age := 22 + b.Occupancy.Residents%38
	work := "Seeking work"
	if sm.Demand != nil && sm.Demand.Jobs > sm.Demand.Population {
		work = "Employed nearby"
	}
	vehicle := "Walking"
	if hh.Vehicles > 0 {
		vehicle = "Personal car"
	}
	activity := "At home"
	if sm.Time != nil && sm.Time.Hour >= 8 && sm.Time.Hour < 17 {
		activity = "Commuting"
	}
	health := float32(0.7)
	if sm.Demand != nil {
		health = sm.Demand.Factors.ServiceScore
	}
	lines := []string{
		fmt.Sprintf("Age: %d", age),
		fmt.Sprintf("Education: %.0f%%", hh.Education*100),
		fmt.Sprintf("Home: %s", zoneTypeName(b.Type)),
		fmt.Sprintf("Workplace: %s", work),
		fmt.Sprintf("Happiness: %.0f%%", hh.Happiness*100),
		fmt.Sprintf("Health: %.0f%%", health*100),
		fmt.Sprintf("Activity: %s", activity),
		"Travel route: —",
		fmt.Sprintf("Vehicle: %s", vehicle),
	}
	if sm.Time != nil {
		lines = append(lines, fmt.Sprintf("Lifetime: %d days in city", sm.Time.DayCount))
	}
	return Selection{
		Kind:        InspCitizen,
		Title:       name,
		Lines:       lines,
		Actions:     []InspectorAction{{ID: "follow", Label: "Follow"}, {ID: "building", Label: "View Building"}},
		buildingIdx: buildingIdx,
		followX:     b.WorldX,
		followZ:     b.WorldZ,
	}
}
