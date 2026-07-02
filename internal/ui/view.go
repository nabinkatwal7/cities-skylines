package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/sim"
)

// ViewStateFromSim gathers read-only presentation data from the simulation.
func ViewStateFromSim(sm *sim.SimulationManager, mouseX, mouseZ float32, onGround bool) ViewState {
	v := ViewState{
		Money:         sm.Money,
		MouseWorldX:   mouseX,
		MouseWorldZ:   mouseZ,
		MouseOnGround: onGround,
	}
	if sm.Time != nil {
		v.TimeStr = sm.Time.TimeString()
		v.DateStr = formatDate(sm.Time.DayCount)
		if sm.Time.IsPaused {
			v.SpeedStr = "Paused"
			v.TimeStr += " ⏸"
		} else if sm.Time.Speed > 1 {
			v.SpeedStr = fmt.Sprintf("x%.0f", sm.Time.Speed)
			v.TimeStr += fmt.Sprintf(" ⏩%s", v.SpeedStr)
		} else {
			v.SpeedStr = "Normal"
		}
	}
	if sm.Demand != nil {
		v.Happiness = sm.Demand.Factors.Happiness
		v.Population = int(sm.Demand.Population)
	}
	if sm.Buildings != nil && sm.Buildings.Stats.Population > v.Population {
		v.Population = sm.Buildings.Stats.Population
	}
	v.Milestone = milestoneFor(v.Population)
	if sm.Transport != nil && sm.Roads != nil {
		v.WeeklyIncome = sm.Transport.TotalIncome()*0.7 - sm.Roads.TotalMaintenance()*0.5
	}
	return v
}

func formatDate(dayCount int32) string {
	year := dayCount/360 + 1
	day := dayCount % 360
	month := day/30 + 1
	dayOfMonth := day%30 + 1
	return fmt.Sprintf("Y%d M%d D%d", year, month, dayOfMonth)
}

func milestoneFor(pop int) string {
	switch {
	case pop >= 5000:
		return "Metropolis"
	case pop >= 2000:
		return "Big Town"
	case pop >= 800:
		return "Town"
	case pop >= 200:
		return "Village"
	case pop >= 50:
		return "Hamlet"
	default:
		return "Settlement"
	}
}
