package road

import (
	"math"

	"github.com/katwate/js-skylines/internal/terrain"
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

func RoadLanes(rt RoadType) int {
	return roadLanes(rt)
}

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

func DefaultSpeed(rt RoadType) float32 {
	return roadSpeed(rt)
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

// HierarchyForRoad returns the hierarchy class for a road type.
func HierarchyForRoad(rt RoadType) RoadHierarchy {
	return roadHierarchy(rt)
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

type TrafficRule uint8

const (
	RuleNone TrafficRule = iota
	RuleStop
	RuleYield
	RuleTrafficLight
	RulePriorityRoad
	RuleRoundabout
)

func (rm *RoadManager) TrafficRuleForNode(nodeIdx uint32) TrafficRule {
	if nodeIdx >= uint32(len(rm.Nodes)) {
		return RuleNone
	}
	n := &rm.Nodes[nodeIdx]
	if len(n.Connected) < 2 {
		return RuleNone
	}
	if n.JunctionType == 2 {
		return RuleRoundabout
	}
	if n.TrafficLight != TrafficLightNone {
		return RuleTrafficLight
	}

	maxHierarchy := HierarchyLocal
	for _, sid := range n.Connected {
		seg := rm.SegmentByID(sid)
		if seg == nil {
			continue
		}
		h := roadHierarchy(seg.RoadType)
		if h > maxHierarchy {
			maxHierarchy = h
		}
	}

	if maxHierarchy >= HierarchyArterial {
		return RulePriorityRoad
	}

	if len(n.Connected) > 3 {
		return RuleTrafficLight
	}

	return RuleYield
}

func elevationHeight(elevation int32) float32 {
	if elevation >= 0 {
		return float32(elevation) * 5.0
	}
	return float32(elevation) * 5.0
}

func isElevated(elevation int32) bool { return elevation > 0 }

func isTunnel(elevation int32) bool { return elevation < 0 }

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

type LaneCategory uint8

const (
	LaneDriving LaneCategory = iota
	LaneParking
	LaneBus
	LaneTram
	LaneBicycle
	LanePedestrian
	LaneEmergency
	LaneTurning
	LaneHighwayMerge
)

func LaneCategoriesForRoad(rt RoadType) []LaneCategory {
	switch rt {
	case RoadOneWay:
		return []LaneCategory{LaneDriving}
	case RoadBus:
		return []LaneCategory{LaneBus}
	case RoadTram:
		return []LaneCategory{LaneTram, LaneTram, LaneTram, LaneTram}
	case RoadBicycle:
		return []LaneCategory{LaneBicycle}
	case RoadPedestrian:
		return []LaneCategory{LanePedestrian}
	case RoadTreeLined:
		return []LaneCategory{LaneDriving, LaneParking}
	case RoadAsymmetric:
		return []LaneCategory{LaneDriving, LaneParking}
	case RoadHighway:
		return []LaneCategory{LaneDriving, LaneDriving, LaneDriving, LaneDriving, LaneDriving, LaneDriving}
	case RoadQuay:
		return []LaneCategory{LaneDriving, LaneDriving}
	case RoadSixLane:
		return []LaneCategory{LaneDriving, LaneDriving, LaneDriving, LaneDriving, LaneDriving, LaneDriving}
	case RoadAvenue:
		return []LaneCategory{LaneDriving, LaneDriving, LaneDriving, LaneDriving, LaneDriving, LaneDriving}
	default:
		return []LaneCategory{LaneDriving, LaneDriving}
	}
}

func laneAllowsVehicle(cat LaneCategory, vt VehicleType) bool {
	switch cat {
	case LaneDriving:
		return vt == VehicleCar || vt == VehicleBus || vt == VehicleTruck || vt == VehicleEmergency
	case LaneParking:
		return vt == VehicleCar
	case LaneBus:
		return vt == VehicleBus
	case LaneTram:
		return vt == VehicleTram
	case LaneBicycle:
		return vt == VehicleBike
	case LanePedestrian:
		return vt == VehiclePedestrian
	case LaneEmergency:
		return vt == VehicleEmergency
	case LaneTurning, LaneHighwayMerge:
		return vt == VehicleCar || vt == VehicleBus || vt == VehicleTruck || vt == VehicleEmergency
	default:
		return true
	}
}

type Lane struct {
	Index      int32
	Direction  int8        // 0=forward (A→B), 1=reverse (B→A)
	SpeedLimit float32
	Category   LaneCategory
	Width      float32
	Priority   int32
}

type LaneTurn int8

const (
	LaneTurnStraight LaneTurn = iota
	LaneTurnLeft
	LaneTurnRight
	LaneTurnUTurn
)

type LaneConnection struct {
	FromSegIdx int32
	FromLane   int32
	ToSegIdx   int32
	ToLane     int32
	Turn       LaneTurn
}

func generateLanes(rt RoadType, direction int8, speedLimit float32, laneCount int32) []Lane {
	lanes := make([]Lane, laneCount)
	half := laneCount / 2
	categories := LaneCategoriesForRoad(rt)

	for i := int32(0); i < laneCount; i++ {
		var dir int8
		if direction == 1 {
			dir = 0
		} else if direction == -1 {
			dir = 1
		} else {
			if i < half {
				dir = 0
			} else {
				dir = 1
			}
		}

		cat := LaneDriving
		if int(i) < len(categories) {
			cat = categories[i]
		}

		lanes[i] = Lane{
			Index:      i,
			Direction:  dir,
			SpeedLimit: speedLimit,
			Category:   cat,
			Width:      3.0,
			Priority:   i,
		}
	}
	return lanes
}

func (rm *RoadManager) JunctionLaneRoutes(nodeIdx uint32) []LaneConnection {
	if rm.junctionCacheDirty {
		rm.junctionCache = make(map[uint32][]LaneConnection)
		rm.junctionCacheDirty = false
	}
	if cached, ok := rm.junctionCache[nodeIdx]; ok {
		return cached
	}

	n := &rm.Nodes[nodeIdx]
	if len(n.Connected) < 2 {
		rm.junctionCache[nodeIdx] = nil
		return nil
	}

	type approach struct {
		segIdx     int
		angle      float32
		dirX, dirZ float32
	}
	approaches := make([]approach, 0, len(n.Connected))
	for _, sid := range n.Connected {
		seg := rm.SegmentByID(sid)
		if seg == nil {
			continue
		}
		other := seg.NodeA
		if other == nodeIdx {
			other = seg.NodeB
		}
		on := &rm.Nodes[other]
		dx := on.X - n.X
		dz := on.Z - n.Z
		l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if l < 0.01 {
			continue
		}
		angle := float32(math.Atan2(float64(dz), float64(dx)))
		approaches = append(approaches, approach{segIdx: int(sid), angle: angle, dirX: dx / l, dirZ: dz / l})
	}
	if len(approaches) < 1 {
		rm.junctionCache[nodeIdx] = nil
		return nil
	}

	for i := 0; i < len(approaches); i++ {
		for j := i + 1; j < len(approaches); j++ {
			if approaches[j].angle < approaches[i].angle {
				approaches[i], approaches[j] = approaches[j], approaches[i]
			}
		}
	}

	connections := make([]LaneConnection, 0)
	for _, src := range approaches {
		seg := rm.SegmentByID(uint32(src.segIdx))
		if seg == nil {
			continue
		}
		lanes := int(seg.LaneCount)
		for li := 0; li < lanes; li++ {
			laneDir := int8(0)
			if li < len(seg.Lanes) {
				laneDir = seg.Lanes[li].Direction
			}
			approachesTowardNode := (seg.NodeA == nodeIdx && laneDir == 0) || (seg.NodeB == nodeIdx && laneDir == 1)
			if !approachesTowardNode {
				continue
			}

			bestTurn := LaneTurnStraight
			bestAngle := float32(math.MaxFloat32)
			var bestDest int

			for _, dst := range approaches {
				if dst.segIdx == src.segIdx {
					continue
				}
				angleDiff := dst.angle - src.angle
				for angleDiff > math.Pi {
					angleDiff -= 2 * math.Pi
				}
				for angleDiff < -math.Pi {
					angleDiff += 2 * math.Pi
				}
				absDiff := float32(math.Abs(float64(angleDiff)))
				if absDiff < bestAngle {
					bestAngle = absDiff
					if absDiff < 0.5 {
						bestTurn = LaneTurnStraight
					} else if angleDiff > 0 {
						bestTurn = LaneTurnRight
					} else {
						bestTurn = LaneTurnLeft
					}
					bestDest = dst.segIdx
				}
			}

			if bestAngle > 2.5 {
				bestTurn = LaneTurnUTurn
			}

			dstLane := int32(0)
			dstSeg := rm.SegmentByID(uint32(bestDest))
			if dstSeg == nil {
				continue
			}
			dstLanes := int(dstSeg.LaneCount)
			if dstLanes > 0 {
				dstLane = int32(li % dstLanes)
			}

			connections = append(connections, LaneConnection{
				FromSegIdx: int32(src.segIdx),
				FromLane:   int32(li),
				ToSegIdx:   int32(bestDest),
				ToLane:     dstLane,
				Turn:       bestTurn,
			})
		}
	}
	rm.junctionCache[nodeIdx] = connections
	return connections
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
	RepairTimer      int32
	Curve            CurveData
	Lanes            []Lane
}

type RoadManager struct {
	Nodes      []RoadNode
	Segments   []RoadSegment
	Models     []rl.Model
	NextID     uint32
	roadTex    rl.Texture2D
	nightMode  bool

	junctionCache      map[uint32][]LaneConnection
	junctionCacheDirty bool
	dirtyModels        []int
	nearestGrid        [][]int
	nearestGridSize    int
	floodTimer         int32
}

func NewRoadManager() *RoadManager {
	return &RoadManager{
		junctionCache:   make(map[uint32][]LaneConnection),
		nearestGridSize: 64,
	}
}

func (rm *RoadManager) invalidateCaches() {
	rm.junctionCacheDirty = true
	rm.nearestGrid = nil
}

func (rm *RoadManager) markDirty(segIdx int) {
	for _, d := range rm.dirtyModels {
		if d == segIdx {
			return
		}
	}
	rm.dirtyModels = append(rm.dirtyModels, segIdx)
}

func (rm *RoadManager) InitOutsideConnections(cs *terrain.ConnectionSystem) {
	for _, c := range cs.GetByType(terrain.ConnHighway) {
		idx := rm.AddNode(c.WorldX, 0, c.WorldZ)
		rm.Nodes[idx].Flags |= RoadFlagOutsideConn
	}
}

func (rm *RoadManager) LoadAssets() {
	tex := rl.LoadTexture("assets/textures/road.png")
	if tex.ID != 0 {
		rm.roadTex = tex
		rm.markAllDirty()
	}
}

func (rm *RoadManager) SetRoadTexture(tex rl.Texture2D) {
	if rm.roadTex.ID != 0 && rm.roadTex.ID != tex.ID {
		rl.UnloadTexture(rm.roadTex)
	}
	rm.roadTex = tex
	rm.markAllDirty()
}

func (rm *RoadManager) ClearRoadTexture() {
	rm.roadTex = rl.Texture2D{}
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

	tx0 := (nb.X - na.X) * 0.3
	tz0 := (nb.Z - na.Z) * 0.3
	tx1 := (na.X - nb.X) * 0.3
	tz1 := (na.Z - nb.Z) * 0.3

	id := rm.NextID
	rm.NextID++
	lanes := generateLanes(rt, 0, roadSpeed(rt), int32(roadLanes(rt)))
	rm.Segments = append(rm.Segments, RoadSegment{
		ID:               id,
		NodeA:            a,
		NodeB:            b,
		RoadType:         rt,
		Length:           length,
		SpeedLimit:       roadSpeed(rt),
		LaneCount:        int32(roadLanes(rt)),
		Direction:        0,
		MaintenanceCost:  rm.CalcSegmentMaintenance(rt, length, 0),
		ConstructionCost: RoadConstructionCost(rt),
		Curve: CurveData{
			P1x: tx0, P1z: tz0,
			P2x: tx1, P2z: tz1,
		},
		Lanes: lanes,
	})

	na.Connected = append(na.Connected, id)
	nb.Connected = append(nb.Connected, id)

	rm.updateJunctionType(a)
	rm.updateJunctionType(b)
	rm.invalidateCaches()

	segIdx := len(rm.Segments) - 1
	rm.queueMeshRebuild(segIdx)

	return id
}

func (rm *RoadManager) SegmentIndex(id uint32) int {
	for i := range rm.Segments {
		if rm.Segments[i].ID == id {
			return i
		}
	}
	return -1
}

func (rm *RoadManager) HasSegmentBetween(a, b uint32) bool {
	for _, s := range rm.Segments {
		if (s.NodeA == a && s.NodeB == b) || (s.NodeA == b && s.NodeB == a) {
			return true
		}
	}
	return false
}

func (rm *RoadManager) queueMeshRebuild(segIdx int) {
	for len(rm.Models) <= segIdx {
		rm.Models = append(rm.Models, rl.Model{})
	}
	rm.markDirty(segIdx)
}

// SyncMeshes rebuilds GPU meshes for any new or dirty segments.
func (rm *RoadManager) SyncMeshes(h *terrain.Heightmap) {
	for len(rm.Models) < len(rm.Segments) {
		rm.queueMeshRebuild(len(rm.Models))
	}
	rm.Rebuild(h)
}

// ClearModels unloads GPU road meshes (e.g. before save load).
func (rm *RoadManager) ClearModels() {
	for _, model := range rm.Models {
		if model.MeshCount > 0 {
			rl.UnloadModel(model)
		}
	}
	rm.Models = nil
	rm.dirtyModels = nil
}

func (rm *RoadManager) ValidNodeIndex(idx uint32) bool {
	return int(idx) >= 0 && int(idx) < len(rm.Nodes)
}

func (rm *RoadManager) PendingMeshRebuilds() int {
	return len(rm.dirtyModels)
}

func (rm *RoadManager) updateJunctionType(nodeIdx uint32) {
	n := &rm.Nodes[nodeIdx]
	if len(n.Connected) < 2 {
		n.TrafficLight = TrafficLightNone
		n.JunctionType = 0
		return
	}
	isRoundabout := false
	for _, sid := range n.Connected {
		seg := rm.SegmentByID(sid)
		if seg == nil {
			// ponytail: tolerate stale segment IDs; upgrade path is storing segment indices not IDs
			continue
		}
		if seg.RoadType == RoadRoundabout {
			isRoundabout = true
			break
		}
	}
	if isRoundabout {
		n.TrafficLight = TrafficLightNone
		n.JunctionType = 2
		return
	}
	if n.JunctionType != 2 {
		if n.TrafficLight == TrafficLightNone {
			n.TrafficLight = TrafficLightRed
		}
		n.JunctionType = 1
	}
}

func (rm *RoadManager) SegmentByID(id uint32) *RoadSegment {
	// ponytail: O(n) scan; upgrade path is an ID->index map updated on add/remove
	for i := range rm.Segments {
		if rm.Segments[i].ID == id {
			return &rm.Segments[i]
		}
	}
	return nil
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

func (rm *RoadManager) CalcSegmentMaintenance(rt RoadType, length float32, elevation int32) float32 {
	base := roadMaintenanceCost(rt)
	lengthFactor := length / 10.0
	elevFactor := float32(1.0)
	switch {
	case elevation > 0:
		elevFactor = 1.5
	case elevation < 0:
		elevFactor = 2.5
	case rm.anyElevated():
		if elevation == 0 {
			break
		}
	}
	decFactor := float32(1.0)
	if roadHasSidewalk(rt) {
		decFactor += 0.2
	}
	if roadHasLighting(rt) {
		decFactor += 0.1
	}
	if rt == RoadTreeLined {
		decFactor += 0.3
	}
	return base * lengthFactor * elevFactor * decFactor
}

func RoadConstructionCost(rt RoadType) float32 {
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
		seg := rm.SegmentByID(n.Connected[0])
		if seg == nil {
			return 0, 0
		}
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
		seg := rm.SegmentByID(sid)
		if seg == nil {
			continue
		}
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

	seg := rm.SegmentByID(n.Connected[0])
	if seg == nil {
		return 0, 0
	}
	if len(n.Connected) > 1 && (int(seg.NodeA) == int(nodeIdx) || int(seg.NodeB) == int(nodeIdx)) {
		last := rm.SegmentByID(n.Connected[len(n.Connected)-1])
		if last != nil {
			seg = last
		}
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
	if steps < 2 {
		steps = 2
	}
	na := &rm.Nodes[seg.NodeA]
	nb := &rm.Nodes[seg.NodeB]

	xs := make([]float32, steps+1)
	zs := make([]float32, steps+1)
	ds := make([]float32, steps+1)

	if seg.RoadType == RoadRoundabout {
		tx0, tz0 := seg.Curve.P1x, seg.Curve.P1z
		tx1, tz1 := seg.Curve.P2x, seg.Curve.P2z
		for si := 0; si <= steps; si++ {
			t := float32(si) / float32(steps)
			xs[si] = cubicBezier(t, na.X, na.X+tx0, nb.X+tx1, nb.X)
			zs[si] = cubicBezier(t, na.Z, na.Z+tz0, nb.Z+tz1, nb.Z)
		}
	} else {
		for si := 0; si <= steps; si++ {
			t := float32(si) / float32(steps)
			xs[si] = na.X + (nb.X-na.X)*t
			zs[si] = na.Z + (nb.Z-na.Z)*t
		}
	}

	ds[0] = 0
	for si := 1; si <= steps; si++ {
		dx := xs[si] - xs[si-1]
		dz := zs[si] - zs[si-1]
		ds[si] = ds[si-1] + float32(math.Sqrt(float64(dx*dx+dz*dz)))
	}

	return xs, zs, ds
}

func (rm *RoadManager) SampleLane(seg RoadSegment, laneIdx int32, steps int) ([]float32, []float32, []float32) {
	xs, zs, ds := rm.SampleSegment(seg, steps)
	if laneIdx < 0 || int(laneIdx) >= len(seg.Lanes) {
		return xs, zs, ds
	}
	lanes := int32(len(seg.Lanes))
	laneW := seg.Lanes[laneIdx].Width
	offset := (float32(laneIdx) - float32(lanes-1)*0.5) * laneW

	lxs := make([]float32, len(xs))
	lzs := make([]float32, len(zs))
	for i := 0; i < len(xs); i++ {
		var perX, perZ float32
		if i < len(xs)-1 {
			dx := xs[i+1] - xs[i]
			dz := zs[i+1] - zs[i]
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l > 0.01 {
				perX = -dz / l
				perZ = dx / l
			}
		} else if i > 0 {
			dx := xs[i] - xs[i-1]
			dz := zs[i] - zs[i-1]
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l > 0.01 {
				perX = -dz / l
				perZ = dx / l
			}
		}
		lxs[i] = xs[i] + perX*offset
		lzs[i] = zs[i] + perZ*offset
	}
	return lxs, lzs, ds
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
	half := float32(terrain.WorldSize / 2)
	size := float32(terrain.WorldSize) / float32(rm.nearestGridSize)
	if rm.nearestGrid == nil {
		rm.buildNearestGrid()
	}
	gx := int((x + half) / size)
	gz := int((z + half) / size)
	if gx >= 0 && gx < rm.nearestGridSize && gz >= 0 && gz < rm.nearestGridSize {
		cellIdx := gz*rm.nearestGridSize + gx
		for _, segIdx := range rm.nearestGrid[cellIdx] {
			seg := rm.Segments[segIdx]
			xs, zs, _ := rm.SampleSegment(seg, 8)
			for i := 0; i < len(xs); i++ {
				dx := x - xs[i]
				dz := z - zs[i]
				if dx*dx+dz*dz < maxDist*maxDist {
					return true
				}
			}
		}
	}
	return false
}

// NearestRoad returns the road type and world distance to the closest sampled point on the nearest segment.
func (rm *RoadManager) NearestRoad(x, z float32) (RoadType, float32, bool) {
	idx := rm.NearestSegment(x, z)
	if idx < 0 {
		return 0, 0, false
	}
	seg := rm.Segments[idx]
	xs, zs, _ := rm.SampleSegment(seg, 16)
	best := float32(math.MaxFloat32)
	for i := range xs {
		dx := x - xs[i]
		dz := z - zs[i]
		d := dx*dx + dz*dz
		if d < best {
			best = d
		}
	}
	return seg.RoadType, float32(math.Sqrt(float64(best))), true
}

func (rm *RoadManager) Rebuild(h *terrain.Heightmap) {
	if len(rm.dirtyModels) == 0 {
		return
	}
	for _, idx := range rm.dirtyModels {
		if idx >= 0 && idx < len(rm.Models) {
			if rm.Models[idx].MeshCount > 0 {
				rl.UnloadModel(rm.Models[idx])
			}
			rm.Models[idx] = rm.buildSurfaceMesh(h, rm.Segments[idx])
		}
	}
	rm.dirtyModels = nil
}

func (rm *RoadManager) markAllDirty() {
	rm.dirtyModels = make([]int, len(rm.Segments))
	for i := range rm.Segments {
		rm.dirtyModels[i] = i
	}
}

func (rm *RoadManager) UploadGPU(h *terrain.Heightmap) {
	rm.Models = make([]rl.Model, len(rm.Segments))
	rm.markAllDirty()
	rm.Rebuild(h)
}

func (rm *RoadManager) buildSurfaceMesh(h *terrain.Heightmap, seg RoadSegment) rl.Model {
	steps := int(seg.Length / 2)
	if steps < 4 {
		steps = 4
	}
	xs, zs, ds := rm.SampleSegment(seg, steps)
	stripSteps := len(xs) - 1
	if stripSteps < 1 {
		return rl.Model{}
	}

	lanes := seg.LaneCount
	if lanes < 1 {
		lanes = int32(roadLanes(seg.RoadType))
	}
	total := float32(lanes) * laneW
	half := total * 0.5

	vertCount := (stripSteps + 1) * 2
	triCount := stripSteps * 2

	verts := make([]float32, 0, vertCount*3)
	normals := make([]float32, 0, vertCount*3)
	tc := make([]float32, 0, vertCount*2)
	indices := make([]uint16, 0, triCount*3)

	const texRepeat = float32(4.0) // world meters per texture repeat

	for si := 0; si <= stripSteps; si++ {
		x := xs[si]
		z := zs[si]

		var perX, perZ float32
		if si < stripSteps {
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
			hgt = h.WorldHeight(x, z) + 0.12
		}

		u := ds[si] / texRepeat
		verts = append(verts, x-perX*half, hgt, z-perZ*half)
		verts = append(verts, x+perX*half, hgt, z+perZ*half)
		normals = append(normals, 0, 1, 0, 0, 1, 0)
		tc = append(tc, 0, u, 1, u)
	}

	for si := 0; si < stripSteps; si++ {
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

func (rm *RoadManager) Update(h *terrain.Heightmap) {
	for i := range rm.Nodes {
		n := &rm.Nodes[i]
		if n.TrafficLight == TrafficLightNone {
			continue
		}
		if n.TrafficLightPhase == 0 {
			maxLanes := int32(2)
			for _, sid := range n.Connected {
				seg := rm.SegmentByID(sid)
				if seg == nil {
					continue
				}
				l := seg.LaneCount
				if l > maxLanes {
					maxLanes = l
				}
			}
			n.TrafficLightPhase = maxLanes * 20
		}

		n.TrafficLightPhase--
		if n.TrafficLightPhase > 0 {
			continue
		}

		switch n.TrafficLight {
		case TrafficLightRed:
			n.TrafficLight = TrafficLightGreen
			maxLanes := int32(2)
			for _, sid := range n.Connected {
				seg := rm.SegmentByID(sid)
				if seg == nil {
					continue
				}
				l := seg.LaneCount
				if l > maxLanes {
					maxLanes = l
				}
			}
			n.TrafficLightPhase = maxLanes * 30
		case TrafficLightGreen:
			n.TrafficLight = TrafficLightYellow
			n.TrafficLightPhase = 30
		default:
			n.TrafficLight = TrafficLightRed
			maxLanes := int32(2)
			for _, sid := range n.Connected {
				seg := rm.SegmentByID(sid)
				if seg == nil {
					continue
				}
				l := seg.LaneCount
				if l > maxLanes {
					maxLanes = l
				}
			}
			n.TrafficLightPhase = maxLanes * 20
		}
	}

	rm.floodTimer++
	if rm.floodTimer%10 != 0 {
		return
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
			waterH := terrain.ActiveSeaLevel() * terrain.MaxHeight
			isFlooded = ah < waterH || bh < waterH
		}
		if isFlooded {
			if !seg.Damaged {
				seg.Damaged = true
				seg.RepairTimer = 600
			}
		}
		if seg.Damaged && !isFlooded && seg.RepairTimer > 0 {
			seg.RepairTimer--
			if seg.RepairTimer <= 0 {
				seg.Damaged = false
			}
		}
	}
}

func (rm *RoadManager) DamageNearby(x, z, radius float32) {
	for i := range rm.Segments {
		seg := &rm.Segments[i]
		na := &rm.Nodes[seg.NodeA]
		nb := &rm.Nodes[seg.NodeB]
		dx := (na.X+nb.X)*0.5 - x
		dz := (na.Z+nb.Z)*0.5 - z
		if dx*dx+dz*dz < radius*radius {
			if !seg.Damaged {
				seg.Damaged = true
				seg.RepairTimer = 900
			}
		}
	}
}

func (rm *RoadManager) RepairSegment(idx int) bool {
	if idx < 0 || idx >= len(rm.Segments) {
		return false
	}
	seg := &rm.Segments[idx]
	if !seg.Damaged {
		return false
	}
	seg.Damaged = false
	seg.RepairTimer = 0
	return true
}

var waterForRoads *terrain.WaterSystem

func SetWaterForRoads(ws *terrain.WaterSystem) {
	waterForRoads = ws
}

func (rm *RoadManager) Draw(h *terrain.Heightmap) {
	if rm.roadTex.ID == 0 {
		rm.drawFallback(h)
	} else {
		if len(rm.Models) < len(rm.Segments) {
			rm.SyncMeshes(h)
		}
		for i := range rm.Models {
			if i < len(rm.Segments) && rm.Models[i].MeshCount == 0 {
				rm.queueMeshRebuild(i)
			}
		}
		if rm.PendingMeshRebuilds() > 0 {
			rm.Rebuild(h)
		}
		for _, model := range rm.Models {
			if model.MeshCount > 0 {
				rl.DrawModel(model, rl.NewVector3(0, 0, 0), 1, rl.White)
			}
		}
		rm.drawMarkings(h)
	}
	rm.drawJunctionMarkings(h)
	rm.drawOutsideConnections(h)
	rm.drawCurbs(h)
	rm.drawSidewalks(h)
	rm.drawStreetLights(h)
	rm.drawRoadsideTrees(h)
	rm.drawBridgePillars(h)
	rm.drawTunnelPortals(h)
	rm.drawDamageOverlay(h)
}

func (rm *RoadManager) drawDamageOverlay(h *terrain.Heightmap) {
	for _, seg := range rm.Segments {
		if !seg.Damaged {
			continue
		}
		xs, zs, _ := rm.SampleSegment(seg, int(seg.Length/4))
		if len(xs) < 2 {
			continue
		}
		for si := 0; si < len(xs)-1; si++ {
			x0, z0 := xs[si], zs[si]
			x1, z1 := xs[si+1], zs[si+1]
			var h0, h1 float32
			if seg.Elevation > 0 {
				h0 = float32(seg.Elevation) * 5
				h1 = float32(seg.Elevation) * 5
			} else {
				h0 = h.WorldHeight(x0, z0) + 0.2
				h1 = h.WorldHeight(x1, z1) + 0.2
			}
			rl.DrawCube(rl.NewVector3((x0+x1)*0.5, (h0+h1)*0.5, (z0+z1)*0.5), 0.8, 0.3, 0.6, rl.NewColor(120, 60, 30, 180))
			if si%3 == 0 {
				rl.DrawCube(rl.NewVector3((x0+x1)*0.5-0.5, (h0+h1)*0.5, (z0+z1)*0.5), 0.5, 0.2, 0.4, rl.NewColor(80, 40, 20, 200))
			}
		}
	}
}

func (rm *RoadManager) drawFallback(h *terrain.Heightmap) {
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
		if seg.Damaged {
			col = rl.NewColor(100, 50, 30, 255)
		}

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

func (rm *RoadManager) drawMarkings(h *terrain.Heightmap) {
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

func (rm *RoadManager) drawJunctionMarkings(h *terrain.Heightmap) {
	for i := range rm.Nodes {
		n := &rm.Nodes[i]
		if len(n.Connected) < 2 {
			continue
		}
		hy := h.WorldHeight(n.X, n.Z) + 0.2
		for _, sid := range n.Connected {
			seg := rm.SegmentByID(sid)
			if seg == nil {
				continue
			}
			if seg.Elevation > 0 {
				hy = float32(seg.Elevation) * 5
				break
			}
		}

		juncCol := rl.NewColor(200, 100, 100, 200)
		if n.JunctionType == 2 {
			juncCol = rl.NewColor(100, 100, 200, 200)
		}
		rl.DrawCube(rl.NewVector3(n.X, hy, n.Z), 2, 0.1, 2, juncCol)

		routes := rm.JunctionLaneRoutes(uint32(i))
		routesBySeg := make(map[int][]LaneConnection)
		for _, rc := range routes {
			routesBySeg[int(rc.FromSegIdx)] = append(routesBySeg[int(rc.FromSegIdx)], rc)
		}

		for _, sid := range n.Connected {
			seg := rm.SegmentByID(sid)
			if seg == nil {
				continue
			}
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

			segRoutes := routesBySeg[int(sid)]
			for _, rc := range segRoutes {
				li := int(rc.FromLane)
				laneOff := (float32(li) - float32(lanes-1)*0.5) * laneW
				arrowX := n.X + dx/l*2.0 + perX*laneOff
				arrowZ := n.Z + dz/l*2.0 + perZ*laneOff
				var arrowCol rl.Color
				switch rc.Turn {
				case LaneTurnStraight:
					arrowCol = rl.NewColor(100, 255, 100, 200)
				case LaneTurnLeft:
					arrowCol = rl.NewColor(255, 200, 100, 200)
				case LaneTurnRight:
					arrowCol = rl.NewColor(100, 200, 255, 200)
				case LaneTurnUTurn:
					arrowCol = rl.NewColor(255, 100, 100, 200)
				}
				rl.DrawCube(rl.NewVector3(arrowX, hy+0.05, arrowZ), 0.5, 0.05, 0.3, arrowCol)
			}
		}

		if n.JunctionType == 2 {
			for _, sid := range n.Connected {
				seg := rm.SegmentByID(sid)
				if seg == nil {
					continue
				}
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
				total := float32(seg.LaneCount) * laneW
				half := total * 0.5
				yx := n.X + dx/l*4.0 + perX*half
				yz := n.Z + dz/l*4.0 + perZ*half

				rl.DrawTriangle3D(
					rl.NewVector3(yx, hy+0.1, yz),
					rl.NewVector3(yx-dx/l*1.5+perX*half*0.5, hy+0.1, yz-dz/l*1.5+perZ*half*0.5),
					rl.NewVector3(yx-dx/l*1.5-perX*half*0.5, hy+0.1, yz-dz/l*1.5-perZ*half*0.5),
					rl.NewColor(255, 200, 0, 200),
				)
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

			pedCol := rl.NewColor(255, 0, 0, 200)
			if n.TrafficLight == TrafficLightGreen {
				pedCol = rl.NewColor(255, 255, 255, 200)
			}
			for _, sid := range n.Connected {
				seg := rm.SegmentByID(sid)
				if seg == nil {
					continue
				}
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
				total := float32(seg.LaneCount) * laneW
				half := total * 0.5
				px := n.X + dx/l*3.5 + perX*(half+1.0)
				pz := n.Z + dz/l*3.5 + perZ*(half+1.0)
				rl.DrawCube(rl.NewVector3(px, hy+0.8, pz), 0.3, 0.8, 0.1, pedCol)
			}
		}
	}
}

func (rm *RoadManager) drawOutsideConnections(h *terrain.Heightmap) {
	for _, n := range rm.Nodes {
		if n.Flags&RoadFlagOutsideConn != 0 {
			hy := h.WorldHeight(n.X, n.Z)
			rl.DrawCube(rl.NewVector3(n.X, hy+1, n.Z), 8, 2, 8, rl.NewColor(150, 100, 50, 200))
		}
	}
}

func (rm *RoadManager) drawCurbs(h *terrain.Heightmap) {
	curbW := float32(0.3)
	curbH := float32(0.15)
	curbCol := rl.NewColor(160, 160, 160, 255)

	for _, seg := range rm.Segments {
		if seg.Damaged {
			continue
		}
		xs, zs, _ := rm.SampleSegment(seg, int(seg.Length/2))
		if len(xs) < 2 {
			continue
		}
		total := float32(seg.LaneCount) * 3.0
		half := total * 0.5

		for si := 0; si < len(xs)-1; si++ {
			dx := xs[si+1] - xs[si]
			dz := zs[si+1] - zs[si]
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
				h0 = h.WorldHeight(xs[si], zs[si])
				h1 = h.WorldHeight(xs[si+1], zs[si+1])
			}

			// left curb
			rm.drawCurbSegment(xs[si], zs[si], xs[si+1], zs[si+1], perX, perZ, h0, h1, -half, -half+curbW, curbH, curbCol)
			// right curb
			rm.drawCurbSegment(xs[si], zs[si], xs[si+1], zs[si+1], perX, perZ, h0, h1, half-curbW, half, curbH, curbCol)
		}
	}
}

func (rm *RoadManager) drawCurbSegment(x0, z0, x1, z1, perX, perZ, h0, h1, inner, outer, height float32, col rl.Color) {
	topIn0 := rl.NewVector3(x0+perX*inner, h0+height, z0+perZ*inner)
	topOut0 := rl.NewVector3(x0+perX*outer, h0+height, z0+perZ*outer)
	topIn1 := rl.NewVector3(x1+perX*inner, h1+height, z1+perZ*inner)
	topOut1 := rl.NewVector3(x1+perX*outer, h1+height, z1+perZ*outer)
	botOut0 := rl.NewVector3(x0+perX*outer, h0, z0+perZ*outer)
	botOut1 := rl.NewVector3(x1+perX*outer, h1, z1+perZ*outer)

	drawQuad(topIn0, topOut0, topIn1, topOut1, col)
	drawQuad(botOut0, botOut1, topOut0, topOut1, col)
}

func (rm *RoadManager) drawSidewalks(h *terrain.Heightmap) {
	swW := float32(2.0)
	swH := float32(0.1)
	swCol := rl.NewColor(180, 180, 180, 255)

	for _, seg := range rm.Segments {
		if seg.Damaged {
			continue
		}
		if !roadHasSidewalk(seg.RoadType) {
			continue
		}
		xs, zs, _ := rm.SampleSegment(seg, int(seg.Length/2))
		if len(xs) < 2 {
			continue
		}
		total := float32(seg.LaneCount) * 3.0
		half := total * 0.5

		for si := 0; si < len(xs)-1; si++ {
			dx := xs[si+1] - xs[si]
			dz := zs[si+1] - zs[si]
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
				h0 = h.WorldHeight(xs[si], zs[si])
				h1 = h.WorldHeight(xs[si+1], zs[si+1])
			}

			rm.drawSwath(xs[si], zs[si], xs[si+1], zs[si+1], perX, perZ, h0, h1, -half-0.3-swW, -half-0.3, swH, swCol)
			rm.drawSwath(xs[si], zs[si], xs[si+1], zs[si+1], perX, perZ, h0, h1, half+0.3, half+0.3+swW, swH, swCol)
		}
	}
}

func (rm *RoadManager) drawSwath(x0, z0, x1, z1, perX, perZ, h0, h1, inner, outer, height float32, col rl.Color) {
	a := rl.NewVector3(x0+perX*inner, h0+height, z0+perZ*inner)
	b := rl.NewVector3(x0+perX*outer, h0+height, z0+perZ*outer)
	c := rl.NewVector3(x1+perX*inner, h1+height, z1+perZ*inner)
	d := rl.NewVector3(x1+perX*outer, h1+height, z1+perZ*outer)
	drawQuad(a, b, c, d, col)
}

func (rm *RoadManager) drawStreetLights(h *terrain.Heightmap) {
	interval := float32(20.0)
	poleCol := rl.NewColor(60, 60, 60, 255)
	lightCol := rl.NewColor(255, 230, 180, 255)
	if rm.nightMode {
		lightCol = rl.NewColor(255, 255, 200, 255)
	}

	for _, seg := range rm.Segments {
		if seg.Damaged {
			continue
		}
		if !roadHasLighting(seg.RoadType) {
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

		placed := float32(interval)
		for placed < totalLen {
			t := placed / totalLen
			idx := int(t * float32(len(xs)-1))
			if idx >= len(xs)-1 {
				break
			}
			frac := t*float32(len(xs)-1) - float32(idx)

			var perX, perZ float32
			if idx < len(xs)-1 {
				dx := xs[idx+1] - xs[idx]
				dz := zs[idx+1] - zs[idx]
				l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
				if l > 0.01 {
					perX = -dz / l
					perZ = dx / l
				}
			}

			lx := xs[idx] + (xs[idx+1]-xs[idx])*frac
			lz := zs[idx] + (zs[idx+1]-zs[idx])*frac

			var hy float32
			if seg.Elevation > 0 {
				hy = float32(seg.Elevation) * 5
			} else {
				hy = h.WorldHeight(lx, lz)
			}

			total := float32(seg.LaneCount) * 3.0
			half := total * 0.5

			for side := float32(-1); side <= 1; side += 2 {
				sx := lx + perX*half*side
				sz := lz + perZ*half*side

				rm.drawLightPole(sx, sz, hy, poleCol, lightCol)
			}

			placed += interval
		}
	}
}

func (rm *RoadManager) drawLightPole(x, z, groundH float32, poleCol, lightCol rl.Color) {
	poleH := float32(3.0)
	rl.DrawCube(rl.NewVector3(x, groundH+poleH*0.5, z), 0.1, poleH, 0.1, poleCol)
	rl.DrawSphere(rl.NewVector3(x, groundH+poleH+0.3, z), 0.25, lightCol)
}

func (rm *RoadManager) drawRoadsideTrees(h *terrain.Heightmap) {
	interval := float32(10.0)

	for _, seg := range rm.Segments {
		if seg.Damaged {
			continue
		}
		if seg.RoadType != RoadTreeLined {
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

		placed := float32(interval)
		side := float32(-1)
		for placed < totalLen {
			t := placed / totalLen
			idx := int(t * float32(len(xs)-1))
			if idx >= len(xs)-1 {
				break
			}
			frac := t*float32(len(xs)-1) - float32(idx)

			var perX, perZ float32
			if idx < len(xs)-1 {
				dx := xs[idx+1] - xs[idx]
				dz := zs[idx+1] - zs[idx]
				l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
				if l > 0.01 {
					perX = -dz / l
					perZ = dx / l
				}
			}

			tx := xs[idx] + (xs[idx+1]-xs[idx])*frac
			tz := zs[idx] + (zs[idx+1]-zs[idx])*frac

			var hy float32
			if seg.Elevation > 0 {
				hy = float32(seg.Elevation) * 5
			} else {
				hy = h.WorldHeight(tx, tz)
			}

			total := float32(seg.LaneCount) * 3.0
			half := total*0.5 + 2.5 + 1.5

			tx += perX * half * side
			tz += perZ * half * side

			trunkCol := rl.NewColor(80, 50, 20, 255)
			leafCol := rl.NewColor(50, 180, 50, 255)
			rl.DrawCube(rl.NewVector3(tx, hy+0.8, tz), 0.2, 1.6, 0.2, trunkCol)
			rl.DrawSphere(rl.NewVector3(tx, hy+2.0, tz), 1.0, leafCol)

			side = -side
			placed += interval
		}
	}
}

func (rm *RoadManager) drawBridgePillars(h *terrain.Heightmap) {
	if !rm.anyElevated() {
		return
	}
	interval := float32(15.0)
	pillarCol := rl.NewColor(140, 140, 140, 255)

	for _, seg := range rm.Segments {
		if seg.Damaged || seg.Elevation <= 0 {
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

		elevH := float32(seg.Elevation) * 5.0
		placed := float32(0)
		for placed < totalLen {
			t := placed / totalLen
			idx := int(t * float32(len(xs)-1))
			if idx >= len(xs)-1 {
				break
			}
			frac := t*float32(len(xs)-1) - float32(idx)
			px := xs[idx] + (xs[idx+1]-xs[idx])*frac
			pz := zs[idx] + (zs[idx+1]-zs[idx])*frac
			gh := h.WorldHeight(px, pz)
			pillarH := elevH - gh
			if pillarH > 0.5 {
				rl.DrawCube(rl.NewVector3(px, gh+pillarH*0.5, pz), 0.4, pillarH, 0.4, pillarCol)
				capCol := rl.NewColor(160, 160, 160, 255)
				rl.DrawCube(rl.NewVector3(px, elevH-0.2, pz), 0.8, 0.2, 0.8, capCol)
			}
			placed += interval
		}
	}
}

func (rm *RoadManager) drawTunnelPortals(h *terrain.Heightmap) {
	portalCol := rl.NewColor(100, 100, 100, 255)
	archCol := rl.NewColor(120, 120, 120, 255)

	for i := range rm.Nodes {
		n := &rm.Nodes[i]
		hasTunnel := false
		hasSurface := false
		for _, sid := range n.Connected {
			seg := rm.SegmentByID(sid)
			if seg == nil {
				continue
			}
			if seg.Elevation < 0 {
				hasTunnel = true
			} else {
				hasSurface = true
			}
		}
		if !hasTunnel || !hasSurface {
			continue
		}

		tunnelSegs := make([]struct{ sid uint32; dx, dz, l float32 }, 0)
		for _, sid := range n.Connected {
			seg := rm.SegmentByID(sid)
			if seg == nil {
				continue
			}
			if seg.Elevation >= 0 {
				continue
			}
			other := seg.NodeA
			if other == uint32(i) {
				other = seg.NodeB
			}
			on := &rm.Nodes[other]
			dx := on.X - n.X
			dz := on.Z - n.Z
			l := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if l > 0.01 {
				tunnelSegs = append(tunnelSegs, struct{ sid uint32; dx, dz, l float32 }{sid, dx / l, dz / l, l})
			}
		}

		tunnelH := float32(0)
		for _, ts := range tunnelSegs {
			seg := rm.SegmentByID(ts.sid)
			if seg == nil {
				continue
			}
			tunnelH = float32(seg.Elevation) * 5.0
			break
		}
		gh := h.WorldHeight(n.X, n.Z)

		for _, ts := range tunnelSegs {
			px := n.X + ts.dx*2.0
			pz := n.Z + ts.dz*2.0
			perX := -ts.dz
			perZ := ts.dx
			rl.DrawCube(rl.NewVector3(px, gh+tunnelH*0.5, pz), 0.8, -tunnelH, 0.8, portalCol)
			rl.DrawCube(rl.NewVector3(px+perX*2.0, gh+tunnelH*0.5, pz+perZ*2.0), 0.8, -tunnelH, 0.8, portalCol)
			rl.DrawCube(rl.NewVector3(px, gh+tunnelH, pz), 5.0, 0.3, 1.0, archCol)
		}
	}
}

func (rm *RoadManager) anyElevated() bool {
	for _, seg := range rm.Segments {
		if seg.Elevation > 0 {
			return true
		}
	}
	return false
}

func (rm *RoadManager) TotalMaintenance() float32 {
	var total float32
	for _, seg := range rm.Segments {
		total += seg.MaintenanceCost
	}
	return total
}

func (rm *RoadManager) AddShortSegment(x1, z1, x2, z2 float32, rt RoadType) {
	na := rm.AddNode(x1, 0, z1)
	nb := rm.AddNode(x2, 0, z2)
	rm.AddSegment(na, nb, rt)
}

func (rm *RoadManager) buildNearestGrid() {
	half := float32(terrain.WorldSize / 2)
	size := float32(terrain.WorldSize) / float32(rm.nearestGridSize)
	gridLen := rm.nearestGridSize * rm.nearestGridSize
	rm.nearestGrid = make([][]int, gridLen)
	for i, seg := range rm.Segments {
		na := &rm.Nodes[seg.NodeA]
		nb := &rm.Nodes[seg.NodeB]
		gx1 := int((na.X + half) / size)
		gz1 := int((na.Z + half) / size)
		gx2 := int((nb.X + half) / size)
		gz2 := int((nb.Z + half) / size)
		if gx1 < 0 {
			gx1 = 0
		}
		if gx1 >= rm.nearestGridSize {
			gx1 = rm.nearestGridSize - 1
		}
		if gz1 < 0 {
			gz1 = 0
		}
		if gz1 >= rm.nearestGridSize {
			gz1 = rm.nearestGridSize - 1
		}
		if gx2 < 0 {
			gx2 = 0
		}
		if gx2 >= rm.nearestGridSize {
			gx2 = rm.nearestGridSize - 1
		}
		if gz2 < 0 {
			gz2 = 0
		}
		if gz2 >= rm.nearestGridSize {
			gz2 = rm.nearestGridSize - 1
		}
		minX, maxX := gx1, gx2
		if gx2 < gx1 {
			minX, maxX = gx2, gx1
		}
		minZ, maxZ := gz1, gz2
		if gz2 < gz1 {
			minZ, maxZ = gz2, gz1
		}
		for gz := minZ; gz <= maxZ; gz++ {
			for gx := minX; gx <= maxX; gx++ {
				idx := gz*rm.nearestGridSize + gx
				rm.nearestGrid[idx] = append(rm.nearestGrid[idx], i)
			}
		}
	}
}

func (rm *RoadManager) NearestSegment(x, z float32) int {
	half := float32(terrain.WorldSize / 2)
	size := float32(terrain.WorldSize) / float32(rm.nearestGridSize)
	bestIdx := -1
	bestDist := float32(math.MaxFloat32)

	if rm.nearestGrid == nil {
		rm.buildNearestGrid()
	}

	gx := int((x + half) / size)
	gz := int((z + half) / size)
	if gx >= 0 && gx < rm.nearestGridSize && gz >= 0 && gz < rm.nearestGridSize {
		cellIdx := gz*rm.nearestGridSize + gx
		for _, segIdx := range rm.nearestGrid[cellIdx] {
			seg := rm.Segments[segIdx]
			xs, zs, _ := rm.SampleSegment(seg, 8)
			for j := 0; j < len(xs); j++ {
				dx := x - xs[j]
				dz := z - zs[j]
				d := dx*dx + dz*dz
				if d < bestDist {
					bestDist = d
					bestIdx = segIdx
				}
			}
		}
		if bestIdx >= 0 {
			goto done
		}
	}

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
done:
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

	rm.updateJunctionType(seg.NodeA)
	rm.updateJunctionType(seg.NodeB)
	rm.invalidateCaches()

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
	s.MaintenanceCost = rm.CalcSegmentMaintenance(newType, s.Length, s.Elevation)
	s.ConstructionCost = RoadConstructionCost(newType)
	s.Lanes = generateLanes(newType, s.Direction, s.SpeedLimit, s.LaneCount)
	rm.queueMeshRebuild(idx)
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

type LanePathStep struct {
	SegIdx  int
	LaneIdx int32
}

func segmentAllowsVehicle(seg RoadSegment, vt VehicleType) bool {
	allowed := roadAllowedVehicles(seg.RoadType)
	if allowed == "all" {
		return true
	}
	switch vt {
	case VehicleCar, VehicleTruck, VehicleEmergency:
		return allowed == "all"
	case VehicleBus:
		return allowed == "bus" || allowed == "all"
	case VehicleTram:
		return allowed == "tram" || allowed == "all"
	case VehicleBike:
		return allowed == "bike" || allowed == "all"
	case VehiclePedestrian:
		return allowed == "pedestrian" || allowed == "all"
	}
	return false
}

func (rm *RoadManager) FindPath(startNode, endNode uint32, vehicleType int) []uint32 {
	return rm.FindPathWithCongestion(startNode, endNode, VehicleType(vehicleType), nil)
}

func (rm *RoadManager) FindPathWithCongestion(startNode, endNode uint32, vt VehicleType, congestion map[int]float32) []uint32 {
	if startNode == endNode {
		return nil
	}
	type nodeDist struct {
		prev    int32
		dist    float32
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
			seg := rm.SegmentByID(sid)
			if seg == nil {
				continue
			}
			if seg.Damaged {
				continue
			}
			if !segmentAllowsVehicle(*seg, vt) {
				continue
			}

			other := seg.NodeA
			if other == uint32(best) {
				other = seg.NodeB
			}
			cost := seg.Length / seg.SpeedLimit

			switch roadHierarchy(seg.RoadType) {
			case HierarchyHighway:
				cost *= 0.6
			case HierarchyArterial:
				cost *= 0.8
			case HierarchyLocal:
				cost *= 1.5
			}
			if congestion != nil {
				if c, ok := congestion[int(sid)]; ok {
					cost *= c
				}
			}

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

func (rm *RoadManager) FindLanePath(startNode, endNode uint32, startLane, endLane int32) []LanePathStep {
	nodePath := rm.FindPath(startNode, endNode, 0)
	if len(nodePath) < 2 {
		return nil
	}

	type segLane struct {
		seg  int
		lane int32
	}
	path := make([]segLane, 0, len(nodePath)-1)

	curLane := startLane
	for ni := 0; ni < len(nodePath)-1; ni++ {
		fromNode := nodePath[ni]
		toNode := nodePath[ni+1]

		var segIdx int
		var found bool
		for _, sid := range rm.Nodes[fromNode].Connected {
			s := rm.SegmentByID(sid)
			if s == nil {
				continue
			}
			if (s.NodeA == fromNode && s.NodeB == toNode) || (s.NodeA == toNode && s.NodeB == fromNode) {
				segIdx = int(sid)
				found = true
				break
			}
		}
		if !found {
			return nil
		}

		path = append(path, segLane{seg: segIdx, lane: curLane})

		routes := rm.JunctionLaneRoutes(toNode)
		nextLane := curLane
		for _, rc := range routes {
			if int(rc.FromSegIdx) == segIdx && rc.FromLane == curLane {
				nextLane = rc.ToLane
				break
			}
		}
		curLane = nextLane
	}

	result := make([]LanePathStep, len(path))
	for i, sl := range path {
		result[i] = LanePathStep{SegIdx: sl.seg, LaneIdx: sl.lane}
	}
	return result
}

func (rm *RoadManager) Unload() {
	for _, model := range rm.Models {
		if model.MeshCount > 0 {
			rl.UnloadModel(model)
		}
	}
	rm.Models = nil
}

func clearModelMeshData(model *rl.Model) {
	meshes := model.GetMeshes()
	for i := range meshes {
		meshes[i].Vertices = nil
		meshes[i].Normals = nil
		meshes[i].Texcoords = nil
		meshes[i].Texcoords2 = nil
		meshes[i].Colors = nil
		meshes[i].Indices = nil
		meshes[i].Tangents = nil
		meshes[i].AnimVertices = nil
		meshes[i].AnimNormals = nil
		meshes[i].BoneIndices = nil
		meshes[i].BoneWeights = nil
	}
}
