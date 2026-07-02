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
	tips    []AdvisorTip
	open    bool
	persist bool
	inited  bool
}

func NewAdvisors() *Advisors { return &Advisors{} }

func (a *Advisors) Toggle() { a.open = !a.open; a.ensureInit() }

func (a *Advisors) ensureInit() {
	if a.inited {
		return
	}
	a.inited = true
}

func (a *Advisors) Refresh(sm *sim.SimulationManager, view ViewState) {
	if !a.open {
		return
	}
	a.ensureInit()
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

func (a *Advisors) Draw() {
	if !a.open || len(a.tips) == 0 {
		return
	}
	w, h := int32(400), int32(36+int32(len(a.tips))*40)
	x := int32(10)
	y := int32(TopBarH + 64)
	drawPanel(x, y, w, h)
	drawLabel(T("advisors.title"), x+14, y+10, FontLg, rl.NewColor(255, 220, 150, 255))
	for i, tip := range a.tips {
		drawLabel(advisorNames[tip.Category]+":", x+14, y+36+int32(i*40), FontMd, csBarLine)
		drawLabel(tip.Message, x+14, y+54+int32(i*40), FontSm, csTextDim)
	}
}

func (a *Advisors) HandleInput() {}
