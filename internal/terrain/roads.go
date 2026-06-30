package terrain

import (
	"math"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RoadType uint8

const (
	RoadTwoLane RoadType = iota
	RoadOneWay
	RoadFourLane
	RoadGravel
)

const laneW = float32(3.0)

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

func (rm *RoadManager) LoadAssets() {
	tex := rl.LoadTexture("assets/road.jpg")
	if tex.ID != 0 {
		rm.roadTex = tex
	}
}

func (rm *RoadManager) AddNode(x, z float32) uint32 {
	id := rm.NextID
	rm.NextID++
	idx := uint32(len(rm.Nodes))
	rm.Nodes = append(rm.Nodes, RoadNode{
		ID:        id,
		X:         x,
		Z:         z,
		Connected: make([]uint32, 0),
	})
	return idx
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

func (rm *RoadManager) NearestNode(x, z float32) (uint32, bool) {
	if len(rm.Nodes) == 0 {
		return 0, false
	}
	bestIdx := uint32(0)
	bestDist := float32(math.MaxFloat32)
	for i, n := range rm.Nodes {
		dx := n.X - x
		dz := n.Z - z
		d := dx*dx + dz*dz
		if d < bestDist {
			bestDist = d
			bestIdx = uint32(i)
		}
	}
	return bestIdx, bestDist < 400
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

func (rm *RoadManager) UploadGPU(h *Heightmap) {
	rm.Models = make([]rl.Model, len(rm.Segments))
	for i, seg := range rm.Segments {
		rm.Models[i] = rm.buildSurfaceMesh(h, seg)
	}
}

func (rm *RoadManager) buildSurfaceMesh(h *Heightmap, seg RoadSegment) rl.Model {
	na := &rm.Nodes[seg.NodeA]
	nb := &rm.Nodes[seg.NodeB]

	dx := nb.X - na.X
	dz := nb.Z - na.Z
	length := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	if length < 0.01 {
		return rl.Model{}
	}

	perX := -dz / length
	perZ := dx / length
	lanes := roadLanes(seg.RoadType)
	total := float32(lanes) * laneW
	half := total * 0.5

	steps := int(length / 2)
	if steps < 1 {
		steps = 1
	}

	vertCount := (steps + 1) * 2
	triCount := steps * 2

	verts := make([]float32, 0, vertCount*3)
	normals := make([]float32, 0, vertCount*3)
	texcoords := make([]float32, 0, vertCount*2)
	indices := make([]uint16, 0, triCount*3)

	for si := 0; si <= steps; si++ {
		t := float32(si) / float32(steps)
		x := na.X + dx*t
		z := na.Z + dz*t
		hgt := h.WorldHeight(x, z) + 0.15

		u := float32(si) / 4.0
		verts = append(verts, x-perX*half, hgt, z-perZ*half)
		verts = append(verts, x+perX*half, hgt, z+perZ*half)
		normals = append(normals, 0, 1, 0, 0, 1, 0)
		texcoords = append(texcoords, 0, u, 1, u)
	}

	for si := 0; si < steps; si++ {
		base := uint16(si * 2)
		indices = append(indices, base, base+2, base+1, base+1, base+2, base+3)
	}

	mesh := rl.Mesh{
		VertexCount:   int32(vertCount),
		TriangleCount: int32(triCount),
		Vertices:      &verts[0],
		Normals:       &normals[0],
		Texcoords:     &texcoords[0],
		Indices:       (*uint16)(unsafe.Pointer(&indices[0])),
	}
	rl.UploadMesh(&mesh, false)
	model := rl.LoadModelFromMesh(mesh)
	mats := model.GetMaterials()
	if len(mats) > 0 && rm.roadTex.ID != 0 {
		rl.SetMaterialTexture(&mats[0], rl.MapAlbedo, rm.roadTex)
	}
	return model
}

func (rm *RoadManager) Draw(h *Heightmap) {
	if rm.roadTex.ID == 0 {
		rm.drawFallback(h)
		return
	}
	if len(rm.Models) == 0 {
		rm.UploadGPU(h)
	}
	for _, model := range rm.Models {
		if model.MeshCount > 0 {
			rl.DrawModel(model, rl.NewVector3(0, 0, 0), 1, rl.White)
		}
	}
	rm.drawMarkings(h)
}

func (rm *RoadManager) drawFallback(h *Heightmap) {
	for _, seg := range rm.Segments {
		na := &rm.Nodes[seg.NodeA]
		nb := &rm.Nodes[seg.NodeB]

		dx := nb.X - na.X
		dz := nb.Z - na.Z
		l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if l < 0.01 {
			continue
		}

		perX := -dz / l
		perZ := dx / l
		lanes := roadLanes(seg.RoadType)
		total := float32(lanes) * laneW
		half := total * 0.5
		col := rl.NewColor(80, 80, 80, 255)

		steps := int(l / 2)
		if steps < 1 {
			steps = 1
		}
		for si := 0; si < steps; si++ {
			t0 := float32(si) / float32(steps)
			t1 := float32(si+1) / float32(steps)
			x0 := na.X + dx*t0
			z0 := na.Z + dz*t0
			h0 := h.WorldHeight(x0, z0) + 0.15
			x1 := na.X + dx*t1
			z1 := na.Z + dz*t1
			h1 := h.WorldHeight(x1, z1) + 0.15

			al := rl.NewVector3(x0-perX*half, h0, z0-perZ*half)
			ar := rl.NewVector3(x0+perX*half, h0, z0+perZ*half)
			bl := rl.NewVector3(x1-perX*half, h1, z1-perZ*half)
			br := rl.NewVector3(x1+perX*half, h1, z1+perZ*half)
			rl.DrawTriangle3D(al, ar, bl, col)
			rl.DrawTriangle3D(bl, ar, br, col)
		}
	}
}

func drawQuad(a, b, c, d rl.Vector3, col rl.Color) {
	rl.DrawTriangle3D(a, b, c, col)
	rl.DrawTriangle3D(c, b, d, col)
}

func (rm *RoadManager) drawMarkings(h *Heightmap) {
	center := rl.NewColor(220, 200, 60, 255)
	edge := rl.NewColor(200, 200, 200, 255)

	for _, seg := range rm.Segments {
		na := &rm.Nodes[seg.NodeA]
		nb := &rm.Nodes[seg.NodeB]

		dx := nb.X - na.X
		dz := nb.Z - na.Z
		l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if l < 0.01 {
			continue
		}

		perX := -dz / l
		perZ := dx / l
		lanes := roadLanes(seg.RoadType)

		steps := int(l / 2)
		if steps < 1 {
			steps = 1
		}

		for si := 0; si < steps; si++ {
			t0 := float32(si) / float32(steps)
			t1 := float32(si+1) / float32(steps)
			x0 := na.X + dx*t0
			z0 := na.Z + dz*t0
			h0 := h.WorldHeight(x0, z0) + 0.15
			x1 := na.X + dx*t1
			z1 := na.Z + dz*t1
			h1 := h.WorldHeight(x1, z1) + 0.15

			for li := 0; li < lanes-1; li++ {
				offset := (float32(li) - float32(lanes-1)*0.5 + 1) * laneW
				gap := float32(0.15)
				cl := rl.NewVector3(x0+perX*(offset-gap), h0, z0+perZ*(offset-gap))
				cr := rl.NewVector3(x0+perX*(offset+gap), h0, z0+perZ*(offset+gap))
				ncl := rl.NewVector3(x1+perX*(offset-gap), h1, z1+perZ*(offset-gap))
				ncr := rl.NewVector3(x1+perX*(offset+gap), h1, z1+perZ*(offset+gap))
				drawQuad(cl, cr, ncl, ncr, center)
			}

			total := float32(lanes) * laneW
			half := total * 0.5
			eW := float32(0.2)
			el1 := rl.NewVector3(x0-perX*half, h0, z0-perZ*half)
			el2 := rl.NewVector3(x0-perX*(half-eW), h0, z0-perZ*(half-eW))
			nel1 := rl.NewVector3(x1-perX*half, h1, z1-perZ*half)
			nel2 := rl.NewVector3(x1-perX*(half-eW), h1, z1-perZ*(half-eW))
			drawQuad(el1, el2, nel1, nel2, edge)
			er1 := rl.NewVector3(x0+perX*(half-eW), h0, z0+perZ*(half-eW))
			er2 := rl.NewVector3(x0+perX*half, h0, z0+perZ*half)
			ner1 := rl.NewVector3(x1+perX*(half-eW), h1, z1+perZ*(half-eW))
			ner2 := rl.NewVector3(x1+perX*half, h1, z1+perZ*half)
			drawQuad(er1, er2, ner1, ner2, edge)
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
