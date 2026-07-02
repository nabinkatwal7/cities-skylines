package terrain

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	HeightmapSize = 1025
	WorldSize     = 4096.0
	MaxHeight     = 60.0
)

type Heightmap struct {
	Data     [HeightmapSize][HeightmapSize]float32
	Min, Max float32
}

func NewHeightmap() *Heightmap {
	return &Heightmap{
		Min: MaxHeight,
		Max: 0,
	}
}

func (h *Heightmap) Get(x, z int) float32 {
	if x < 0 || x >= HeightmapSize || z < 0 || z >= HeightmapSize {
		return 0
	}
	return h.Data[z][x]
}

func (h *Heightmap) Set(x, z int, val float32) {
	if x < 0 || x >= HeightmapSize || z < 0 || z >= HeightmapSize {
		return
	}
	h.Data[z][x] = val
	if val < h.Min {
		h.Min = val
	}
	if val > h.Max {
		h.Max = val
	}
}

func (h *Heightmap) GetInterpolated(worldX, worldZ float32) float32 {
	tx := (worldX + WorldSize/2) / WorldSize * float32(HeightmapSize-1)
	tz := (worldZ + WorldSize/2) / WorldSize * float32(HeightmapSize-1)
	x0 := int(tx)
	z0 := int(tz)
	if x0 < 0 {
		x0 = 0
	}
	if z0 < 0 {
		z0 = 0
	}
	if x0 >= HeightmapSize {
		x0 = HeightmapSize - 1
	}
	if z0 >= HeightmapSize {
		z0 = HeightmapSize - 1
	}
	x1 := min(x0+1, HeightmapSize-1)
	z1 := min(z0+1, HeightmapSize-1)

	fx := tx - float32(x0)
	fz := tz - float32(z0)

	h00 := h.Get(x0, z0)
	h10 := h.Get(x1, z0)
	h01 := h.Get(x0, z1)
	h11 := h.Get(x1, z1)

	return lerp(lerp(h00, h10, fx), lerp(h01, h11, fx), fz)
}

func (h *Heightmap) WorldHeight(worldX, worldZ float32) float32 {
	return h.GetInterpolated(worldX, worldZ) * MaxHeight
}

func (h *Heightmap) IsUnderwater(worldX, worldZ float32) bool {
	return h.GetInterpolated(worldX, worldZ) < ActiveSeaLevel()
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

// PickXZ returns world X/Z where a screen ray hits the terrain surface.
func (h *Heightmap) PickXZ(ray rl.Ray) (x, z float32, ok bool) {
	const step = 6.0
	const maxDist = 3000.0
	prevAbove := true
	prevT := float32(0)
	for t := float32(0); t <= maxDist; t += step {
		px := ray.Position.X + ray.Direction.X*t
		py := ray.Position.Y + ray.Direction.Y*t
		pz := ray.Position.Z + ray.Direction.Z*t
		th := h.WorldHeight(px, pz)
		above := py > th
		if t > 0 && above != prevAbove {
			lo, hi := prevT, t
			for i := 0; i < 10; i++ {
				mid := (lo + hi) * 0.5
				mx := ray.Position.X + ray.Direction.X*mid
				my := ray.Position.Y + ray.Direction.Y*mid
				mz := ray.Position.Z + ray.Direction.Z*mid
				if my > h.WorldHeight(mx, mz) {
					hi = mid
				} else {
					lo = mid
				}
			}
			x = ray.Position.X + ray.Direction.X*lo
			z = ray.Position.Z + ray.Direction.Z*lo
			return x, z, true
		}
		prevAbove = above
		prevT = t
	}
	if ray.Direction.Y != 0 {
		t := -ray.Position.Y / ray.Direction.Y
		if t > 0 {
			return ray.Position.X + ray.Direction.X*t, ray.Position.Z + ray.Direction.Z*t, true
		}
	}
	return 0, 0, false
}
