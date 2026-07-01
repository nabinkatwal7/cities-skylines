package terrain

import "math"

type SimEvent string

const (
	EventRoadPlaced         SimEvent = "road:placed"
	EventRoadRemoved        SimEvent = "road:removed"
	EventRoadUpgraded       SimEvent = "road:upgraded"
	EventZonePlaced         SimEvent = "zone:placed"
	EventZoneCleared        SimEvent = "zone:cleared"
	EventParkPlaced         SimEvent = "park:placed"
	EventBuildingUpgraded   SimEvent = "building:upgraded"
	EventBuildingAbandoned  SimEvent = "building:abandoned"
	EventBuildingDemolished SimEvent = "building:demolished"
	EventBuildingConstructed SimEvent = "building:constructed"
	EventDayNightCycle      SimEvent = "time:daynight"
	EventTaxCollected       SimEvent = "economy:tax"
	EventTimeMinute          SimEvent = "time:minute"
	EventTimeHour            SimEvent = "time:hour"
	EventTimeDay             SimEvent = "time:day"
	EventFloodStarted        SimEvent = "flood:started"
	EventFloodReceded        SimEvent = "flood:receded"
	EventParkingLotPlaced    SimEvent = "parkinglot:placed"
	EventParkingGaragePlaced SimEvent = "parkinggarage:placed"
	EventParkingLotRemoved   SimEvent = "parkinglot:removed"
)

type SimulationManager struct {
	Generator   *Generator
	Heightmap   *Heightmap
	Water       *WaterSystem
	Terraform   *TerraformSystem
	Trees       *TreeSystem
	Resources   *ResourceSystem
	Roads       *RoadManager
	Zones       *ZoneManager
	Buildings   *BuildingManager
	Services    *ServiceManager
	Connections *ConnectionSystem
	Vehicles    *VehicleManager
	Transport   *TransportManager
	Districts   *DistrictManager
	Buildability *BuildabilityChecker
	Parking     *ParkingManager

	Seed      int64
	Night     bool
	Money     float32
	TaxTimer  int32

	EventBus  *EventBus
	scheduler *Scheduler
	Time      *GameTime
	Jobs      *JobQueue
}

func NewSimulationManager(seed int64) *SimulationManager {
	gen := NewGenerator(seed)
	h := gen.Generate()
	water := NewWaterSystem()
	water.Init(h)
	trees := NewTreeSystem(seed)
	res := NewResourceSystem(seed, h)
	trees.Generate(h, water)
	roads := NewRoadManager()
	roads.LoadAssets()
	conn := NewConnectionSystem()
	roads.InitOutsideConnections(conn)
	zones := NewZoneManager(128, 128)
	buildings := NewBuildingManager()
	services := NewServiceManager()
	vehicles := NewVehicleManager()
	parking := NewParkingManager()

	sm := &SimulationManager{
		Generator:   gen,
		Heightmap:   h,
		Water:       water,
		Trees:       trees,
		Resources:   res,
		Roads:       roads,
		Zones:       zones,
		Buildings:   buildings,
		Services:    services,
		Connections: conn,
		Vehicles:    vehicles,
		Parking:     parking,
		Transport:   NewTransportManager(),
		Districts:   NewDistrictManager(),
		Seed:        seed,
		EventBus:    NewEventBus(),
		scheduler:   NewScheduler(),
		Time:        NewGameTime(),
		Jobs:        NewJobQueue(),
	}

	sm.initScheduler()
	sm.initEventListeners()
	sm.Water.SetEventBus(sm.EventBus)
	SetWaterForBuildings(sm.Water)
	SetRoadsForBuildings(sm.Roads)
	SetWaterForRoads(sm.Water)
	sm.Buildability = NewBuildabilityChecker(sm.Heightmap, sm.Water, sm.Trees, sm.Buildings, sm.Roads, sm.Zones, sm.Resources)
	SetBuildabilityChecker(sm.Buildability)
	SetZoneBuildabilityCheck(sm.Buildability)
	SetLandValueTrees(sm.Trees)
	SetResourceForBuildings(sm.Resources)
	sm.Trees.SetResources(sm.Resources)
	sm.Resources.SetTrees(sm.Trees)
	sm.Parking.GenerateRoadsideSpots(sm.Roads)
	return sm
}

