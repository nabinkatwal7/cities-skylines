package terrain

import "math"

type BuildabilityInfo struct {
	Score           float32
	SlopeScore      float32
	WaterScore      float32
	PollutionScore  float32
	ObjectScore     float32
	RoadScore       float32
	IsUnderwater    bool
	IsSlopeTooSteep bool
	IsBlocked       bool
	HasRoadAccess   bool
	HasObjects      bool
}

type RoadAccess interface {
	HasNearbyRoad(x, z, max float32) bool
}

type BuildabilityChecker struct {
	heightmap *Heightmap
	water     *WaterSystem
	trees     *TreeSystem
	roads     RoadAccess
	resources *ResourceSystem
}

func NewBuildabilityChecker(h *Heightmap, water *WaterSystem, trees *TreeSystem, roads RoadAccess, resources *ResourceSystem) *BuildabilityChecker {
	return &BuildabilityChecker{
		heightmap: h,
		water:     water,
		trees:     trees,
		roads:     roads,
		resources: resources,
	}
}

func (bc *BuildabilityChecker) GetBuildability(worldX, worldZ float32) BuildabilityInfo {
	cellX := int((worldX + WorldSize/2) / WorldSize * float32(HeightmapSize-1))
	cellZ := int((worldZ + WorldSize/2) / WorldSize * float32(HeightmapSize-1))
	if cellX < 0 || cellX >= HeightmapSize || cellZ < 0 || cellZ >= HeightmapSize {
		return BuildabilityInfo{}
	}
	return bc.GetCellBuildability(cellX, cellZ)
}

func (bc *BuildabilityChecker) GetCellBuildability(x, z int) BuildabilityInfo {
	info := BuildabilityInfo{Score: 1.0}

	h := bc.heightmap.Get(x, z)
	worldX := (float32(x)/float32(HeightmapSize-1) - 0.5) * WorldSize
	worldZ := (float32(z)/float32(HeightmapSize-1) - 0.5) * WorldSize

	info.WaterScore = 1.0
	if h < ActiveSeaLevel() {
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
		info.PollutionScore = 1.0
	} else {
		info.PollutionScore = 1.0
	}

	if bc.trees != nil && bc.trees.HasTreeAt(worldX, worldZ) {
		info.IsBlocked = true
		info.HasObjects = true
	}

	cellSize := float32(WorldSize / 128)
	if bc.roads != nil && bc.roads.HasNearbyRoad(worldX, worldZ, cellSize*2) {
		info.HasRoadAccess = true
		info.RoadScore = 1.0
	} else {
		info.RoadScore = 0.3
	}

	info.ObjectScore = 1.0
	if info.HasObjects {
		info.ObjectScore = 0.8
	}
	if !info.HasRoadAccess {
		info.ObjectScore = float32(math.Min(float64(info.ObjectScore), 0.7))
	}

	info.Score = info.WaterScore * info.SlopeScore * info.PollutionScore * info.ObjectScore
	info.Score = float32(math.Max(0, math.Min(1, float64(info.Score))))
	return info
}

var buildabilityChecker *BuildabilityChecker

func SetBuildabilityChecker(bc *BuildabilityChecker) {
	buildabilityChecker = bc
}

func (bc *BuildabilityChecker) slopeAt(x, z int) float32 {
	if x <= 0 || x >= HeightmapSize-1 || z <= 0 || z >= HeightmapSize-1 {
		return 0
	}
	dx := (bc.heightmap.Get(x+1, z) - bc.heightmap.Get(x-1, z)) / 2
	dz := (bc.heightmap.Get(x, z+1) - bc.heightmap.Get(x, z-1)) / 2
	return float32(math.Sqrt(float64(dx*dx + dz*dz)))
}
