package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	WaterGridSize = 129
	SeaLevel      = 0.05
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
	Grid       [WaterGridSize][WaterGridSize]WaterCell
	Mesh       rl.Mesh
	Model      rl.Model
	Dirty      bool
	rebuildTimer float64
}

func NewWaterSystem() *WaterSystem {
	return &WaterSystem{Dirty: true}
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

	ws.rebuildTimer += dt
	if ws.rebuildTimer >= 0.25 {
		ws.Dirty = true
		ws.rebuildTimer = 0
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

func (ws *WaterSystem) buildMesh() {
	verts := WaterGridSize * WaterGridSize
	quads := (WaterGridSize - 1) * (WaterGridSize - 1)
	scale := WorldSize / float32(WaterGridSize-1)

	vertices := make([]float32, verts*3)
	normals := make([]float32, verts*3)
	texcoords := make([]float32, verts*2)
	indices := make([]uint16, quads*6)

	idx := 0
	for z := 0; z < WaterGridSize; z++ {
		for x := 0; x < WaterGridSize; x++ {
			worldX := float32(x)*scale - WorldSize/2
			worldZ := float32(z)*scale - WorldSize/2
			h := ws.Grid[z][x].Height * MaxHeight

			vertices[idx*3] = worldX
			vertices[idx*3+1] = h + 0.1
			vertices[idx*3+2] = worldZ

			normals[idx*3] = 0
			normals[idx*3+1] = 1
			normals[idx*3+2] = 0

			texcoords[idx*2] = float32(x) / float32(WaterGridSize-1)
			texcoords[idx*2+1] = float32(z) / float32(WaterGridSize-1)

			idx++
		}
	}

	qi := 0
	for z := 0; z < WaterGridSize-1; z++ {
		for x := 0; x < WaterGridSize-1; x++ {
			a := z*WaterGridSize + x
			b := z*WaterGridSize + x + 1
			c := (z+1)*WaterGridSize + x
			d := (z+1)*WaterGridSize + x + 1
			indices[qi*6] = uint16(a)
			indices[qi*6+1] = uint16(c)
			indices[qi*6+2] = uint16(b)
			indices[qi*6+3] = uint16(b)
			indices[qi*6+4] = uint16(c)
			indices[qi*6+5] = uint16(d)
			qi++
		}
	}

	mesh := rl.Mesh{
		VertexCount:   int32(verts),
		TriangleCount: int32(quads * 2),
		Vertices:      &vertices[0],
		Normals:       &normals[0],
		Texcoords:     &texcoords[0],
		Indices:       &indices[0],
	}

	if ws.Model.MeshCount > 0 {
		rl.UnloadModel(ws.Model)
	}
	ws.Model = rl.LoadModelFromMesh(mesh)
	clearModelMeshData(&ws.Model)
	ws.Dirty = false
}

func (ws *WaterSystem) Draw() {
	if ws.Dirty {
		ws.buildMesh()
	}
	if ws.Model.MeshCount == 0 {
		return
	}
	waterColor := rl.NewColor(30, 120, 200, 180)
	rl.DrawModel(ws.Model, rl.NewVector3(0, 0, 0), 1, waterColor)
}

func (ws *WaterSystem) UploadGPU() {
	ws.buildMesh()
}

func (ws *WaterSystem) Unload() {
	if ws.Model.MeshCount > 0 {
		rl.UnloadModel(ws.Model)
	}
}
