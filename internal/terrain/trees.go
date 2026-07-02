package terrain

import (
	"math"
	"math/rand"

	"github.com/katwate/js-skylines/internal/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type TreeSpecies int

const (
	TreeOak   TreeSpecies = 0
	TreePine  TreeSpecies = 1
	TreeBirch TreeSpecies = 2
	TreePalm  TreeSpecies = 3
)

type Tree struct {
	X, Z     float32
	Species  TreeSpecies
	Age      float32
	Health   float32
	Scale    float32
	Yaw      float32
	Lifecycle core.LifecycleState
	RemovalTimer int32
}

type TreeState int

const (
	TreeSeedling TreeState = 0
	TreeYoung    TreeState = 1
	TreeMature   TreeState = 2
	TreeOld      TreeState = 3
	TreeDead     TreeState = 4
)

const TreePoolSize = 20000

type TreeSystem struct {
	Pool          [TreePoolSize]Tree
	FreeList      []int32
	activeCount   int32
	seed          int64
	Model         rl.Model
	colorMap      map[TreeSpecies]rl.Color
	RemovedCount  int
	forestTimer   int32
	resources     *ResourceSystem
	Deforestation float32
}

func (ts *TreeSystem) DeforestPct() float32 {
	total := ts.activeCount + int32(ts.RemovedCount)
	if total == 0 {
		return 0
	}
	return float32(ts.RemovedCount) / float32(total)
}

func NewTreeSystem(seed int64) *TreeSystem {
	ts := &TreeSystem{
		seed:     seed,
		colorMap: map[TreeSpecies]rl.Color{
			TreeOak:   rl.NewColor(34, 139, 34, 255),
			TreePine:  rl.NewColor(0, 100, 0, 255),
			TreeBirch: rl.NewColor(144, 200, 80, 255),
			TreePalm:  rl.NewColor(85, 160, 50, 255),
		},
		FreeList: make([]int32, TreePoolSize),
	}
	for i := 0; i < TreePoolSize; i++ {
		ts.Pool[i].Lifecycle = core.LifecycleUnallocated
		ts.FreeList[i] = int32(TreePoolSize - 1 - i)
	}
	return ts
}

func (ts *TreeSystem) Alloc() int32 {
	if len(ts.FreeList) == 0 {
		return -1
	}
	idx := ts.FreeList[len(ts.FreeList)-1]
	ts.FreeList = ts.FreeList[:len(ts.FreeList)-1]
	ts.Pool[idx].Lifecycle = core.LifecycleAllocated
	ts.activeCount++
	return idx
}

func (ts *TreeSystem) Free(slot int32) {
	if slot < 0 || int(slot) >= TreePoolSize {
		return
	}
	ts.Pool[slot] = Tree{}
	ts.Pool[slot].Lifecycle = core.LifecycleReturnedToPool
	ts.FreeList = append(ts.FreeList, slot)
	ts.activeCount--
}

func (ts *TreeSystem) ForEach(fn func(*Tree, int32)) {
	for i := 0; i < TreePoolSize; i++ {
		if ts.Pool[i].Lifecycle == core.LifecycleActive {
			fn(&ts.Pool[i], int32(i))
		}
	}
}

func (ts *TreeSystem) Generate(h *Heightmap, water *WaterSystem) {
	rng := rand.New(rand.NewSource(ts.seed + 1))
	created := 0

	spacing := 4.0
	for z := spacing / 2; z < float64(HeightmapSize-1); z += spacing {
		for x := spacing / 2; x < float64(HeightmapSize-1); x += spacing {
			if created >= TreePoolSize {
				return
			}

			ix := int(x)
			iz := int(z)
			hVal := h.Get(ix, iz)

			if hVal < SeaLevel+0.02 || hVal > 0.7 {
				continue
			}

			if water != nil && water.IsWet(
				float32(float64(x)/float64(HeightmapSize-1)*WorldSize-WorldSize/2),
				float32(float64(z)/float64(HeightmapSize-1)*WorldSize-WorldSize/2),
			) {
				continue
			}

			noise := rng.Float64()
			if noise > 0.35 {
				continue
			}

			density := float64(1.0)
			if hVal > 0.5 {
				density = float64(1.0 - (hVal-0.5)/0.2)
			}
			if rng.Float64() > density {
				continue
			}

			worldX := float64(x)/float64(HeightmapSize-1)*WorldSize - WorldSize/2
			worldZ := float64(z)/float64(HeightmapSize-1)*WorldSize - WorldSize/2

			var species TreeSpecies
			switch {
			case hVal < 0.15:
				species = TreePalm
			case hVal < 0.35:
				species = TreeBirch
			case hVal < 0.55:
				species = TreeOak
			default:
				species = TreePine
			}

			slot := ts.Alloc()
			if slot < 0 {
				return
			}
			t := &ts.Pool[slot]
			t.X = float32(worldX)
			t.Z = float32(worldZ)
			t.Species = species
			t.Age = rng.Float32() * 100
			t.Health = 0.8 + rng.Float32()*0.2
			t.Scale = 0.5 + rng.Float32()*1.0
			t.Yaw = rng.Float32() * 6.2832
			t.Lifecycle = core.LifecycleActive
			created++
		}
	}
}

func (ts *TreeSystem) Update(dt float64) {
	for i := 0; i < TreePoolSize; i++ {
		t := &ts.Pool[i]
		switch t.Lifecycle {
		case core.LifecycleActive:
			t.Age += float32(dt) * 0.01
			if t.Age > 100 && t.Health > 0 {
				t.Health -= float32(dt) * 0.001
			}
			t.Health = float32(math.Max(0, math.Min(1, float64(t.Health))))
			if t.Health <= 0 {
				t.RemovalTimer = 0
				t.Lifecycle = core.LifecycleMarkedForRemoval
			}
		case core.LifecycleMarkedForRemoval:
			t.RemovalTimer++
			if t.RemovalTimer > 600 {
				t.Lifecycle = core.LifecycleDestroyed
			}
		case core.LifecycleDestroyed:
			ts.Free(int32(i))
		}
	}

	ts.forestTimer++
	if ts.forestTimer > 120 && ts.resources != nil {
		ts.forestTimer = 0
		regrow := 0
		for i := 0; i < TreePoolSize; i++ {
			t := &ts.Pool[i]
			if t.Lifecycle != core.LifecycleActive {
				continue
			}
			rx := int((t.X/WorldSize + 0.5) * float32(HeightmapSize-1))
			rz := int((t.Z/WorldSize + 0.5) * float32(HeightmapSize-1))
			if rx >= 0 && rx < HeightmapSize && rz >= 0 && rz < HeightmapSize {
				ts.resources.RegenerateForest(rx, rz, 0.01)
				regrow++
			}
		}
	}
}

func (ts *TreeSystem) SetResources(rs *ResourceSystem) {
	ts.resources = rs
}

func (ts *TreeSystem) RemoveNear(x, z, radius float32) int {
	radiusSq := radius * radius
	removed := 0
	for i := 0; i < TreePoolSize; i++ {
		t := &ts.Pool[i]
		if t.Lifecycle != core.LifecycleActive {
			continue
		}
		dx := t.X - x
		dz := t.Z - z
		if dx*dx+dz*dz <= radiusSq {
			t.RemovalTimer = 0
			t.Lifecycle = core.LifecycleMarkedForRemoval
			removed++
			if ts.resources != nil {
				rx := int((t.X/WorldSize + 0.5) * float32(HeightmapSize-1))
				rz := int((t.Z/WorldSize + 0.5) * float32(HeightmapSize-1))
				if rx >= 0 && rx < HeightmapSize && rz >= 0 && rz < HeightmapSize {
					ts.resources.Map.NoiseAbsorption[rz][rx] *= 0.5
				}
			}
		}
	}
	ts.RemovedCount += removed
	return removed
}

func (ts *TreeSystem) Draw(h *Heightmap, camX, camZ float32) {
	maxDist := float32(400)
	if ts.Model.MeshCount == 0 {
		ts.drawFallback(h, camX, camZ, maxDist)
		return
	}
	ts.ForEach(func(t *Tree, _ int32) {
		state := ts.getState(*t)
		dx := t.X - camX
		dz := t.Z - camZ
		if dx*dx+dz*dz > maxDist*maxDist {
			return
		}
		if state == TreeDead {
			ts.drawDeadMesh(t, h)
			return
		}
		height := h.WorldHeight(t.X, t.Z)
		stageScale := ts.growthScale(state)
		col := ts.stageColor(t.Species, state)
		scale := t.Scale * 0.4 * stageScale
		pos := rl.NewVector3(t.X, height+scale*0.5, t.Z)
		axis := rl.NewVector3(0, 1, 0)
		rl.DrawModelEx(ts.Model, pos, axis, t.Yaw*57.3, rl.NewVector3(scale, scale, scale), col)
	})
}

func (ts *TreeSystem) drawDeadMesh(t *Tree, h *Heightmap) {
	height := h.WorldHeight(t.X, t.Z)
	scale := t.Scale * 0.4 * 0.6
	pos := rl.NewVector3(t.X, height+scale*0.5, t.Z)
	axis := rl.NewVector3(0, 1, 0)
	deadCol := rl.NewColor(80, 60, 40, 255)
	rl.DrawModelEx(ts.Model, pos, axis, t.Yaw*57.3, rl.NewVector3(scale, scale*0.5, scale), deadCol)
}

func (ts *TreeSystem) drawFallback(h *Heightmap, camX, camZ, maxDist float32) {
	ts.ForEach(func(t *Tree, _ int32) {
		state := ts.getState(*t)
		dx := t.X - camX
		dz := t.Z - camZ
		if dx*dx+dz*dz > maxDist*maxDist {
			return
		}
		height := h.WorldHeight(t.X, t.Z)
		stageScale := ts.growthScale(state)
		scale := t.Scale * stageScale
		col := ts.stageColor(t.Species, state)

		if state == TreeDead {
			rl.DrawCube(rl.NewVector3(t.X, height+scale*0.3, t.Z), 0.15*scale, scale*0.6, 0.15*scale, rl.NewColor(80, 60, 40, 255))
			return
		}

		rl.DrawCube(rl.NewVector3(t.X, height+scale*0.5, t.Z), 0.2*scale, scale, 0.2*scale, rl.NewColor(101, 67, 33, 255))
		crownSize := scale * 0.8
		crownY := height + scale + crownSize*0.3
		rl.DrawCube(rl.NewVector3(t.X, crownY, t.Z), crownSize, crownSize*0.6, crownSize, col)
	})
}

func (ts *TreeSystem) growthScale(state TreeState) float32 {
	switch state {
	case TreeSeedling:
		return 0.3
	case TreeYoung:
		return 0.6
	case TreeMature:
		return 1.0
	case TreeOld:
		return 1.1
	case TreeDead:
		return 0.8
	default:
		return 0.5
	}
}

func (ts *TreeSystem) stageColor(species TreeSpecies, state TreeState) rl.Color {
	base := ts.colorMap[species]
	if state == TreeOld {
		return rl.NewColor(
			uint8(float32(base.R)*0.7),
			uint8(float32(base.G)*0.7),
			uint8(float32(base.B)*0.6),
			255,
		)
	}
	if state == TreeSeedling {
		return rl.NewColor(
			uint8(float32(base.R)*0.5+100),
			uint8(float32(base.G)*0.5+100),
			uint8(float32(base.B)*0.5+80),
			255,
		)
	}
	return base
}

func (ts *TreeSystem) getState(t Tree) TreeState {
	if t.Health <= 0 {
		return TreeDead
	}
	switch {
	case t.Age < 20:
		return TreeSeedling
	case t.Age < 50:
		return TreeYoung
	case t.Age < 80:
		return TreeMature
	default:
		return TreeOld
	}
}

func (ts *TreeSystem) Count() int {
	return int(ts.activeCount)
}

func (ts *TreeSystem) PlantAt(worldX, worldZ float32, species TreeSpecies) bool {
	slot := ts.Alloc()
	if slot < 0 {
		return false
	}
	t := &ts.Pool[slot]
	t.X = worldX
	t.Z = worldZ
	t.Species = species
	t.Age = 1
	t.Health = 1.0
	t.Scale = 0.3
	t.Yaw = rand.Float32() * 6.2832
	t.Lifecycle = core.LifecycleActive
	return true
}

func (ts *TreeSystem) TreeCountAt(x, z, radius float32) int {
	radiusSq := radius * radius
	count := 0
	for i := 0; i < TreePoolSize; i++ {
		t := &ts.Pool[i]
		if t.Lifecycle != core.LifecycleActive {
			continue
		}
		dx := t.X - x
		dz := t.Z - z
		if dx*dx+dz*dz <= radiusSq {
			count++
		}
	}
	return count
}

func (ts *TreeSystem) HasTreeAt(worldX, worldZ float32) bool {
	for i := 0; i < TreePoolSize; i++ {
		t := &ts.Pool[i]
		if t.Lifecycle != core.LifecycleActive {
			continue
		}
		dx := t.X - worldX
		dz := t.Z - worldZ
		if dx*dx+dz*dz < 4 {
			return true
		}
	}
	return false
}
