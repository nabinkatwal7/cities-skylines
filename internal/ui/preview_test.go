package ui

import (
	"testing"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/sim"
)

func TestEvalPlacementRoadElevationCost(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	sm.Money = 10_000
	ts := NewToolSystem()
	ts.Activate(ToolRoad)
	ts.RoadType = road.RoadTwoLane
	ctx := WorldContext{
		Sim:           sm,
		OnGround:      true,
		PreviewX:      120,
		PreviewZ:      120,
		RoadElevation: 2,
		Tools:         ts,
	}
	p := EvalPlacement(ctx)
	want := sm.RoadPlacementCost(road.RoadTwoLane, 2)
	if p.Cost != want {
		t.Fatalf("preview cost=%v want %v", p.Cost, want)
	}
}

func TestEvalPlacementRejectsBrokeTransport(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	sm.Money = 100
	ts := NewToolSystem()
	ts.Activate(ToolTransport)
	ctx := WorldContext{
		Sim:      sm,
		OnGround: true,
		PreviewX: 50,
		PreviewZ: 50,
		Tools:    ts,
	}
	p := EvalPlacement(ctx)
	if p.Valid {
		t.Fatal("transport preview should be invalid when broke")
	}
}

func TestEvalPlacementRejectsShortRoadSegment(t *testing.T) {
	sm := sim.NewSimulationManager(1)
	sm.Money = 10_000
	ts := NewToolSystem()
	ts.Activate(ToolRoad)
	ctx := WorldContext{
		Sim:        sm,
		OnGround:   true,
		PreviewX:   120,
		PreviewZ:   120,
		RoadActive: true,
		RoadStartX: 120,
		RoadStartZ: 120,
		Tools:      ts,
	}
	p := EvalPlacement(ctx)
	if p.Valid {
		t.Fatal("zero-length road segment should invalidate preview")
	}
}
