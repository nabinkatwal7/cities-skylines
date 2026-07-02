package ui

import (
	"math"
	"testing"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/sim"
)

func TestSnapRoadXZGrid(t *testing.T) {
	x, z := SnapRoadXZ(10.2, -7.8)
	if x != 12 || z != -8 {
		t.Fatalf("snap=(%v,%v) want (12,-8)", x, z)
	}
}

func TestSnapRoadFromNodeAngleAndDistance(t *testing.T) {
	x, z := snapRoadFromNode(0, 0, 20, 0)
	if x != 20 || z != 0 {
		t.Fatalf("east snap=(%v,%v) want (20,0)", x, z)
	}
	x, z = snapRoadFromNode(0, 0, 10, 10)
	dist := math.Hypot(float64(x), float64(z))
	if math.Abs(dist-16) > 0.01 {
		t.Fatalf("diagonal dist=%v want 16m (4m grid)", dist)
	}
	angle := math.Atan2(float64(z), float64(x))
	snapped := math.Round(angle/roadAngleSnap) * roadAngleSnap
	if math.Abs(angle-snapped) > 0.01 {
		t.Fatalf("angle %v not snapped to 15° steps (nearest %v)", angle, snapped)
	}
}

func TestSnapRoadReusesNearbyNode(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	n := sm.Roads.AddNode(40, 0, 40)
	ctx := SnapContext{
		Sim:  sm,
		Tool: ToolRoad,
	}
	x, z := snapRoad(ctx, 41.5, 40.5)
	if x != 40 || z != 40 {
		t.Fatalf("snap=(%v,%v) want existing node (40,40) idx=%d", x, z, n)
	}
}

func TestSnapRoadChainsFromStartNode(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	start := sm.Roads.AddNode(0, 0, 0)
	ctx := SnapContext{
		Sim:           sm,
		Tool:          ToolRoad,
		RoadActive:    true,
		RoadStartNode: start,
	}
	x, z := snapRoad(ctx, 11, 3)
	dx := x
	dz := z
	dist := math.Hypot(float64(dx), float64(dz))
	if dist < roadGrid {
		t.Fatalf("chained snap too short: %v", dist)
	}
	angle := math.Atan2(float64(dz), float64(dx))
	snapped := math.Round(angle/roadAngleSnap) * roadAngleSnap
	if math.Abs(angle-snapped) > 0.01 {
		t.Fatalf("chained angle not snapped: %v", angle)
	}
}

func TestRoadToolbarMapsToRoadType(t *testing.T) {
	item := roadOptions()
	if len(item.Options) != len(road.RoadTypeOptions) {
		t.Fatalf("toolbar options=%d road types=%d", len(item.Options), len(road.RoadTypeOptions))
	}
	for i, name := range item.Options {
		if name != road.RoadTypeOptionNames[i] {
			t.Fatalf("option %d name=%q want %q", i, name, road.RoadTypeOptionNames[i])
		}
	}
	if road.RoadTypeFromOptionIndex(5) != road.RoadRoundabout {
		t.Fatal("toolbar index 5 should be roundabout")
	}
	if road.RoadTypeFromOptionIndex(6) != road.RoadSixLane {
		t.Fatal("toolbar index 6 should be six-lane")
	}
}
