package terrain

import "math"

type BuildabilityInfo struct {
	Score          float32
	SlopeScore     float32
	WaterScore     float32
	PollutionScore float32
	ObjectScore    float32
	RoadScore      float32
	IsUnderwater   bool
	IsSlopeTooSteep bool
	IsBlocked      bool
	HasRoadAccess  bool
	HasObjects     bool
}

type BuildabilityChecker struct {
	heightmap *Heightmap
	water     *WaterSystem
	trees     *TreeSystem
	buildings *BuildingManager
	roads     *RoadManager
	zones     *ZoneManager
	resources *ResourceSystem
}

func NewBuildabilityChecker(h *Heightmap, water *WaterSystem, trees *TreeSystem, buildings *BuildingManager, roads *RoadManager, zones *ZoneManager, resources *ResourceSystem) *BuildabilityChecker {
	return &BuildabilityChecker{
		heightmap: h,
		water:     water,
		trees:     trees,
		buildings: buildings,
		roads:     roads,
		zones:     zones,
		resources: resources,
	}
}

func (bc *BuildabilityChecker) GetBuildability(worldX, worldZ float32) BuildabilityInfo {
	info := BuildabilityInfo{Score: 1.0}
	cellX := (int)((worldX + WorldSize/2) / WorldSize * float32(HeightmapSize-1))
	cellZ := (int)((worldZ + WorldSize/2) / WorldSize * float32(HeightmapSize-1))
	if cellX < 0 || cellX >= HeightmapSize || cellZ < 0 || cellZ >= HeightmapSize {
		info.Score = 0
		return info
	}
	return bc.GetCellBuildability(cellX, cellZ)
}

func (bc *BuildabilityChecker) GetCellBuildability(x, z int) BuildabilityInfo {
	info := BuildabilityInfo{Score: 1.0}

	h := bc.heightmap.Get(x, z)

	worldX := (float32(x)/float32(HeightmapSize-1) - 0.5) * WorldSize
	worldZ := (float32(z)/float32(HeightmapSize-1) - 0.5) * WorldSize

	info.WaterScore = 1.0
	if h <= SeaLevel {
		info.IsUnderwater = true
		info.WaterScore = 0
	} else if bc.water != nil && bc.water.IsFlooded(worldX, worldZ) {
		info.IsUnderwater = true
		info.WaterScore = 0.1
	}

	slope := bc.slopeAt(x, z)
	info.SlopeScore = 1.0
	if slope > 0.3 {
		info.IsSlopeTooSteep = true
		info.SlopeScore = float32(math.Max(0, float64(1-slope/0.5)))
	} else if slope > 0.15 {
		info.SlopeScore = 0.5
	}

	if bc.resources != nil {
		info.PollutionScore = bc.pollutionAt(x, z)
	} else {
		info.PollutionScore = 1.0
	}

	if bc.trees != nil {
		if bc.trees.HasTreeAt(worldX, worldZ) {
			info.IsBlocked = true
			info.HasObjects = true
		}
	}

	if bc.buildings != nil {
		bc.buildings.ForEach(func(b *Building, _ int32) {
			if b.Lifecycle != LifecycleActive {
				return
			}
			bx := (b.Position.X/WorldSize + 0.5) * float32(HeightmapSize-1)
			bz := (b.Position.Z/WorldSize + 0.5) * float32(HeightmapSize-1)
			dx := float32(x) - bx
			dz := float32(z) - bz
			if dx*dx+dz*dz < 4 {
				info.IsBlocked = true
				info.HasObjects = true
			}
		})
	}

	cellSize := WorldSize / float32(bc.zones.width)
	if bc.roads != nil && bc.roads.HasNearbyRoad(worldX, worldZ, cellSize*2) {
		info.HasRoadAccess = true
		info.RoadScore = 1.0
	} else {
		info.RoadScore = 0.3
	}

	info.ObjectScore = 1.0
	if info.HasObjects {
		penalty := float32(0)
		if bc.trees != nil && info.HasObjects {
			penalty = -0.2
		}
		if bc.buildings != nil && info.HasObjects {
			penalty = -0.4
		}
		if penalty < 0 {
			info.ObjectScore = 1.0 + penalty
		}
	}
	if !info.HasRoadAccess {
		info.ObjectScore = float32(math.Min(float64(info.ObjectScore), 0.7))
	}

	info.Score = info.WaterScore * info.SlopeScore * info.PollutionScore * info.ObjectScore
	info.Score = float32(math.Max(0, math.Min(1, float64(info.Score))))

	return info
}

func (bc *BuildabilityChecker) slopeAt(x, z int) float32 {
	if x <= 0 || x >= HeightmapSize-1 || z <= 0 || z >= HeightmapSize-1 {
		return 0
	}
	dx := (bc.heightmap.Get(x+1, z) - bc.heightmap.Get(x-1, z)) / 2
	dz := (bc.heightmap.Get(x, z+1) - bc.heightmap.Get(x, z-1)) / 2
	return float32(math.Sqrt(float64(dx*dx + dz*dz)))
}

func (bc *BuildabilityChecker) pollutionAt(x, z int) float32 {
	if bc.zones == nil {
		return 1.0
	}
	zoneW := bc.zones.width
	zoneH := bc.zones.height

	cx := int(float32(x) / float32(HeightmapSize-1) * float32(zoneW-1))
	cz := int(float32(z) / float32(HeightmapSize-1) * float32(zoneH-1))

	search := 5
	totalWeight := float32(0)
	pollutionWeight := float32(0)

	for dz := -search; dz <= search; dz++ {
		for dx := -search; dx <= search; dx++ {
			nx := cx + dx
			nz := cz + dz
			if nx < 0 || nx >= zoneW || nz < 0 || nz >= zoneH {
				continue
			}
			cell := &bc.zones.Cells[nz][nx]
			if cell.Type != ZoneIndustrial {
				continue
			}
			dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if dist < 0.5 {
				dist = 0.5
			}
			weight := 1.0 / dist
			totalWeight += weight
			pollutionWeight += weight * cell.Density
		}
	}

	if totalWeight < 0.001 {
		return 1.0
	}
	avgPollution := pollutionWeight / totalWeight
	score := 1.0 - avgPollution*0.5
	if score < 0 {
		score = 0
	}
	return score
}
