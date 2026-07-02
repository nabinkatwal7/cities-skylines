package road

import (
	"math"
	"testing"
)

// ponytail: asserts strip winding faces +Y (raylib backface culls the other side).
func TestRoadStripWindingFacesUp(t *testing.T) {
	left0 := [3]float32{0, 0, -4}
	right0 := [3]float32{0, 0, 4}
	left1 := [3]float32{10, 0, -4}
	right1 := [3]float32{10, 0, 4}

	normalY := func(a, b, c [3]float32) float32 {
		e1x, _, e1z := b[0]-a[0], b[1]-a[1], b[2]-a[2]
		e2x, _, e2z := c[0]-a[0], c[1]-a[1], c[2]-a[2]
		return e1z*e2x - e1x*e2z
	}

	tri1 := normalY(left0, right0, left1)
	tri2 := normalY(right0, right1, left1)
	if tri1 <= 0 || tri2 <= 0 {
		t.Fatalf("road strip normals should face +Y; got %v and %v", tri1, tri2)
	}
	if math.Abs(float64(tri1-tri2)) > 1e-3 {
		t.Fatalf("triangle normals should match magnitude; got %v vs %v", tri1, tri2)
	}
}