func (sm *SimulationManager) initScheduler() {
	sm.scheduler.Register(GroupFast, UpdateTask{
		Name:     "water",
		Priority: SchedPriorityHigh,
		BudgetMs: 1,
		Callback: func(dt float64) { sm.Water.Update(dt) },
	})
	sm.scheduler.Register(GroupFast, UpdateTask{
		Name:     "trees",
		Priority: SchedPriorityLow,
		BudgetMs: 0.5,
		Callback: func(dt float64) { sm.Trees.Update(dt) },
	})
	sm.scheduler.Register(GroupMedium, UpdateTask{
		Name:     "roads",
		Priority: SchedPriorityCritical,
		BudgetMs: 2,
		Callback: func(dt float64) { sm.Roads.Update(sm.Heightmap) },
	})
	sm.scheduler.Register(GroupMedium, UpdateTask{
		Name:     "vehicles",
		Priority: SchedPriorityHigh,
		BudgetMs: 3,
		Callback: func(dt float64) { sm.Vehicles.Update(sm.Roads, sm.Heightmap, sm.Parking) },
	})
	sm.scheduler.Register(GroupMedium, UpdateTask{
		Name:     "transport",
		Priority: SchedPriorityHigh,
		BudgetMs: 2,
		Callback: func(dt float64) { sm.Transport.Update(sm.Roads, sm.Heightmap) },
	})
	sm.scheduler.Register(GroupFast, UpdateTask{
		Name:     "parking",
		Priority: SchedPriorityLow,
		BudgetMs: 0.5,
		Callback: func(dt float64) { sm.Parking.Timer++ },
	})
	sm.scheduler.Register(GroupSlow, UpdateTask{
		Name:     "buildings",
		Priority: SchedPriorityHigh,
		BudgetMs: 1,
		Callback: func(dt float64) { sm.Buildings.Update(sm.Zones, sm.Heightmap, sm.Roads, sm.Districts) },
	})
	sm.scheduler.Register(GroupVerySlow, UpdateTask{
		Name:     "tax",
		Priority: SchedPriorityMedium,
		BudgetMs: 1,
		Callback: func(dt float64) { sm.collectTax(dt) },
	})

}

func (sm *SimulationManager) initEventListeners() {
	sm.EventBus.On(string(EventRoadRemoved), func(data any) {
		idx, ok := data.(int)
		if ok && sm.Vehicles != nil {
			sm.Vehicles.OnRoadRemoved(idx)
			sm.Parking.GenerateRoadsideSpots(sm.Roads)
		}
	})
	sm.EventBus.On(string(EventRoadPlaced), func(data any) {
		if sm.Parking != nil {
			sm.Parking.GenerateRoadsideSpots(sm.Roads)
		}
	})
	sm.EventBus.On(string(EventRoadUpgraded), func(data any) {
		if sm.Parking != nil {
			sm.Parking.GenerateRoadsideSpots(sm.Roads)
		}
	})
	sm.EventBus.On(string(EventBuildingDemolished), func(data any) {
		if sm.Zones == nil {
			return
		}
		if cellX, ok := data.(int); ok {
			_ = cellX
		}
	})
	sm.EventBus.On(string(EventDayNightCycle), func(data any) {
		isNight, ok := data.(bool)
		if ok && sm.Roads != nil {
			sm.Roads.SetNightMode(isNight)
		}
	})
	sm.EventBus.On(string(EventTimeHour), func(data any) {
		if sm.Buildings != nil {
			sm.Buildings.FlushStats()
		}
	})
	sm.EventBus.On(string(EventZoneCleared), func(data any) {
		if sm.Zones == nil {
			return
		}
	})
}

func (sm *SimulationManager) InitDefaultRoads() {
	n0 := sm.Roads.AddNode(-50, 0, -50)
	n1 := sm.Roads.AddNode(-50, 0, 50)
	n2 := sm.Roads.AddNode(50, 0, 50)
	n3 := sm.Roads.AddNode(50, 0, -50)
	n4 := sm.Roads.AddNode(-80, 0, 0)
	n5 := sm.Roads.AddNode(80, 0, 0)
	n6 := sm.Roads.AddNode(0, 0, -80)
	n7 := sm.Roads.AddNode(0, 0, 80)
	sm.Roads.AddSegment(n0, n1, RoadTwoLane)
	sm.Roads.AddSegment(n1, n2, RoadTwoLane)
	sm.Roads.AddSegment(n2, n3, RoadTwoLane)
	sm.Roads.AddSegment(n3, n0, RoadTwoLane)
	sm.Roads.AddSegment(n4, n0, RoadFourLane)
	sm.Roads.AddSegment(n5, n2, RoadFourLane)
	sm.Roads.AddSegment(n6, n3, RoadTwoLane)
	sm.Roads.AddSegment(n7, n1, RoadTwoLane)
}

