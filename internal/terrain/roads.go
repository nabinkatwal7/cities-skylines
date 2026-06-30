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
	NextID   uint32
}

func NewRoadManager() *RoadManager {
	return &RoadManager{}
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

func (rm *RoadManager) Draw(h *Heightmap) {
	for _, seg := range rm.Segments {
		na := &rm.Nodes[seg.NodeA]
		nb := &rm.Nodes[seg.NodeB]

		dx := nb.X - na.X
		dz := nb.Z - na.Z
		length := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if length < 0.01 {
			continue
		}

		perX := -dz / length
		perZ := dx / length
		half := float32(3.0)
		px := perX * half
		pz := perZ * half

		ha := h.WorldHeight(na.X, na.Z) + 0.15
		hb := h.WorldHeight(nb.X, nb.Z) + 0.15

		a0 := rl.NewVector3(na.X-px, ha, na.Z-pz)
		a1 := rl.NewVector3(na.X+px, ha, na.Z+pz)
		b0 := rl.NewVector3(nb.X-px, hb, nb.Z-pz)
		b1 := rl.NewVector3(nb.X+px, hb, nb.Z+pz)

		col := rl.NewColor(80, 80, 80, 255)
		rl.DrawTriangle3D(a0, a1, b0, col)
		rl.DrawTriangle3D(b0, a1, b1, col)
	}
}

func (rm *RoadManager) Unload() {
}
