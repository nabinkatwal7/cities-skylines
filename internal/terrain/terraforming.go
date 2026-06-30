package terrain

import "math"

type TerraformTool int

const (
	ToolRaise  TerraformTool = 0
	ToolLower  TerraformTool = 1
	ToolLevel  TerraformTool = 2
	ToolSmooth TerraformTool = 3
	ToolFlatten TerraformTool = 4
)

type Brush struct {
	WorldX, WorldZ float32
	Radius         float32
	Strength       float32
	Tool           TerraformTool
	TargetHeight   float32
}

type TerraformSystem struct {
	active     bool
	brush      Brush
	heightmap  *Heightmap
	manager    *Manager
}

func NewTerraformSystem(m *Manager) *TerraformSystem {
	return &TerraformSystem{
		heightmap: m.Heightmap,
		manager:   m,
		brush: Brush{
			Radius:   10,
			Strength: 0.02,
			Tool:     ToolRaise,
		},
	}
}

func (ts *TerraformSystem) SetTool(tool TerraformTool) {
	ts.brush.Tool = tool
}

func (ts *TerraformSystem) SetRadius(r float32) {
	ts.brush.Radius = float32(math.Max(2, math.Min(50, float64(r))))
}

func (ts *TerraformSystem) SetStrength(s float32) {
	ts.brush.Strength = float32(math.Max(0.005, math.Min(0.1, float64(s))))
}

func (ts *TerraformSystem) Brush() *Brush {
	return &ts.brush
}

func (ts *TerraformSystem) Apply(worldX, worldZ float32) {
	if !ts.active {
		return
	}

	ts.brush.WorldX = worldX
	ts.brush.WorldZ = worldZ

	tx := (worldX + WorldSize/2) / WorldSize * float32(HeightmapSize-1)
	tz := (worldZ + WorldSize/2) / WorldSize * float32(HeightmapSize-1)
	radiusVerts := ts.brush.Radius / WorldSize * float32(HeightmapSize-1)

	cx := int(tx)
	cz := int(tz)
	r := int(radiusVerts) + 2

	affectedChunks := make(map[int]bool)

	minX := max(0, cx-r)
	maxX := min(HeightmapSize-1, cx+r)
	minZ := max(0, cz-r)
	maxZ := min(HeightmapSize-1, cz+r)

	for z := minZ; z <= maxZ; z++ {
		for x := minX; x <= maxX; x++ {
			dx := float32(x) - tx
			dz := float32(z) - tz
			dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

			if dist > radiusVerts {
				continue
			}

			falloff := 1.0 - float32(math.Min(1, float64(dist/radiusVerts)))
			falloff = falloff * falloff

			h := ts.heightmap.Get(x, z)

			switch ts.brush.Tool {
			case ToolRaise:
				h += ts.brush.Strength * falloff
			case ToolLower:
				h -= ts.brush.Strength * falloff
			case ToolLevel:
				diff := ts.brush.TargetHeight - h
				h += diff * 0.3 * falloff
			case ToolSmooth:
				avg := ts.sampleAverage(x, z)
				h += (avg - h) * 0.4 * falloff
			case ToolFlatten:
				diff := ts.brush.TargetHeight - h
				h += diff * 0.5 * falloff
			}

			h = float32(math.Max(0, math.Min(1, float64(h))))
			ts.heightmap.Set(x, z, h)

			chunkX := x / (ChunkSize - 1)
			chunkZ := z / (ChunkSize - 1)
			idx := chunkZ*ChunksPerSide + chunkX
			affectedChunks[idx] = true
		}
	}

	for idx := range affectedChunks {
		ts.rebuildChunkData(idx)
	}

	if ts.manager != nil {
		ts.manager.Water.Init(ts.heightmap)
	}
}

func (ts *TerraformSystem) sampleAverage(x, z int) float32 {
	var sum float32
	var count int
	for dz := -1; dz <= 1; dz++ {
		for dx := -1; dx <= 1; dx++ {
			sx := x + dx
			sz := z + dz
			if sx >= 0 && sx < HeightmapSize && sz >= 0 && sz < HeightmapSize {
				sum += ts.heightmap.Get(sx, sz)
				count++
			}
		}
	}
	return sum / float32(count)
}

func (ts *TerraformSystem) rebuildChunkData(chunkIdx int) {
	if chunkIdx < 0 || chunkIdx >= NumChunks {
		return
	}
	c := ts.manager.Chunks[chunkIdx]
	baseX := c.IndexX * (ChunkSize - 1)
	baseZ := c.IndexZ * (ChunkSize - 1)
	for lz := 0; lz < ChunkSize; lz++ {
		for lx := 0; lx < ChunkSize; lx++ {
			hz := min(baseZ+lz, HeightmapSize-1)
			hx := min(baseX+lx, HeightmapSize-1)
			c.Heights[lz][lx] = ts.heightmap.Get(hx, hz)
		}
	}
	c.Dirty = true
	ts.manager.RebuildChunk(chunkIdx)
}

func (ts *TerraformSystem) SetActive(a bool) {
	ts.active = a
}

func (ts *TerraformSystem) IsActive() bool {
	return ts.active
}

func (ts *TerraformSystem) GetBuildability(x, z int) float32 {
	if x < 0 || x >= HeightmapSize || z < 0 || z >= HeightmapSize {
		return 0
	}

	h := ts.heightmap.Get(x, z)
	if h < SeaLevel {
		return 0
	}

	slope := ts.GetSlope(x, z)
	if slope > 0.3 {
		return float32(math.Max(0, float64(1-slope/0.5)))
	}
	if slope > 0.15 {
		return 0.5
	}

	return 1.0
}

func (ts *TerraformSystem) GetSlope(x, z int) float32 {
	if x <= 0 || x >= HeightmapSize-1 || z <= 0 || z >= HeightmapSize-1 {
		return 0
	}

	dx := (ts.heightmap.Get(x+1, z) - ts.heightmap.Get(x-1, z)) / 2
	dz := (ts.heightmap.Get(x, z+1) - ts.heightmap.Get(x, z-1)) / 2

	return float32(math.Sqrt(float64(dx*dx + dz*dz)))
}
