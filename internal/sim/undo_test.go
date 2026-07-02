package sim

import (
	"testing"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/zoning"
)

func TestUndoStackRoadPlace(t *testing.T) {
	sm := NewSimulationManager(1)
	nSeg := len(sm.Roads.Segments)
	moneyBefore := sm.Money
	sm.Money = 100_000
	n0 := sm.PlaceRoadNode(120, 120)
	_, segID, ok := sm.PlaceRoadSegment(n0, 136, 120, road.RoadTwoLane, 0)
	if !ok {
		t.Fatal("road placement in open area should succeed")
	}
	if len(sm.Roads.Segments) != nSeg+1 {
		t.Fatalf("expected %d segments, got %d", nSeg+1, len(sm.Roads.Segments))
	}
	spent := moneyBefore - sm.Money
	if spent <= 0 {
		spent = moneyBefore - (sm.Money - road.RoadConstructionCost(road.RoadTwoLane))
	}
	sm.PushRoadPlace(segID, road.RoadConstructionCost(road.RoadTwoLane))
	if !sm.Undo.Undo(sm) {
		t.Fatal("undo failed")
	}
	if len(sm.Roads.Segments) != nSeg {
		t.Fatalf("undo should restore segment count, got %d", len(sm.Roads.Segments))
	}
}

func TestUndoZoneRestore(t *testing.T) {
	sm := NewSimulationManager(1)
	if sm.Zones == nil {
		t.Fatal("zones missing")
	}
	cx, cz := 10, 10
	before := sm.Zones.Cells[cz][cx]
	sm.Zones.SetZoneCell(cx, cz, zoning.ZoneResidentialLow)
	sm.PushZoneChange(cx, cz, before)
	if sm.Zones.Cells[cz][cx].Type != zoning.ZoneResidentialLow {
		t.Fatal("zone should be set")
	}
	if !sm.Undo.Undo(sm) {
		t.Fatal("undo failed")
	}
	if sm.Zones.Cells[cz][cx].Type != before.Type {
		t.Fatalf("zone type=%v want %v", sm.Zones.Cells[cz][cx].Type, before.Type)
	}
}

func TestUndoClearOnLimit(t *testing.T) {
	u := NewUndoStack()
	sm := NewSimulationManager(2)
	for i := 0; i < undoLimit+5; i++ {
		v := i
		u.Push(func(s *SimulationManager) { _ = v })
	}
	if len(u.entries) != undoLimit {
		t.Fatalf("stack len=%d want %d", len(u.entries), undoLimit)
	}
	u.Clear()
	if u.CanUndo() {
		t.Fatal("clear should empty stack")
	}
	_ = sm
}
