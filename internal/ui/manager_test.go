package ui

import (
	"testing"

	"github.com/katwate/js-skylines/internal/core"
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
	if !m.HUD.dirty {
		t.Fatal("HUD should mark dirty on flood event")
	}
	if len(m.Notifications.queue) == 0 {
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
	m.SyncView(ViewState{Money: 5000, TimeStr: "Y1 M1 D1"})
	if m.Money != 5000 || m.HUD.view.Money != 5000 {
		t.Fatalf("sync money: mgr=%v hud=%v", m.Money, m.HUD.view.Money)
	}
}
