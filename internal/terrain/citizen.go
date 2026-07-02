package terrain

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const CitizenPoolSize = 500

type CitizenState uint8

const (
	CivHome CitizenState = iota
	CivWalkingToStop
	CivWaiting
	CivRiding
	CivWalkingToDest
	CivAtDest
)

type LegType uint8

const (
	LegWalk LegType = iota
	LegBus
	LegTram
	LegMetro
	LegTrain
	LegFerry
	LegMonorail
	LegCableCar
	LegTaxi
	LegShip
	LegAir
)

type JourneyLeg struct {
	LegType   LegType
	Mode      TransportType
	FromStopID uint32
	ToStopID   uint32
	FromX, FromZ float32
	ToX, ToZ     float32
	VehicleID    uint32
	Distance     float32
	Progress     float32
	WaitTimer    int32
	Complete     bool
}

type Journey struct {
	Legs     []JourneyLeg
	LegIndex int
	Active   bool
}

type Citizen struct {
	ID         uint32
	HomeX, HomeZ float32
	WorkX, WorkZ float32
	State      CitizenState
	Journey    Journey
	X, Z       float32
	TargetX, TargetZ float32
	Speed      float32
	Wealth     int32
	Patience   int32
	Timer      int32
	Lifecycle  LifecycleState
}

const routeCacheSize = 64

type routeCacheKey struct {
	fromX, fromZ, toX, toZ int16
}

type routeCacheEntry struct {
	key   routeCacheKey
	legs  [][]JourneyLeg
	tick  int32
}

type CitizenManager struct {
	Pool     [CitizenPoolSize]Citizen
	FreeList []int32
	Count    int32
	NextID   uint32
	Timer    int32

	routeCache     [routeCacheSize]routeCacheEntry
	routeCacheNext int
	cacheTick      int32
}

func NewCitizenManager() *CitizenManager {
	cm := &CitizenManager{
		FreeList: make([]int32, CitizenPoolSize),
	}
	for i := 0; i < CitizenPoolSize; i++ {
		cm.Pool[i].Lifecycle = LifecycleUnallocated
		cm.FreeList[i] = int32(CitizenPoolSize - 1 - i)
	}
	return cm
}

func (cm *CitizenManager) alloc() int32 {
	if len(cm.FreeList) == 0 {
		return -1
	}
	idx := cm.FreeList[len(cm.FreeList)-1]
	cm.FreeList = cm.FreeList[:len(cm.FreeList)-1]
	cm.Pool[idx] = Citizen{}
	cm.Pool[idx].Lifecycle = LifecycleAllocated
	cm.Count++
	return int32(idx)
}

func (cm *CitizenManager) free(slot int32) {
	if slot < 0 || int(slot) >= CitizenPoolSize {
		return
	}
	cm.Pool[slot].Lifecycle = LifecycleReturnedToPool
	cm.FreeList = append(cm.FreeList, slot)
	cm.Count--
}

func (cm *CitizenManager) ForEach(fn func(*Citizen, int32)) {
	for i := 0; i < CitizenPoolSize; i++ {
		if cm.Pool[i].Lifecycle == LifecycleAllocated {
			fn(&cm.Pool[i], int32(i))
		}
	}
}

func (cm *CitizenManager) SpawnCitizen(homeX, homeZ, workX, workZ float32) {
	slot := cm.alloc()
	if slot < 0 {
		return
	}
	c := &cm.Pool[slot]
	c.ID = cm.NextID
	cm.NextID++
	c.HomeX = homeX
	c.HomeZ = homeZ
	c.WorkX = workX
	c.WorkZ = workZ
	c.X = homeX
	c.Z = homeZ
	c.State = CivHome
	c.Speed = 6
	c.Wealth = int32(rand.Intn(100))
	c.Patience = 30 + int32(rand.Intn(60))
	c.Lifecycle = LifecycleAllocated
}

func (cm *CitizenManager) StartJourney(c *Citizen, tm *TransportManager) {
	legs := cm.evaluateRoute(c, tm)
	if len(legs) == 0 {
		c.Journey = Journey{Active: false}
		c.State = CivAtDest
		c.Timer = 120
		return
	}
	c.Journey = Journey{Legs: legs, LegIndex: 0, Active: true}
	c.State = CivWalkingToStop
	c.TargetX = legs[0].FromX
	c.TargetZ = legs[0].FromZ
}

