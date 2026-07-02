package ui

import (
	"testing"

	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/sim"
)

func TestUIManagerSubscribesToEvents(t *testing.T) {
	bus := core.NewEventBus()
	m := NewManager()
	m.Attach(bus)

	m.HUD.dirty = false
	bus.Emit(string(core.EventTimeMinute), nil)
	if !m.HUD.dirty {
		t.Fatal("HUD should mark dirty on time minute event")
	}

	m.HUD.dirty = false
	bus.Emit(string(core.EventFloodStarted), nil)
	if len(m.Notifications.Active()) == 0 {
		t.Fatal("flood should enqueue notification")
	}

	m.Detach()
	m.HUD.dirty = false
	bus.Emit(string(core.EventTimeHour), nil)
	if m.HUD.dirty {
		t.Fatal("detached HUD should not receive events")
	}
}

func TestUIManagerSyncView(t *testing.T) {
	m := NewManager()
	m.SyncView(ViewState{Money: 5000, TimeStr: "12:00 PM", Population: 120})
	if m.Money != 5000 || m.HUD.view.Money != 5000 {
		t.Fatalf("sync money: mgr=%v hud=%v", m.Money, m.HUD.view.Money)
	}
	if !m.Unlocks().Unlocked(CatRoads) {
		t.Fatal("roads should always be unlocked")
	}
}

func TestUnlockRegistry_hidesLateCategories(t *testing.T) {
	u := NewUnlockRegistry()
	u.SyncPopulation(10)
	if u.Unlocked(CatHealthcare) {
		t.Fatal("healthcare should be locked at pop 10")
	}
	u.SyncPopulation(400)
	if !u.Unlocked(CatHealthcare) {
		t.Fatal("healthcare should unlock at pop 400")
	}
}

func TestBuildMenus_filterSearch(t *testing.T) {
	u := NewUnlockRegistry()
	u.SyncPopulation(1000)
	b := NewBuildMenus(u)
	b.OpenCategory(CatHealthcare)
	if len(b.filtered) < 2 {
		t.Fatalf("expected health assets, got %d", len(b.filtered))
	}
	b.search = "hospital"
	b.refilter()
	if len(b.filtered) != 1 {
		t.Fatalf("search should narrow to hospital, got %d", len(b.filtered))
	}
}

func TestViewStateFromSim(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	v := ViewStateFromSim(sm, 1, 2, true)
	if v.Money != sm.Money {
		t.Fatalf("money=%v want %v", v.Money, sm.Money)
	}
	if v.DateStr == "" || v.TimeStr == "" {
		t.Fatal("expected date and time strings")
	}
}
