package ui

import (
	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AdvisorCategory int

const (
	AdvEconomy AdvisorCategory = iota
	AdvTransport
	AdvUtilities
	AdvHealthcare
	AdvEducation
	AdvIndustry
	AdvEnvironment
)

var advisorNames = []string{"Economy", "Transport", "Utilities", "Health", "Education", "Industry", "Environment"}

// AdvisorTip is one contextual recommendation (24.14).
type AdvisorTip struct {
	Category AdvisorCategory
	Message  string
}

// Advisors surfaces city guidance from live simulation data (24.14).
type Advisors struct {
	tips []AdvisorTip
	open bool
}

func NewAdvisors() *Advisors { return &Advisors{} }

func (a *Advisors) Refresh(sm *sim.SimulationManager, view ViewState) {
	a.tips = a.tips[:0]
	if sm == nil {
		return
	}
	if view.WeeklyIncome < 0 {
		a.add(AdvEconomy, "Raise taxes or cut road maintenance to balance the budget.")
	}
	if view.Money < 5000 {
		a.add(AdvEconomy, "Treasury is low — avoid large projects until income recovers.")
	}
	if sm.Demand != nil && sm.Demand.Factors.WorkerShortage > 0.4 {
		a.add(AdvIndustry, "Worker shortage — zone more residential or improve education.")
	}
	if sm.Demand != nil && sm.Demand.Factors.FreightCongestion > 0.5 {
		a.add(AdvTransport, "Freight congestion — add industrial road links or cargo stations.")
	}
	if sm.Services != nil && !sm.Services.Electricity {
		a.add(AdvUtilities, "Build power production to restore electricity.")
	}
	if sm.Services != nil && !sm.Services.Water {
		a.add(AdvUtilities, "Expand water pumping and pipe coverage.")
	}
	if sm.Demand != nil && sm.Demand.Factors.ServiceScore < 0.4 {
		a.add(AdvHealthcare, "Healthcare coverage is weak — place clinics near residential zones.")
	}
	if sm.Demand != nil && sm.Demand.Education < 0.35 {
		a.add(AdvEducation, "Education level is low — build schools to unlock office demand.")
	}
	if sm.Demand != nil && sm.Demand.Industrial < 0.2 {
		a.add(AdvIndustry, "Industrial demand is soft — ensure freight and worker supply.")
	}
	if sm.Demand != nil && sm.Demand.Factors.Pollution > 0.55 {
		a.add(AdvEnvironment, "Pollution is high — add parks and relocate heavy industry.")
	}
	if len(a.tips) == 0 {
		a.add(AdvEconomy, "City is stable. Keep monitoring utilities and demand.")
	}
}

func (a *Advisors) add(cat AdvisorCategory, msg string) {
	a.tips = append(a.tips, AdvisorTip{Category: cat, Message: msg})
	if len(a.tips) > 5 {
		a.tips = a.tips[len(a.tips)-5:]
	}
}

func (a *Advisors) Toggle() { a.open = !a.open }

func (a *Advisors) Draw() {
	if !a.open || len(a.tips) == 0 {
		return
	}
	w, h := int32(360), int32(28+int32(len(a.tips))*36)
	x := int32(8)
	y := int32(TopBarH + 60)
	rl.DrawRectangle(x, y, w, h, rl.NewColor(0, 0, 0, 210))
	rl.DrawRectangleLines(x, y, w, h, rl.NewColor(180, 160, 100, 200))
	DrawUIText(T("advisors.title"), x+10, y+8, 15, rl.NewColor(255, 220, 150, 230))
	for i, tip := range a.tips {
		DrawUIText(advisorNames[tip.Category]+":", x+10, y+28+int32(i*36), 13, rl.NewColor(180, 200, 255, 220))
		DrawUIText(tip.Message, x+10, y+44+int32(i*36), 12, rl.LightGray)
	}
}

func (a *Advisors) HandleInput() {}
