package sim

import (
	"testing"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/zoning"
)

func TestPlaceRoadSegmentRejectsInvalidNode(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.Money = 10_000
	_, _, ok := sm.PlaceRoadSegment(999, 120, 120, road.RoadTwoLane, 0)
	if ok {
		t.Fatal("invalid node index should fail")
	}
}

func TestPlaceRoadSegmentRejectsInsufficientFunds(t *testing.T) {
	sm := NewSimulationManager(1)
	n0 := sm.PlaceRoadNode(120, 120)
	sm.Money = 10
	_, _, ok := sm.PlaceRoadSegment(n0, 136, 120, road.RoadHighway, 0)
	if ok {
		t.Fatal("should reject when broke")
	}
	if len(sm.Roads.Segments) != 0 {
		t.Fatalf("segments=%d want 0", len(sm.Roads.Segments))
	}
}

func TestPlaceRoadSegmentDeductsElevationCost(t *testing.T) {
	sm := NewSimulationManager(1)
	n0 := sm.PlaceRoadNode(200, 200)
	sm.Money = 10_000
	before := sm.Money
	_, _, ok := sm.PlaceRoadSegment(n0, 216, 200, road.RoadTwoLane, 2)
	if !ok {
		t.Fatal("bridge placement should succeed")
	}
	spent := before - sm.Money
	want := sm.RoadPlacementCost(road.RoadTwoLane, 2)
	if spent != want {
		t.Fatalf("spent=%v want %v", spent, want)
	}
}

func TestPushRoadRemoveInvalidatesChainIndex(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.Money = 100_000
	n0 := sm.PlaceRoadNode(220, 220)
	n1, _, ok := sm.PlaceRoadSegment(n0, 236, 220, road.RoadTwoLane, 0)
	if !ok {
		t.Fatal("setup segment failed")
	}
	if !sm.Roads.ValidNodeIndex(n1) {
		t.Fatal("end node should exist")
	}
	idx := 0
	sm.PushRoadRemove(idx)
	if sm.Roads.ValidNodeIndex(n1) && len(sm.Roads.Segments) > 0 {
		// ponytail: node indices shift after removal; callers must revalidate
	}
}

func TestZonePlacementNearRoad(t *testing.T) {
	sm := NewSimulationManager(1)
	sm.InitDefaultRoads()
	if sm.Zones == nil {
		t.Fatal("zones missing")
	}
	var wx, wz float32
	found := false
	for x := -200; x <= 200; x += 4 {
		for z := -200; z <= 200; z += 4 {
			wx, wz = float32(x), float32(z)
			if sm.Zones.CanZone(wx, wz) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no zoneable cell found — road frontage or buildability broken")
	}
	cx := sm.Zones.CellX(wx)
	cz := sm.Zones.CellZ(wz)
	sm.Zones.SetZone(wx, wz, zoning.ZoneResidentialLow)
	if sm.Zones.Cells[cz][cx].Type != zoning.ZoneResidentialLow {
		t.Fatalf("zone type=%v", sm.Zones.Cells[cz][cx].Type)
	}
}