func (sm *SimulationManager) Update(dt float64) {
	if sm.Time.IsPaused {
		sm.EventBus.ProcessQueue()
		return
	}

	simDt := dt * sm.Time.Speed
	sm.Time.Tick(simDt)
	sm.Jobs.Process()
	sm.scheduler.RunAll(simDt)

	if sm.Time.MinuteChanged() {
		sm.EventBus.Emit(string(EventTimeMinute), sm.Time.Minute)
	}
	if sm.Time.HourChanged() {
		sm.EventBus.Emit(string(EventTimeHour), sm.Time.Hour)
	}
	if sm.Time.DayChanged() {
		sm.EventBus.Emit(string(EventTimeDay), sm.Time.DayCount)
	}
	sm.syncDayNight()
	sm.Time.Snapshot()

	sm.EventBus.ProcessQueue()
}

func (sm *SimulationManager) syncDayNight() {
	isDay := sm.Time.IsDaytime()
	if sm.Night == isDay {
		sm.Night = !isDay
		sm.EventBus.Emit(string(EventDayNightCycle), sm.Night)
	}
}

func (sm *SimulationManager) SetSpeed(speed float64) {
	if speed < 0 {
		speed = 0
	}
	if speed > 3 {
		speed = 3
	}
	sm.Time.Speed = speed
	sm.Time.IsPaused = speed == 0
}

func (sm *SimulationManager) TogglePause() {
	sm.Time.IsPaused = !sm.Time.IsPaused
}

func (sm *SimulationManager) IsPaused() bool {
	return sm.Time.IsPaused
}

func (sm *SimulationManager) Speed() float64 {
	return sm.Time.Speed
}

func (sm *SimulationManager) collectTax(dt float64) {
	sm.TaxTimer++
	if sm.TaxTimer > 60 {
		pop := sm.Buildings.Population()
		tax := float32(sm.Buildings.Count) * 0.5
		if pop > 0 {
			tax += float32(pop) * 0.1
		}
		sm.Money += tax

		if sm.Roads != nil {
			maint := sm.Roads.TotalMaintenance()
			sm.Money -= maint
		}
		if sm.Transport != nil {
			sm.Money -= sm.Transport.TotalMaintenance()
			sm.Money += sm.Transport.TotalIncome() * 0.1
		}

		sm.EventBus.Emit(string(EventTaxCollected), tax)
		sm.TaxTimer = 0
	}
}

const (
	MaxRoadSlope      = 0.25
	MinCurveRadius    = 8.0
	RoadProximityDist = 4.0
	BuildingProximity = 3.0
)

func (sm *SimulationManager) CanPlaceRoad(x1, z1, x2, z2 float32, rt RoadType, elevation int32, excludeSegID uint32) string {
	half := float32(WorldSize / 2)
	if x1 < -half || x1 > half || z1 < -half || z1 > half {
		return "outside map boundary"
	}
	if x2 < -half || x2 > half || z2 < -half || z2 > half {
		return "outside map boundary"
	}

	dx := x2 - x1
	dz := z2 - z1
	length := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	if length < 2.0 {
		return "segment too short"
	}

	if elevation >= 0 {
		h1 := sm.Heightmap.WorldHeight(x1, z1)
		h2 := sm.Heightmap.WorldHeight(x2, z2)
		slope := float32(math.Abs(float64(h2-h1))) / length
		if slope > MaxRoadSlope {
			return "terrain slope too steep"
		}

		if sm.Heightmap.IsUnderwater(x1, z1) || sm.Heightmap.IsUnderwater(x2, z2) {
			return "endpoint underwater"
		}

		steps := int(length / 4)
		if steps < 4 {
			steps = 4
		}
		for si := 1; si < steps; si++ {
			t := float32(si) / float32(steps)
			px := x1 + dx*t
			pz := z1 + dz*t
			if sm.Heightmap.IsUnderwater(px, pz) {
				return "segment crosses water"
			}
		}
	}

	for _, seg := range sm.Roads.Segments {
		if seg.ID == excludeSegID {
			continue
		}
		xs, zs, _ := sm.Roads.SampleSegment(seg, 8)
		for i := 0; i < len(xs); i++ {
			sdx := x1 - xs[i]
			sdz := z1 - zs[i]
			if sdx*sdx+sdz*sdz < RoadProximityDist*RoadProximityDist {
				return "too close to existing road"
			}
			sdx = x2 - xs[i]
			sdz = z2 - zs[i]
			if sdx*sdx+sdz*sdz < RoadProximityDist*RoadProximityDist {
				return "too close to existing road"
			}
		}
	}

	if sm.Buildings != nil {
		for i := 0; i < BuildingPoolSize; i++ {
			b := &sm.Buildings.Pool[i]
			if b.Lifecycle != LifecycleActive {
				continue
			}
			hw := b.Width * 0.5
			hd := b.Depth * 0.5
			if lineIntersectsRect(x1, z1, x2, z2, b.Position.X-hw, b.Position.Z-hd, b.Position.X+hw, b.Position.Z+hd) {
				return "intersects building"
			}
		}
	}

	return ""
}

