package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type TransportType uint8

const (
	TransBus TransportType = iota
	TransTram
	TransMetro
	TransTrain
	TransFerry
	TransMonorail
	TransCableCar
	TransTaxi
	TransAir
	TransShip
)

type TransportStop struct {
	ID          uint32
	X, Z        float32
	Name        string
	TransType   TransportType
	Passengers  int32
	IsStation   bool
	Underground bool
}

type TransportLine struct {
	ID        uint32
	Name      string
	TransType TransportType
	Stops     []uint32
	Active    bool
	Color     rl.Color
}

type TransportVehicle struct {
	ID         uint32
	LineID     uint32
	TransType  TransportType
	X, Z       float32
	Speed      float32
	StopIdx    int
	Passengers int32
	Capacity   int32
	Forward    bool
	Timer      int32
	Moving     bool
	TargetX    float32
	TargetZ    float32
}

type TransportNetwork struct {
	Type             TransportType
	Active           bool
	VehicleCount     int32
	RouteCount       int32
	StopCount        int32
	StationCount     int32
	PassengersPerDay int32
	TotalIncome      float32
	MaintenanceCost  float32
	Capacity         int32
}

const TransportVehiclePoolSize = 200

type TransportManager struct {
	Stops    []TransportStop
	Lines    []TransportLine
	Vehicles []TransportVehicle
	NextID   uint32

	Networks [10]TransportNetwork

	Pool     [TransportVehiclePoolSize]TransportVehicle
	FreeList []int32
	PoolNext uint32
}

func NewTransportManager() *TransportManager {
	tm := &TransportManager{
		FreeList: make([]int32, TransportVehiclePoolSize),
	}
	for i := 0; i < TransportVehiclePoolSize; i++ {
		tm.Pool[i].ID = math.MaxUint32
		tm.FreeList[i] = int32(TransportVehiclePoolSize - 1 - i)
	}
	for i := range tm.Networks {
		tm.Networks[i].Type = TransportType(i)
		tm.Networks[i].Active = true
	}
	return tm
}

func (tm *TransportManager) allocVehicle() int32 {
	if len(tm.FreeList) == 0 {
		return -1
	}
	idx := tm.FreeList[len(tm.FreeList)-1]
	tm.FreeList = tm.FreeList[:len(tm.FreeList)-1]
	tm.Pool[idx] = TransportVehicle{}
	return int32(idx)
}

func (tm *TransportManager) freeVehicle(slot int32) {
	if slot < 0 || int(slot) >= TransportVehiclePoolSize {
		return
	}
	tm.Pool[slot].ID = math.MaxUint32
	tm.FreeList = append(tm.FreeList, slot)
}

func (tm *TransportManager) forEachVehicle(fn func(*TransportVehicle, int32)) {
	for i := 0; i < TransportVehiclePoolSize; i++ {
		if tm.Pool[i].ID != math.MaxUint32 {
			fn(&tm.Pool[i], int32(i))
		}
	}
}

func (tm *TransportManager) AddStop(x, z float32, tt TransportType) uint32 {
	id := tm.NextID
	tm.NextID++
	isStation := tt == TransMetro || tt == TransTrain || tt == TransMonorail || tt == TransAir || tt == TransShip
	tm.Stops = append(tm.Stops, TransportStop{
		ID:          id,
		X:           x, Z: z,
		Name:        "Stop",
		TransType:   tt,
		IsStation:   isStation,
		Underground: tt == TransMetro,
	})
	if int(tt) < len(tm.Networks) {
		tm.Networks[tt].StopCount++
		if isStation {
			tm.Networks[tt].StationCount++
		}
	}
	return id
}

func (tm *TransportManager) AddLine(name string, tt TransportType, stopIDs []uint32, col rl.Color) uint32 {
	id := tm.NextID
	tm.NextID++
	tm.Lines = append(tm.Lines, TransportLine{
		ID:        id,
		Name:      name,
		TransType: tt,
		Stops:     stopIDs,
		Active:    true,
		Color:     col,
	})
	if int(tt) < len(tm.Networks) {
		tm.Networks[tt].RouteCount++
	}
	return id
}

func (tm *TransportManager) AddStopToLine(lineID, stopID uint32) {
	for i := range tm.Lines {
		if tm.Lines[i].ID == lineID {
			tm.Lines[i].Stops = append(tm.Lines[i].Stops, stopID)
			tm.Lines[i].Active = len(tm.Lines[i].Stops) >= 2
			return
		}
	}
}

