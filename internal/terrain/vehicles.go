package terrain

import (
	"math"

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

type Vehicle struct {
	Entity
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
}

const VehiclePoolSize = 500

type VehicleManager struct {
	Pool     [VehiclePoolSize]Vehicle
	FreeList []int32
	Count    int32
	NextID   uint32
	Timer    int32
}

func NewVehicleManager() *VehicleManager {
	vm := &VehicleManager{
		FreeList: make([]int32, VehiclePoolSize),
	}
	for i := 0; i < VehiclePoolSize; i++ {
		vm.Pool[i].Lifecycle = LifecycleUnallocated
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
	vm.Pool[idx].Lifecycle = LifecycleAllocated
	vm.Count++
	return idx
}

func (vm *VehicleManager) Free(slot int32) {
	if slot < 0 || int(slot) >= VehiclePoolSize {
		return
	}
	vm.Pool[slot] = Vehicle{}
	vm.Pool[slot].Lifecycle = LifecycleReturnedToPool
	vm.FreeList = append(vm.FreeList, slot)
	vm.Count--
}

func (vm *VehicleManager) ForEach(fn func(*Vehicle, int32)) {
	for i := 0; i < VehiclePoolSize; i++ {
		if vm.Pool[i].Lifecycle == LifecycleActive {
			fn(&vm.Pool[i], int32(i))
		}
	}
}

func (vm *VehicleManager) buildCongestionMap(rm *RoadManager) map[int]float32 {
	cong := make(map[int]float32)
	for i := 0; i < VehiclePoolSize; i++ {
		if vm.Pool[i].Lifecycle == LifecycleActive {
			cong[vm.Pool[i].RoadSeg]++
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
	v.Entity = NewEntity(vm.NextID, na.X, 0, na.Z, OwnerVehicle)
	v.Lifecycle = LifecycleActive
	v.Type = VehicleCar
	v.TargetSpeed = seg.SpeedLimit * 0.8
	v.RoadSeg = segIdx
	v.Lane = lane
	v.Color = rl.NewColor(uint8(100+vm.NextID%100), uint8(50+vm.NextID%80), uint8(80+vm.NextID%100), 255)
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

func (vm *VehicleManager) evictOldest() {
	oldestFrame := int32(math.MaxInt32)
	oldestSlot := int32(-1)
	for i := 0; i < VehiclePoolSize; i++ {
		if vm.Pool[i].Lifecycle == LifecycleActive && vm.Pool[i].CreatedAt < oldestFrame {
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
	lookAhead := func(n *RoadNode) (float32, TrafficRule) {
		dx := n.X - v.Position.X
		dz := n.Z - v.Position.Z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		rule := rm.TrafficRuleForNode(n.ID)
		return dist, rule
	}

	distA, ruleA := lookAhead(&rm.Nodes[seg.NodeA])
	distB, ruleB := lookAhead(&rm.Nodes[seg.NodeB])

	var distToNode float32
	var rule TrafficRule
	if distA < distB {
		distToNode = distA
		rule = ruleA
	} else {
		distToNode = distB
		rule = ruleB
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

	case RuleYield, RuleRoundabout:
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

	case RulePriorityRoad:
		hasPriority := true
		for _, sid := range rm.Nodes[seg.NodeA].Connected {
			otherSeg := rm.Segments[sid]
			if otherSeg.ID == seg.ID {
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

func (vm *VehicleManager) Update(rm *RoadManager, h *Heightmap) {
	vm.Timer++
	if vm.Timer%30 == 0 && vm.Count < 50 {
		vm.SpawnCar(rm)
	}

	for i := 0; i < VehiclePoolSize; i++ {
		v := &vm.Pool[i]
		if v.Lifecycle != LifecycleActive {
			continue
		}
		if v.Lifecycle == LifecycleSuspended {
			continue
		}

		if v.RoadSeg < 0 || v.RoadSeg >= len(rm.Segments) {
			v.RemovalTimer = 0
			v.Lifecycle = LifecycleMarkedForRemoval
			continue
		}

		seg := rm.Segments[v.RoadSeg]
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
					s := rm.Segments[sid]
					if (s.NodeA == seg.NodeA && s.NodeB == nextNode) || (s.NodeA == nextNode && s.NodeB == seg.NodeA) ||
						(s.NodeA == seg.NodeB && s.NodeB == nextNode) || (s.NodeA == nextNode && s.NodeB == seg.NodeB) {
						v.RoadSeg = int(sid)
						v.SegProgress = 0
						seg = s
						found = true
						break
					}
				}
				if !found {
					v.RemovalTimer = 0
					v.Lifecycle = LifecycleMarkedForRemoval
				}
			} else {
				v.RemovalTimer = 0
				v.Lifecycle = LifecycleMarkedForRemoval
			}
		}
	}

	vm.processRemovals()
}

func (vm *VehicleManager) processRemovals() {
	for i := 0; i < VehiclePoolSize; i++ {
		v := &vm.Pool[i]
		switch v.Lifecycle {
		case LifecycleMarkedForRemoval:
			v.RemovalTimer++
			if v.RemovalTimer > 10 {
				v.Lifecycle = LifecycleDestroyed
			}
		case LifecycleDestroyed:
			vm.Free(int32(i))
		}
	}
}

func (vm *VehicleManager) Draw(h *Heightmap) {
	vm.ForEach(func(v *Vehicle, _ int32) {
		hy := h.WorldHeight(v.Position.X, v.Position.Z) + 0.5
		rl.DrawCube(rl.NewVector3(v.Position.X, hy, v.Position.Z), 1.5, 0.5, 1, v.Color)
		rl.DrawCube(rl.NewVector3(v.Position.X, hy+0.4, v.Position.Z+0.6), 0.8, 0.3, 0.1, rl.NewColor(200, 50, 50, 255))
		rl.DrawCube(rl.NewVector3(v.Position.X, hy+0.4, v.Position.Z-0.6), 0.8, 0.3, 0.1, rl.NewColor(200, 50, 50, 255))
	})
}

func (vm *VehicleManager) OnRoadRemoved(segIdx int) {
	for i := 0; i < VehiclePoolSize; i++ {
		v := &vm.Pool[i]
		if v.Lifecycle == LifecycleActive && v.RoadSeg == segIdx {
			v.Lifecycle = LifecycleMarkedForRemoval
			v.RemovalTimer = 0
		}
	}
}

func (vm *VehicleManager) Unload() {
	for i := 0; i < VehiclePoolSize; i++ {
		if vm.Pool[i].Lifecycle == LifecycleActive {
			vm.Pool[i].Lifecycle = LifecycleReturnedToPool
		}
	}
	vm.Count = 0
}