func (cm *CitizenManager) evaluateRoute(c *Citizen, tm *TransportManager) []JourneyLeg {
	if tm == nil {
		return nil
	}

	candidates := cm.findRoutes(c.HomeX, c.HomeZ, c.WorkX, c.WorkZ, tm)
	if len(candidates) == 0 {
		walkLeg := JourneyLeg{
			LegType:  LegWalk,
			FromX:    c.HomeX, FromZ: c.HomeZ,
			ToX:      c.WorkX, ToZ: c.WorkZ,
			Distance: dist(c.HomeX, c.HomeZ, c.WorkX, c.WorkZ),
		}
		return []JourneyLeg{walkLeg}
	}

	bestLegs := candidates[0]
	bestCost := generalizedCost(candidates[0], c.Wealth, tm)
	for i := 1; i < len(candidates); i++ {
		cost := generalizedCost(candidates[i], c.Wealth, tm)
		if cost < bestCost {
			bestCost = cost
			bestLegs = candidates[i]
		}
	}
	return bestLegs
}

const walkSpeed float32 = 6
const valueOfTime float32 = 0.5
const transferFixedCost float32 = 120
const waitCostPerMinute float32 = 20

func generalizedCost(legs []JourneyLeg, wealth int32, tm *TransportManager) float32 {
	var walkTime, vehicleTime, waitTime float32
	transferCount := 0

	for _, leg := range legs {
		switch leg.LegType {
		case LegWalk:
			walkTime += leg.Distance / walkSpeed
		default:
			speed := float32(30)
			if tm != nil && int(leg.Mode) < len(transportModeConfigs) {
				speed = transportModeConfigs[leg.Mode].Speed
			}
			vehicleTime += leg.Distance / speed
			waitTime += 30 / speed
			transferCount++
		}
	}

	if transferCount > 0 {
		transferCount--
	}

	timeCost := (walkTime + vehicleTime) * valueOfTime * 60
	waitCost := waitTime * waitCostPerMinute
	transferCost := float32(transferCount) * transferFixedCost
	moneyCost := float32(transferCount) * 2.0 * (1.0 - float32(wealth)/150.0)

	congestionPenalty := float32(0)
	if tm != nil {
		for _, leg := range legs {
			if leg.LegType != LegWalk && (leg.Mode == TransBus || leg.Mode == TransTaxi) {
				congestionPenalty += tm.RoadCongestion * 30
			}
		}
	}

	noise := float32(rand.Intn(20)) - 10
	return timeCost + waitCost + transferCost + moneyCost + congestionPenalty + noise
}

func (cm *CitizenManager) findRoutes(fromX, fromZ, toX, toZ float32, tm *TransportManager) [][]JourneyLeg {
	key := routeCacheKey{
		fromX: int16(fromX), fromZ: int16(fromZ),
		toX: int16(toX), toZ: int16(toZ),
	}
	for i := range cm.routeCache {
		e := &cm.routeCache[i]
		if e.key == key && cm.cacheTick-e.tick < 120 {
			return e.legs
		}
	}

	routes := cm.buildRoutes(fromX, fromZ, toX, toZ, tm)

	entry := &cm.routeCache[cm.routeCacheNext]
	entry.key = key
	entry.legs = routes
	entry.tick = cm.cacheTick
	cm.routeCacheNext = (cm.routeCacheNext + 1) % routeCacheSize

	return routes
}