func (tm *TransportManager) SpawnVehicle(lineIdx int) {
	if lineIdx < 0 || lineIdx >= len(tm.Lines) {
		return
	}
	line := &tm.Lines[lineIdx]
	if len(line.Stops) < 2 {
		return
	}
	stop0 := &tm.Stops[line.Stops[0]]
	cap := int32(30)
	spd := float32(20)
	switch line.TransType {
	case TransTram:
		cap, spd = 60, 25
	case TransMetro:
		cap, spd = 120, 40
	case TransTrain:
		cap, spd = 200, 60
	case TransFerry:
		cap, spd = 80, 15
	case TransMonorail:
		cap, spd = 100, 35
	case TransCableCar:
		cap, spd = 20, 8
	case TransTaxi:
		cap, spd = 4, 30
	case TransAir:
		cap, spd = 180, 150
	case TransShip:
		cap, spd = 250, 25
	}

	slot := tm.allocVehicle()
	if slot < 0 {
		tm.Vehicles = append(tm.Vehicles, TransportVehicle{
			ID:        tm.NextID,
			LineID:    line.ID,
			TransType: line.TransType,
			X:         stop0.X,
			Z:         stop0.Z,
			Capacity:  cap,
			Speed:     spd,
			Forward:   true,
			Moving:    true,
		})
		tm.NextID++
		netIdx := int(line.TransType)
		if netIdx < len(tm.Networks) {
			tm.Networks[netIdx].VehicleCount++
			tm.Networks[netIdx].Capacity += cap
		}
		return
	}
	v := &tm.Pool[slot]
	v.ID = tm.PoolNext
	tm.PoolNext++
	v.LineID = line.ID
	v.TransType = line.TransType
	v.X = stop0.X
	v.Z = stop0.Z
	v.Capacity = cap
	v.Speed = spd
	v.Forward = true
	v.Moving = true
	v.StopIdx = 0

	netIdx := int(line.TransType)
	if netIdx < len(tm.Networks) {
		tm.Networks[netIdx].VehicleCount++
		tm.Networks[netIdx].Capacity += cap
	}
}

func (tm *TransportManager) Update(rm *RoadManager, h *Heightmap) {
	for li := range tm.Lines {
		line := &tm.Lines[li]
		if !line.Active {
			continue
		}
		count := 0
		for _, v := range tm.Vehicles {
			if v.LineID == line.ID {
				count++
			}
		}
		tm.forEachVehicle(func(v *TransportVehicle, _ int32) {
			if v.LineID == line.ID {
				count++
			}
		})
		desired := 1
		switch line.TransType {
		case TransBus:
			desired = 2
		case TransMetro:
			desired = 3
		case TransTrain:
			desired = 2
		case TransFerry:
			desired = 1
		case TransTram:
			desired = 2
		case TransMonorail:
			desired = 2
		case TransCableCar:
			desired = 1
		}
		if count < desired {
			tm.SpawnVehicle(li)
		}
	}

	tm.updatePoolVehicles(rm, h)
	tm.updateSliceVehicles(h)

	for si := range tm.Stops {
		s := &tm.Stops[si]
		if s.Passengers < 5 {
			s.Passengers++
		}
	}

	for i := range tm.Networks {
		net := &tm.Networks[i]
		if net.TotalIncome > 0 {
			net.TotalIncome *= 0.95
		}
	}
}

func (tm *TransportManager) updatePoolVehicles(rm *RoadManager, h *Heightmap) {
	for i := 0; i < TransportVehiclePoolSize; i++ {
		v := &tm.Pool[i]
		if v.ID == math.MaxUint32 {
			continue
		}
		tm.moveVehicle(v, rm, h)
	}
}

func (tm *TransportManager) updateSliceVehicles(h *Heightmap) {
	for i := range tm.Vehicles {
		v := &tm.Vehicles[i]
		tm.moveVehicle(v, nil, h)
	}
}

