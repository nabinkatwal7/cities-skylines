package terrain

// ponytail: test map toggled in code; flip FlatTestMap in simulation init when done testing.
var FlatTestMap bool

// Test land/river heights in world meters (flat map).
const (
	TestLandMeters  = 2.0
	TestRiverMeters = 1.0
)

func TestLandNorm() float32  { return TestLandMeters / MaxHeight }
func TestRiverNorm() float32 { return TestRiverMeters / MaxHeight }

// ActiveSeaLevel returns the underwater threshold for the active map profile.
func ActiveSeaLevel() float32 {
	if FlatTestMap {
		return (TestRiverMeters + 0.25) / MaxHeight
	}
	return SeaLevel
}