func lineIntersectsRect(x1, z1, x2, z2, rx0, rz0, rx1, rz1 float32) bool {
	minX := float32(math.Min(float64(x1), float64(x2)))
	maxX := float32(math.Max(float64(x1), float64(x2)))
	minZ := float32(math.Min(float64(z1), float64(z2)))
	maxZ := float32(math.Max(float64(z1), float64(z2)))
	return !(maxX < rx0 || minX > rx1 || maxZ < rz0 || minZ > rz1)
}

func (sm *SimulationManager) PlaceRoadNode(x, z float32) uint32 {
	if sm.Connections != nil {
		for _, c := range sm.Connections.GetByType(ConnHighway) {
			dx := c.WorldX - x
			dz := c.WorldZ - z
			if dx*dx+dz*dz < 64 {
				for idx := range sm.Roads.Nodes {
					n := &sm.Roads.Nodes[idx]
					if n.Flags&RoadFlagOutsideConn != 0 {
						nx := n.X - c.WorldX
						nz := n.Z - c.WorldZ
						if nx*nx+nz*nz < 0.01 {
							return uint32(idx)
						}
					}
				}
			}
		}
	}
	return sm.Roads.AddNode(x, 0, z)
}

func (sm *SimulationManager) PlaceRoadSegment(nodeA uint32, x, z float32, roadType RoadType, elevation int32) (uint32, uint32, bool) {
	na := sm.Roads.Nodes[nodeA]
	snapX, snapZ := x, z
	if sm.Connections != nil {
		for _, c := range sm.Connections.GetByType(ConnHighway) {
			dx := c.WorldX - x
			dz := c.WorldZ - z
			if dx*dx+dz*dz < 64 {
				for idx := range sm.Roads.Nodes {
					n := &sm.Roads.Nodes[idx]
					if n.Flags&RoadFlagOutsideConn != 0 {
						nx := n.X - c.WorldX
						nz := n.Z - c.WorldZ
						if nx*nx+nz*nz < 0.01 {
							snapX = n.X
							snapZ = n.Z
							nodeB := uint32(idx)
							segID := sm.Roads.AddSegment(nodeA, nodeB, roadType)
							sm.finalizeSegment(segID, nodeA, nodeB, roadType, elevation)
							return nodeB, segID, true
						}
					}
				}
			}
		}
	}
	reason := sm.CanPlaceRoad(na.X, na.Z, snapX, snapZ, roadType, elevation, math.MaxUint32)
	if reason != "" {
		return math.MaxUint32, math.MaxUint32, false
	}
	nodeB := sm.Roads.AddNode(snapX, 0, snapZ)
	segID := sm.Roads.AddSegment(nodeA, nodeB, roadType)
	sm.finalizeSegment(segID, nodeA, nodeB, roadType, elevation)
	return nodeB, segID, true
}

func (sm *SimulationManager) finalizeSegment(segID, nodeA, nodeB uint32, roadType RoadType, elevation int32) {
	if elevation != 0 {
		for i := range sm.Roads.Segments {
			if sm.Roads.Segments[i].ID == segID {
				sm.Roads.Segments[i].Elevation = elevation
				sm.Roads.Segments[i].MaintenanceCost = sm.Roads.calcSegmentMaintenance(roadType, sm.Roads.Segments[i].Length, elevation)
				if elevation > 0 {
					sm.Roads.Nodes[nodeA].Flags |= RoadFlagBridge
					sm.Roads.Nodes[nodeB].Flags |= RoadFlagBridge
				} else if elevation < 0 {
					sm.Roads.Nodes[nodeA].Flags |= RoadFlagTunnel
					sm.Roads.Nodes[nodeB].Flags |= RoadFlagTunnel
				}
				break
			}
		}
	}
	cost := roadConstructionCost(roadType)
	if elevation > 0 {
		cost += float32(elevation) * 50
	} else if elevation < 0 {
		cost += float32(-elevation) * 100
	}
	sm.Roads.Rebuild(sm.Heightmap)
	sm.Money -= cost
	sm.EventBus.Emit(string(EventRoadPlaced), segID)
}

