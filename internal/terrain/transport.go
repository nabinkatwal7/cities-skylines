package terrain

import (
	"math"
	"math/rand"

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
	TransWalk
	TransBicycle
	TransCar
	TransBlimp
)

const TransportModeCount = 14

type TransportModeConfig struct {
	Speed           float32
	Capacity        int32
	OperatingCost   float32
	TicketRevenue   float32
	Pollution       float32
	Noise           float32
	SizeX, SizeY, SizeZ float32
	NeedsRoad       bool
	NeedsRails      bool
	NeedsTrack      bool
	NeedsWater      bool
	NeedsUnderground bool
}

var transportModeConfigs = [TransportModeCount]TransportModeConfig{
	TransBus:     {40, 30, 2, 2, 0.8, 0.6, 2.5, 0.8, 1.2, true, false, false, false, false},
	TransTram:    {30, 60, 3, 3, 0.3, 0.8, 4.0, 0.6, 1.0, false, false, true, false, false},
	TransMetro:   {50, 120, 6, 5, 0.1, 0.4, 5.0, 1.0, 1.5, false, false, false, false, true},
	TransTrain:   {70, 200, 10, 8, 0.3, 0.9, 6.0, 1.2, 1.5, false, true, false, false, false},
	TransFerry:   {20, 80, 5, 4, 0.5, 0.5, 5.0, 0.5, 2.5, false, false, false, true, false},
	TransMonorail: {45, 100, 5, 4, 0.2, 0.5, 4.0, 0.6, 1.0, false, false, true, false, false},
	TransCableCar: {12, 20, 2, 2, 0.0, 0.2, 1.5, 0.8, 1.0, false, false, false, false, false},
	TransTaxi:    {50, 4, 1, 10, 0.9, 0.4, 1.5, 0.5, 1.0, true, false, false, false, false},
	TransAir:     {200, 180, 30, 20, 2.0, 2.5, 8.0, 1.5, 6.0, false, false, false, false, false},
	TransShip:    {30, 250, 15, 12, 1.5, 1.0, 8.0, 1.0, 3.0, false, false, false, true, false},
	TransWalk:    {6, 1, 0, 0, 0.0, 0.0, 0.5, 1.0, 0.5, true, false, false, false, false},
	TransBicycle: {15, 1, 0, 0, 0.0, 0.1, 1.0, 0.5, 0.3, true, false, false, false, false},
	TransCar:     {50, 4, 0, 0, 1.0, 0.7, 2.0, 0.6, 1.0, true, false, false, false, false},
	TransBlimp:   {25, 60, 8, 6, 0.4, 0.3, 6.0, 1.2, 3.0, false, false, false, false, false},
}

type TransportStop struct {
	ID          uint32
	X, Z        float32
	Name        string
	TransType   TransportType
	ConnectedNetworks []TransportType
	Passengers  int32
	IsStation   bool
	Underground bool
	DistrictID  int32
	Accessibility float32
	Capacity    int32
	BoardingQueue []uint32
	TransferStationID uint32
}

type TransportLine struct {
	ID             uint32
	Name           string
	TransType      TransportType
	Stops          []uint32
	Active         bool
	Color          rl.Color
	VehicleCount   int32
	PassengerCount int32
	Budget         float32
	TotalPassengers int64
	TotalIncome    float32
	IsCircular     bool
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
	StandingCapacity int32
	Forward    bool
	Timer      int32
	Moving     bool
	TargetX    float32
	TargetZ    float32
	Path       []uint32
	PathIdx    int
	FuelType   uint8
	Maintenance float32
	Delay      int32
	CurrentStopID uint32
	HomeDepotType   uint8
	HomeDepotSlot   int32
	MaintenanceTimer int32
}

type TransportNetwork struct {
	Type              TransportType
	Active            bool
	VehicleCount      int32
	RouteCount        int32
	StopCount         int32
	StationCount      int32
	PassengersPerDay  int32
	WeeklyPassengers  int32
	LifetimePassengers int64
	TotalIncome       float32
	TotalExpenses     float32
	MaintenanceCost   float32
	Capacity          int32
	AvgWaitTime       float32
	CapacityUsage     float32
	VehicleUtilization float32
	Pollution         float32
	Noise             float32
}

type TransferStation struct {
	ID      uint32
	Name    string
	StopIDs []uint32
	X, Z    float32
}

const TransportVehiclePoolSize = 200
const MetroTrackPoolSize = 500

type MetroTrack struct {
	ID         uint32
	StartX, StartZ float32
	EndX, EndZ float32
	Length     float32
	Lifecycle  LifecycleState
	Elevated   bool
}

type TrackManager struct {
	Pool     [MetroTrackPoolSize]MetroTrack
	FreeList []int32
	Count    int32
	NextID   uint32
}

type TransportManager struct {
	Stops    []TransportStop
	Lines    []TransportLine
	Vehicles []TransportVehicle
	NextID   uint32

	Networks [TransportModeCount]TransportNetwork

	Pool     [TransportVehiclePoolSize]TransportVehicle
	FreeList []int32
	PoolNext uint32

	Parking *ParkingManager
	Tracks  *TrackManager
	Rails   *RailManager
	Cargo   *CargoManager
	Citizens *CitizenManager

	CableConnections []CableConnection
	TransferStations []TransferStation
	TransferNextID   uint32

	RoadCongestion float32
}

type CableConnection struct {
	ID         uint32
	StopA, StopB uint32
	StartX, StartZ float32
	EndX, EndZ float32
}

