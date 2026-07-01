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
	RoadHighway
	RoadRoundabout
	RoadSixLane
	RoadAvenue
	RoadBus
	RoadTram
	RoadBicycle
	RoadTreeLined
	RoadAsymmetric
	RoadPedestrian
	RoadQuay
)

func RoadTypeName(rt RoadType) string {
	switch rt {
	case RoadTwoLane:
		return "2-Lane"
	case RoadOneWay:
		return "1-Way"
	case RoadFourLane:
		return "4-Lane"
	case RoadSixLane:
		return "6-Lane"
	case RoadAvenue:
		return "Avenue"
	case RoadGravel:
		return "Gravel"
	case RoadHighway:
		return "Highway"
	case RoadRoundabout:
		return "Roundabout"
	case RoadBus:
		return "Bus Rd"
	case RoadTram:
		return "Tram Rd"
	case RoadBicycle:
		return "Bike Rd"
	case RoadTreeLined:
		return "Tree Rd"
	case RoadAsymmetric:
		return "Asym Rd"
	case RoadPedestrian:
		return "Pedestrian"
	case RoadQuay:
		return "Quay"
	default:
		return "Road"
	}
}

func roadWidth(rt RoadType) float32 {
	lanes := roadLanes(rt)
	if lanes < 1 {
		return 4.0
	}
	return float32(lanes)*3.0 + 2.0
}

func roadHasSidewalk(rt RoadType) bool {
	switch rt {
	case RoadHighway, RoadBicycle, RoadPedestrian:
		return false
	default:
		return true
	}
}

func roadHasLighting(rt RoadType) bool {
	switch rt {
	case RoadGravel, RoadBicycle, RoadPedestrian:
		return false
	default:
		return true
	}
}

func roadAllowedVehicles(rt RoadType) string {
	switch rt {
	case RoadBus:
		return "bus"
	case RoadTram:
		return "tram"
	case RoadBicycle:
		return "bike"
	case RoadPedestrian:
		return "pedestrian"
	default:
		return "all"
	}
}

func roadNoise(rt RoadType) float32 {
	switch rt {
	case RoadPedestrian, RoadBicycle:
		return 0.1
	case RoadGravel, RoadQuay, RoadTreeLined:
		return 0.3
	case RoadTwoLane, RoadOneWay, RoadBus, RoadTram, RoadAsymmetric:
		return 0.5
	case RoadFourLane:
		return 0.7
	case RoadSixLane, RoadAvenue:
		return 0.8
	case RoadHighway:
		return 1.0
	default:
		return 0.5
	}
}

type RoadHierarchy uint8

const (
	HierarchyLocal RoadHierarchy = iota
	HierarchyCollector
	HierarchyArterial
	HierarchyHighway
)

const laneW = float32(3.0)

func roadLanes(rt RoadType) int {
	switch rt {
	case RoadOneWay, RoadBus:
		return 1
	case RoadBicycle:
		return 1
	case RoadPedestrian:
		return 1
	case RoadTwoLane, RoadGravel, RoadRoundabout, RoadTreeLined, RoadAsymmetric:
		return 2
	case RoadFourLane, RoadTram:
		return 4
	case RoadSixLane, RoadAvenue, RoadHighway:
		return 6
	case RoadQuay:
		return 2
	default:
		return 2
	}
}

func roadSpeed(rt RoadType) float32 {
	switch rt {
	case RoadGravel, RoadRoundabout:
		return 30
	case RoadPedestrian, RoadQuay:
		return 20
	case RoadBicycle:
		return 25
	case RoadTwoLane, RoadTreeLined, RoadAsymmetric:
		return 50
	case RoadOneWay:
		return 60
	case RoadBus:
		return 50
	case RoadTram:
		return 40
	case RoadFourLane:
		return 70
	case RoadSixLane, RoadAvenue:
		return 80
	case RoadHighway:
		return 100
	default:
		return 50
	}
}

