package road

import "testing"

func TestRoadTypeFromOptionIndex(t *testing.T) {
	cases := []struct {
		idx  int
		want RoadType
	}{
		{0, RoadTwoLane},
		{4, RoadHighway},
		{5, RoadRoundabout},
		{6, RoadSixLane},
		{7, RoadAvenue},
		{14, RoadQuay},
	}
	for _, c := range cases {
		if got := RoadTypeFromOptionIndex(c.idx); got != c.want {
			t.Fatalf("index %d: got %v want %v", c.idx, got, c.want)
		}
	}
	if got := RoadTypeFromOptionIndex(99); got != RoadTwoLane {
		t.Fatalf("out of range should default to two-lane, got %v", got)
	}
}

func TestOptionIndexForRoadType(t *testing.T) {
	if OptionIndexForRoadType(RoadRoundabout) != 5 {
		t.Fatalf("roundabout index=%d want 5", OptionIndexForRoadType(RoadRoundabout))
	}
	if OptionIndexForRoadType(RoadSixLane) != 6 {
		t.Fatalf("six-lane index=%d want 6", OptionIndexForRoadType(RoadSixLane))
	}
}

func TestRoadTypeOptionNamesAlignWithOptions(t *testing.T) {
	if len(RoadTypeOptionNames) != len(RoadTypeOptions) {
		t.Fatalf("names=%d options=%d", len(RoadTypeOptionNames), len(RoadTypeOptions))
	}
	for i, rt := range RoadTypeOptions {
		if RoadTypeName(rt) == "" || RoadTypeName(rt) == "Road" {
			t.Fatalf("option %d (%s) missing name", i, RoadTypeOptionNames[i])
		}
	}
}

func TestHasSegmentBetween(t *testing.T) {
	rm := NewRoadManager()
	a := rm.AddNode(0, 0, 0)
	b := rm.AddNode(8, 0, 0)
	if rm.HasSegmentBetween(a, b) {
		t.Fatal("no segment yet")
	}
	rm.AddSegment(a, b, RoadTwoLane)
	if !rm.HasSegmentBetween(a, b) {
		t.Fatal("should find forward segment")
	}
	if !rm.HasSegmentBetween(b, a) {
		t.Fatal("should find reverse segment")
	}
}

func TestAddSegmentQueuesMeshRebuild(t *testing.T) {
	rm := NewRoadManager()
	a := rm.AddNode(0, 0, 0)
	b := rm.AddNode(12, 0, 0)
	rm.AddSegment(a, b, RoadTwoLane)
	if rm.PendingMeshRebuilds() != 1 {
		t.Fatalf("pending rebuilds=%d want 1", rm.PendingMeshRebuilds())
	}
	if len(rm.Models) != 1 {
		t.Fatalf("models=%d want 1", len(rm.Models))
	}
}

func TestClearModelsResetsSlice(t *testing.T) {
	rm := NewRoadManager()
	a := rm.AddNode(0, 0, 0)
	b := rm.AddNode(8, 0, 0)
	rm.AddSegment(a, b, RoadTwoLane)
	if len(rm.Models) != 1 {
		t.Fatalf("models=%d want 1", len(rm.Models))
	}
	rm.ClearModels()
	if len(rm.Models) != 0 {
		t.Fatalf("models=%d want 0 after clear", len(rm.Models))
	}
	if rm.PendingMeshRebuilds() != 0 {
		t.Fatal("dirty list should be cleared")
	}
}

func TestValidNodeIndex(t *testing.T) {
	rm := NewRoadManager()
	idx := rm.AddNode(1, 0, 1)
	if !rm.ValidNodeIndex(idx) {
		t.Fatal("valid node")
	}
	if rm.ValidNodeIndex(99) {
		t.Fatal("out of range")
	}
}

func TestSegmentIndex(t *testing.T) {
	rm := NewRoadManager()
	a := rm.AddNode(0, 0, 0)
	b := rm.AddNode(4, 0, 0)
	id := rm.AddSegment(a, b, RoadTwoLane)
	if rm.SegmentIndex(id) != 0 {
		t.Fatalf("index=%d want 0", rm.SegmentIndex(id))
	}
	if rm.SegmentIndex(9999) != -1 {
		t.Fatal("missing id should be -1")
	}
}