func NewTransportManager() *TransportManager {
	tm := &TransportManager{
		FreeList: make([]int32, TransportVehiclePoolSize),
		Tracks:   NewTrackManager(),
		Rails:    NewRailManager(),
		Cargo:    NewCargoManager(),
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

func NewTrackManager() *TrackManager {
	tm := &TrackManager{
		FreeList: make([]int32, MetroTrackPoolSize),
	}
	for i := 0; i < MetroTrackPoolSize; i++ {
		tm.Pool[i].Lifecycle = LifecycleUnallocated
		tm.FreeList[i] = int32(MetroTrackPoolSize - 1 - i)
	}
	return tm
}

func (trk *TrackManager) allocTrack() int32 {
	if len(trk.FreeList) == 0 {
		return -1
	}
	idx := trk.FreeList[len(trk.FreeList)-1]
	trk.FreeList = trk.FreeList[:len(trk.FreeList)-1]
	trk.Pool[idx] = MetroTrack{}
	trk.Pool[idx].Lifecycle = LifecycleAllocated
	trk.Count++
	return int32(idx)
}

func (trk *TrackManager) freeTrack(slot int32) {
	if slot < 0 || int(slot) >= MetroTrackPoolSize {
		return
	}
	trk.Pool[slot].Lifecycle = LifecycleReturnedToPool
	trk.FreeList = append(trk.FreeList, slot)
	trk.Count--
}

func (trk *TrackManager) AddTrack(startX, startZ, endX, endZ float32) int32 {
	return trk.addTrackInternal(startX, startZ, endX, endZ, false)
}

func (trk *TrackManager) AddTrackElevated(startX, startZ, endX, endZ float32) int32 {
	return trk.addTrackInternal(startX, startZ, endX, endZ, true)
}

func (trk *TrackManager) addTrackInternal(startX, startZ, endX, endZ float32, elevated bool) int32 {
	slot := trk.allocTrack()
	if slot < 0 {
		return -1
	}
	t := &trk.Pool[slot]
	t.ID = trk.NextID
	trk.NextID++
	t.StartX = startX
	t.StartZ = startZ
	t.EndX = endX
	t.EndZ = endZ
	dx := endX - startX
	dz := endZ - startZ
	t.Length = float32(math.Sqrt(float64(dx*dx + dz*dz)))
	t.Elevated = elevated
	return slot
}

func (trk *TrackManager) ForEachTrack(fn func(*MetroTrack, int32)) {
	for i := 0; i < MetroTrackPoolSize; i++ {
		if trk.Pool[i].Lifecycle == LifecycleAllocated {
			fn(&trk.Pool[i], int32(i))
		}
	}
}

func (trk *TrackManager) FindTrackPath(startX, startZ, endX, endZ float32) []int32 {
	type trackDist struct {
		prev    int32
		dist    float32
		visited bool
	}
	count := trk.Count
	if count == 0 {
		return nil
	}
	nodes := make([]trackDist, MetroTrackPoolSize)
	for i := range nodes {
		nodes[i].prev = -1
		nodes[i].dist = math.MaxFloat32
	}

	bestStart := int32(-1)
	bestEnd := int32(-1)
	for i := 0; i < MetroTrackPoolSize; i++ {
		t := &trk.Pool[i]
		if t.Lifecycle != LifecycleAllocated {
			continue
		}
		dx1 := t.StartX - startX
		dz1 := t.StartZ - startZ
		d1 := dx1*dx1 + dz1*dz1
		dx2 := t.EndX - startX
		dz2 := t.EndZ - startZ
		d2 := dx2*dx2 + dz2*dz2
		if d1 < d2 && d1 < 2500 {
			if bestStart < 0 || d1 < nodes[bestStart].dist {
				bestStart = int32(i)
				nodes[i].dist = d1
			}
		} else if d2 < 2500 {
			if bestStart < 0 || d2 < nodes[bestStart].dist {
				bestStart = int32(i)
				nodes[i].dist = d2
			}
		}
		dx1 = t.StartX - endX
		dz1 = t.StartZ - endZ
		d1 = dx1*dx1 + dz1*dz1
		dx2 = t.EndX - endX
		dz2 = t.EndZ - endZ
		d2 = dx2*dx2 + dz2*dz2
		if d1 < d2 && d1 < 2500 {
			if bestEnd < 0 || d1 < nodes[bestEnd].dist {
				bestEnd = int32(i)
			}
		} else if d2 < 2500 {
			if bestEnd < 0 || d2 < nodes[bestEnd].dist {
				bestEnd = int32(i)
			}
		}
	}
	if bestStart < 0 || bestEnd < 0 {
		return nil
	}

	for {
		best := -1
		bestD := float32(math.MaxFloat32)
		for i := 0; i < MetroTrackPoolSize; i++ {
			if trk.Pool[i].Lifecycle != LifecycleAllocated {
				continue
			}
			if !nodes[i].visited && nodes[i].dist < bestD {
				best = i
				bestD = nodes[i].dist
			}
		}
		if best < 0 || int32(best) == bestEnd {
			break
		}
		nodes[best].visited = true
		t := &trk.Pool[best]

		for j := 0; j < MetroTrackPoolSize; j++ {
			if trk.Pool[j].Lifecycle != LifecycleAllocated || j == best {
				continue
			}
			other := &trk.Pool[j]
			shared := false
			if (t.StartX == other.StartX && t.StartZ == other.StartZ) ||
				(t.StartX == other.EndX && t.StartZ == other.EndZ) ||
				(t.EndX == other.StartX && t.EndZ == other.StartZ) ||
				(t.EndX == other.EndX && t.EndZ == other.EndZ) {
				shared = true
			} else {
				dx := t.EndX - t.StartX
				dz := t.EndZ - t.StartZ
				ox := other.StartX - t.StartX
				oz := other.StartZ - t.StartZ
				if dx*oz-dz*ox < 1 && dx*oz-dz*ox > -1 {
					dot := dx*ox + dz*oz
					len2 := dx*dx + dz*dz
					if dot > 0 && dot < len2 {
						shared = true
					}
				}
			}
			if shared {
				cost := other.Length
				nd := nodes[best].dist + cost
				if nd < nodes[j].dist {
					nodes[j].dist = nd
					nodes[j].prev = int32(best)
				}
			}
		}
	}

	if nodes[bestEnd].prev < 0 {
		return nil
	}

	path := make([]int32, 0)
	cur := bestEnd
	for cur >= 0 {
		path = append(path, cur)
		if cur == bestStart {
			break
		}
		cur = nodes[cur].prev
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func (tm *TransportManager) AddStop(x, z float32, tt TransportType) uint32 {
	id := tm.NextID
	tm.NextID++
	isStation := tt == TransMetro || tt == TransTrain || tt == TransMonorail || tt == TransAir || tt == TransShip || tt == TransBlimp
	tm.Stops = append(tm.Stops, TransportStop{
		ID:                id,
		X:                 x, Z: z,
		Name:             "Stop",
		TransType:        tt,
		ConnectedNetworks: []TransportType{tt},
		IsStation:        isStation,
		Underground:      tt == TransMetro,
		DistrictID:       -1,
		Accessibility:    0,
		Capacity:         50,
		TransferStationID: math.MaxUint32,
	})
	if int(tt) < len(tm.Networks) {
		tm.Networks[tt].StopCount++
		if isStation {
			tm.Networks[tt].StationCount++
		}
	}

	linkDist := float32(15)
	for i := range tm.Stops {
		s := &tm.Stops[i]
		if s.ID == id || s.TransType == tt {
			continue
		}
		dx := s.X - x
		dz := s.Z - z
		if dx*dx+dz*dz <= linkDist*linkDist {
			tm.linkStops(s.ID, id)
			break
		}
	}

	return id
}

func (tm *TransportManager) linkStops(stopA, stopB uint32) {
	for i := range tm.TransferStations {
		ts := &tm.TransferStations[i]
		for _, sid := range ts.StopIDs {
			if sid == stopA {
				for _, sid2 := range ts.StopIDs {
					if sid2 == stopB {
						return
					}
				}
				ts.StopIDs = append(ts.StopIDs, stopB)
				for si := range tm.Stops {
					if tm.Stops[si].ID == stopB {
						tm.Stops[si].TransferStationID = ts.ID
						break
					}
				}
				return
			}
		}
	}

	tsID := tm.TransferNextID
	tm.TransferNextID++
	ts := TransferStation{
		ID:      tsID,
		Name:    "Transfer Hub",
		StopIDs: []uint32{stopA, stopB},
	}
	for si := range tm.Stops {
		s := &tm.Stops[si]
		if s.ID == stopA || s.ID == stopB {
			s.TransferStationID = tsID
			ts.X += s.X / 2
			ts.Z += s.Z / 2
		}
	}
	tm.TransferStations = append(tm.TransferStations, ts)
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
		IsCircular: tt != TransBus && tt != TransTram && tt != TransTaxi,
	})
	if int(tt) < len(tm.Networks) {
		tm.Networks[tt].RouteCount++
	}
	if tt == TransMetro && tm.Tracks != nil && len(stopIDs) >= 2 {
		for i := 0; i < len(stopIDs)-1; i++ {
			a := tm.StopByID(stopIDs[i])
			b := tm.StopByID(stopIDs[i+1])
			if a != nil && b != nil {
				tm.Tracks.AddTrack(a.X, a.Z, b.X, b.Z)
			}
		}
	}
	if tt == TransTrain && tm.Rails != nil && len(stopIDs) >= 2 {
		for i := 0; i < len(stopIDs)-1; i++ {
			a := tm.StopByID(stopIDs[i])
			b := tm.StopByID(stopIDs[i+1])
			if a != nil && b != nil {
				tm.Rails.AddTrack(a.X, a.Z, b.X, b.Z)
			}
		}
	}
	if tt == TransMonorail && tm.Tracks != nil && len(stopIDs) >= 2 {
		for i := 0; i < len(stopIDs)-1; i++ {
			a := tm.StopByID(stopIDs[i])
			b := tm.StopByID(stopIDs[i+1])
			if a != nil && b != nil {
				tm.Tracks.AddTrackElevated(a.X, a.Z, b.X, b.Z)
			}
		}
	}
	if tt == TransCableCar && len(stopIDs) >= 2 {
		for i := 0; i < len(stopIDs)-1; i++ {
			a := tm.StopByID(stopIDs[i])
			b := tm.StopByID(stopIDs[i+1])
			if a != nil && b != nil {
				tm.CableConnections = append(tm.CableConnections, CableConnection{
					ID:      tm.NextID,
					StopA:   stopIDs[i],
					StopB:   stopIDs[i+1],
					StartX:  a.X, StartZ: a.Z,
					EndX:    b.X, EndZ:   b.Z,
				})
				tm.NextID++
			}
		}
	}
	return id
}

func (tm *TransportManager) AddNetworkToStop(stopID uint32, tt TransportType) bool {
	s := tm.StopByID(stopID)
	if s == nil {
		return false
	}
	for _, existing := range s.ConnectedNetworks {
		if existing == tt {
			return true
		}
	}
	s.ConnectedNetworks = append(s.ConnectedNetworks, tt)
	if int(tt) < len(tm.Networks) {
		tm.Networks[tt].StopCount++
	}
	return true
}

func (tm *TransportManager) AddStopToLine(lineID, stopID uint32) {
	for i := range tm.Lines {
		if tm.Lines[i].ID == lineID {
			tm.Lines[i].Stops = append(tm.Lines[i].Stops, stopID)
			tm.Lines[i].Active = len(tm.Lines[i].Stops) >= 2
			tt := tm.Lines[i].TransType
			if tt == TransMetro && tm.Tracks != nil && len(tm.Lines[i].Stops) >= 2 {
				prev := tm.StopByID(tm.Lines[i].Stops[len(tm.Lines[i].Stops)-2])
				cur := tm.StopByID(stopID)
				if prev != nil && cur != nil {
					tm.Tracks.AddTrack(prev.X, prev.Z, cur.X, cur.Z)
				}
			}
			if tt == TransTrain && tm.Rails != nil && len(tm.Lines[i].Stops) >= 2 {
				prev := tm.StopByID(tm.Lines[i].Stops[len(tm.Lines[i].Stops)-2])
				cur := tm.StopByID(stopID)
				if prev != nil && cur != nil {
					tm.Rails.AddTrack(prev.X, prev.Z, cur.X, cur.Z)
				}
			}
			if tt == TransMonorail && tm.Tracks != nil && len(tm.Lines[i].Stops) >= 2 {
				prev := tm.StopByID(tm.Lines[i].Stops[len(tm.Lines[i].Stops)-2])
				cur := tm.StopByID(stopID)
				if prev != nil && cur != nil {
					tm.Tracks.AddTrackElevated(prev.X, prev.Z, cur.X, cur.Z)
				}
			}
			if tt == TransCableCar && len(tm.Lines[i].Stops) >= 2 {
				prev := tm.StopByID(tm.Lines[i].Stops[len(tm.Lines[i].Stops)-2])
				cur := tm.StopByID(stopID)
				if prev != nil && cur != nil {
					tm.CableConnections = append(tm.CableConnections, CableConnection{
						ID:      tm.NextID,
						StopA:   tm.Lines[i].Stops[len(tm.Lines[i].Stops)-2],
						StopB:   stopID,
						StartX:  prev.X, StartZ: prev.Z,
						EndX:    cur.X, EndZ:   cur.Z,
					})
					tm.NextID++
				}
			}
			return
		}
	}
}

func (tm *TransportManager) RemoveLine(id uint32) {
	for li, line := range tm.Lines {
		if line.ID != id {
			continue
		}
		tm.Lines = append(tm.Lines[:li], tm.Lines[li+1:]...)

		netIdx := int(line.TransType)
		if netIdx < len(tm.Networks) {
			tm.Networks[netIdx].RouteCount--
		}

		tm.forEachVehicle(func(v *TransportVehicle, slot int32) {
			if v.LineID == line.ID {
				v.ID = math.MaxUint32
				tm.Networks[netIdx].VehicleCount--
				tm.Networks[netIdx].Capacity -= v.Capacity
				tm.freeVehicle(slot)
			}
		})
		for vi := 0; vi < len(tm.Vehicles); vi++ {
			if tm.Vehicles[vi].LineID == line.ID {
				tm.Networks[netIdx].VehicleCount--
				tm.Networks[netIdx].Capacity -= tm.Vehicles[vi].Capacity
				tm.Vehicles = append(tm.Vehicles[:vi], tm.Vehicles[vi+1:]...)
				vi--
			}
		}
		return
	}
}

func (tm *TransportManager) RemoveStop(id uint32) {
	stopIdx := -1
	for si, s := range tm.Stops {
		if s.ID == id {
			stopIdx = si
			break
		}
	}
	if stopIdx < 0 {
		return
	}
	removed := tm.Stops[stopIdx]
	tm.Stops = append(tm.Stops[:stopIdx], tm.Stops[stopIdx+1:]...)

	if int(removed.TransType) < len(tm.Networks) {
		tm.Networks[removed.TransType].StopCount--
		if removed.IsStation {
			tm.Networks[removed.TransType].StationCount--
		}
	}

	for li := range tm.Lines {
		line := &tm.Lines[li]
		pos := -1
		for si, sid := range line.Stops {
			if sid == id {
				pos = si
				break
			}
		}
		if pos < 0 {
			continue
		}
		line.Stops = append(line.Stops[:pos], line.Stops[pos+1:]...)
		line.Active = len(line.Stops) >= 2

		tm.forEachVehicle(func(v *TransportVehicle, _ int32) {
			if v.LineID == line.ID {
				if v.StopIdx > pos {
					v.StopIdx--
				} else if v.StopIdx == pos {
					if len(line.Stops) > 0 {
						if v.StopIdx >= len(line.Stops) {
							v.StopIdx = len(line.Stops) - 1
						}
					} else {
						v.StopIdx = 0
					}
				}
			}
		})
		for vi := range tm.Vehicles {
			v := &tm.Vehicles[vi]
			if v.LineID == line.ID {
				if v.StopIdx > pos {
					v.StopIdx--
				} else if v.StopIdx == pos {
					if len(line.Stops) > 0 {
						if v.StopIdx >= len(line.Stops) {
							v.StopIdx = len(line.Stops) - 1
						}
					} else {
						v.StopIdx = 0
					}
				}
			}
		}
	}
}

func (tm *TransportManager) findNearestDepot(tt TransportType, x, z, maxDist float32) (uint8, int32, float32, float32) {
	if tm.Parking == nil {
		return 0, -1, x, z
	}
	switch tt {
	case TransBus:
		slot, _ := tm.Parking.NearestBusDepot(x, z, maxDist)
		if slot >= 0 { return DepotBus, slot, tm.Parking.BusDepots[slot].X, tm.Parking.BusDepots[slot].Z }
	case TransTram:
		slot, _ := tm.Parking.NearestTramDepot(x, z, maxDist)
		if slot >= 0 { return DepotTram, slot, tm.Parking.TramDepots[slot].X, tm.Parking.TramDepots[slot].Z }
	case TransMetro:
		slot, _ := tm.Parking.NearestMetroDepot(x, z, maxDist)
		if slot >= 0 { return DepotMetro, slot, tm.Parking.MetroDepots[slot].X, tm.Parking.MetroDepots[slot].Z }
	case TransFerry:
		slot, _ := tm.Parking.NearestFerryDepot(x, z, maxDist)
		if slot >= 0 { return DepotFerry, slot, tm.Parking.FerryDepots[slot].X, tm.Parking.FerryDepots[slot].Z }
	case TransMonorail:
		slot, _ := tm.Parking.NearestMonorailDepot(x, z, maxDist)
		if slot >= 0 { return DepotMonorail, slot, tm.Parking.MonorailDepots[slot].X, tm.Parking.MonorailDepots[slot].Z }
	case TransCableCar:
		slot, _ := tm.Parking.NearestCableCarDepot(x, z, maxDist)
		if slot >= 0 { return DepotCableCar, slot, tm.Parking.CableCarDepots[slot].X, tm.Parking.CableCarDepots[slot].Z }
	case TransTaxi:
		slot, _ := tm.Parking.NearestTaxiDepot(x, z, maxDist)
		if slot >= 0 { return DepotTaxi, slot, tm.Parking.TaxiDepots[slot].X, tm.Parking.TaxiDepots[slot].Z }
	case TransAir:
		slot, _ := tm.Parking.NearestAirportDepot(x, z, maxDist)
		if slot >= 0 { return DepotAirport, slot, tm.Parking.AirportDepots[slot].X, tm.Parking.AirportDepots[slot].Z }
	case TransShip:
		slot, _ := tm.Parking.NearestPortDepot(x, z, maxDist)
		if slot >= 0 { return DepotPort, slot, tm.Parking.PortDepots[slot].X, tm.Parking.PortDepots[slot].Z }
	}
	return 0, -1, x, z
}

func (tm *TransportManager) SpawnVehicle(lineIdx int) {
	if lineIdx < 0 || lineIdx >= len(tm.Lines) {
		return
	}
	line := &tm.Lines[lineIdx]
	if len(line.Stops) < 2 {
		return
	}
	if int(line.TransType) >= len(transportModeConfigs) {
		return
	}
	cfg := &transportModeConfigs[line.TransType]
	if cfg.Capacity <= 1 {
		return
	}
	stop0 := &tm.Stops[line.Stops[0]]
	spawnX, spawnZ := stop0.X, stop0.Z
	depotType := uint8(0)
	depotSlot := int32(-1)
	dt, ds, dx, dz := tm.findNearestDepot(line.TransType, spawnX, spawnZ, 5000)
	if ds >= 0 {
		depotType = dt
		depotSlot = ds
		spawnX = dx
		spawnZ = dz
	}
	cap := cfg.Capacity
	spd := cfg.Speed
	standCap := cap * 3 / 2

	slot := tm.allocVehicle()
	if slot < 0 {
		tm.Vehicles = append(tm.Vehicles, TransportVehicle{
			ID:        tm.NextID,
			LineID:    line.ID,
			TransType: line.TransType,
			X:         spawnX,
			Z:         spawnZ,
			Capacity:  cap,
			StandingCapacity: standCap,
			Speed:     spd,
			Forward:   true,
			Moving:    true,
			FuelType:  0,
			Maintenance: 1.0,
			CurrentStopID: line.Stops[0],
			HomeDepotType: depotType,
			HomeDepotSlot: depotSlot,
		})
		tm.NextID++
		line.VehicleCount++
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
	v.X = spawnX
	v.Z = spawnZ
	v.Capacity = cap
	v.StandingCapacity = standCap
	v.Speed = spd
	v.Forward = true
	v.Moving = true
	v.StopIdx = 0
	v.FuelType = 0
	v.Maintenance = 1.0
	v.CurrentStopID = line.Stops[0]
	v.HomeDepotType = depotType
	v.HomeDepotSlot = depotSlot

	line.VehicleCount++
	netIdx := int(line.TransType)
	if netIdx < len(tm.Networks) {
		tm.Networks[netIdx].VehicleCount++
		tm.Networks[netIdx].Capacity += cap
	}
}

func (tm *TransportManager) InitExternalConnections(cs *ConnectionSystem) {
	if cs == nil {
		return
	}
	for _, c := range cs.GetByType(ConnAir) {
		tm.AddStop(c.WorldX, c.WorldZ, TransAir)
		if len(tm.Stops) > 0 {
			s := &tm.Stops[len(tm.Stops)-1]
			s.Name = "Airport (External)"
			s.IsStation = true
			s.Capacity = 500
		}
	}
	for _, c := range cs.GetByType(ConnShip) {
		tm.AddStop(c.WorldX, c.WorldZ, TransShip)
		if len(tm.Stops) > 0 {
			s := &tm.Stops[len(tm.Stops)-1]
			s.Name = "Port (External)"
			s.IsStation = true
			s.Capacity = 500
		}
	}
	for _, c := range cs.GetByType(ConnRail) {
		tm.AddStop(c.WorldX, c.WorldZ, TransTrain)
		if len(tm.Stops) > 0 {
			s := &tm.Stops[len(tm.Stops)-1]
			s.Name = "Rail (External)"
			s.IsStation = true
			s.Capacity = 500
		}
	}
}

func (tm *TransportManager) Update(rm *RoadManager, dm *DistrictManager, h *Heightmap) {
	for i := 0; i < TransportVehiclePoolSize; i++ {
		if tm.Pool[i].ID != math.MaxUint32 && tm.Pool[i].Maintenance <= 0 {
			tm.freeVehicle(int32(i))
		}
	}
	for vi := len(tm.Vehicles) - 1; vi >= 0; vi-- {
		if tm.Vehicles[vi].Maintenance <= 0 {
			tm.Vehicles = append(tm.Vehicles[:vi], tm.Vehicles[vi+1:]...)
		}
	}

	if rm != nil {
		roadVehCount := 0
		tm.forEachVehicle(func(v *TransportVehicle, _ int32) {
			if v.TransType == TransBus || v.TransType == TransTaxi {
				roadVehCount++
			}
		})
		tm.RoadCongestion = float32(roadVehCount) / 50.0
		if tm.RoadCongestion > 1.0 {
			tm.RoadCongestion = 1.0
		}
	}

	for li := range tm.Lines {
		line := &tm.Lines[li]
		if !line.Active {
			continue
		}
		netIdx := int(line.TransType)
		if netIdx < len(tm.Networks) {
			tm.Networks[netIdx].TotalExpenses += transportModeConfigs[line.TransType].OperatingCost * float32(line.VehicleCount) * (line.Budget / 100.0)
		}

		line.VehicleCount = 0
		line.PassengerCount = 0
		for _, v := range tm.Vehicles {
			if v.LineID == line.ID {
				line.VehicleCount++
				line.PassengerCount += v.Passengers
			}
		}
		tm.forEachVehicle(func(v *TransportVehicle, _ int32) {
			if v.LineID == line.ID {
				line.VehicleCount++
				line.PassengerCount += v.Passengers
			}
		})
		desired := 0
		if int(line.TransType) < len(transportModeConfigs) {
			cfg := &transportModeConfigs[line.TransType]
			if cfg.Capacity > 1 {
				desired = 1
				switch line.TransType {
				case TransBus, TransTram, TransMonorail:
					desired = 2
				case TransMetro:
					desired = 3
				}
				budgetFactor := line.Budget / 100.0
				if budgetFactor < 0.5 {
					budgetFactor = 0.5
				}
				if budgetFactor > 2.0 {
					budgetFactor = 2.0
				}
				desired = int(float32(desired) * budgetFactor)
				if desired < 1 {
					desired = 1
				}
			}
		}
		if line.VehicleCount < int32(desired) {
			tm.SpawnVehicle(li)
		}
	}

	tm.updatePoolVehicles(rm, h)
	tm.updateSliceVehicles(h)

	for si := range tm.Stops {
		s := &tm.Stops[si]
		if s.Passengers < s.Capacity && s.Passengers < 5 {
			s.Passengers++
		}
		if dm != nil {
			idx := dm.DistrictAt(s.X, s.Z)
			if idx >= 0 {
				s.DistrictID = int32(dm.Districts[idx].ID)
			}
		}
		s.Accessibility = 0
		if rm != nil {
			near := rm.NearestSegment(s.X, s.Z)
			if near >= 0 {
				seg := rm.Segments[near]
				na := &rm.Nodes[seg.NodeA]
				nb := &rm.Nodes[seg.NodeB]
				mx := (na.X + nb.X) / 2
				mz := (na.Z + nb.Z) / 2
				dx := mx - s.X
				dz := mz - s.Z
				if dx*dx+dz*dz < 100 {
					s.Accessibility = 0.6
				}
			}
		}
		if s.IsStation && s.Underground {
			s.Accessibility += 0.3
		}
		if s.DistrictID >= 0 {
			s.Accessibility += 0.2
		}
		if s.Accessibility > 1 {
			s.Accessibility = 1
		}
	}

	for i := range tm.Networks {
		net := &tm.Networks[i]
		if net.TotalIncome > 0 {
			net.TotalIncome *= 0.98
		}
		net.WeeklyPassengers += net.PassengersPerDay
		if net.WeeklyPassengers > 999999 {
			net.WeeklyPassengers = 0
		}
		net.LifetimePassengers += int64(net.PassengersPerDay)
		net.PassengersPerDay = 0

		var totalCap int32
		var totalVeh int32
		tm.forEachVehicle(func(v *TransportVehicle, _ int32) {
			if v.TransType == TransportType(i) {
				totalCap += v.Capacity + v.StandingCapacity
				totalVeh++
			}
		})
		if totalCap > 0 {
			passengers := tm.countPassengers(TransportType(i))
			net.CapacityUsage = float32(passengers) / float32(totalCap)
		}
		if totalVeh > 0 {
			net.VehicleUtilization = float32(totalVeh) / float32(net.VehicleCount+1)
		}
	}
}

func (tm *TransportManager) countPassengers(tt TransportType) int32 {
	var total int32
	tm.forEachVehicle(func(v *TransportVehicle, _ int32) {
		if v.TransType == tt {
			total += v.Passengers
		}
	})
	return total
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

	nextIdx := (v.StopIdx + 1) % len(line.Stops)
	if !v.Forward {
		nextIdx = (v.StopIdx - 1 + len(line.Stops)) % len(line.Stops)
	}
	nextStop := tm.StopByID(line.Stops[nextIdx])
	if nextStop == nil {
		return
	}

	currentStop := tm.StopByID(line.Stops[v.StopIdx])

	if v.Moving {
		if v.TransType == TransBus && rm != nil && len(rm.Segments) > 0 {
			tm.moveBusOnRoad(v, rm, h, nextStop, currentStop, line, nextIdx)
		} else if v.TransType == TransTram && rm != nil && len(rm.Segments) > 0 {
			tm.moveTramOnTracks(v, rm, h, nextStop, currentStop, line, nextIdx)
		} else if (v.TransType == TransMetro || v.TransType == TransMonorail) && tm.Tracks != nil && tm.Tracks.Count > 0 {
			tm.moveMetroOnTracks(v, h, nextStop, currentStop, line, nextIdx)
		} else if v.TransType == TransTrain && tm.Rails != nil && tm.Rails.Count > 0 {
			tm.moveTrainOnRails(v, h, nextStop, currentStop, line, nextIdx)
		} else {
			tm.moveDirect(v, rm, h, nextStop, currentStop, line, nextIdx)
		}
	} else {
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
	}
}

func (tm *TransportManager) moveDirect(v *TransportVehicle, rm *RoadManager, h *Heightmap, nextStop, currentStop *TransportStop, line *TransportLine, nextIdx int) {
	v.TargetX = nextStop.X
	v.TargetZ = nextStop.Z

	dx := nextStop.X - v.X
	dz := nextStop.Z - v.Z
	dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

	if dist < 3 {
		tm.arriveAtStop(v, currentStop, nextStop, line, nextIdx)
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
		netIdx := int(v.TransType)
		if netIdx < len(tm.Networks) && netIdx < len(transportModeConfigs) {
			tm.Networks[netIdx].Pollution += transportModeConfigs[v.TransType].Pollution * 0.001
			tm.Networks[netIdx].Noise += transportModeConfigs[v.TransType].Noise * 0.001
		}
	}
}

func (tm *TransportManager) moveBusOnRoad(v *TransportVehicle, rm *RoadManager, h *Heightmap, nextStop, currentStop *TransportStop, line *TransportLine, nextIdx int) {
	// Compute path if needed
	if len(v.Path) == 0 {
		startNode, startOK := rm.NearestNode(v.X, v.Z)
		endNode, endOK := rm.NearestNode(nextStop.X, nextStop.Z)
		if !startOK || !endOK {
			tm.moveDirect(v, rm, h, nextStop, currentStop, line, nextIdx)
			return
		}
		v.Path = rm.FindPath(startNode, endNode, int(VehicleBus))
		v.PathIdx = 0
		if len(v.Path) == 0 {
			v.Path = append(v.Path, startNode, endNode)
		}
	}

	// Follow path
	for v.PathIdx < len(v.Path)-1 {
		nodeIdx := v.Path[v.PathIdx]
		nextNodeIdx := v.Path[v.PathIdx+1]
		if int(nodeIdx) >= len(rm.Nodes) || int(nextNodeIdx) >= len(rm.Nodes) {
			v.Path = nil
			return
		}
		node := &rm.Nodes[nodeIdx]
		nextNode := &rm.Nodes[nextNodeIdx]

		dx := nextNode.X - v.X
		dz := nextNode.Z - v.Z
		targetDist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if targetDist < 2 {
			v.PathIdx++
			v.Maintenance -= 0.001
			if v.Maintenance < 0 {
				v.Maintenance = 0
			}
			// Traffic light check
			if node.TrafficLight != TrafficLightNone && int(nodeIdx) < len(rm.Nodes) {
				if node.TrafficLight == TrafficLightRed || node.TrafficLight == TrafficLightYellow {
					v.Delay++
					return
				}
			}
			continue
		}

		spd := v.Speed
		spd = tm.applyCongestion(rm, int(nodeIdx), int(nextNodeIdx), spd)
		v.Delay += int32(spd / v.Speed * 10)

		v.X += (dx / targetDist) * spd * 0.02
		v.Z += (dz / targetDist) * spd * 0.02

		netIdx := int(v.TransType)
		if netIdx < len(tm.Networks) && netIdx < len(transportModeConfigs) {
			tm.Networks[netIdx].Pollution += transportModeConfigs[v.TransType].Pollution * 0.001
			tm.Networks[netIdx].Noise += transportModeConfigs[v.TransType].Noise * 0.001
		}
		return
	}

	// Reached end of path — arrived at next stop
	v.Path = nil
	tm.arriveAtStop(v, currentStop, nextStop, line, nextIdx)
	v.Path = nil
}

func (tm *TransportManager) moveTramOnTracks(v *TransportVehicle, rm *RoadManager, h *Heightmap, nextStop, currentStop *TransportStop, line *TransportLine, nextIdx int) {
	if len(v.Path) == 0 {
		startNode, startOK := rm.NearestNode(v.X, v.Z)
		endNode, endOK := rm.NearestNode(nextStop.X, nextStop.Z)
		if !startOK || !endOK {
			tm.moveDirect(v, rm, h, nextStop, currentStop, line, nextIdx)
			return
		}
		v.Path = rm.FindPath(startNode, endNode, int(VehicleTram))
		v.PathIdx = 0
		if len(v.Path) == 0 {
			v.Path = append(v.Path, startNode, endNode)
		}
	}

	for v.PathIdx < len(v.Path)-1 {
		nodeIdx := v.Path[v.PathIdx]
		nextNodeIdx := v.Path[v.PathIdx+1]
		if int(nodeIdx) >= len(rm.Nodes) || int(nextNodeIdx) >= len(rm.Nodes) {
			v.Path = nil
			return
		}
		node := &rm.Nodes[nodeIdx]
		nextNode := &rm.Nodes[nextNodeIdx]

		dx := nextNode.X - v.X
		dz := nextNode.Z - v.Z
		targetDist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if targetDist < 2 {
			v.PathIdx++
			v.Maintenance -= 0.002
			if v.Maintenance < 0 {
				v.Maintenance = 0
			}
			if node.TrafficLight != TrafficLightNone && int(nodeIdx) < len(rm.Nodes) {
				if node.TrafficLight == TrafficLightRed || node.TrafficLight == TrafficLightYellow {
					v.Delay++
					return
				}
			}
			continue
		}

		spd := v.Speed
		v.Delay += int32(spd / v.Speed * 10)

		v.X += (dx / targetDist) * spd * 0.02
		v.Z += (dz / targetDist) * spd * 0.02

		netIdx := int(v.TransType)
		if netIdx < len(tm.Networks) && netIdx < len(transportModeConfigs) {
			tm.Networks[netIdx].Pollution += transportModeConfigs[v.TransType].Pollution * 0.001
			tm.Networks[netIdx].Noise += transportModeConfigs[v.TransType].Noise * 0.001
		}
		return
	}

	v.Path = nil
	tm.arriveAtStop(v, currentStop, nextStop, line, nextIdx)
	v.Path = nil
}

func (tm *TransportManager) moveMetroOnTracks(v *TransportVehicle, h *Heightmap, nextStop, currentStop *TransportStop, line *TransportLine, nextIdx int) {
	if len(v.Path) == 0 {
		path := tm.Tracks.FindTrackPath(v.X, v.Z, nextStop.X, nextStop.Z)
		v.Path = make([]uint32, len(path))
		for i, idx := range path {
			v.Path[i] = uint32(idx)
		}
		v.PathIdx = 0
	}

	for v.PathIdx < len(v.Path)-1 {
		trackIdx := int(v.Path[v.PathIdx])
		nextTrackIdx := int(v.Path[v.PathIdx+1])
		if trackIdx >= MetroTrackPoolSize || nextTrackIdx >= MetroTrackPoolSize {
			v.Path = nil
			return
		}
		track := &tm.Tracks.Pool[trackIdx]
		nextTrack := &tm.Tracks.Pool[nextTrackIdx]

		tx, tz := track.EndX, track.EndZ
		dx := nextTrack.StartX - track.EndX
		dz := nextTrack.StartZ - track.EndZ
		if dx*dx+dz*dz > nextTrack.StartX-track.StartX*(nextTrack.StartZ-track.StartZ) {
			tx, tz = track.StartX, track.StartZ
		}

		dx = tx - v.X
		dz = tz - v.Z
		targetDist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if targetDist < 2 {
			v.PathIdx++
			v.Maintenance -= 0.001
			if v.Maintenance < 0 {
				v.Maintenance = 0
			}
			continue
		}

		v.X += (dx / targetDist) * v.Speed * 0.02
		v.Z += (dz / targetDist) * v.Speed * 0.02

		netIdx := int(v.TransType)
		if netIdx < len(tm.Networks) && netIdx < len(transportModeConfigs) {
			tm.Networks[netIdx].Pollution += transportModeConfigs[v.TransType].Pollution * 0.001
			tm.Networks[netIdx].Noise += transportModeConfigs[v.TransType].Noise * 0.001
		}
		return
	}

	v.Path = nil
	tm.arriveAtStop(v, currentStop, nextStop, line, nextIdx)
	v.Path = nil
}

func (tm *TransportManager) moveTrainOnRails(v *TransportVehicle, h *Heightmap, nextStop, currentStop *TransportStop, line *TransportLine, nextIdx int) {
	if len(v.Path) == 0 {
		path := tm.Rails.FindTrackPath(v.X, v.Z, nextStop.X, nextStop.Z)
		v.Path = make([]uint32, len(path))
		for i, idx := range path {
			v.Path[i] = uint32(idx)
		}
		v.PathIdx = 0
	}

	for v.PathIdx < len(v.Path)-1 {
		trackIdx := int(v.Path[v.PathIdx])
		nextTrackIdx := int(v.Path[v.PathIdx+1])
		if trackIdx >= RailTrackPoolSize || nextTrackIdx >= RailTrackPoolSize {
			v.Path = nil
			return
		}
		track := &tm.Rails.Pool[trackIdx]
		nextTrack := &tm.Rails.Pool[nextTrackIdx]

		tx, tz := track.EndX, track.EndZ
		dx := nextTrack.StartX - track.EndX
		dz := nextTrack.StartZ - track.EndZ
		if dx*dx+dz*dz > nextTrack.StartX-track.StartX*(nextTrack.StartZ-track.StartZ) {
			tx, tz = track.StartX, track.StartZ
		}

		if track.Occupied && track.OccupierID != v.ID {
			if track.SignalA == SignalRed || track.SignalB == SignalRed {
				v.Delay++
				return
			}
		}
		if !track.Occupied {
			track.Occupied = true
			track.OccupierID = v.ID
		}

		dx = tx - v.X
		dz = tz - v.Z
		targetDist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if targetDist < 2 {
			track.Occupied = false
			track.OccupierID = 0
			v.PathIdx++
			v.Maintenance -= 0.002
			if v.Maintenance < 0 {
				v.Maintenance = 0
			}
			continue
		}

		v.X += (dx / targetDist) * v.Speed * 0.02
		v.Z += (dz / targetDist) * v.Speed * 0.02

		netIdx := int(v.TransType)
		if netIdx < len(tm.Networks) && netIdx < len(transportModeConfigs) {
			tm.Networks[netIdx].Pollution += transportModeConfigs[v.TransType].Pollution * 0.001
			tm.Networks[netIdx].Noise += transportModeConfigs[v.TransType].Noise * 0.001
		}
		return
	}

	for _, idx := range v.Path {
		if int(idx) < RailTrackPoolSize {
			tm.Rails.Pool[idx].Occupied = false
			tm.Rails.Pool[idx].OccupierID = 0
		}
	}
	v.Path = nil
	tm.arriveAtStop(v, currentStop, nextStop, line, nextIdx)
	v.Path = nil
}

func (tm *TransportManager) applyCongestion(rm *RoadManager, fromNode, toNode int, baseSpeed float32) float32 {
	for _, sid := range rm.Nodes[fromNode].Connected {
		seg := rm.Segments[sid]
		if int(seg.NodeA) == toNode || int(seg.NodeB) == toNode {
			// Slow down based on road type
			switch roadHierarchy(seg.RoadType) {
			case HierarchyLocal:
				return baseSpeed * 0.8
			case HierarchyArterial:
				return baseSpeed * 0.9
			}
			break
		}
	}
	return baseSpeed
}

func (tm *TransportManager) arriveAtStop(v *TransportVehicle, currentStop, nextStop *TransportStop, line *TransportLine, nextIdx int) {
	v.Moving = false
	v.Timer = 0
	v.StopIdx = nextIdx
	v.CurrentStopID = line.Stops[nextIdx]
	if !line.IsCircular {
		if v.StopIdx == 0 || v.StopIdx == len(line.Stops)-1 {
			v.Forward = !v.Forward
		}
	}

	totalCap := v.Capacity + v.StandingCapacity
	available := totalCap - v.Passengers
	boarded := int32(0)
	if currentStop != nil && available > 0 {
		maxBoard := v.Capacity / 10
		if maxBoard > available {
			maxBoard = available
		}
		if maxBoard > currentStop.Passengers {
			maxBoard = currentStop.Passengers
		}
		boarded = maxBoard
	}

	v.Passengers += boarded
	line.PassengerCount += boarded
	line.TotalPassengers += int64(boarded)

	distFactor := float32(1.0)
	if nextStop != nil && currentStop != nil {
		dx := nextStop.X - currentStop.X
		dz := nextStop.Z - currentStop.Z
		segmentDist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		distFactor = 1.0 + segmentDist/500.0
		if distFactor > 3.0 {
			distFactor = 3.0
		}
	}
	ticketPrice := float32(0.5)
	netIdx := int(v.TransType)
	if netIdx < len(transportModeConfigs) {
		ticketPrice = transportModeConfigs[v.TransType].TicketRevenue
	}
	income := float32(boarded) * ticketPrice * distFactor
	touristBonus := income * 0.2
	income += touristBonus
	line.TotalIncome += income
	if currentStop != nil {
		currentStop.Passengers -= boarded
		if currentStop.Passengers < 0 {
			currentStop.Passengers = 0
		}
	}
	netIdx = int(v.TransType)
	if netIdx < len(tm.Networks) {
		tm.Networks[netIdx].PassengersPerDay += boarded
		tm.Networks[netIdx].TotalIncome += income
	}

	if tm.Citizens == nil {
		return
	}
	currentStopID := line.Stops[nextIdx]
	wasFull := available <= 0
	tm.Citizens.ForEach(func(c *Citizen, slot int32) {
		if c.State == CivWaiting && c.Journey.LegIndex < len(c.Journey.Legs) {
			leg := &c.Journey.Legs[c.Journey.LegIndex]
			if leg.LegType != LegWalk && leg.Mode == v.TransType && leg.FromStopID == currentStopID {
				if wasFull {
					c.Timer++
					if c.Timer > c.Patience*2 {
						cm := tm.Citizens
						cm.StartJourney(c, tm)
						c.Timer = 0
					}
				} else {
					leg.VehicleID = v.ID
					c.Timer = 0
				}
			}
		}
		if c.State == CivRiding && c.Journey.LegIndex < len(c.Journey.Legs) {
			leg := &c.Journey.Legs[c.Journey.LegIndex]
			if leg.VehicleID == v.ID && leg.ToStopID == currentStopID {
				leg.Complete = true
				c.Patience = 30 + int32(rand.Intn(60))
			}
		}
	})
}

func (tm *TransportManager) StopByID(id uint32) *TransportStop {
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
	case TransWalk:
		return "Walk"
	case TransBicycle:
		return "Bicycle"
	case TransCar:
		return "Car"
	case TransBlimp:
		return "Blimp"
	default:
		return "Unknown"
	}
}

func (tm *TransportManager) Draw(h *Heightmap) {
	if tm.Tracks != nil {
		tm.Tracks.ForEachTrack(func(t *MetroTrack, slot int32) {
			y := float32(0.2)
			if t.Elevated {
				hy1 := h.WorldHeight(t.StartX, t.StartZ) + 2
				hy2 := h.WorldHeight(t.EndX, t.EndZ) + 2
				rl.DrawLine3D(rl.NewVector3(t.StartX, hy1, t.StartZ), rl.NewVector3(t.EndX, hy2, t.EndZ), rl.NewColor(100, 200, 200, 100))
				return
			}
			rl.DrawLine3D(rl.NewVector3(t.StartX, y, t.StartZ), rl.NewVector3(t.EndX, y, t.EndZ), rl.NewColor(80, 80, 200, 100))
		})
	}
	if tm.Rails != nil {
		tm.Rails.Draw(h)
	}
	if tm.Cargo != nil {
		tm.Cargo.Draw(h)
	}
	for _, cc := range tm.CableConnections {
		hy1 := h.WorldHeight(cc.StartX, cc.StartZ) + 3
		hy2 := h.WorldHeight(cc.EndX, cc.EndZ) + 3
		rl.DrawLine3D(rl.NewVector3(cc.StartX, hy1, cc.StartZ), rl.NewVector3(cc.EndX, hy2, cc.EndZ), rl.NewColor(200, 200, 100, 120))
	}

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
	case TransWalk:
		return rl.NewColor(200, 200, 200, 200)
	case TransBicycle:
		return rl.NewColor(150, 200, 150, 200)
	case TransCar:
		return rl.NewColor(100, 200, 255, 200)
	case TransBlimp:
		return rl.NewColor(200, 180, 100, 200)
	default:
		return rl.NewColor(150, 150, 150, 200)
	}
}

func (tm *TransportManager) drawVehicle(v *TransportVehicle, h *Heightmap) {
	hy := h.WorldHeight(v.X, v.Z) + 0.6
	if int(v.TransType) < len(transportModeConfigs) {
		cfg := &transportModeConfigs[v.TransType]
		col := TransportStopColor(v.TransType)
		if v.TransType == TransMetro || v.TransType == TransFerry || v.TransType == TransShip {
			hy = 0.5
		}
		if v.TransType == TransMonorail {
			rl.DrawCube(rl.NewVector3(v.X, hy+2, v.Z), cfg.SizeX, cfg.SizeY, cfg.SizeZ, col)
			rl.DrawCube(rl.NewVector3(v.X, hy+1.7, v.Z), 0.3, 0.5, 0.3, rl.NewColor(150, 150, 150, 255))
			return
		}
		if v.TransType == TransCableCar {
			rl.DrawCube(rl.NewVector3(v.X, hy+3, v.Z), cfg.SizeX, cfg.SizeY, cfg.SizeZ, col)
			return
		}
		if v.TransType == TransAir {
			rl.DrawCube(rl.NewVector3(v.X, hy+5, v.Z), cfg.SizeX, cfg.SizeY, cfg.SizeZ, col)
			return
		}
		rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), cfg.SizeX, cfg.SizeY, cfg.SizeZ, col)
		if v.TransType == TransBus || v.TransType == TransTram || v.TransType == TransTrain {
			rl.DrawCube(rl.NewVector3(v.X, hy+cfg.SizeY*0.6, v.Z+cfg.SizeZ*0.7), cfg.SizeX*0.6, cfg.SizeY*0.3, 0.1, rl.NewColor(255, 255, 200, 255))
		}
		return
	}
	rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 2, 0.6, 1, rl.Gray)
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

