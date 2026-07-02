package sim

import (
	"testing"

	"github.com/katwate/js-skylines/internal/zoning"
)

func TestUndoStackRoadPlace(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.InitDefaultRoads()
	if len(sm.Roads.Segments) == 0 || len(sm.Roads.Nodes) < 2 {
		t.Fatal("need default roads")
	}
	nSeg := len(sm.Roads.Segments)
	money := sm.Money
	seg := sm.Roads.Segments[0]
	a, b := seg.NodeA, seg.NodeB
	na := sm.Roads.Nodes[a]
	nb := sm.Roads.Nodes[b]
	midX := (na.X + nb.X) * 0.5
	midZ := (na.Z + nb.Z) * 0.5
	moneyBefore := sm.Money
	_, segID, ok := sm.PlaceRoadSegment(a, midX, midZ, seg.RoadType, 0)
	if !ok {
		t.Skip("road placement blocked in test terrain; skipping placement undo")
	}
	if len(sm.Roads.Segments) != nSeg+1 {
		t.Fatalf("expected %d segments, got %d", nSeg+1, len(sm.Roads.Segments))
	}
	spent := moneyBefore - sm.Money
	sm.PushRoadPlace(segID, spent)
	if !sm.Undo.Undo(sm) {
		t.Fatal("undo failed")
	}
	if len(sm.Roads.Segments) != nSeg {
		t.Fatalf("undo should restore segment count, got %d", len(sm.Roads.Segments))
	}
	_ = money
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
