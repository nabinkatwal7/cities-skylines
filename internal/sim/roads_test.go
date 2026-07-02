package sim

import (
	"testing"

	"github.com/katwate/js-skylines/internal/road"
)

func TestPlaceRoadStartNodeDoesNotReuseNearby(t *testing.T) {
	sm := NewSimulationManager(1)
	existing := sm.Roads.AddNode(120, 0, 120)
	before := len(sm.Roads.Nodes)
	got := sm.PlaceRoadStartNode(121.5, 119.5)
	if got == existing {
		t.Fatalf("PlaceRoadStartNode should not reuse nearby node %d", existing)
	}
	if len(sm.Roads.Nodes) != before+1 {
		t.Fatalf("nodes=%d want %d", len(sm.Roads.Nodes), before+1)
	}
}

func TestPlaceRoadNodeReusesNearbyNode(t *testing.T) {
	sm := NewSimulationManager(1)
	before := len(sm.Roads.Nodes)
	existing := sm.Roads.AddNode(120, 0, 120)
	got := sm.PlaceRoadNode(121.5, 119.5)
	if got != existing {
		t.Fatalf("PlaceRoadNode=%d want existing %d", got, existing)
	}
	if len(sm.Roads.Nodes) != before+1 {
		t.Fatalf("nodes=%d want %d (no duplicate)", len(sm.Roads.Nodes), before+1)
	}
}

func TestPlaceRoadSegmentOpenArea(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.Money = 100_000
	n0 := sm.PlaceRoadNode(120, 120)
	n1, segID, ok := sm.PlaceRoadSegment(n0, 136, 120, road.RoadTwoLane, 0)
	if !ok {
		t.Fatal("expected placement in open area")
	}
	if segID == ^uint32(0) {
		t.Fatal("invalid segment id")
	}
	if n1 == n0 {
		t.Fatal("end node should differ from start")
	}
	if len(sm.Roads.Segments) != 1 {
		t.Fatalf("segments=%d want 1", len(sm.Roads.Segments))
	}
	if sm.Roads.PendingMeshRebuilds() != 1 {
		t.Fatalf("pending mesh rebuilds=%d want 1", sm.Roads.PendingMeshRebuilds())
	}
	_ = n1
}

func TestPlaceRoadSegmentRejectsDuplicate(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.Money = 100_000
	n0 := sm.PlaceRoadNode(140, 140)
	_, _, ok := sm.PlaceRoadSegment(n0, 156, 140, road.RoadTwoLane, 0)
	if !ok {
		t.Fatal("first segment should place")
	}
	_, _, ok = sm.PlaceRoadSegment(n0, 156, 140, road.RoadTwoLane, 0)
	if ok {
		t.Fatal("duplicate segment between same nodes should fail")
	}
}

func TestPlaceRoadSegmentRejectsTooShort(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.Money = 100_000
	n0 := sm.PlaceRoadNode(160, 160)
	_, _, ok := sm.PlaceRoadSegment(n0, 160.5, 160.5, road.RoadTwoLane, 0)
	if ok {
		t.Fatal("sub-2m segment should fail")
	}
}

func TestRoadPlacementCost(t *testing.T) {
	sm := NewSimulationManager(1)
	base := sm.RoadPlacementCost(road.RoadHighway, 0)
	if base != 500 {
		t.Fatalf("highway cost=%v want 500", base)
	}
	bridge := sm.RoadPlacementCost(road.RoadTwoLane, 2)
	if bridge != 200 {
		t.Fatalf("bridge cost=%v want 200", bridge)
	}
	tunnel := sm.RoadPlacementCost(road.RoadTwoLane, -1)
	if tunnel != 200 {
		t.Fatalf("tunnel cost=%v want 200", tunnel)
	}
}

func TestCanPlaceRoadAllowsJunctionAtNode(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.InitDefaultRoads()
	if len(sm.Roads.Nodes) < 2 {
		t.Fatal("need default roads")
	}
	n0 := sm.Roads.Nodes[0]
	n1 := sm.Roads.Nodes[1]
	reason := sm.CanPlaceRoad(n0.X, n0.Z, n1.X, n1.Z, road.RoadTwoLane, 0, ^uint32(0))
	if reason != "" {
		t.Fatalf("junction along existing nodes blocked: %q", reason)
	}
}

func TestRoadUpgradeMarksMeshDirty(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.Money = 100_000
	n0 := sm.PlaceRoadNode(180, 180)
	_, _, ok := sm.PlaceRoadSegment(n0, 196, 180, road.RoadTwoLane, 0)
	if !ok {
		t.Fatal("setup segment failed")
	}
	sm.Roads.UpgradeSegment(0, road.RoadFourLane)
	if sm.Roads.Segments[0].RoadType != road.RoadFourLane {
		t.Fatalf("type=%v", sm.Roads.Segments[0].RoadType)
	}
	if sm.Roads.PendingMeshRebuilds() != 1 {
		t.Fatalf("pending rebuilds=%d want 1", sm.Roads.PendingMeshRebuilds())
	}
}

func TestRoadTypeToolbarMatchesEnum(t *testing.T) {
	if road.RoadTypeFromOptionIndex(5) != road.RoadRoundabout {
		t.Fatal("index 5 must be roundabout (was off-by-one before fix)")
	}
	if road.RoadTypeFromOptionIndex(7) != road.RoadAvenue {
		t.Fatal("index 7 must be avenue")
	}
}
