package ui

import (
	"testing"

	"github.com/katwate/js-skylines/internal/sim"
)

func TestToolActivate_singleMode(t *testing.T) {
	ts := NewToolSystem()
	ts.Activate(ToolRoad)
	if ts.Mode != ModePlace || ts.Selected != ToolRoad {
		t.Fatalf("road: mode=%v tool=%v", ts.Mode, ts.Selected)
	}
	ts.Activate(ToolRemove)
	if ts.Mode != ModeBulldoze {
		t.Fatalf("remove: mode=%v", ts.Mode)
	}
	ts.Activate(ToolZone)
	if ts.Mode != ModePaint {
		t.Fatalf("zone: mode=%v", ts.Mode)
	}
	ts.Activate(ToolMeasure)
	if ts.Mode != ModeMeasure {
		t.Fatalf("measure: mode=%v", ts.Mode)
	}
}

func TestInspectorPick_building(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	p := NewInspectorPanel()
	if sm.Buildings == nil {
		t.Skip("no buildings manager")
	}
	p.Pick(sm, 0, 0)
	// no buildings yet — panel stays hidden
	if p.Visible() {
		t.Fatal("expected no selection on empty map")
	}
}