func roadHierarchy(rt RoadType) RoadHierarchy {
	switch rt {
	case RoadGravel, RoadRoundabout, RoadPedestrian, RoadQuay, RoadBicycle:
		return HierarchyLocal
	case RoadTwoLane, RoadOneWay, RoadTreeLined, RoadAsymmetric, RoadBus, RoadTram:
		return HierarchyCollector
	case RoadFourLane, RoadSixLane, RoadAvenue:
		return HierarchyArterial
	case RoadHighway:
		return HierarchyHighway
	default:
		return HierarchyLocal
	}
}

type TrafficLightState uint8

const (
	TrafficLightNone TrafficLightState = iota
	TrafficLightRed
	TrafficLightYellow
	TrafficLightGreen
)

type RoadFlags uint32

const (
	RoadFlagNone        RoadFlags = 0
	RoadFlagOutsideConn RoadFlags = 1 << 0
	RoadFlagBridge      RoadFlags = 1 << 1
	RoadFlagTunnel      RoadFlags = 1 << 2
)

type RoadNode struct {
	ID                uint32
	X, Y, Z           float32
	Connected         []uint32
	TrafficLight      TrafficLightState
	TrafficLightPhase int32
	JunctionType      uint8 // 0=normal, 1=traffic_light, 2=roundabout
	Flags             RoadFlags
}

type CurveData struct {
	P1x, P1z float32
	P2x, P2z float32
}

type RoadSegment struct {
	ID               uint32
	NodeA            uint32
	NodeB            uint32
	RoadType         RoadType
	Length           float32
	SpeedLimit       float32
	LaneCount        int32
	Direction        int8 // 0=two-way, 1=forward (A→B), -1=reverse (B→A)
	Elevation        int32
	MaintenanceCost  float32
	ConstructionCost float32
	Damaged          bool
	Curve            CurveData
}

type RoadManager struct {
	Nodes      []RoadNode
	Segments   []RoadSegment
	Models     []rl.Model
	NextID     uint32
	roadTex    rl.Texture2D
	lightTimer int32
	nightMode  bool
}

func NewRoadManager() *RoadManager {
	return &RoadManager{}
}

func (rm *RoadManager) InitOutsideConnections(cs *ConnectionSystem) {
	for _, c := range cs.GetByType(ConnHighway) {
		idx := rm.AddNode(c.WorldX, 0, c.WorldZ)
		rm.Nodes[idx].Flags |= RoadFlagOutsideConn
	}
}

func (rm *RoadManager) LoadAssets() {
	tex := rl.LoadTexture("assets/road.jpg")
	if tex.ID != 0 {
		rm.roadTex = tex
	}
}

