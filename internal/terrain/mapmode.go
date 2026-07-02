package terrain

// ponytail: test map toggled in code; flip FlatTestMap in simulation init when done testing.
var FlatTestMap bool

// Test land/river heights in world meters (flat map).
const (
	TestLandMeters  = 2.0
	TestRiverMeters = 1.0
)

func TestLandNorm() float32  { return TestLandMeters / MaxHeight }
func TestRiverNorm() float32  { return TestRiverMeters / MaxHeight }

// ActiveSeaLevel is the normalized height below which terrain counts as underwater.
func ActiveSeaLevel() float32 {
	if FlatTestMap {
		// Land at 2m stays dry; river channels at 1m are wet.
		return (TestRiverMeters + 0.15) / MaxHeight
	}
	return SeaLevel
}

// ActiveWaterSurfaceY is the world Y used to render open water.
func ActiveWaterSurfaceY() float32 {
	if FlatTestMap {
		return TestRiverMeters + 0.2
	}
	return SeaLevel*MaxHeight + 0.1
}
