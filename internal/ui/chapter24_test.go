package ui

import (
	"testing"

	"github.com/katwate/js-skylines/internal/sim"
)

func TestNotificationsResolveOnRecovery(t *testing.T) {
	n := NewNotifications()
	n.Raise(NotifBudget, "Budget deficit")
	if len(n.Active()) != 1 {
		t.Fatalf("expected one active notification, got %d", len(n.Active()))
	}
	n.Resolve(NotifBudget)
	if len(n.Active()) != 0 {
		t.Fatal("resolved notification should leave active list empty")
	}

	sm := sim.NewSimulationManager(1)
	view := ViewState{Money: -100, WeeklyIncome: -1000, Population: 100}
	n.Refresh(sm, view)
	found := false
	for _, it := range n.Active() {
		if it.Kind == NotifBudget {
			found = true
		}
	}
	if !found {
		t.Fatal("expected budget deficit from Refresh")
	}
}

func TestSearchFindsStreet(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	sm.InitDefaultRoads()
	s := NewSearchSystem()
	s.open = true
	s.query = "road"
	s.Update(sm)
	if len(s.results) == 0 {
		t.Fatal("expected road search hits after InitDefaultRoads")
	}
}

func TestOverlayLayersIndependent(t *testing.T) {
	o := NewOverlayManager()
	o.SetLayer(LayerHeatmap, false)
	if o.LayerOn(LayerHeatmap) {
		t.Fatal("heatmap layer should be off")
	}
	if !o.LayerOn(LayerSelection) {
		t.Fatal("selection layer should remain on")
	}
}

func TestSelectionFromInspector(t *testing.T) {
	sel := NewSelectionSystem()
	sel.FromInspector(Selection{Kind: InspBuilding, followX: 10, followZ: 20})
	if !sel.Active() || sel.kind != SelectBuilding {
		t.Fatalf("selection=%v active=%v", sel.kind, sel.Active())
	}
	x, z, ok := sel.Target()
	if !ok || x != 10 || z != 20 {
		t.Fatalf("target=(%v,%v) ok=%v", x, z, ok)
	}
}
