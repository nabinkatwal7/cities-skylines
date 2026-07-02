package road

import (
	"math"

	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/gameassets"
	"github.com/katwate/js-skylines/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type VehicleType uint8

const (
	VehicleCar VehicleType = iota
	VehicleBus
	VehicleTruck
	VehicleEmergency
	VehicleBike
	VehiclePedestrian
	VehicleTram
)

func VehicleTypeName(vt VehicleType) string {
	switch vt {
	case VehicleCar:
		return "Car"
	case VehicleBus:
		return "Bus"
	case VehicleTruck:
		return "Truck"
	case VehicleEmergency:
		return "Emergency"
	case VehicleBike:
		return "Bicycle"
	case VehiclePedestrian:
		return "Pedestrian"
	case VehicleTram:
		return "Tram"
	default:
		return "Unknown"
	}
}

type Vehicle struct {
	core.Entity
	Type        VehicleType
	Speed       float32
	TargetSpeed float32
	Path        []uint32
	PathIdx     int
	SegProgress float32
	RoadSeg     int
	Waiting     int32
	Color       rl.Color
	Lane        int
	TurnSignal  int
	SignalTimer int32
	ParkTimer   int32
	ParkSpotIdx int
}

const VehiclePoolSize = 500

type VehicleManager struct {
	Pool     [VehiclePoolSize]Vehicle
	FreeList []int32
	Count    int32
	NextID   uint32
	Timer    int32
	assets   *gameassets.Catalog
}

func NewVehicleManager() *VehicleManager {
	vm := &VehicleManager{
		FreeList: make([]int32, VehiclePoolSize),
	}
	for i := 0; i < VehiclePoolSize; i++ {
		vm.Pool[i].Lifecycle = core.LifecycleUnallocated
		vm.FreeList[i] = int32(VehiclePoolSize - 1 - i)
	}
	return vm
}

func (vm *VehicleManager) Alloc() int32 {
	if len(vm.FreeList) == 0 {
		return -1
	}
	idx := vm.FreeList[len(vm.FreeList)-1]
	vm.FreeList = vm.FreeList[:len(vm.FreeList)-1]
	vm.Pool[idx].Lifecycle = core.LifecycleAllocated
	vm.Count++
	return idx
}

func (vm *VehicleManager) Free(slot int32) {
	if slot < 0 || int(slot) >= VehiclePoolSize {
		return
	}
	vm.Pool[slot] = Vehicle{}
	vm.Pool[slot].Lifecycle = core.LifecycleReturnedToPool
	vm.FreeList = append(vm.FreeList, slot)
	vm.Count--
}

func (vm *VehicleManager) ForEach(fn func(*Vehicle, int32)) {
	for i := 0; i < VehiclePoolSize; i++ {
		if vm.Pool[i].Lifecycle == core.LifecycleActive {
			fn(&vm.Pool[i], int32(i))
		}
	}
}

func (vm *VehicleManager) buildCongestionMap(rm *RoadManager) map[int]float32 {
	cong := make(map[int]float32)
	for i := 0; i < VehiclePoolSize; i++ {
		v := &vm.Pool[i]
		if v.Lifecycle == core.LifecycleActive && !v.HasFlag(core.FlagParked) {
			cong[v.RoadSeg]++
		}
	}
	for k, v := range cong {
		cap := float32(20)
		if v >= cap {
			cong[k] = 3.0
		} else {
			cong[k] = 1.0 + v/cap*2.0
		}
	}
	return cong
}

func (vm *VehicleManager) SpawnCar(rm *RoadManager) {
	if len(rm.Segments) == 0 {
		return
	}
	segIdx := int(vm.NextID) % len(rm.Segments)
	seg := rm.Segments[segIdx]
	na := &rm.Nodes[seg.NodeA]

	lanes := int(seg.LaneCount)
	lane := int(vm.NextID)
	if lanes > 0 {
		lane %= lanes
	} else {
		lane = 0
	}

	slot := vm.Alloc()
	if slot < 0 {
		return
	}
	v := &vm.Pool[slot]
	v.Entity = core.NewEntity(vm.NextID, na.X, 0, na.Z, core.OwnerVehicle)
	v.Lifecycle = core.LifecycleActive
	v.Type = VehicleCar
	v.TargetSpeed = seg.SpeedLimit * 0.8
	v.RoadSeg = int(seg.ID)
	v.Lane = lane
	v.Color = rl.NewColor(uint8(100+vm.NextID%100), uint8(50+vm.NextID%80), uint8(80+vm.NextID%100), 255)
	v.ParkTimer = 0
	v.ParkSpotIdx = -1
	vm.NextID++

	destIdx := int(vm.NextID) % len(rm.Nodes)
	if destIdx >= len(rm.Nodes) {
		destIdx = 0
	}
	cong := vm.buildCongestionMap(rm)
	v.Path = rm.FindPathWithCongestion(seg.NodeA, uint32(destIdx), v.Type, cong)
	v.PathIdx = 0
	if len(v.Path) > 0 {
		v.Path = v.Path[1:]
	} else {
		v.Path = nil
	}

	if vm.Count > VehiclePoolSize-10 {
		vm.evictOldest()
	}
}

func (vm *VehicleManager) SpawnOutsideCar(rm *RoadManager) {
	var outsideNodeIdx uint32 = math.MaxUint32
	for idx := range rm.Nodes {
		if rm.Nodes[idx].Flags&RoadFlagOutsideConn != 0 && len(rm.Nodes[idx].Connected) > 0 {
			outsideNodeIdx = uint32(idx)
			break
		}
	}
	if outsideNodeIdx == math.MaxUint32 {
		return
	}
	n := &rm.Nodes[outsideNodeIdx]
	firstSeg := rm.SegmentByID(n.Connected[0])
	if firstSeg == nil {
		return
	}
	otherNode := firstSeg.NodeA
	if otherNode == outsideNodeIdx {
		otherNode = firstSeg.NodeB
	}
	if int(otherNode) >= len(rm.Nodes) {
		return
	}

	destIdx := int(vm.NextID) % len(rm.Nodes)
	if destIdx >= len(rm.Nodes) {
		destIdx = 0
	}
	if uint32(destIdx) == outsideNodeIdx {
		destIdx = (destIdx + 1) % len(rm.Nodes)
	}
	cong := vm.buildCongestionMap(rm)
	path := rm.FindPathWithCongestion(outsideNodeIdx, uint32(destIdx), VehicleCar, cong)
	if len(path) < 2 {
		return
	}

	slot := vm.Alloc()
	if slot < 0 {
		return
	}
	v := &vm.Pool[slot]
	v.Entity = core.NewEntity(vm.NextID, n.X, 0, n.Z, core.OwnerVehicle)
	v.Lifecycle = core.LifecycleActive
	v.Type = VehicleCar
	v.TargetSpeed = firstSeg.SpeedLimit * 0.8
	v.RoadSeg = int(n.Connected[0])
	v.Lane = 0
	v.Color = rl.NewColor(uint8(150+vm.NextID%80), uint8(100+vm.NextID%60), uint8(50+vm.NextID%100), 255)
	v.ParkTimer = 0
	v.ParkSpotIdx = -1
	v.Path = path[1:]
	v.PathIdx = 0
	vm.NextID++
}

func (vm *VehicleManager) evictOldest() {
	oldestFrame := int32(math.MaxInt32)
	oldestSlot := int32(-1)
	for i := 0; i < VehiclePoolSize; i++ {
		if vm.Pool[i].Lifecycle == core.LifecycleActive && vm.Pool[i].CreatedAt < oldestFrame {
			oldestFrame = vm.Pool[i].CreatedAt
			oldestSlot = int32(i)
		}
	}
	if oldestSlot >= 0 {
		vm.Free(oldestSlot)
	}
}

func (vm *VehicleManager) chooseLane(v *Vehicle, rm *RoadManager, seg RoadSegment) int {
	lanes := int(seg.LaneCount)
	if lanes <= 1 {
		return 0
	}

	if v.Path == nil || v.PathIdx >= len(v.Path) {
		rightmost := lanes - 1
		if rightmost >= 0 {
			return rightmost
		}
		return 0
	}

	nextNode := v.Path[v.PathIdx]

	if rm.Nodes[nextNode].JunctionType == 2 {
		exitCount := 0
		for ni := v.PathIdx; ni < len(v.Path); ni++ {
			for _, sid := range rm.Nodes[v.Path[ni]].Connected {
				s := rm.SegmentByID(sid)
				if s == nil {
					continue
				}
				other := s.NodeA
				if other == v.Path[ni] {
					other = s.NodeB
				}
				if other != v.Path[ni] && ni+1 < len(v.Path) && other == v.Path[ni+1] {
					exitCount++
					break
				}
			}
		}
		switch {
		case exitCount <= 1:
			if lanes > 0 {
				return lanes - 1
			}
		case exitCount == 2:
			if lanes >= 3 {
				return lanes / 2
			}
			if lanes > 0 {
				return lanes - 1
			}
		default:
			return 0
		}
	}

	upcomingTurn := LaneTurnStraight

	var turnSegIdx int32 = -1
	for _, sid := range rm.Nodes[nextNode].Connected {
		if int(sid) == v.RoadSeg {
			continue
		}
		turnSegIdx = int32(sid)
		break
	}
	if turnSegIdx >= 0 {
		routes := rm.JunctionLaneRoutes(nextNode)
		for _, rc := range routes {
			if int(rc.FromSegIdx) == v.RoadSeg && int(rc.ToSegIdx) == int(turnSegIdx) {
				upcomingTurn = rc.Turn
				break
			}
		}
	}

	var preferredLane int
	switch upcomingTurn {
	case LaneTurnLeft:
		preferredLane = 0
	case LaneTurnRight:
		preferredLane = lanes - 1
	default:
		preferredLane = lanes / 2
	}

	if preferredLane >= lanes {
		preferredLane = lanes - 1
	}
	if preferredLane < 0 {
		preferredLane = 0
	}

	if v.Lane == preferredLane {
		return preferredLane
	}

	distToNode := float32(0)
	if seg.NodeA == nextNode || seg.NodeB == nextNode {
		na := &rm.Nodes[seg.NodeA]
		nb := &rm.Nodes[seg.NodeB]
		var nodeX, nodeZ float32
		if seg.NodeA == nextNode {
			nodeX = na.X
			nodeZ = na.Z
		} else {
			nodeX = nb.X
			nodeZ = nb.Z
		}
		dx := nodeX - v.Position.X
		dz := nodeZ - v.Position.Z
		distToNode = float32(math.Sqrt(float64(dx*dx + dz*dz)))
	}

	if distToNode > 25 {
		return v.Lane
	}

	return preferredLane
}

func (vm *VehicleManager) applyTrafficRules(v *Vehicle, rm *RoadManager, seg RoadSegment) {
	lookAhead := func(nodeIdx uint32) (float32, TrafficRule) {
		n := &rm.Nodes[nodeIdx]
		dx := n.X - v.Position.X
		dz := n.Z - v.Position.Z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		rule := rm.TrafficRuleForNode(nodeIdx)
		return dist, rule
	}

	distA, ruleA := lookAhead(seg.NodeA)
	distB, ruleB := lookAhead(seg.NodeB)

	var distToNode float32
	var rule TrafficRule
	var approachedNodeIdx uint32
	if distA < distB {
		distToNode = distA
		rule = ruleA
		approachedNodeIdx = seg.NodeA
	} else {
		distToNode = distB
		rule = ruleB
		approachedNodeIdx = seg.NodeB
	}

	approachDist := float32(15)
	if seg.SpeedLimit > 50 {
		approachDist = 25
	}

	switch rule {
	case RuleTrafficLight:
		on := &rm.Nodes[seg.NodeA]
		if rm.Nodes[seg.NodeB].TrafficLight != TrafficLightNone {
			if distB < distA {
				on = &rm.Nodes[seg.NodeB]
			}
		}
		if distToNode < approachDist && on.TrafficLight >= TrafficLightYellow {
			v.Speed *= 0.9
			if v.Speed < 2 {
				v.Speed = 2
			}
			v.Waiting++
		} else {
			v.Waiting = 0
		}

	case RuleStop:
		if distToNode < approachDist {
			stopDist := float32(6)
			if distToNode < stopDist && v.Speed > 0.5 {
				v.Speed *= 0.85
				if v.Speed < 0.5 {
					v.Speed = 0.5
				}
				v.Waiting++
			} else if distToNode < approachDist*0.5 {
				v.Speed *= 0.9
				v.Waiting++
			} else {
				v.Waiting = 0
			}
		} else {
			v.Waiting = 0
		}

	case RuleYield:
		if distToNode < approachDist {
			yieldDist := float32(10)
			if distToNode < yieldDist {
				v.Speed *= 0.92
				if v.Speed < 1 {
					v.Speed = 1
				}
				v.Waiting++
			} else {
				v.Speed *= 0.97
				v.Waiting = 0
			}
		} else {
			v.Waiting = 0
		}

	case RuleRoundabout:
		if distToNode < approachDist {
			vehicleInside := false
			for i := 0; i < VehiclePoolSize; i++ {
				other := &vm.Pool[i]
				if other == v || other.Lifecycle != core.LifecycleActive {
					continue
				}
				if other.RoadSeg == v.RoadSeg {
					continue
				}
				os := rm.SegmentByID(uint32(other.RoadSeg))
				if os == nil {
					continue
				}
				if os.NodeA == approachedNodeIdx || os.NodeB == approachedNodeIdx {
					vehicleInside = true
					break
				}
			}
			yieldDist := float32(10)
			if distToNode < yieldDist && vehicleInside {
				v.Speed *= 0.85
				if v.Speed < 0.5 {
					v.Speed = 0.5
				}
				v.Waiting++
			} else if distToNode < yieldDist {
				v.Speed *= 0.95
				if v.Speed < 1 {
					v.Speed = 1
				}
			} else {
				v.Waiting = 0
			}
		} else {
			v.Waiting = 0
		}

	case RulePriorityRoad:
		hasPriority := true
		for _, sid := range rm.Nodes[seg.NodeA].Connected {
			otherSeg := rm.SegmentByID(sid)
			if otherSeg == nil || otherSeg.ID == seg.ID {
				continue
			}
			if roadHierarchy(otherSeg.RoadType) >= roadHierarchy(seg.RoadType) {
				continue
			}
			hasPriority = false
			break
		}
		if !hasPriority && distToNode < approachDist {
			v.Speed *= 0.95
			if v.Speed < 1 {
				v.Speed = 1
			}
		} else {
			v.Waiting = 0
		}

	default:
		v.Waiting = 0
	}
}

func (vm *VehicleManager) avoidCollisions(v *Vehicle, rm *RoadManager) {
	for i := 0; i < VehiclePoolSize; i++ {
		other := &vm.Pool[i]
		if other == v || other.Lifecycle != core.LifecycleActive {
			continue
		}
		if other.RoadSeg != v.RoadSeg || other.Lane != v.Lane {
			continue
		}
		if other.SegProgress <= v.SegProgress {
			continue
		}
		gap := other.SegProgress - v.SegProgress
		if gap < 5.0 {
			v.Speed *= 0.85
			if v.Speed < 0.5 {
				v.Speed = 0.5
			}
			if gap < 2.0 {
				v.Speed *= 0.5
				if v.Speed < 0.1 {
					v.Speed = 0.1
				}
			}
		}
	}
}

func (vm *VehicleManager) Update(rm *RoadManager, h *terrain.Heightmap, pm *ParkingManager) {
	vm.Timer++
	if vm.Timer%30 == 0 && vm.Count < 50 {
		vm.SpawnCar(rm)
	}
	if vm.Timer%60 == 0 && vm.Count < 40 {
		vm.SpawnOutsideCar(rm)
	}

	for i := 0; i < VehiclePoolSize; i++ {
		v := &vm.Pool[i]
		if v.Lifecycle != core.LifecycleActive {
			continue
		}
		if v.Lifecycle == core.LifecycleSuspended {
			continue
		}

		if v.HasFlag(core.FlagParked) {
			v.ParkTimer--
			if v.ParkTimer <= 0 {
				vm.unpark(v, rm, pm)
			}
			continue
		}

		if v.RoadSeg < 0 {
			v.RemovalTimer = 0
			v.Lifecycle = core.LifecycleMarkedForRemoval
			continue
		}

		segPtr := rm.SegmentByID(uint32(v.RoadSeg))
		if segPtr == nil {
			v.RemovalTimer = 0
			v.Lifecycle = core.LifecycleMarkedForRemoval
			continue
		}
		seg := *segPtr
		if seg.Damaged {
			v.Speed *= 0.95
			if v.Speed < 1 {
				v.Speed = 1
			}
		}

		v.Lane = vm.chooseLane(v, rm, seg)

		vm.applyTrafficRules(v, rm, seg)

		v.Speed += (v.TargetSpeed - v.Speed) * 0.05
		if v.Speed < 0.5 {
			v.Speed = 0.5
		}

		vm.avoidCollisions(v, rm)

		xs, zs, ds := rm.SampleSegment(seg, int(seg.Length/2))
		if len(xs) < 2 {
			continue
		}
		totalLen := ds[len(ds)-1]
		if totalLen < 0.01 {
			continue
		}

		v.SegProgress += v.Speed * 0.02
		if v.SegProgress > totalLen {
			v.SegProgress = totalLen
		}
		if v.SegProgress < 0 {
			v.SegProgress = 0
		}

		lanes := int(seg.LaneCount)
		if v.Lane >= lanes {
			v.Lane = 0
		}
		laneW := float32(3.0)
		if v.Lane < len(seg.Lanes) {
			laneW = seg.Lanes[v.Lane].Width
		}
		laneOffset := (float32(v.Lane) - float32(lanes)*0.5) * laneW

		t := v.SegProgress / totalLen
		if t > 1 {
			t = 1
		}
		idx := int(t * float32(len(xs)-1))
		if idx >= len(xs)-1 {
			idx = len(xs) - 2
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

		v.Position.X = xs[idx] + (xs[idx+1]-xs[idx])*frac + perX*laneOffset
		v.Position.Z = zs[idx] + (zs[idx+1]-zs[idx])*frac + perZ*laneOffset

		if v.SegProgress >= totalLen {
			if v.Path != nil && v.PathIdx < len(v.Path) {
				nextNode := v.Path[v.PathIdx]
				v.PathIdx++

				found := false
				for _, sid := range rm.Nodes[seg.NodeA].Connected {
					if int(sid) == v.RoadSeg {
						continue
					}
					s := rm.SegmentByID(sid)
					if s == nil {
						continue
					}
					if (s.NodeA == seg.NodeA && s.NodeB == nextNode) || (s.NodeA == nextNode && s.NodeB == seg.NodeA) ||
						(s.NodeA == seg.NodeB && s.NodeB == nextNode) || (s.NodeA == nextNode && s.NodeB == seg.NodeB) {
						v.RoadSeg = int(sid)
						v.SegProgress = 0
						seg = *s
						found = true
						break
					}
				}
				if !found {
					vm.tryPark(v, rm, pm)
				}
			} else {
				vm.tryPark(v, rm, pm)
			}
		}
	}

	vm.processRemovals()
}

func (vm *VehicleManager) tryPark(v *Vehicle, rm *RoadManager, pm *ParkingManager) {
	if pm == nil {
		v.RemovalTimer = 0
		v.Lifecycle = core.LifecycleMarkedForRemoval
		return
	}
	spotIdx := pm.FindSpot(v.Position.X, v.Position.Z, 50)
	if spotIdx >= 0 {
		ok := pm.OccupySpot(spotIdx, v.ID)
		if ok {
			v.SetFlag(core.FlagParked)
			v.ParkSpotIdx = spotIdx
			v.ParkTimer = int32(120 + vm.NextID%180)
			sp := pm.Spots[spotIdx]
			v.Position.X = sp.X
			v.Position.Z = sp.Z
			return
		}
	}
	v.RemovalTimer = 0
	v.Lifecycle = core.LifecycleMarkedForRemoval
}

func (vm *VehicleManager) unpark(v *Vehicle, rm *RoadManager, pm *ParkingManager) {
	v.ClearFlag(core.FlagParked)
	if v.ParkSpotIdx >= 0 && pm != nil {
		pm.FreeSpot(v.ParkSpotIdx)
	}
	v.ParkSpotIdx = -1
	if len(rm.Segments) == 0 {
		v.RemovalTimer = 0
		v.Lifecycle = core.LifecycleMarkedForRemoval
		return
	}
	segIdx := int(vm.NextID) % len(rm.Segments)
	seg := rm.Segments[segIdx]
	v.RoadSeg = int(seg.ID)
	v.SegProgress = 0
	v.Speed = 0.5
	v.TargetSpeed = seg.SpeedLimit * 0.8

	destIdx := int(vm.NextID+1) % len(rm.Nodes)
	if destIdx >= len(rm.Nodes) {
		destIdx = 0
	}
	cong := vm.buildCongestionMap(rm)
	v.Path = rm.FindPathWithCongestion(seg.NodeA, uint32(destIdx), v.Type, cong)
	v.PathIdx = 0
	if len(v.Path) > 0 {
		v.Path = v.Path[1:]
	} else {
		v.Path = nil
	}
	na := &rm.Nodes[seg.NodeA]
	v.Position.X = na.X
	v.Position.Z = na.Z
}

func (vm *VehicleManager) processRemovals() {
	for i := 0; i < VehiclePoolSize; i++ {
		v := &vm.Pool[i]
		switch v.Lifecycle {
		case core.LifecycleMarkedForRemoval:
			v.RemovalTimer++
			if v.RemovalTimer > 10 {
				v.Lifecycle = core.LifecycleDestroyed
			}
		case core.LifecycleDestroyed:
			vm.Free(int32(i))
		}
	}
}

func (vm *VehicleManager) SetAssets(c *gameassets.Catalog) {
	vm.assets = c
}

func (vm *VehicleManager) Draw(h *terrain.Heightmap) {
	vm.ForEach(func(v *Vehicle, _ int32) {
		hy := h.WorldHeight(v.Position.X, v.Position.Z)
		if vm.assets != nil {
			if model, ok := vm.assets.VehicleModel(int(v.Type)); ok {
				scale := float32(0.35)
				if v.Type == VehicleTruck || v.Type == VehicleBus {
					scale = 0.42
				}
				if v.HasFlag(core.FlagParked) {
					scale *= 0.9
				}
				yOff := gameassets.GroundOffset(model, scale)
				pos := rl.NewVector3(v.Position.X, hy+yOff, v.Position.Z)
				rl.DrawModelEx(model, pos, rl.NewVector3(0, 1, 0), 0, rl.NewVector3(scale, scale, scale), rl.White)
				return
			}
		}
		if v.HasFlag(core.FlagParked) {
			hy += 0.2
			rl.DrawCube(rl.NewVector3(v.Position.X, hy, v.Position.Z), 1.5, 0.4, 1, v.Color)
			return
		}
		hy += 0.5
		rl.DrawCube(rl.NewVector3(v.Position.X, hy, v.Position.Z), 1.5, 0.5, 1, v.Color)
		rl.DrawCube(rl.NewVector3(v.Position.X, hy+0.4, v.Position.Z+0.6), 0.8, 0.3, 0.1, rl.NewColor(200, 50, 50, 255))
		rl.DrawCube(rl.NewVector3(v.Position.X, hy+0.4, v.Position.Z-0.6), 0.8, 0.3, 0.1, rl.NewColor(200, 50, 50, 255))
	})
}

func (vm *VehicleManager) OnRoadRemoved(segID int) {
	for i := 0; i < VehiclePoolSize; i++ {
		v := &vm.Pool[i]
		if v.Lifecycle == core.LifecycleActive && v.RoadSeg == segID {
			v.Lifecycle = core.LifecycleMarkedForRemoval
			v.RemovalTimer = 0
		}
	}
}

func (vm *VehicleManager) Unload() {
	for i := 0; i < VehiclePoolSize; i++ {
		if vm.Pool[i].Lifecycle == core.LifecycleActive {
			vm.Pool[i].Lifecycle = core.LifecycleReturnedToPool
		}
	}
	vm.Count = 0
}
