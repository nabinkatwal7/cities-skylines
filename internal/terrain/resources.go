package terrain

import (
	"math"
	"math/rand"
)

type ResourceMap struct {
	Ore            [HeightmapSize][HeightmapSize]float32
	Oil            [HeightmapSize][HeightmapSize]float32
	Fertility      [HeightmapSize][HeightmapSize]float32
	Forest         [HeightmapSize][HeightmapSize]float32
	NoiseAbsorption [HeightmapSize][HeightmapSize]float32
}

type ResourceSystem struct {
	Map      ResourceMap
	overlay  bool
	trees    *TreeSystem
}

func NewResourceSystem(seed int64, h *Heightmap) *ResourceSystem {
	rs := &ResourceSystem{}
	rng := rand.New(rand.NewSource(seed + 2))

	noise := NewNoise(seed + 10)

	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			nx := float64(x) * 0.005
			nz := float64(z) * 0.005

			terrainH := h.Get(x, z)

			oreBase := noise.Fbm(nx*2+100, nz*2+100, 3, 2.0, 0.5)
			oreMask := float64(0)
			if terrainH > 0.4 && terrainH < 0.85 {
				oreMask = 1.0 - math.Abs(float64(terrainH-0.6))*4
			}
			rs.Map.Ore[z][x] = float32(math.Max(0, math.Min(1, (oreBase*0.7+rng.Float64()*0.3)*oreMask)))

			oilBase := noise.Fbm(nx*1.5+200, nz*1.5+200, 3, 2.0, 0.5)
			oilMask := float64(0)
			if terrainH < 0.3 {
				oilMask = 1.0 - float64(terrainH)*3
			}
			rs.Map.Oil[z][x] = float32(math.Max(0, math.Min(1, (oilBase*0.6+rng.Float64()*0.4)*oilMask)))

			fertBase := noise.Fbm(nx+300, nz+300, 4, 2.0, 0.5)
			fertMask := float64(1.0)
			if terrainH < SeaLevel || terrainH > 0.6 {
				fertMask = 0
			} else if terrainH > 0.4 {
				fertMask = 1.0 - (float64(terrainH)-0.4)/0.2
			}
			rs.Map.Fertility[z][x] = float32(math.Max(0, math.Min(1, fertBase*fertMask)))

			forestBase := noise.Fbm(nx+400, nz+400, 3, 2.0, 0.5)
			forestMask := float64(1.0)
			if terrainH < 0.1 || terrainH > 0.65 {
				forestMask = 0
			} else if terrainH > 0.5 {
				forestMask = 1.0 - (float64(terrainH)-0.5)/0.15
			}
			rs.Map.Forest[z][x] = float32(math.Max(0, math.Min(1, forestBase*forestMask)))

			absorpBase := noise.Fbm(nx+500, nz+500, 3, 2.0, 0.5)
			rs.Map.NoiseAbsorption[z][x] = rs.Map.Forest[z][x] * float32(math.Max(0, absorpBase))
		}
	}

	return rs
}

func (rs *ResourceSystem) SetTrees(ts *TreeSystem) {
	rs.trees = ts
}

func (rs *ResourceSystem) NoiseAt(worldX, worldZ float32) float32 {
	rx := int((worldX/WorldSize + 0.5) * float32(HeightmapSize-1))
	rz := int((worldZ/WorldSize + 0.5) * float32(HeightmapSize-1))
	if rx < 0 || rx >= HeightmapSize || rz < 0 || rz >= HeightmapSize {
		return 0
	}
	baseNoise := float32(0.3)
	absorp := rs.Map.NoiseAbsorption[rz][rx]
	if absorp > 0 {
		baseNoise *= 1.0 - absorp*0.6
	}
	if baseNoise < 0 {
		baseNoise = 0
	}
	return baseNoise
}

func (rs *ResourceSystem) ToggleOverlay() {
	rs.overlay = !rs.overlay
}

func (rs *ResourceSystem) DrawOverlay() bool {
	return rs.overlay
}

func (rs *ResourceSystem) ExtractOre(x, z int, amount float32) {
	rs.Map.Ore[z][x] = float32(math.Max(0, float64(rs.Map.Ore[z][x]-amount)))
}

func (rs *ResourceSystem) ExtractOil(x, z int, amount float32) {
	rs.Map.Oil[z][x] = float32(math.Max(0, float64(rs.Map.Oil[z][x]-amount)))
}

func (rs *ResourceSystem) RegenerateForest(x, z int, amount float32) {
	rs.Map.Forest[z][x] = float32(math.Min(1, float64(rs.Map.Forest[z][x]+amount)))
}