func (tm *TransportManager) CoverageScore(x, z float32) float32 {
	if tm == nil {
		return 0
	}
	var score float32
	modesFound := make(map[TransportType]bool)
	for i := range tm.Stops {
		s := &tm.Stops[i]
		dx := s.X - x
		dz := s.Z - z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if dist < 200 && !modesFound[s.TransType] {
			modesFound[s.TransType] = true
			walkTime := dist / 6
			access := 1.0 - walkTime/33.0
			if access > 0 {
				score += access * 0.25
			}
		}
	}
	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (tm *TransportManager) TotalMonthlyCost() float32 {
	var total float32
	for li := range tm.Lines {
		line := &tm.Lines[li]
		netIdx := int(line.TransType)
		if netIdx < len(transportModeConfigs) {
			cfg := transportModeConfigs[line.TransType]
			opCost := cfg.OperatingCost * float32(line.VehicleCount) * (line.Budget / 100.0)
			total += opCost
		}
	}
	return total
}

func (tm *TransportManager) SetLineBudget(lineID uint32, budget float32) {
	for i := range tm.Lines {
		if tm.Lines[i].ID == lineID {
			if budget < 10 {
				budget = 10
			}
			if budget > 200 {
				budget = 200
			}
			tm.Lines[i].Budget = budget
			return
		}
	}
}

func (tm *TransportManager) NearestStop(x, z float32, maxDist float32) *TransportStop {
	best := float32(maxDist)
	var found *TransportStop
	for i := range tm.Stops {
		s := &tm.Stops[i]
		dx := s.X - x
		dz := s.Z - z
		d := dx*dx + dz*dz
		if d < best {
			best = d
			found = s
		}
	}
	return found
}

func (tm *TransportManager) NearestStopOfType(x, z float32, tt TransportType, maxDist float32) *TransportStop {
	best := float32(maxDist)
	var found *TransportStop
	for i := range tm.Stops {
		s := &tm.Stops[i]
		if s.TransType != tt {
			continue
		}
		dx := s.X - x
		dz := s.Z - z
		d := dx*dx + dz*dz
		if d < best {
			best = d
			found = s
		}
	}
	return found
}

func (tm *TransportManager) HasLineBetween(tt TransportType, stopA, stopB uint32) bool {
	for li := range tm.Lines {
		line := &tm.Lines[li]
		if line.TransType != tt || !line.Active {
			continue
		}
		hasA, hasB := false, false
		for _, sid := range line.Stops {
			if sid == stopA {
				hasA = true
			}
			if sid == stopB {
				hasB = true
			}
		}
		if hasA && hasB {
			return true
		}
	}
	return false
}

func (tm *TransportManager) NearestLine(x, z float32, maxDist float32) *TransportLine {
	best := float32(maxDist)
	var found *TransportLine
	for li := range tm.Lines {
		line := &tm.Lines[li]
		for _, sid := range line.Stops {
			s := 	tm.StopByID(sid)
			if s == nil {
				continue
			}
			dx := s.X - x
			dz := s.Z - z
			d := dx*dx + dz*dz
			if d < best {
				best = d
				found = line
			}
		}
	}
	return found
}
