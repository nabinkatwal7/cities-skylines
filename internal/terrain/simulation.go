package terrain



type SimEvent string

const (
	EventRoadPlaced        SimEvent = "road:placed"
	EventRoadRemoved       SimEvent = "road:removed"
	EventRoadUpgraded      SimEvent = "road:upgraded"
	EventZonePlaced        SimEvent = "zone:placed"
	EventZoneCleared       SimEvent = "zone:cleared"
	EventParkPlaced        SimEvent = "park:placed"
	EventBuildingUpgraded  SimEvent = "building:upgraded"
	EventBuildingAbandoned SimEvent = "building:abandoned"
	EventBuildingDemolished SimEvent = "building:demolished"
	EventBuildingConstructed SimEvent = "building:constructed"
	EventDayNightCycle     SimEvent = "time:daynight"
	EventTaxCollected      SimEvent = "economy:tax"
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

	Seed      int64
	Night     bool
	Money     float32
	TaxTimer  int32

	EventBus  *EventBus
	scheduler *Scheduler
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
		Transport:   NewTransportManager(),
		Districts:   NewDistrictManager(),
		Seed:        seed,
		EventBus:    NewEventBus(),
		scheduler:   NewScheduler(),
	}

	sm.initScheduler()
	return sm
}

func (sm *SimulationManager) initScheduler() {
	sm.scheduler.Register(GroupFast, UpdateTask{
		Name:     "water",
		Priority: SchedPriorityHigh,
		Callback: func(dt float64) { sm.Water.Update(dt) },
	})
	sm.scheduler.Register(GroupFast, UpdateTask{
		Name:     "trees",
		Priority: SchedPriorityLow,
		Callback: func(dt float64) { sm.Trees.Update(dt) },
	})
	sm.scheduler.Register(GroupMedium, UpdateTask{
		Name:     "roads",
		Priority: SchedPriorityMedium,
		Callback: func(dt float64) { sm.Roads.Update(sm.Heightmap) },
	})
	sm.scheduler.Register(GroupMedium, UpdateTask{
		Name:     "vehicles",
		Priority: SchedPriorityMedium,
		Callback: func(dt float64) { sm.Vehicles.Update(sm.Roads, sm.Heightmap) },
	})
	sm.scheduler.Register(GroupMedium, UpdateTask{
		Name:     "transport",
		Priority: SchedPriorityMedium,
		Callback: func(dt float64) { sm.Transport.Update(sm.Roads, sm.Heightmap) },
	})
	sm.scheduler.Register(GroupSlow, UpdateTask{
		Name:     "buildings",
		Priority: SchedPriorityHigh,
		Callback: func(dt float64) { sm.Buildings.Update(sm.Zones, sm.Heightmap, sm.Roads, sm.Districts) },
	})
	sm.scheduler.Register(GroupSlow, UpdateTask{
		Name:     "tax",
		Priority: SchedPriorityMedium,
		Callback: func(dt float64) { sm.collectTax(dt) },
	})
}

func (sm *SimulationManager) InitDefaultRoads() {
	n0 := sm.Roads.AddNode(-50, -50)
	n1 := sm.Roads.AddNode(-50, 50)
	n2 := sm.Roads.AddNode(50, 50)
	n3 := sm.Roads.AddNode(50, -50)
	n4 := sm.Roads.AddNode(-80, 0)
	n5 := sm.Roads.AddNode(80, 0)
	n6 := sm.Roads.AddNode(0, -80)
	n7 := sm.Roads.AddNode(0, 80)
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
	sm.scheduler.RunGroup(GroupFast, dt)
	sm.scheduler.RunGroup(GroupMedium, dt)
	sm.scheduler.RunGroup(GroupSlow, dt)
	sm.scheduler.RunGroup(GroupVerySlow, dt)
	sm.EventBus.ProcessQueue()
}

func (sm *SimulationManager) collectTax(dt float64) {
	sm.TaxTimer++
	if sm.TaxTimer > 60 {
		pop := sm.Buildings.Population()
		tax := float32(len(sm.Buildings.Buildings)) * 0.5
		if pop > 0 {
			tax += float32(pop) * 0.1
		}
		sm.Money += tax
		sm.EventBus.Emit(string(EventTaxCollected), tax)
		sm.TaxTimer = 0
	}
}

func (sm *SimulationManager) PlaceRoadNode(x, z float32) uint32 {
	return sm.Roads.AddNode(x, z)
}

func (sm *SimulationManager) PlaceRoadSegment(nodeA uint32, x, z float32, roadType RoadType, elevation int32) (uint32, uint32) {
	nodeB := sm.Roads.AddNode(x, z)
	segID := sm.Roads.AddSegment(nodeA, nodeB, roadType)
	if elevation > 0 {
		for i := range sm.Roads.Segments {
			if sm.Roads.Segments[i].ID == segID {
				sm.Roads.Segments[i].Elevation = elevation
				break
			}
		}
	}
	sm.Roads.Rebuild(sm.Heightmap)
	sm.Money -= 100
	sm.EventBus.Emit(string(EventRoadPlaced), segID)
	return nodeB, segID
}

func (sm *SimulationManager) RemoveSegment(idx int) {
	sm.Roads.RemoveSegment(idx)
	sm.Roads.Rebuild(sm.Heightmap)
	sm.EventBus.Emit(string(EventRoadRemoved), idx)
}

func (sm *SimulationManager) UpgradeSegment(idx int, newType RoadType) {
	sm.Roads.UpgradeSegment(idx, newType)
	sm.Roads.Rebuild(sm.Heightmap)
	sm.EventBus.Emit(string(EventRoadUpgraded), idx)
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