func (rm *RoadManager) AddNode(x, y, z float32) uint32 {
	id := rm.NextID
	rm.NextID++
	idx := uint32(len(rm.Nodes))
	rm.Nodes = append(rm.Nodes, RoadNode{
		ID:        id,
		X:         x,
		Y:         y,
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

	tx0, tz0 := rm.TangentAtNodeInternal(a, b)
	tx1, tz1 := rm.TangentAtNodeInternal(b, a)

	id := rm.NextID
	rm.NextID++
	rm.Segments = append(rm.Segments, RoadSegment{
		ID:               id,
		NodeA:            a,
		NodeB:            b,
		RoadType:         rt,
		Length:           length,
		SpeedLimit:       roadSpeed(rt),
		LaneCount:        int32(roadLanes(rt)),
		Direction:        0,
		MaintenanceCost:  roadMaintenanceCost(rt),
		ConstructionCost: roadConstructionCost(rt),
		Curve: CurveData{
			P1x: tx0, P1z: tz0,
			P2x: tx1, P2z: tz1,
		},
	})

	na.Connected = append(na.Connected, id)
	nb.Connected = append(nb.Connected, id)

	if len(na.Connected) >= 2 && na.TrafficLight == TrafficLightNone {
		na.TrafficLight = TrafficLightRed
		na.JunctionType = 1
	}
	if len(nb.Connected) >= 2 && nb.TrafficLight == TrafficLightNone {
		nb.TrafficLight = TrafficLightRed
		nb.JunctionType = 1
	}

	return id
}

func roadMaintenanceCost(rt RoadType) float32 {
	switch rt {
	case RoadGravel:
		return 0.5
	case RoadBicycle, RoadPedestrian:
		return 0.3
	case RoadTwoLane, RoadTreeLined, RoadAsymmetric:
		return 1.0
	case RoadOneWay:
		return 1.2
	case RoadFourLane:
		return 2.0
	case RoadSixLane, RoadAvenue:
		return 2.5
	case RoadHighway:
		return 3.5
	case RoadRoundabout, RoadBus, RoadTram, RoadQuay:
		return 1.5
	default:
		return 1.0
	}
}

func roadConstructionCost(rt RoadType) float32 {
	switch rt {
	case RoadGravel:
		return 50
	case RoadBicycle:
		return 40
	case RoadPedestrian:
		return 60
	case RoadTwoLane, RoadTreeLined, RoadAsymmetric:
		return 100
	case RoadOneWay:
		return 80
	case RoadFourLane:
		return 200
	case RoadSixLane:
		return 300
	case RoadAvenue:
		return 350
	case RoadHighway:
		return 500
	case RoadRoundabout:
		return 150
	case RoadBus, RoadTram:
		return 150
	case RoadQuay:
		return 120
	default:
		return 100
	}
}

func (rm *RoadManager) TangentAtNodeInternal(nodeIdx uint32, towards uint32) (float32, float32) {
	n := &rm.Nodes[nodeIdx]
	other := &rm.Nodes[towards]
	dx := other.X - n.X
	dz := other.Z - n.Z
	l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	if l < 0.01 {
		return 0, 0
	}
	return dx / l * l * 0.3, dz / l * l * 0.3
}

func (rm *RoadManager) TangentAtNode(nodeIdx uint32, incomingSegIdx int) (float32, float32) {
	n := &rm.Nodes[nodeIdx]
	if len(n.Connected) == 0 {
		return 0, 0
	}
	if len(n.Connected) == 1 || incomingSegIdx < 0 {
		seg := rm.Segments[n.Connected[0]]
		other := seg.NodeA
		if other == nodeIdx {
			other = seg.NodeB
		}
		on := &rm.Nodes[other]
		dx := on.X - n.X
		dz := on.Z - n.Z
		l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if l < 0.01 {
			return 0, 0
		}
		return dx / l * l * 0.3, dz / l * l * 0.3
	}

	var ax, az float32
	for _, sid := range n.Connected {
		if incomingSegIdx >= 0 && int(sid) == incomingSegIdx {
			continue
		}
		seg := rm.Segments[sid]
		other := seg.NodeA
		if other == nodeIdx {
			other = seg.NodeB
		}
		on := &rm.Nodes[other]
		dx := on.X - n.X
		dz := on.Z - n.Z
		l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if l > 0.01 {
			ax += dx / l
			az += dz / l
		}
	}
	al := float32(math.Sqrt(float64(ax*ax + az*az)))
	if al < 0.01 {
		return 0, 0
	}
	ax /= al
	az /= al

	seg := rm.Segments[n.Connected[0]]
	if int(seg.NodeA) == int(nodeIdx) || int(seg.NodeB) == int(nodeIdx) {
		seg = rm.Segments[n.Connected[len(n.Connected)-1]]
	}
	dx := float32(0)
	dz := float32(0)
	if int(seg.NodeA) == int(nodeIdx) {
		dx = rm.Nodes[seg.NodeB].X - n.X
		dz = rm.Nodes[seg.NodeB].Z - n.Z
	} else {
		dx = rm.Nodes[seg.NodeA].X - n.X
		dz = rm.Nodes[seg.NodeA].Z - n.Z
	}
	l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	scale := l * 0.3
	return ax * scale, az * scale
}

func cubicBezier(t float32, p0, p1, p2, p3 float32) float32 {
	u := 1 - t
	return u*u*u*p0 + 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t*p3
}

func (rm *RoadManager) SampleSegment(seg RoadSegment, steps int) ([]float32, []float32, []float32) {
	na := &rm.Nodes[seg.NodeA]
	nb := &rm.Nodes[seg.NodeB]

	tx0, tz0 := rm.TangentAtNode(seg.NodeA, int(seg.ID))
	tx1, tz1 := rm.TangentAtNode(seg.NodeB, int(seg.ID))

	xs := make([]float32, steps+1)
	zs := make([]float32, steps+1)
	ds := make([]float32, steps+1)

	for si := 0; si <= steps; si++ {
		t := float32(si) / float32(steps)
		x := cubicBezier(t, na.X, na.X+tx0, nb.X+tx1, nb.X)
		z := cubicBezier(t, na.Z, na.Z+tz0, nb.Z+tz1, nb.Z)
		xs[si] = x
		zs[si] = z
	}

	ds[0] = 0
	for si := 1; si <= steps; si++ {
		dx := xs[si] - xs[si-1]
		dz := zs[si] - zs[si-1]
		ds[si] = ds[si-1] + float32(math.Sqrt(float64(dx*dx+dz*dz)))
	}

	return xs, zs, ds
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

func (rm *RoadManager) HasNearbyRoad(x, z, maxDist float32) bool {
	for _, seg := range rm.Segments {
		xs, zs, _ := rm.SampleSegment(seg, 8)
		for i := 0; i < len(xs); i++ {
			dx := x - xs[i]
			dz := z - zs[i]
			if dx*dx+dz*dz < maxDist*maxDist {
				return true
			}
		}
	}
	return false
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
	xs, zs, _ := rm.SampleSegment(seg, int(seg.Length/2))
	steps := len(xs) - 1
	if steps < 1 {
		return rl.Model{}
	}

	lanes := seg.LaneCount
	if lanes < 1 {
		lanes = int32(int(seg.LaneCount))
	}
	total := float32(lanes) * laneW
	half := total * 0.5

	vertCount := (steps + 1) * 2
	triCount := steps * 2

	verts := make([]float32, 0, vertCount*3)
	normals := make([]float32, 0, vertCount*3)
	tc := make([]float32, 0, vertCount*2)
	indices := make([]uint16, 0, triCount*3)

	for si := 0; si <= steps; si++ {
		x := xs[si]
		z := zs[si]

		var perX, perZ float32
		if si < steps {
			dx := xs[si+1] - x
			dz := zs[si+1] - z
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l > 0.01 {
				perX = -dz / l
				perZ = dx / l
			}
		} else {
			dx := x - xs[si-1]
			dz := z - zs[si-1]
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l > 0.01 {
				perX = -dz / l
				perZ = dx / l
			}
		}

		var hgt float32
		if seg.Elevation > 0 {
			hgt = float32(seg.Elevation) * 5
		} else {
			hgt = h.WorldHeight(x, z) + 0.15
		}

		u := float32(si) / 4.0
		verts = append(verts, x-perX*half, hgt, z-perZ*half)
		verts = append(verts, x+perX*half, hgt, z+perZ*half)
		normals = append(normals, 0, 1, 0, 0, 1, 0)
		tc = append(tc, 0, u, 1, u)
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
		Texcoords:     &tc[0],
		Indices:       (*uint16)(unsafe.Pointer(&indices[0])),
	}
	rl.UploadMesh(&mesh, false)
	model := rl.LoadModelFromMesh(mesh)
	clearModelMeshData(&model)
	mats := model.GetMaterials()
	if len(mats) > 0 && rm.roadTex.ID != 0 {
		rl.SetMaterialTexture(&mats[0], rl.MapAlbedo, rm.roadTex)
	}
	return model
}

func (rm *RoadManager) SetNightMode(isNight bool) {
	rm.nightMode = isNight
}

func (rm *RoadManager) Update(h *Heightmap) {
	rm.lightTimer++
	if rm.lightTimer > 120 {
		rm.lightTimer = 0
		for i := range rm.Nodes {
			n := &rm.Nodes[i]
			if n.TrafficLight != TrafficLightNone {
				switch n.TrafficLight {
				case TrafficLightRed:
					n.TrafficLight = TrafficLightGreen
				case TrafficLightGreen:
					n.TrafficLight = TrafficLightYellow
				default:
					n.TrafficLight = TrafficLightRed
				}
			}
		}
	}

	for i := range rm.Segments {
		seg := &rm.Segments[i]
		na := &rm.Nodes[seg.NodeA]
		nb := &rm.Nodes[seg.NodeB]

		isFlooded := false
		if waterForRoads != nil {
			isFlooded = waterForRoads.IsFlooded(na.X, na.Z) || waterForRoads.IsFlooded(nb.X, nb.Z)
		}
		if !isFlooded {
			ah := h.WorldHeight(na.X, na.Z)
			bh := h.WorldHeight(nb.X, nb.Z)
			waterH := float32(SeaLevel*MaxHeight + 0.1)
			isFlooded = ah < waterH || bh < waterH
		}
		if isFlooded {
			if !seg.Damaged {
				seg.Damaged = true
			}
		} else {
			seg.Damaged = false
		}
	}
}

var waterForRoads *WaterSystem

func SetWaterForRoads(ws *WaterSystem) {
	waterForRoads = ws
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
	rm.drawJunctionMarkings(h)
	rm.drawOutsideConnections(h)
}

func (rm *RoadManager) drawFallback(h *Heightmap) {
	for _, seg := range rm.Segments {
		xs, zs, ds := rm.SampleSegment(seg, int(seg.Length/2))
		if len(xs) < 2 {
			continue
		}
		totalLen := ds[len(ds)-1]
		if totalLen < 0.01 {
			continue
		}

		lanes := int(seg.LaneCount)
		total := float32(lanes) * laneW
		half := total * 0.5
		col := rl.NewColor(80, 80, 80, 255)

		for si := 0; si < len(xs)-1; si++ {
			x0, z0 := xs[si], zs[si]
			x1, z1 := xs[si+1], zs[si+1]
			dx := x1 - x0
			dz := z1 - z0
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l < 0.01 {
				continue
			}
			perX := -dz / l
			perZ := dx / l
			var h0, h1 float32
			if seg.Elevation > 0 {
				h0 = float32(seg.Elevation) * 5
				h1 = float32(seg.Elevation) * 5
			} else {
				h0 = h.WorldHeight(x0, z0) + 0.15
				h1 = h.WorldHeight(x1, z1) + 0.15
			}

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
		if seg.Damaged {
			continue
		}
		xs, zs, ds := rm.SampleSegment(seg, int(seg.Length/2))
		if len(xs) < 2 {
			continue
		}
		totalLen := ds[len(ds)-1]
		if totalLen < 0.01 {
			continue
		}

		lanes := int(seg.LaneCount)
		total := float32(lanes) * laneW
		half := total * 0.5

		for si := 0; si < len(xs)-1; si++ {
			x0, z0 := xs[si], zs[si]
			x1, z1 := xs[si+1], zs[si+1]
			dx := x1 - x0
			dz := z1 - z0
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l < 0.01 {
				continue
			}
			perX := -dz / l
			perZ := dx / l

			var h0, h1 float32
			if seg.Elevation > 0 {
				h0 = float32(seg.Elevation) * 5
				h1 = float32(seg.Elevation) * 5
			} else {
				h0 = h.WorldHeight(x0, z0) + 0.15
				h1 = h.WorldHeight(x1, z1) + 0.15
			}

			for li := 0; li < lanes-1; li++ {
				offset := (float32(li) - float32(lanes-1)*0.5 + 1) * laneW
				gap := float32(0.15)
				cl := rl.NewVector3(x0+perX*(offset-gap), h0, z0+perZ*(offset-gap))
				cr := rl.NewVector3(x0+perX*(offset+gap), h0, z0+perZ*(offset+gap))
				ncl := rl.NewVector3(x1+perX*(offset-gap), h1, z1+perZ*(offset-gap))
				ncr := rl.NewVector3(x1+perX*(offset+gap), h1, z1+perZ*(offset+gap))
				drawQuad(cl, cr, ncl, ncr, center)
			}

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

func (rm *RoadManager) drawJunctionMarkings(h *Heightmap) {
	for i := range rm.Nodes {
		n := &rm.Nodes[i]
		if n.TrafficLight == TrafficLightNone || len(n.Connected) < 2 {
			continue
		}
		hy := h.WorldHeight(n.X, n.Z) + 0.2
		for _, sid := range n.Connected {
			if rm.Segments[sid].Elevation > 0 {
				hy = float32(rm.Segments[sid].Elevation) * 5
				break
			}
		}
		col := rl.NewColor(200, 100, 100, 200)
		if n.JunctionType == 2 {
			col = rl.NewColor(100, 100, 200, 200)
		}

		rl.DrawCube(rl.NewVector3(n.X, hy, n.Z), 2, 0.1, 2, col)

		for _, sid := range n.Connected {
			seg := rm.Segments[sid]
			other := seg.NodeA
			if other == uint32(i) {
				other = seg.NodeB
			}
			on := &rm.Nodes[other]
			dx := on.X - n.X
			dz := on.Z - n.Z
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l < 0.01 {
				continue
			}

			perX := -dz / l
			perZ := dx / l

			lanes := int(seg.LaneCount)
			total := float32(lanes) * laneW
			half := total * 0.5

			stopX := n.X + dx/l*1.5
			stopZ := n.Z + dz/l*1.5

			rl.DrawCube(rl.NewVector3(stopX, hy, stopZ), 0.3, 0.05, total*0.8, rl.NewColor(255, 255, 255, 200))

			crossX := n.X + dx/l*3
			crossZ := n.Z + dz/l*3
			for ci := 0; ci < 3; ci++ {
				off := float32(ci)*0.5 - 0.5
				rl.DrawCube(rl.NewVector3(crossX+perX*half*off, hy, crossZ+perZ*half*off), 0.2, 0.05, total*0.3, rl.NewColor(255, 255, 255, 180))
			}
		}

		if n.TrafficLight != TrafficLightNone {
			tlCol := rl.NewColor(255, 200, 50, 255)
			switch n.TrafficLight {
			case TrafficLightRed:
				tlCol = rl.NewColor(255, 0, 0, 255)
			case TrafficLightYellow:
				tlCol = rl.NewColor(255, 200, 0, 255)
			case TrafficLightGreen:
				tlCol = rl.NewColor(0, 255, 0, 255)
			}
			rl.DrawSphere(rl.NewVector3(n.X, hy+2.5, n.Z), 0.4, tlCol)
			rl.DrawCube(rl.NewVector3(n.X, hy+1.8, n.Z), 0.1, 1.5, 0.1, rl.NewColor(60, 60, 60, 200))
		}
	}
}

func (rm *RoadManager) drawOutsideConnections(h *Heightmap) {
	for _, n := range rm.Nodes {
		if n.Flags&RoadFlagOutsideConn != 0 {
			hy := h.WorldHeight(n.X, n.Z)
			rl.DrawCube(rl.NewVector3(n.X, hy+1, n.Z), 8, 2, 8, rl.NewColor(150, 100, 50, 200))
		}
	}
}

func (rm *RoadManager) AddShortSegment(x1, z1, x2, z2 float32, rt RoadType) {
	na := rm.AddNode(x1, 0, z1)
	nb := rm.AddNode(x2, 0, z2)
	rm.AddSegment(na, nb, rt)
}

func (rm *RoadManager) NearestSegment(x, z float32) int {
	bestIdx := -1
	bestDist := float32(math.MaxFloat32)
	for i, seg := range rm.Segments {
		xs, zs, _ := rm.SampleSegment(seg, 8)
		for j := 0; j < len(xs); j++ {
			dx := x - xs[j]
			dz := z - zs[j]
			d := dx*dx + dz*dz
			if d < bestDist {
				bestDist = d
				bestIdx = i
			}
		}
	}
	if bestDist > 100 {
		return -1
	}
	return bestIdx
}

func (rm *RoadManager) RemoveSegment(idx int) {
	if idx < 0 || idx >= len(rm.Segments) {
		return
	}
	seg := rm.Segments[idx]

	if idx < len(rm.Models) {
		if rm.Models[idx].MeshCount > 0 {
			rl.UnloadModel(rm.Models[idx])
		}
	}

	na := &rm.Nodes[seg.NodeA]
	nb := &rm.Nodes[seg.NodeB]

	filter := func(s []uint32, id uint32) []uint32 {
		out := s[:0]
		for _, v := range s {
			if v != id {
				out = append(out, v)
			}
		}
		return out
	}
	na.Connected = filter(na.Connected, seg.ID)
	nb.Connected = filter(nb.Connected, seg.ID)

	if len(na.Connected) < 2 {
		na.TrafficLight = TrafficLightNone
		na.JunctionType = 0
	}
	if len(nb.Connected) < 2 {
		nb.TrafficLight = TrafficLightNone
		nb.JunctionType = 0
	}

	rm.Segments = append(rm.Segments[:idx], rm.Segments[idx+1:]...)
	rm.Models = append(rm.Models[:idx], rm.Models[idx+1:]...)

	removeA := len(na.Connected) == 0 && na.Flags&RoadFlagOutsideConn == 0
	removeB := len(nb.Connected) == 0 && nb.Flags&RoadFlagOutsideConn == 0
	if removeA && removeB {
		if seg.NodeA > seg.NodeB {
			rm.removeNodeByIndex(seg.NodeA)
			rm.removeNodeByIndex(seg.NodeB)
		} else {
			rm.removeNodeByIndex(seg.NodeB)
			rm.removeNodeByIndex(seg.NodeA)
		}
	} else if removeA {
		rm.removeNodeByIndex(seg.NodeA)
	} else if removeB {
		rm.removeNodeByIndex(seg.NodeB)
	}
}

func (rm *RoadManager) UpgradeSegment(idx int, newType RoadType) {
	if idx < 0 || idx >= len(rm.Segments) {
		return
	}
	s := &rm.Segments[idx]
	s.RoadType = newType
	s.SpeedLimit = roadSpeed(newType)
	s.LaneCount = int32(roadLanes(newType))
	s.MaintenanceCost = roadMaintenanceCost(newType)
	s.ConstructionCost = roadConstructionCost(newType)
}

func (rm *RoadManager) removeNodeByIndex(idx uint32) {
	rm.Nodes = append(rm.Nodes[:idx], rm.Nodes[idx+1:]...)
	for i := range rm.Segments {
		if rm.Segments[i].NodeA > idx {
			rm.Segments[i].NodeA--
		}
		if rm.Segments[i].NodeB > idx {
			rm.Segments[i].NodeB--
		}
	}
}

func (rm *RoadManager) FindPath(startNode, endNode uint32, vehicleType int) []uint32 {
	if startNode == endNode {
		return nil
	}
	type nodeDist struct {
		prev  int32
		dist  float32
		visited bool
	}
	nodes := make([]nodeDist, len(rm.Nodes))
	for i := range nodes {
		nodes[i].prev = -1
		nodes[i].dist = math.MaxFloat32
	}
	nodes[startNode].dist = 0

	for {
		best := -1
		bestD := float32(math.MaxFloat32)
		for i := range nodes {
			if !nodes[i].visited && nodes[i].dist < bestD {
				best = i
				bestD = nodes[i].dist
			}
		}
		if best < 0 || uint32(best) == endNode {
			break
		}
		nodes[best].visited = true
		cn := &rm.Nodes[best]

		for _, sid := range cn.Connected {
			seg := rm.Segments[sid]
			if seg.Damaged {
				continue
			}
			other := seg.NodeA
			if other == uint32(best) {
				other = seg.NodeB
			}
			cost := seg.Length / seg.SpeedLimit
			nd := nodes[best].dist + cost
			if nd < nodes[other].dist {
				nodes[other].dist = nd
				nodes[other].prev = int32(best)
			}
		}
	}

	if nodes[endNode].prev < 0 {
		return nil
	}

	path := make([]uint32, 0)
	cur := int32(endNode)
	for cur >= 0 {
		path = append(path, uint32(cur))
		cur = nodes[cur].prev
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
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
