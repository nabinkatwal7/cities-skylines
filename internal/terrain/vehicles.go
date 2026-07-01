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

type VehicleManager struct {
	Vehicles []Vehicle
	NextID   uint32
	Timer    int32
}

func NewVehicleManager() *VehicleManager {
	return &VehicleManager{}
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

	v := Vehicle{
		Entity:      NewEntity(vm.NextID, na.X, 0, na.Z, OwnerVehicle),
		Type:        VehicleCar,
		TargetSpeed: roadSpeed(seg.RoadType) * 0.8,
		RoadSeg:     segIdx,
		Lane:        lane,
		Color:       rl.NewColor(uint8(100+vm.NextID%100), uint8(50+vm.NextID%80), uint8(80+vm.NextID%100), 255),
	}
	vm.Vehicles = append(vm.Vehicles, v)
	vm.NextID++

	if len(vm.Vehicles) > 200 {
		vm.Vehicles = vm.Vehicles[1:]
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
	if vm.Timer%30 == 0 && len(vm.Vehicles) < 50 {
		vm.SpawnCar(rm)
	}

	for i := range vm.Vehicles {
		v := &vm.Vehicles[i]
		if v.HasFlag(FlagParked) {
			continue
		}

		if v.RoadSeg < 0 || v.RoadSeg >= len(rm.Segments) {
			v.SetFlag(FlagParked)
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
			v.SetFlag(FlagParked)
		}
	}
}

func (vm *VehicleManager) Draw(h *Heightmap) {
	for _, v := range vm.Vehicles {
		if v.HasFlag(FlagParked) {
			continue
		}
		hy := h.WorldHeight(v.Position.X, v.Position.Z) + 0.5
		rl.DrawCube(rl.NewVector3(v.Position.X, hy, v.Position.Z), 1.5, 0.5, 1, v.Color)
		rl.DrawCube(rl.NewVector3(v.Position.X, hy+0.4, v.Position.Z+0.6), 0.8, 0.3, 0.1, rl.NewColor(200, 50, 50, 255))
		rl.DrawCube(rl.NewVector3(v.Position.X, hy+0.4, v.Position.Z-0.6), 0.8, 0.3, 0.1, rl.NewColor(200, 50, 50, 255))
	}
}

func (vm *VehicleManager) Unload() {
	vm.Vehicles = nil
}
