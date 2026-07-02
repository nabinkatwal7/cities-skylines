package ui

import (
	"testing"

	"github.com/katwate/js-skylines/internal/sim"
)

func TestInfoViews_singleActive(t *testing.T) {
	v := NewInfoViews()
	v.Set(ViewPollution)
	if v.Active() != ViewPollution {
		t.Fatal("expected pollution view")
	}
	v.Set(ViewTraffic)
	if v.Active() != ViewTraffic {
		t.Fatal("only one view should be active")
	}
	v.Clear()
	if v.Active() != ViewNone {
		t.Fatal("clear should reset view")
	}
}

func TestInfoViews_cycle(t *testing.T) {
	v := NewInfoViews()
	v.Cycle()
	if v.Active() != ViewElectricity {
		t.Fatalf("cycle start: got %v", v.Active())
	}
}

func TestBuildBuildingSelection(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	sel := buildBuildingSelection(sm, 0)
	if sel.Kind != InspNone && sel.Title == "" {
		t.Fatal("unexpected empty title")
	}
	_ = sel
}
