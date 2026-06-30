package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RoadType uint8

const (
	RoadTwoLane RoadType = iota
	RoadOneWay
	RoadFourLane
	RoadGravel
)

func roadLanes(rt RoadType) int {
	switch rt {
	case RoadOneWay:
		return 1
	case RoadTwoLane, RoadGravel:
		return 2
	case RoadFourLane:
		return 4
	default:
		return 2
	}
}

const laneWidth = float32(3.0)

type RoadNode struct {
	ID              uint32
	X, Z            float32
	Connected       []uint32
	HasTrafficLight bool
}

type RoadSegment struct {
	ID       uint32
	NodeA    uint32
	NodeB    uint32
	RoadType RoadType
	Length   float32
	Elevated bool
}

type RoadManager struct {
	Nodes    []RoadNode
	Segments []RoadSegment
	Models   []rl.Model
	NextID   uint32
	roadTex  rl.Texture2D
}

func NewRoadManager() *RoadManager {
	return &RoadManager{}
}

func (rm *RoadManager) AddNode(x, z float32) uint32 {
	id := rm.NextID
	rm.NextID++
	rm.Nodes = append(rm.Nodes, RoadNode{
		ID:        id,
		X:         x,
		Z:         z,
		Connected: make([]uint32, 0),
	})
	return id
}

func (rm *RoadManager) AddSegment(a, b uint32, rt RoadType) uint32 {
	na := &rm.Nodes[a]
	nb := &rm.Nodes[b]
	dx := nb.X - na.X
	dz := nb.Z - na.Z
	length := float32(math.Sqrt(float64(dx*dx + dz*dz)))

	id := rm.NextID
	rm.NextID++
	rm.Segments = append(rm.Segments, RoadSegment{
		ID:       id,
		NodeA:    a,
		NodeB:    b,
		RoadType: rt,
		Length:   length,
	})

	na.Connected = append(na.Connected, id)
	nb.Connected = append(nb.Connected, id)

	return id
}

func (rm *RoadManager) Prepare() {
	img := rl.GenImageCellular(128, 128, 8)
	rl.ImageColorTint(img, rl.NewColor(80, 80, 80, 255))
	rl.ImageColorBrightness(img, -20)
	rm.roadTex = rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
}

func (rm *RoadManager) UploadGPU(h *Heightmap) {
	rm.Models = make([]rl.Model, len(rm.Segments))
	for i, seg := range rm.Segments {
		rm.Models[i] = rm.buildSegmentMesh(h, seg)
	}
}

func (rm *RoadManager) buildSegmentMesh(h *Heightmap, seg RoadSegment) rl.Model {
	na := &rm.Nodes[seg.NodeA]
	nb := &rm.Nodes[seg.NodeB]

	dx := nb.X - na.X
	dz := nb.Z - na.Z
	length := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	if length < 0.01 {
		return rl.Model{}
	}

	nx := dx / length
	nz := dz / length
	perX := -nz
	perZ := nx

	lanes := roadLanes(seg.RoadType)
	ha := h.WorldHeight(na.X, na.Z) + 0.15
	hb := h.WorldHeight(nb.X, nb.Z) + 0.15

	triCount := lanes * 2
	vertCount := lanes * 4

	verts := make([]float32, 0, vertCount*3)
	normals := make([]float32, 0, vertCount*3)
	texcoords := make([]float32, 0, vertCount*2)
	indices := make([]uint16, 0, triCount*3)

	for li := 0; li < lanes; li++ {
		offset := (float32(li) - float32(lanes-1)*0.5) * laneWidth
		lpx := perX * (offset - laneWidth*0.5)
		lpz := perZ * (offset - laneWidth*0.5)
		rpx := perX * (offset + laneWidth*0.5)
		rpz := perZ * (offset + laneWidth*0.5)

		base := uint16(li * 4)
		verts = append(verts,
			na.X+lpx, ha, na.Z+lpz,
			na.X+rpx, ha, na.Z+rpz,
			nb.X+lpx, hb, nb.Z+lpz,
			nb.X+rpx, hb, nb.Z+rpz,
		)
		normals = append(normals,
			0, 1, 0,
			0, 1, 0,
			0, 1, 0,
			0, 1, 0,
		)
		texcoords = append(texcoords,
			0, 0,
			1, 0,
			0, length/4.0,
			1, length/4.0,
		)
		indices = append(indices, base, base+2, base+1, base+1, base+2, base+3)
	}

	mesh := rl.Mesh{
		VertexCount:   int32(vertCount),
		TriangleCount: int32(triCount),
		Vertices:      &verts[0],
		Normals:       &normals[0],
		Texcoords:     &texcoords[0],
		Indices:       &indices[0],
	}
	rl.UploadMesh(&mesh, false)
	model := rl.LoadModelFromMesh(mesh)
	mats := model.GetMaterials()
	if len(mats) > 0 {
		rl.SetMaterialTexture(&mats[0], rl.MapAlbedo, rm.roadTex)
	}
	return model
}

func (rm *RoadManager) Draw(h *Heightmap) {
	if len(rm.Models) == 0 {
		rm.UploadGPU(h)
	}
	for _, model := range rm.Models {
		if model.MeshCount > 0 {
			rl.DrawModel(model, rl.NewVector3(0, 0, 0), 1, rl.White)
		}
	}
}

func (rm *RoadManager) Unload() {
	for _, model := range rm.Models {
		if model.MeshCount > 0 {
			rl.UnloadModel(model)
		}
	}
	rm.Models = nil
	if rm.roadTex.ID != 0 {
		rl.UnloadTexture(rm.roadTex)
	}
}