func (tm *TransportManager) moveVehicle(v *TransportVehicle, rm *RoadManager, h *Heightmap) {
	lineIdx := -1
	for li, line := range tm.Lines {
		if line.ID == v.LineID {
			lineIdx = li
			break
		}
	}
	if lineIdx < 0 {
		return
	}
	line := &tm.Lines[lineIdx]
	if len(line.Stops) < 2 {
		return
	}

	currentStop := tm.stopByID(line.Stops[v.StopIdx])
	nextIdx := (v.StopIdx + 1) % len(line.Stops)
	if !v.Forward {
		nextIdx = (v.StopIdx - 1 + len(line.Stops)) % len(line.Stops)
	}
	nextStop := tm.stopByID(line.Stops[nextIdx])
	if nextStop == nil {
		return
	}

	if v.Moving {
		v.TargetX = nextStop.X
		v.TargetZ = nextStop.Z
	}

	dx := nextStop.X - v.X
	dz := nextStop.Z - v.Z
	dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

	if dist < 3 && v.Moving {
		v.Moving = false
		v.Timer = 0
		v.StopIdx = nextIdx
		if v.StopIdx == 0 || v.StopIdx == len(line.Stops)-1 {
			v.Forward = !v.Forward
		}
		boarded := v.Capacity / 10
		if v.Passengers+boarded > v.Capacity {
			boarded = v.Capacity - v.Passengers
		}
		v.Passengers += boarded
		if currentStop != nil {
			currentStop.Passengers -= boarded
			if currentStop.Passengers < 0 {
				currentStop.Passengers = 0
			}
		}
		netIdx := int(v.TransType)
		if netIdx < len(tm.Networks) {
			tm.Networks[netIdx].PassengersPerDay += boarded
			tm.Networks[netIdx].TotalIncome += float32(boarded) * 0.5
		}
	}

	if !v.Moving {
		v.Timer++
		wait := int32(60)
		switch v.TransType {
		case TransMetro, TransTrain, TransFerry, TransMonorail:
			wait = 90
		case TransAir, TransShip:
			wait = 120
		}
		if v.Timer > wait {
			v.Moving = true
			if currentStop != nil {
				currentStop.Passengers += 2
			}
		}
		return
	}

	if dist > 0.5 {
		spd := v.Speed
		if rm != nil && v.TransType == TransBus && len(rm.Segments) > 0 {
			nearest := -1
			bestD := float32(math.MaxFloat32)
			for si, seg := range rm.Segments {
				xs, zs, _ := rm.SampleSegment(seg, 8)
				for j := 0; j < len(xs); j++ {
					dx2 := v.X - xs[j]
					dz2 := v.Z - zs[j]
					d := dx2*dx2 + dz2*dz2
					if d < bestD {
						bestD = d
						nearest = si
					}
				}
			}
			if nearest >= 0 {
				seg := rm.Segments[nearest]
				if seg.Damaged {
					spd *= 0.7
				}
			}
		}
		v.X += (dx / dist) * spd * 0.02
		v.Z += (dz / dist) * spd * 0.02
	}
}

func (tm *TransportManager) stopByID(id uint32) *TransportStop {
	for i := range tm.Stops {
		if tm.Stops[i].ID == id {
			return &tm.Stops[i]
		}
	}
	return nil
}

func transportTypeName(tt TransportType) string {
	switch tt {
	case TransBus:
		return "Bus"
	case TransTram:
		return "Tram"
	case TransMetro:
		return "Metro"
	case TransTrain:
		return "Train"
	case TransFerry:
		return "Ferry"
	case TransMonorail:
		return "Monorail"
	case TransCableCar:
		return "Cable Car"
	case TransTaxi:
		return "Taxi"
	case TransAir:
		return "Air"
	case TransShip:
		return "Ship"
	default:
		return "Unknown"
	}
}

func (tm *TransportManager) Draw(h *Heightmap) {
	for _, s := range tm.Stops {
		hy := h.WorldHeight(s.X, s.Z) + 0.5
		col := TransportStopColor(s.TransType)
		switch {
		case s.Underground:
			rl.DrawCube(rl.NewVector3(s.X, 0, s.Z), 3, 0.5, 3, col)
			rl.DrawCube(rl.NewVector3(s.X, 1.5, s.Z), 1, 3, 1, rl.NewColor(150, 150, 220, 200))
		case s.TransType == TransTram:
			rl.DrawCube(rl.NewVector3(s.X, hy, s.Z), 1.5, 0.5, 1.5, col)
		case s.IsStation:
			rl.DrawCube(rl.NewVector3(s.X, hy+1, s.Z), 4, 2, 4, col)
			rl.DrawCubeWires(rl.NewVector3(s.X, hy+1, s.Z), 4, 2, 4, rl.NewColor(60, 60, 60, 100))
		default:
			rl.DrawCube(rl.NewVector3(s.X, hy, s.Z), 1, 1, 1, col)
		}
	}

	for i := 0; i < TransportVehiclePoolSize; i++ {
		v := &tm.Pool[i]
		if v.ID == math.MaxUint32 {
			continue
		}
		tm.drawVehicle(v, h)
	}
	for _, v := range tm.Vehicles {
		tm.drawVehicle(&v, h)
	}
}

