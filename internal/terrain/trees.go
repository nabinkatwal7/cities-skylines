package terrain

import (
	"math"
	"math/rand"

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
	X, Z    float32
	Species TreeSpecies
	Age     float32
	Health  float32
	Scale   float32
	Yaw     float32
}

type TreeState int

const (
	TreeSeedling TreeState = 0
	TreeYoung    TreeState = 1
	TreeMature   TreeState = 2
	TreeOld      TreeState = 3
	TreeDead     TreeState = 4
)

type TreeSystem struct {
	Trees    []Tree
	MaxTrees int
	seed     int64
	Models   []rl.Model
	colorMap map[TreeSpecies]rl.Color
}

func NewTreeSystem(seed int64) *TreeSystem {
	return &TreeSystem{
		seed:     seed,
		MaxTrees: 20000,
		colorMap: map[TreeSpecies]rl.Color{
			TreeOak:   rl.NewColor(34, 139, 34, 255),
			TreePine:  rl.NewColor(0, 100, 0, 255),
			TreeBirch: rl.NewColor(144, 200, 80, 255),
			TreePalm:  rl.NewColor(85, 160, 50, 255),
		},
	}
}

func (ts *TreeSystem) Generate(h *Heightmap, water *WaterSystem) {
	rng := rand.New(rand.NewSource(ts.seed + 1))
	ts.Trees = make([]Tree, 0, ts.MaxTrees)

	spacing := 4.0
	for z := spacing / 2; z < float64(HeightmapSize-1); z += spacing {
		for x := spacing / 2; x < float64(HeightmapSize-1); x += spacing {
			if len(ts.Trees) >= ts.MaxTrees {
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

			ts.Trees = append(ts.Trees, Tree{
				X:       float32(worldX),
				Z:       float32(worldZ),
				Species: species,
				Age:     rng.Float32() * 100,
				Health:  0.8 + rng.Float32()*0.2,
				Scale:   0.5 + rng.Float32()*1.0,
				Yaw:     rng.Float32() * 6.2832,
			})
		}
	}
}

func (ts *TreeSystem) Update(dt float64) {
	for i := range ts.Trees {
		t := &ts.Trees[i]
		t.Age += float32(dt) * 0.01
		if t.Age > 100 && t.Health > 0 {
			t.Health -= float32(dt) * 0.001
		}
		t.Health = float32(math.Max(0, math.Min(1, float64(t.Health))))
	}
}

func (ts *TreeSystem) RemoveNear(x, z, radius float32) {
	remaining := ts.Trees[:0]
	for _, t := range ts.Trees {
		dx := t.X - x
		dz := t.Z - z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if dist > radius {
			remaining = append(remaining, t)
		}
	}
	ts.Trees = remaining
}

func (ts *TreeSystem) Draw(h *Heightmap, camX, camZ float32) {
	maxDist := float32(400)
	for _, t := range ts.Trees {
		state := ts.getState(t)
		if state == TreeDead {
			continue
		}
		dx := t.X - camX
		dz := t.Z - camZ
		if dx*dx+dz*dz > maxDist*maxDist {
			continue
		}
		height := h.WorldHeight(t.X, t.Z)
		var model rl.Model
		if int(t.Species) < len(ts.Models) && ts.Models[t.Species].MeshCount > 0 {
			model = ts.Models[t.Species]
		} else if len(ts.Models) > 0 && ts.Models[0].MeshCount > 0 {
			model = ts.Models[0]
		} else {
			ts.drawFallbackTree(t, height)
			continue
		}
		scale := t.Scale * 0.8
		pos := rl.NewVector3(t.X, height+scale*0.5, t.Z)
		axis := rl.NewVector3(0, 1, 0)
		rl.DrawModelEx(model, pos, axis, t.Yaw*57.3, rl.NewVector3(scale, scale, scale), rl.White)
	}
}

func (ts *TreeSystem) drawFallbackTree(t Tree, height float32) {
	scale := t.Scale
	col := ts.colorMap[t.Species]
	rl.DrawCube(rl.NewVector3(t.X, height+scale*0.5, t.Z), 0.2*scale, scale, 0.2*scale, rl.NewColor(101, 67, 33, 255))
	crownSize := scale * 0.8
	crownY := height + scale + crownSize*0.3
	rl.DrawCube(rl.NewVector3(t.X, crownY, t.Z), crownSize, crownSize*0.6, crownSize, col)
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
	return len(ts.Trees)
}