func (cm *CitizenManager) buildRoutes(fromX, fromZ, toX, toZ float32, tm *TransportManager) [][]JourneyLeg {
	var routes [][]JourneyLeg

	walkOnly := JourneyLeg{
		LegType:  LegWalk,
		FromX:    fromX, FromZ: fromZ,
		ToX:      toX, ToZ: toZ,
		Distance: dist(fromX, fromZ, toX, toZ),
	}
	walkOnly.Distance = dist(fromX, fromZ, toX, toZ)
	routes = append(routes, []JourneyLeg{walkOnly})

	modePriority := []TransportType{TransBus, TransTram, TransMetro, TransTrain, TransMonorail, TransFerry, TransCableCar, TransTaxi}
	for _, mode := range modePriority {
		fromStop := tm.NearestStopOfType(fromX, fromZ, mode, 200)
		toStop := tm.NearestStopOfType(toX, toZ, mode, 200)
		if fromStop == nil || toStop == nil || fromStop.ID == toStop.ID {
			continue
		}
		if !tm.HasLineBetween(mode, fromStop.ID, toStop.ID) {
			continue
		}

		legs := []JourneyLeg{
			{
				LegType:   LegWalk,
				FromX:     fromX, FromZ: fromZ,
				ToX:       fromStop.X, ToZ: fromStop.Z,
				FromStopID: fromStop.ID,
				Distance:   dist(fromX, fromZ, fromStop.X, fromStop.Z),
			},
			{
				LegType:   legTypeFromTransport(mode),
				Mode:      mode,
				FromStopID: fromStop.ID,
				ToStopID:   toStop.ID,
				FromX:      fromStop.X, FromZ: fromStop.Z,
				ToX:        toStop.X, ToZ: toStop.Z,
				Distance:   dist(fromStop.X, fromStop.Z, toStop.X, toStop.Z),
			},
			{
				LegType:   LegWalk,
				FromX:     toStop.X, FromZ: toStop.Z,
				ToX:       toX, ToZ: toZ,
				Distance:  dist(toStop.X, toStop.Z, toX, toZ),
			},
		}
		routes = append(routes, legs)
	}

	for _, modeA := range modePriority {
		for _, modeB := range modePriority {
			if modeA == modeB {
				continue
			}
			fromStop := tm.NearestStopOfType(fromX, fromZ, modeA, 200)
			toStop := tm.NearestStopOfType(toX, toZ, modeB, 200)
			if fromStop == nil || toStop == nil || fromStop.ID == toStop.ID {
				continue
			}

			transferMid := cm.findTransferStop(fromStop, toStop, modeB, tm)
			if transferMid == nil {
				continue
			}

			midStopA := tm.StopByID(transferMid.StopA)
			midStopB := tm.StopByID(transferMid.StopB)
			if midStopA == nil || midStopB == nil {
				continue
			}
			if !tm.HasLineBetween(modeA, fromStop.ID, midStopA.ID) {
				continue
			}
			if !tm.HasLineBetween(modeB, midStopB.ID, toStop.ID) {
				continue
			}

			transferDist := dist(midStopA.X, midStopA.Z, midStopB.X, midStopB.Z)
			if transferDist > 100 {
				continue
			}

			legs := []JourneyLeg{
				{LegType: LegWalk, FromX: fromX, FromZ: fromZ, ToX: fromStop.X, ToZ: fromStop.Z, FromStopID: fromStop.ID, Distance: dist(fromX, fromZ, fromStop.X, fromStop.Z)},
				{LegType: legTypeFromTransport(modeA), Mode: modeA, FromStopID: fromStop.ID, ToStopID: midStopA.ID, FromX: fromStop.X, FromZ: fromStop.Z, ToX: midStopA.X, ToZ: midStopA.Z, Distance: dist(fromStop.X, fromStop.Z, midStopA.X, midStopA.Z)},
				{LegType: LegWalk, FromX: midStopA.X, FromZ: midStopA.Z, ToX: midStopB.X, ToZ: midStopB.Z, Distance: transferDist},
				{LegType: legTypeFromTransport(modeB), Mode: modeB, FromStopID: midStopB.ID, ToStopID: toStop.ID, FromX: midStopB.X, FromZ: midStopB.Z, ToX: toStop.X, ToZ: toStop.Z, Distance: dist(midStopB.X, midStopB.Z, toStop.X, toStop.Z)},
				{LegType: LegWalk, FromX: toStop.X, FromZ: toStop.Z, ToX: toX, ToZ: toZ, Distance: dist(toStop.X, toStop.Z, toX, toZ)},
			}
			routes = append(routes, legs)
		}
	}

	return routes
}

type transferMatch struct {
	StopA, StopB uint32
}

func (cm *CitizenManager) findTransferStop(fromStop, toStop *TransportStop, modeB TransportType, tm *TransportManager) *transferMatch {
	if fromStop.TransferStationID != math.MaxUint32 {
		for _, ts := range tm.TransferStations {
			if ts.ID != fromStop.TransferStationID {
				continue
			}
			for _, sid := range ts.StopIDs {
				s := tm.StopByID(sid)
				if s != nil && s.TransType == modeB {
					return &transferMatch{StopA: fromStop.ID, StopB: sid}
				}
			}
		}
	}

	midA := tm.NearestStopOfType(fromStop.X, fromStop.Z, modeB, 100)
	if midA != nil {
		return &transferMatch{StopA: fromStop.ID, StopB: midA.ID}
	}
	return nil
}

