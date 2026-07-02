package terrain

import "testing"

func TestWaterWorldToGridCenter(t *testing.T) {
	x, z, ok := waterWorldToGrid(0, 0)
	if !ok {
		t.Fatal("center should map to grid")
	}
	mid := WaterGridSize / 2
	if x != mid || z != mid {
		t.Fatalf("grid=(%d,%d) want (%d,%d)", x, z, mid, mid)
	}
}

func TestFlatMapLandNotFlooded(t *testing.T) {
	FlatTestMap = true
	defer func() { FlatTestMap = false }()
	h := NewGenerator(1).Generate()
	ws := NewWaterSystem()
	ws.Init(h)
	if ws.IsFlooded(200, 200) {
		t.Fatal("dry land should not be flooded")
	}
	if ws.IsFlooded(-200, -200) {
		t.Fatal("dry land should not be flooded")
	}
}

func TestFlatMapRiverIsWet(t *testing.T) {
	FlatTestMap = true
	defer func() { FlatTestMap = false }()
	h := NewGenerator(1).Generate()
	if !h.IsUnderwater(0, 0) {
		t.Fatal("map center river channel should be underwater")
	}
	ws := NewWaterSystem()
	ws.Init(h)
	x, z, ok := waterWorldToGrid(0, 0)
	if !ok {
		t.Fatal("grid lookup failed")
	}
	if ws.Grid[z][x].Height <= 0 {
		t.Fatal("water grid should have depth in river channel")
	}
}
