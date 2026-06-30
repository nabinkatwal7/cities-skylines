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
	m := lanes + lanes - 1
	if lanes == 1 {
		m = 1
	}

	ha := h.WorldHeight(na.X, na.Z) + 0.15
	hb := h.WorldHeight(nb.X, nb.Z) + 0.15

	verts := make([]float32, 0, m*4*3)
	colors := make([]uint8, 0, m*4*4)
	indices := make([]uint16, 0, m*6)

	asphalt := rl.NewColor(70, 70, 70, 255)
	centerLine := rl.NewColor(220, 200, 60, 255)
	edgeLine := rl.NewColor(200, 200, 200, 255)

	for li := 0; li < lanes; li++ {
		offset := (float32(li) - float32(lanes-1)*0.5) * laneWidth
		lpx := perX * (offset - laneWidth*0.5 + 0.1)
		lpz := perZ * (offset - laneWidth*0.5 + 0.1)
		rpx := perX * (offset + laneWidth*0.5 - 0.1)
		rpz := perZ * (offset + laneWidth*0.5 - 0.1)

		base := uint16(len(verts) / 3)
		verts = append(verts,
			na.X+lpx, ha, na.Z+lpz,
			na.X+rpx, ha, na.Z+rpz,
			nb.X+lpx, hb, nb.Z+lpz,
			nb.X+rpx, hb, nb.Z+rpz,
		)
		for i := 0; i < 4; i++ {
			colors = append(colors, asphalt.R, asphalt.G, asphalt.B, asphalt.A)
		}
		indices = append(indices, base, base+2, base+1, base+1, base+2, base+3)
	}

	for li := 0; li < lanes-1; li++ {
		offset := (float32(li) - float32(lanes-1)*0.5 + 1) * laneWidth
		left := offset - 0.15
		right := offset + 0.15
		lpx := perX * left
		lpz := perZ * left
		rpx := perX * right
		rpz := perZ * right

		base := uint16(len(verts) / 3)
		verts = append(verts,
			na.X+lpx, ha, na.Z+lpz,
			na.X+rpx, ha, na.Z+rpz,
			nb.X+lpx, hb, nb.Z+lpz,
			nb.X+rpx, hb, nb.Z+rpz,
		)
		col := centerLine
		for i := 0; i < 4; i++ {
			colors = append(colors, col.R, col.G, col.B, col.A)
		}
		indices = append(indices, base, base+2, base+1, base+1, base+2, base+3)
	}

	{
		totalWidth := float32(lanes) * laneWidth
		half := totalWidth * 0.5

		lpx := perX * -(half)
		lpz := perZ * -(half)
		rpx := perX * -(half - 0.2)
		rpz := perZ * -(half - 0.2)

		base := uint16(len(verts) / 3)
		verts = append(verts,
			na.X+lpx, ha, na.Z+lpz,
			na.X+rpx, ha, na.Z+rpz,
			nb.X+lpx, hb, nb.Z+lpz,
			nb.X+rpx, hb, nb.Z+rpz,
		)
		col := edgeLine
		for i := 0; i < 4; i++ {
			colors = append(colors, col.R, col.G, col.B, col.A)
		}
		indices = append(indices, base, base+2, base+1, base+1, base+2, base+3)

		lpx = perX * (half - 0.2)
		lpz = perZ * (half - 0.2)
		rpx = perX * half
		rpz = perZ * half

		base = uint16(len(verts) / 3)
		verts = append(verts,
			na.X+lpx, ha, na.Z+lpz,
			na.X+rpx, ha, na.Z+rpz,
			nb.X+lpx, hb, nb.Z+lpz,
			nb.X+rpx, hb, nb.Z+rpz,
		)
		for i := 0; i < 4; i++ {
			colors = append(colors, col.R, col.G, col.B, col.A)
		}
		indices = append(indices, base, base+2, base+1, base+1, base+2, base+3)
	}

	mesh := rl.Mesh{
		VertexCount:   int32(len(verts) / 3),
		TriangleCount: int32(len(indices) / 3),
		Vertices:      &verts[0],
		Colors:        &colors[0],
		Indices:       &indices[0],
	}
	rl.UploadMesh(&mesh, false)
	return rl.LoadModelFromMesh(mesh)
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

func (rm *RoadManager) NearestNode(x, z float32) (uint32, bool) {
	if len(rm.Nodes) == 0 {
		return 0, false
	}
	best := uint32(0)
	bestDist := float32(math.MaxFloat32)
	for _, n := range rm.Nodes {
		dx := n.X - x
		dz := n.Z - z
		d := dx*dx + dz*dz
		if d < bestDist {
			bestDist = d
			best = n.ID
		}
	}
	return best, bestDist < 400
}

func (rm *RoadManager) Rebuild(h *Heightmap) {
	for _, model := range rm.Models {
		if model.MeshCount > 0 {
			rl.UnloadModel(model)
		}
	}
	rm.Models = nil
	rm.UploadGPU(h)
}

func (rm *RoadManager) Unload() {
	for _, model := range rm.Models {
		if model.MeshCount > 0 {
			rl.UnloadModel(model)
		}
	}
	rm.Models = nil
}
