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

func (vm *VehicleManager) SpawnCar(rm *RoadManager) {
	if len(rm.Segments) == 0 {
		return
	}
	segIdx := int(vm.NextID) % len(rm.Segments)
	seg := rm.Segments[segIdx]
	na := &rm.Nodes[seg.NodeA]

	lanes := roadLanes(seg.RoadType)
	lane := int(vm.NextID) % lanes

	slot := vm.Alloc()
	if slot < 0 {
		return
	}
	v := &vm.Pool[slot]
	v.Entity = NewEntity(vm.NextID, na.X, 0, na.Z, OwnerVehicle)
	v.Lifecycle = LifecycleActive
	v.Type = VehicleCar
	v.TargetSpeed = roadSpeed(seg.RoadType) * 0.8
	v.RoadSeg = segIdx
	v.Lane = lane
	v.Color = rl.NewColor(uint8(100+vm.NextID%100), uint8(50+vm.NextID%80), uint8(80+vm.NextID%100), 255)
	vm.NextID++

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
	lanes := roadLanes(seg.RoadType)
	if lanes <= 1 {
		return 0
	}
	na := &rm.Nodes[seg.NodeA]
	nb := &rm.Nodes[seg.NodeB]

	connCount := len(na.Connected)
	if connCount <= 2 {
		_ = nb
		return v.Lane
	}

	rightmost := lanes - 1
	if connCount >= 3 {
		return rightmost
	}
	return 0
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

		na := &rm.Nodes[seg.NodeA]

		if na.HasTrafficLight && na.JunctionType == 1 {
			dx := na.X - v.Position.X
			dz := na.Z - v.Position.Z
			distToNode := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if distToNode < 15 && na.TrafficLightPhase >= 2 {
				v.Speed *= 0.9
				if v.Speed < 2 {
					v.Speed = 2
				}
				v.Waiting++
			} else {
				v.Waiting = 0
			}
		}

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

		lanes := roadLanes(seg.RoadType)
		halfLane := float32(lanes)*laneW*0.5 - float32(v.Lane)*laneW - laneW*0.5

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

		v.Position.X = xs[idx] + (xs[idx+1]-xs[idx])*frac + perX*halfLane
		v.Position.Z = zs[idx] + (zs[idx+1]-zs[idx])*frac + perZ*halfLane

		if v.SegProgress >= totalLen {
			v.RemovalTimer = 0
			v.Lifecycle = LifecycleMarkedForRemoval
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
