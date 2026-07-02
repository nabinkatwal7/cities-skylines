package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type NotificationKind string

const (
	NotifNoPower      NotificationKind = "no_power"
	NotifNoWater      NotificationKind = "no_water"
	NotifCrime        NotificationKind = "crime"
	NotifFire         NotificationKind = "fire"
	NotifDeathWave    NotificationKind = "death_wave"
	NotifTraffic      NotificationKind = "traffic"
	NotifAbandoned    NotificationKind = "abandoned"
	NotifBudget       NotificationKind = "budget"
	NotifMilestone    NotificationKind = "milestone"
	NotifFlood        NotificationKind = "flood"
)

// Notification is a persistent alert until resolved or dismissed (24.13).
type Notification struct {
	Kind      NotificationKind
	Message   string
	CreatedAt int32
	Resolved  bool
	Dismissed bool
}

// Notifications reports simulation events (24.13).
type Notifications struct {
	items      []Notification
	lastPop    int
	pending    int
	batchEvery int
}

func NewNotifications() *Notifications {
	return &Notifications{batchEvery: 1}
}

func (n *Notifications) MarkDirty() { n.pending++ }

func (n *Notifications) Subscribe(bus *core.EventBus) []func() {
	if bus == nil {
		return nil
	}
	add := func(kind NotificationKind, msg string) {
		n.Raise(kind, msg)
	}
	return []func(){
		bus.On(string(core.EventFloodStarted), func(any) { add(NotifFlood, "Flood warning") }),
		bus.On(string(core.EventFloodReceded), func(any) { n.Resolve(NotifFlood) }),
		bus.On(string(core.EventTimeDay), func(any) { n.tickDay() }),
	}
}

func (n *Notifications) Raise(kind NotificationKind, msg string) {
	for i := range n.items {
		if n.items[i].Kind == kind && !n.items[i].Dismissed {
			n.items[i].Message = msg
			n.items[i].Resolved = false
			return
		}
	}
	n.items = append(n.items, Notification{Kind: kind, Message: msg, CreatedAt: int32(time.Now().Unix())})
	n.pending++
}

func (n *Notifications) Resolve(kind NotificationKind) {
	for i := range n.items {
		if n.items[i].Kind == kind {
			n.items[i].Resolved = true
		}
	}
}

func (n *Notifications) Dismiss(kind NotificationKind) {
	for i := range n.items {
		if n.items[i].Kind == kind {
			n.items[i].Dismissed = true
		}
	}
}

func (n *Notifications) Active() []Notification {
	out := make([]Notification, 0, len(n.items))
	for _, it := range n.items {
		if !it.Dismissed && !it.Resolved {
			out = append(out, it)
		}
	}
	return out
}

// Refresh evaluates live simulation conditions (presentation only).
// Batched: skips work unless pending or forced (24.26).
func (n *Notifications) Refresh(sm *sim.SimulationManager, view ViewState, force bool) {
	if sm == nil {
		return
	}
	if !force && n.pending == 0 {
		return
	}
	n.pending = 0
	if sm.Services != nil && !sm.Services.Electricity {
		n.Raise(NotifNoPower, "No electricity")
	} else {
		n.Resolve(NotifNoPower)
	}
	if sm.Services != nil && !sm.Services.Water {
		n.Raise(NotifNoWater, "No water")
	} else {
		n.Resolve(NotifNoWater)
	}
	if sm.Demand != nil {
		if sm.Demand.Factors.Crime > 0.55 {
			n.Raise(NotifCrime, "Crime rising")
		} else {
			n.Resolve(NotifCrime)
		}
		if sm.Demand.Factors.DeathWave > 0.25 {
			n.Raise(NotifDeathWave, "Death wave active")
		} else {
			n.Resolve(NotifDeathWave)
		}
		if sm.Demand.Factors.FreightCongestion > 0.6 {
			n.Raise(NotifTraffic, "Traffic congestion")
		} else {
			n.Resolve(NotifTraffic)
		}
	}
	if sm.Buildings != nil && sm.Buildings.Stats.Abandonment > 0 {
		n.Raise(NotifAbandoned, fmt.Sprintf("%d buildings abandoned", sm.Buildings.Stats.Abandonment))
	} else {
		n.Resolve(NotifAbandoned)
	}
	if view.Money < 0 || view.WeeklyIncome < -500 {
		n.Raise(NotifBudget, "Budget deficit")
	} else {
		n.Resolve(NotifBudget)
	}
	if view.Population > n.lastPop && milestoneFor(view.Population) != milestoneFor(n.lastPop) {
		n.Raise(NotifMilestone, "Milestone: "+view.Milestone)
	}
	n.lastPop = view.Population
}

func (n *Notifications) tickDay() {
	// ponytail: fire warnings from disaster system when wired
}

func (n *Notifications) HandleClick(mx, my int32) bool {
	active := n.Active()
	if len(active) == 0 {
		return false
	}
	x := int32(ScreenW - 300)
	y := int32(TopBarH + 4)
	for i, it := range active {
		rowY := y + int32(i*22)
		if mx >= x+270 && mx < x+290 && my >= rowY && my < rowY+18 {
			n.Dismiss(it.Kind)
			return true
		}
	}
	return false
}

func (n *Notifications) DrawHUD(y int32) {
	active := n.Active()
	if len(active) == 0 {
		return
	}
	drawLabel(active[len(active)-1].Message, ScreenW-360, y, FontMd, rl.NewColor(255, 210, 120, 255))
}

func (n *Notifications) Draw() {
	active := n.Active()
	if len(active) == 0 {
		return
	}
	x := int32(ScreenW - 320)
	y := int32(TopBarH + 8)
	drawPanel(x, y, 312, int32(len(active)*26+12))
	for i, it := range active {
		rowY := y + int32(i*26) + 8
		col := rl.NewColor(255, 200, 110, 255)
		if strings.Contains(it.Message, "Milestone") {
			col = csHappy
		}
		drawLabel(it.Message, x+10, rowY, FontMd, col)
		drawLabel("×", x+286, rowY, FontLg, csTextDim)
	}
}

func (n *Notifications) Push(msg string) {
	n.Raise(NotificationKind("event"), msg)
}
