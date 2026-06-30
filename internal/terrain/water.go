package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	WaterGridSize = 129
	SeaLevel      = 0.15
	LakeThreshold = 0.25
)

type WaterCell struct {
	Height   float32
	Velocity float32
	FlowX    float32
	FlowZ    float32
	Base     float32
}

type WaterSystem struct {
	Grid [WaterGridSize][WaterGridSize]WaterCell
}

func NewWaterSystem() *WaterSystem {
	return &WaterSystem{}
}

func (ws *WaterSystem) Init(h *Heightmap) {
	for z := 0; z < WaterGridSize; z++ {
		for x := 0; x < WaterGridSize; x++ {
			hx := float64(x) / float64(WaterGridSize-1) * float64(HeightmapSize-1)
			hz := float64(z) / float64(WaterGridSize-1) * float64(HeightmapSize-1)

			terrainH := h.Get(int(hx), int(hz))
			cell := &ws.Grid[z][x]
			cell.Base = terrainH

			if terrainH < SeaLevel {
				cell.Height = SeaLevel - terrainH
			} else if terrainH < LakeThreshold {
				cell.Height = float32(math.Max(0, float64(LakeThreshold-terrainH)*0.3))
			} else {
				cell.Height = 0
			}
		}
	}

	ws.carveLake(h)
}

func (ws *WaterSystem) carveLake(h *Heightmap) {
	center := HeightmapSize / 2
	radius := 20.0
	for z := center - 30; z <= center+30; z++ {
		for x := center - 30; x <= center+30; x++ {
			if x < 0 || x >= HeightmapSize || z < 0 || z >= HeightmapSize {
				continue
			}
			dist := math.Sqrt(float64((x-center)*(x-center) + (z-center)*(z-center)))
			if dist < radius {
				val := float64(h.Get(x, z))
				lower := (1 - dist/radius) * 0.15
				h.Set(x, z, float32(math.Max(0, val-lower)))
			}
		}
	}
}

func (ws *WaterSystem) Update(dt float64) {
	iterations := 3
	for iter := 0; iter < iterations; iter++ {
		for z := 1; z < WaterGridSize-1; z++ {
			for x := 1; x < WaterGridSize-1; x++ {
				cell := &ws.Grid[z][x]
				if cell.Height <= 0.001 {
					continue
				}

				flow := cell.Height * 0.25 * float32(dt)
				neighbors := [4]*WaterCell{
					&ws.Grid[z-1][x],
					&ws.Grid[z+1][x],
					&ws.Grid[z][x-1],
					&ws.Grid[z][x+1],
				}
				dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

				for i, n := range neighbors {
					nTotal := n.Height + n.Base
					cTotal := cell.Height + cell.Base
					if nTotal < cTotal && cell.Height > 0.001 {
						diff := cTotal - nTotal
						amount := math.Min(float64(flow), float64(diff)*0.1)
						amount = math.Min(amount, float64(cell.Height)*0.5)

						cell.Height -= float32(amount)
						n.Height += float32(amount)
						cell.FlowX += float32(dirs[i][0]) * float32(amount)
						cell.FlowZ += float32(dirs[i][1]) * float32(amount)
					}
				}
			}
		}
	}
}

func (ws *WaterSystem) IsWet(worldX, worldZ float32) bool {
	tx := worldX / WorldSize * float32(WaterGridSize-1)
	tz := worldZ / WorldSize * float32(WaterGridSize-1)
	x := int(tx + float32(WaterGridSize)/2)
	z := int(tz + float32(WaterGridSize)/2)
	if x < 0 || x >= WaterGridSize || z < 0 || z >= WaterGridSize {
		return false
	}
	return ws.Grid[z][x].Height > 0.01
}

func (ws *WaterSystem) Draw() {
	h := float32(SeaLevel*MaxHeight + 0.1)
	rl.DrawPlane(rl.NewVector3(0, h, 0), rl.NewVector2(WorldSize, WorldSize), rl.NewColor(30, 120, 210, 160))
}

func (ws *WaterSystem) Unload() {
}
