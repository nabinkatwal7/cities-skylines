package terrain

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
	tx := worldX / WorldSize * float32(HeightmapSize-1)
	tz := worldZ / WorldSize * float32(HeightmapSize-1)
	x0 := int(tx)
	z0 := int(tz)
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

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}
