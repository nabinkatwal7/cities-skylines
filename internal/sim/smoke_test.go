package sim

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/terrain"
)

func TestHeightmapPickXZ(t *testing.T) {
	h := terrain.NewHeightmap()
	for z := 0; z < terrain.HeightmapSize; z++ {
		for x := 0; x < terrain.HeightmapSize; x++ {
			h.Set(x, z, 0.3)
		}
	}

	ray := rl.Ray{
		Position:  rl.NewVector3(0, 100, 0),
		Direction: rl.NewVector3(0, -1, 0),
	}
	x, z, ok := h.PickXZ(ray)
	if !ok {
		t.Fatal("PickXZ expected hit on flat terrain")
	}
	if x < -terrain.WorldSize/2 || x > terrain.WorldSize/2 {
		t.Fatalf("PickXZ x out of bounds: %v", x)
	}
	if z < -terrain.WorldSize/2 || z > terrain.WorldSize/2 {
		t.Fatalf("PickXZ z out of bounds: %v", z)
	}
}

func TestTrafficRuleForNodeBounds(t *testing.T) {
	rm := road.NewRoadManager()
	if got := rm.TrafficRuleForNode(0); got != road.RuleNone {
		t.Fatalf("empty manager: got %v want RuleNone", got)
	}
	if got := rm.TrafficRuleForNode(999); got != road.RuleNone {
		t.Fatalf("out of range node: got %v want RuleNone", got)
	}

	n0 := rm.AddNode(0, 0, 0)
	n1 := rm.AddNode(10, 0, 0)
	n2 := rm.AddNode(0, 0, 10)
	segID := rm.AddSegment(n0, n1, road.RoadTwoLane)
	_ = rm.AddSegment(n1, n2, road.RoadFourLane)

	rm.Nodes[n1].JunctionType = 2
	if got := rm.TrafficRuleForNode(n1); got != road.RuleRoundabout {
		t.Fatalf("roundabout junction: got %v want RuleRoundabout", got)
	}

	rm.Nodes[n1].JunctionType = 1
	rm.Nodes[n1].TrafficLight = road.TrafficLightRed
	if got := rm.TrafficRuleForNode(n1); got != road.RuleTrafficLight {
		t.Fatalf("traffic light node: got %v want RuleTrafficLight", got)
	}

	_ = segID
}

func TestSegmentByID(t *testing.T) {
	rm := road.NewRoadManager()
	if seg := rm.SegmentByID(42); seg != nil {
		t.Fatal("SegmentByID on empty manager should return nil")
	}

	n0 := rm.AddNode(0, 0, 0)
	n1 := rm.AddNode(20, 0, 0)
	id := rm.AddSegment(n0, n1, road.RoadTwoLane)

	seg := rm.SegmentByID(id)
	if seg == nil {
		t.Fatal("SegmentByID expected segment")
	}
	if seg.NodeA != n0 || seg.NodeB != n1 {
		t.Fatalf("segment nodes: got %d,%d want %d,%d", seg.NodeA, seg.NodeB, n0, n1)
	}
	if rm.SegmentByID(id+1000) != nil {
		t.Fatal("SegmentByID missing ID should return nil")
	}
}