func (sm *SimulationManager) RemoveSegment(idx int) {
	sm.Roads.RemoveSegment(idx)
	sm.Roads.Rebuild(sm.Heightmap)
	sm.EventBus.Emit(string(EventRoadRemoved), idx)
}

func (sm *SimulationManager) RemoveTrees(worldX, worldZ, radius float32) int {
	cost := float32(10)
	if sm.Money < cost {
		return 0
	}
	removed := sm.Trees.RemoveNear(worldX, worldZ, radius)
	if removed > 0 {
		sm.Money -= cost * float32(removed)
	}
	return removed
}

func (sm *SimulationManager) UpgradeSegment(idx int, newType RoadType) bool {
	if idx < 0 || idx >= len(sm.Roads.Segments) {
		return false
	}
	old := sm.Roads.Segments[idx]
	if old.RoadType == newType {
		if old.Damaged {
			cost := sm.Roads.Segments[idx].MaintenanceCost * 5
			if sm.Money < cost {
				return false
			}
			if sm.Roads.RepairSegment(idx) {
				sm.Money -= cost
				return true
			}
		}
		return true
	}
	oldCost := roadConstructionCost(old.RoadType)
	newCost := roadConstructionCost(newType)
	diff := newCost - oldCost
	if diff > 0 && sm.Money < diff {
		return false
	}
	sm.Roads.UpgradeSegment(idx, newType)
	if diff > 0 {
		sm.Money -= diff
	} else if diff < 0 {
		sm.Money -= diff // diff is negative, so this adds money (refund)
	}
	sm.Roads.Rebuild(sm.Heightmap)
	sm.EventBus.Emit(string(EventRoadUpgraded), idx)
	return true
}

func (sm *SimulationManager) SetZone(worldX, worldZ float32, zt ZoneType) {
	sm.Zones.SetZone(worldX, worldZ, zt, sm.Roads)
	sm.Money -= 20
	sm.EventBus.Emit(string(EventZonePlaced), zt)
}

func (sm *SimulationManager) PlacePark(x, z float32) {
	sm.Services.AddPark(x, z)
	sm.Money -= 500
	sm.EventBus.Emit(string(EventParkPlaced), nil)
}

func (sm *SimulationManager) PlaceParkingLot(x, z float32, isGarage bool) bool {
	if sm.Money < 1000 {
		return false
	}
	ok := sm.Parking.PlaceParkingLot(x, z, 20, 15, isGarage)
	if ok {
		cost := float32(1000)
		if isGarage {
			cost = 3000
		}
		sm.Money -= cost
		if isGarage {
			sm.EventBus.Emit(string(EventParkingGaragePlaced), nil)
		} else {
			sm.EventBus.Emit(string(EventParkingLotPlaced), nil)
		}
	}
	return ok
}

func (sm *SimulationManager) RemoveParkingLot(x, z float32) bool {
	bestSlot := int32(-1)
	bestDist := float32(100.0)
	sm.Parking.ForEachLot(func(lot *ParkingLot, slot int32) {
		dx := lot.Position.X - x
		dz := lot.Position.Z - z
		d := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if d < bestDist {
			bestDist = d
			bestSlot = slot
		}
	})
	if bestSlot >= 0 {
		sm.Parking.RemoveParkingLot(bestSlot)
		sm.Money += 500
		sm.EventBus.Emit(string(EventParkingLotRemoved), bestSlot)
		return true
	}
	return false
}

func (sm *SimulationManager) InitTerraform(chunks []*Chunk, rebuildChunk func(idx int)) {
	sm.Terraform = NewTerraformSystem(sm.Heightmap, sm.Water, chunks, rebuildChunk)
}

func (sm *SimulationManager) SetNight(night bool) {
	sm.Night = night
	sm.EventBus.Emit(string(EventDayNightCycle), sm.Night)
}

func (sm *SimulationManager) ToggleDayNight() {
	sm.SetNight(!sm.Night)
}