func TransportStopColor(tt TransportType) rl.Color {
	switch tt {
	case TransBus:
		return rl.NewColor(0, 150, 200, 200)
	case TransTram:
		return rl.NewColor(200, 100, 200, 200)
	case TransMetro:
		return rl.NewColor(100, 100, 200, 200)
	case TransTrain:
		return rl.NewColor(200, 150, 50, 200)
	case TransFerry:
		return rl.NewColor(50, 150, 200, 200)
	case TransMonorail:
		return rl.NewColor(100, 200, 200, 200)
	case TransCableCar:
		return rl.NewColor(200, 200, 100, 200)
	case TransTaxi:
		return rl.NewColor(200, 200, 50, 200)
	case TransAir:
		return rl.NewColor(200, 100, 100, 200)
	case TransShip:
		return rl.NewColor(50, 100, 200, 200)
	default:
		return rl.NewColor(150, 150, 150, 200)
	}
}

func (tm *TransportManager) drawVehicle(v *TransportVehicle, h *Heightmap) {
	hy := h.WorldHeight(v.X, v.Z) + 0.6
	switch v.TransType {
	case TransBus:
		rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 2.5, 0.8, 1.2, rl.NewColor(0, 100, 200, 255))
		rl.DrawCube(rl.NewVector3(v.X, hy+0.5, v.Z+0.8), 1.5, 0.3, 0.1, rl.NewColor(200, 200, 100, 255))
	case TransTram:
		rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 4, 0.6, 1, rl.NewColor(200, 50, 200, 255))
		rl.DrawCube(rl.NewVector3(v.X, hy+0.4, v.Z+0.6), 3, 0.2, 0.1, rl.NewColor(255, 255, 100, 255))
	case TransMetro:
		rl.DrawCube(rl.NewVector3(v.X, 0.5, v.Z), 5, 1, 1.5, rl.NewColor(100, 100, 220, 255))
		rl.DrawCube(rl.NewVector3(v.X, 1, v.Z), 4.5, 0.3, 1.3, rl.NewColor(200, 200, 255, 255))
	case TransTrain:
		rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 6, 1.2, 1.5, rl.NewColor(200, 150, 50, 255))
		rl.DrawCube(rl.NewVector3(v.X, hy+0.8, v.Z+1), 4, 0.3, 0.1, rl.NewColor(255, 200, 100, 255))
	case TransFerry:
		rl.DrawCube(rl.NewVector3(v.X, 0.3, v.Z), 5, 0.5, 2.5, rl.NewColor(50, 150, 200, 255))
	case TransMonorail:
		rl.DrawCube(rl.NewVector3(v.X, hy+2, v.Z), 4, 0.6, 1, rl.NewColor(100, 200, 200, 255))
		rl.DrawCube(rl.NewVector3(v.X, hy+1.7, v.Z), 0.3, 0.5, 0.3, rl.NewColor(150, 150, 150, 255))
	case TransCableCar:
		rl.DrawCube(rl.NewVector3(v.X, hy+3, v.Z), 1.5, 0.8, 1, rl.NewColor(200, 200, 100, 255))
	case TransTaxi:
		rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 1.5, 0.5, 1, rl.NewColor(200, 200, 50, 255))
	case TransAir:
		rl.DrawCube(rl.NewVector3(v.X, hy+5, v.Z), 8, 1.5, 6, rl.NewColor(200, 100, 100, 255))
	case TransShip:
		rl.DrawCube(rl.NewVector3(v.X, 0.5, v.Z), 8, 1, 3, rl.NewColor(50, 100, 200, 255))
	}
}

func (tm *TransportManager) Unload() {
	tm.Vehicles = nil
	tm.Lines = nil
	tm.Stops = nil
}

func (tm *TransportManager) TotalMaintenance() float32 {
	var total float32
	for i := range tm.Networks {
		total += tm.Networks[i].MaintenanceCost
	}
	return total
}

func (tm *TransportManager) TotalIncome() float32 {
	var total float32
	for i := range tm.Networks {
		total += tm.Networks[i].TotalIncome
	}
	return total
}