func scoreRoute(legs []JourneyLeg, wealth int32, patience int32) float32 {
	var total float32
	for _, leg := range legs {
		switch leg.LegType {
		case LegWalk:
			total += leg.Distance * 0.1
		default:
			total += leg.Distance * 0.02
			total += 5
		}
	}
	transferPenalty := float32(len(legs)-1) * 10
	waitPenalty := float32(len(legs)) * 3
	costFactor := 1.0 - float32(wealth)/200.0
	total = (total + transferPenalty + waitPenalty) * (1.0 + costFactor*0.5)
	return total
}

func legTypeFromTransport(tt TransportType) LegType {
	switch tt {
	case TransBus:
		return LegBus
	case TransTram:
		return LegTram
	case TransMetro:
		return LegMetro
	case TransTrain:
		return LegTrain
	case TransFerry:
		return LegFerry
	case TransMonorail:
		return LegMonorail
	case TransCableCar:
		return LegCableCar
	case TransTaxi:
		return LegTaxi
	case TransShip:
		return LegShip
	case TransAir:
		return LegAir
	default:
		return LegWalk
	}
}

func (cm *CitizenManager) Update(tm *TransportManager, h *Heightmap, buildings *BuildingManager) {
	cm.Timer++

	if cm.Timer%120 == 0 && buildings != nil {
		buildings.ForEach(func(b *Building, slot int32) {
			if b.Residents > 0 && cm.Count < CitizenPoolSize/2 {
				cm.SpawnCitizen(b.Position.X, b.Position.Z, b.Position.X+float32(rand.Intn(200)-100), b.Position.Z+float32(rand.Intn(200)-100))
			}
		})
	}

	cm.ForEach(func(c *Citizen, slot int32) {
		if c.State == CivHome {
			c.Timer++
			if c.Timer > 300+int32(rand.Intn(300)) {
				cm.StartJourney(c, tm)
				c.Timer = 0
			}
			return
		}

		if c.State == CivAtDest {
			c.Timer++
			if c.Timer > 180+int32(rand.Intn(180)) {
				c.State = CivHome
				c.X = c.HomeX
				c.Z = c.HomeZ
				c.Timer = 0
				c.Journey = Journey{}
			}
			return
		}

		if c.State == CivWalkingToStop || c.State == CivWalkingToDest {
			dx := c.TargetX - c.X
			dz := c.TargetZ - c.Z
			distRem := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if distRem < 1 {
				c.X = c.TargetX
				c.Z = c.TargetZ
				if c.State == CivWalkingToStop {
					c.State = CivWaiting
					c.Timer = 0
				} else {
					c.State = CivAtDest
					c.Timer = 0
				}
				return
			}
			step := c.Speed * 0.02
			if step > distRem {
				step = distRem
			}
			c.X += (dx / distRem) * step
			c.Z += (dz / distRem) * step
			return
		}

		if c.State == CivWaiting {
			c.Timer++
			if c.Timer > c.Patience*2 {
				cm.StartJourney(c, tm)
				c.Timer = 0
				return
			}
			leg := &c.Journey.Legs[c.Journey.LegIndex]
			if leg.VehicleID != 0 {
				c.State = CivRiding
				c.Timer = 0
			}
			return
		}

		if c.State == CivRiding {
			leg := &c.Journey.Legs[c.Journey.LegIndex]
			if leg.Complete {
				c.Journey.LegIndex++
				if c.Journey.LegIndex >= len(c.Journey.Legs) {
					c.State = CivAtDest
					c.Timer = 0
				} else {
					nextLeg := c.Journey.Legs[c.Journey.LegIndex]
					c.State = CivWalkingToDest
					c.TargetX = nextLeg.ToX
					c.TargetZ = nextLeg.ToZ
				}
			}
			return
		}
	})
	cm.cacheTick++
}

func (cm *CitizenManager) Draw(h *Heightmap) {
	cm.ForEach(func(c *Citizen, slot int32) {
		if c.State == CivHome || c.State == CivAtDest {
			return
		}
		if c.State == CivRiding {
			return
		}
		hy := h.WorldHeight(c.X, c.Z) + 0.3
		rl.DrawCube(rl.NewVector3(c.X, hy, c.Z), 0.4, 0.8, 0.4, rl.NewColor(60, 60, 200, 220))
	})
}

func dist(x1, z1, x2, z2 float32) float32 {
	dx := x2 - x1
	dz := z2 - z1
	return float32(math.Sqrt(float64(dx*dx + dz*dz)))
}
