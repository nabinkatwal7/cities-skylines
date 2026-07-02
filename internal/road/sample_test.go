package road

import (
	"math"
	"testing"
)

func TestSampleSegmentIsStraightForTwoLane(t *testing.T) {
	rm := NewRoadManager()
	a := rm.AddNode(0, 0, 0)
	b := rm.AddNode(40, 0, 0)
	rm.AddSegment(a, b, RoadTwoLane)
	seg := rm.Segments[0]
	xs, zs, ds := rm.SampleSegment(seg, 8)
	if len(xs) != 9 {
		t.Fatalf("samples=%d want 9", len(xs))
	}
	for i := range xs {
		if math.Abs(float64(zs[i])) > 0.01 {
			t.Fatalf("sample %d z=%v want ~0", i, zs[i])
		}
		if i > 0 && xs[i] <= xs[i-1] {
			t.Fatalf("x should increase along segment")
		}
	}
	if ds[len(ds)-1] < 39 || ds[len(ds)-1] > 41 {
		t.Fatalf("length=%v want ~40", ds[len(ds)-1])
	}
}
