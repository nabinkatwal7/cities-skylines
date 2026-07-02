package sim

import (
	"math"

	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/transport"
)

type SimulationManager struct {
	Generator   *terrain.Generator
	Heightmap   *terrain.Heightmap
	Water       *terrain.WaterSystem
	Terraform   *terrain.TerraformSystem
	Trees       *terrain.TreeSystem
	Resources   *terrain.ResourceSystem
	Roads        *road.RoadManager
	Connections  *terrain.ConnectionSystem
	Vehicles     *road.VehicleManager
	Transport    *transport.TransportManager
	Buildability *terrain.BuildabilityChecker
	Parking      *road.ParkingManager

	Seed      int64
	Night     bool
	Money     float32
	TaxTimer  int32

	EventBus  *core.EventBus
	scheduler *core.Scheduler
	Time      *core.GameTime
	Jobs      *core.JobQueue
}

func NewSimulationManager(seed int64) *SimulationManager {
	gen := terrain.NewGenerator(seed)
	h := gen.Generate()
	water := terrain.NewWaterSystem()
	water.Init(h)
	trees := terrain.NewTreeSystem(seed)
	res := terrain.NewResourceSystem(seed, h)
	trees.Generate(h, water)
	roads := road.NewRoadManager()
	roads.LoadAssets()
	conn := terrain.NewConnectionSystem()
	roads.InitOutsideConnections(conn)
	vehicles := road.NewVehicleManager()
	parking := road.NewParkingManager()

	sm := &SimulationManager{
		Money:       100000,
		Generator:   gen,
		Heightmap:   h,
		Water:       water,
		Trees:       trees,
		Resources:   res,
		Roads:       roads,
		Connections: conn,
		Vehicles:    vehicles,
		Parking:     parking,
		Transport:   transport.NewTransportManager(),
		Seed:        seed,
		EventBus:    core.NewEventBus(),
		scheduler:   core.NewScheduler(),
		Time:        core.NewGameTime(),
		Jobs:        core.NewJobQueue(),
	}
	sm.Transport.Parking = parking
	sm.Transport.Rails.InitOutsideConnections(conn)
	sm.Transport.InitExternalConnections(conn)

	sm.initScheduler()
	sm.initEventListeners()
	sm.Water.SetEventBus(sm.EventBus)
	road.SetWaterForRoads(sm.Water)
	sm.Buildability = terrain.NewBuildabilityChecker(sm.Heightmap, sm.Water, sm.Trees, sm.Roads, sm.Resources)
	terrain.SetBuildabilityChecker(sm.Buildability)
	sm.Trees.SetResources(sm.Resources)
	sm.Resources.SetTrees(sm.Trees)
	sm.Parking.GenerateRoadsideSpots(sm.Roads)
	return sm
}

func (sm *SimulationManager) initScheduler() {
	sm.scheduler.Register(core.GroupFast, core.UpdateTask{
		Name:     "water",
		Priority: core.SchedPriorityHigh,
		BudgetMs: 1,
		Callback: func(dt float64) { sm.Water.Update(dt) },
	})
	sm.scheduler.Register(core.GroupFast, core.UpdateTask{
		Name:     "trees",
		Priority: core.SchedPriorityLow,
		BudgetMs: 0.5,
		Callback: func(dt float64) { sm.Trees.Update(dt) },
	})
	sm.scheduler.Register(core.GroupMedium, core.UpdateTask{
		Name:     "roads",
		Priority: core.SchedPriorityCritical,
		BudgetMs: 2,
		Callback: func(dt float64) { sm.Roads.Update(sm.Heightmap) },
	})
	sm.scheduler.Register(core.GroupMedium, core.UpdateTask{
		Name:     "vehicles",
		Priority: core.SchedPriorityHigh,
		BudgetMs: 3,
		Callback: func(dt float64) { sm.Vehicles.Update(sm.Roads, sm.Heightmap, sm.Parking) },
	})
	sm.scheduler.Register(core.GroupMedium, core.UpdateTask{
		Name:     "transport",
		Priority: core.SchedPriorityHigh,
		BudgetMs: 2,
		Callback: func(dt float64) { sm.Transport.Update(sm.Roads, sm.Heightmap) },
	})
	sm.scheduler.Register(core.GroupMedium, core.UpdateTask{
		Name:     "cargo",
		Priority: core.SchedPriorityMedium,
		BudgetMs: 1,
		Callback: func(dt float64) {
			if sm.Transport.Cargo != nil {
				sm.Transport.Cargo.Update(sm.Transport.Rails)
			}
		},
	})
	sm.scheduler.Register(core.GroupFast, core.UpdateTask{
		Name:     "parking",
		Priority: core.SchedPriorityLow,
		BudgetMs: 0.5,
		Callback: func(dt float64) { sm.Parking.Update(sm.Transport) },
	})
	sm.scheduler.Register(core.GroupVerySlow, core.UpdateTask{
		Name:     "tax",
		Priority: core.SchedPriorityMedium,
		BudgetMs: 1,
		Callback: func(dt float64) { sm.collectTax(dt) },
	})
	sm.scheduler.Register(core.GroupVerySlow, core.UpdateTask{
		Name:     "connections",
		Priority: core.SchedPriorityLow,
		BudgetMs: 0.5,
		Callback: func(dt float64) { sm.Connections.Update() },
	})

}

func (sm *SimulationManager) initEventListeners() {
	sm.EventBus.On(string(core.EventRoadRemoved), func(data any) {
		idx, ok := data.(int)
		if ok && sm.Vehicles != nil {
			sm.Vehicles.OnRoadRemoved(idx)
			sm.Parking.GenerateRoadsideSpots(sm.Roads)
		}
	})
	sm.EventBus.On(string(core.EventRoadPlaced), func(data any) {
		if sm.Parking != nil {
			sm.Parking.GenerateRoadsideSpots(sm.Roads)
		}
	})
	sm.EventBus.On(string(core.EventRoadUpgraded), func(data any) {
		if sm.Parking != nil {
			sm.Parking.GenerateRoadsideSpots(sm.Roads)
		}
	})
	sm.EventBus.On(string(core.EventDayNightCycle), func(data any) {
		isNight, ok := data.(bool)
		if ok && sm.Roads != nil {
			sm.Roads.SetNightMode(isNight)
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
	sm.Roads.AddSegment(n0, n1, road.RoadTwoLane)
	sm.Roads.AddSegment(n1, n2, road.RoadTwoLane)
	sm.Roads.AddSegment(n2, n3, road.RoadTwoLane)
	sm.Roads.AddSegment(n3, n0, road.RoadTwoLane)
	sm.Roads.AddSegment(n4, n0, road.RoadFourLane)
	sm.Roads.AddSegment(n5, n2, road.RoadFourLane)
	sm.Roads.AddSegment(n6, n3, road.RoadTwoLane)
	sm.Roads.AddSegment(n7, n1, road.RoadTwoLane)
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
		sm.EventBus.Emit(string(core.EventTimeMinute), sm.Time.Minute)
	}
	if sm.Time.HourChanged() {
		sm.EventBus.Emit(string(core.EventTimeHour), sm.Time.Hour)
	}
	if sm.Time.DayChanged() {
		sm.EventBus.Emit(string(core.EventTimeDay), sm.Time.DayCount)
	}
	sm.syncDayNight()
	sm.Time.Snapshot()

	sm.EventBus.ProcessQueue()
}

func (sm *SimulationManager) syncDayNight() {
	isDay := sm.Time.IsDaytime()
	if sm.Night == isDay {
		sm.Night = !isDay
		sm.EventBus.Emit(string(core.EventDayNightCycle), sm.Night)
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
		if sm.Roads != nil {
			sm.Money -= sm.Roads.TotalMaintenance()
		}
		if sm.Transport != nil {
			sm.Money -= sm.Transport.TotalMaintenance()
			sm.Money += sm.Transport.TotalIncome() * 0.1
		}
		sm.TaxTimer = 0
	}
}

const (
	MaxRoadSlope      = 0.25
	MinCurveRadius    = 8.0
	RoadProximityDist = 4.0
)

func (sm *SimulationManager) CanPlaceRoad(x1, z1, x2, z2 float32, rt road.RoadType, elevation int32, excludeSegID uint32) string {
	half := float32(terrain.WorldSize / 2)
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
			if isNearNode(sm, xs[i], zs[i]) {
				continue
			}
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

	return ""
}

func isNearNode(sm *SimulationManager, x, z float32) bool {
	for _, n := range sm.Roads.Nodes {
		dx := n.X - x
		dz := n.Z - z
		if dx*dx+dz*dz < RoadProximityDist*RoadProximityDist {
			return true
		}
	}
	return false
}

func (sm *SimulationManager) PlaceRoadNode(x, z float32) uint32 {
	if sm.Connections != nil {
		for _, c := range sm.Connections.GetByType(terrain.ConnHighway) {
			dx := c.WorldX - x
			dz := c.WorldZ - z
			if dx*dx+dz*dz < 64 {
				for idx := range sm.Roads.Nodes {
					n := &sm.Roads.Nodes[idx]
					if n.Flags&road.RoadFlagOutsideConn != 0 {
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

func (sm *SimulationManager) PlaceRoadSegment(nodeA uint32, x, z float32, roadType road.RoadType, elevation int32) (uint32, uint32, bool) {
	na := sm.Roads.Nodes[nodeA]
	snapX, snapZ := x, z
	if sm.Connections != nil {
		for _, c := range sm.Connections.GetByType(terrain.ConnHighway) {
			dx := c.WorldX - x
			dz := c.WorldZ - z
			if dx*dx+dz*dz < 64 {
				for idx := range sm.Roads.Nodes {
					n := &sm.Roads.Nodes[idx]
					if n.Flags&road.RoadFlagOutsideConn != 0 {
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

	nodeB := uint32(math.MaxUint32)
	for i := range sm.Roads.Nodes {
		n := &sm.Roads.Nodes[i]
		dx := n.X - snapX
		dz := n.Z - snapZ
		if dx*dx+dz*dz < RoadProximityDist*RoadProximityDist && uint32(i) != nodeA {
			nodeB = uint32(i)
			break
		}
	}
	if nodeB == math.MaxUint32 {
		nodeB = sm.Roads.AddNode(snapX, 0, snapZ)
	}
	segID := sm.Roads.AddSegment(nodeA, nodeB, roadType)
	sm.finalizeSegment(segID, nodeA, nodeB, roadType, elevation)
	return nodeB, segID, true
}

func (sm *SimulationManager) finalizeSegment(segID, nodeA, nodeB uint32, roadType road.RoadType, elevation int32) {
	if elevation != 0 {
		for i := range sm.Roads.Segments {
			if sm.Roads.Segments[i].ID == segID {
				sm.Roads.Segments[i].Elevation = elevation
				sm.Roads.Segments[i].MaintenanceCost = sm.Roads.CalcSegmentMaintenance(roadType, sm.Roads.Segments[i].Length, elevation)
				if elevation > 0 {
					sm.Roads.Nodes[nodeA].Flags |= road.RoadFlagBridge
					sm.Roads.Nodes[nodeB].Flags |= road.RoadFlagBridge
				} else if elevation < 0 {
					sm.Roads.Nodes[nodeA].Flags |= road.RoadFlagTunnel
					sm.Roads.Nodes[nodeB].Flags |= road.RoadFlagTunnel
				}
				break
			}
		}
	}
	cost := road.RoadConstructionCost(roadType)
	if elevation > 0 {
		cost += float32(elevation) * 50
	} else if elevation < 0 {
		cost += float32(-elevation) * 100
	}
	sm.Roads.Rebuild(sm.Heightmap)
	sm.Money -= cost
	sm.EventBus.Emit(string(core.EventRoadPlaced), segID)
}

func (sm *SimulationManager) RemoveSegment(idx int) {
	if idx < 0 || idx >= len(sm.Roads.Segments) {
		return
	}
	segID := int(sm.Roads.Segments[idx].ID)
	sm.Roads.RemoveSegment(idx)
	sm.Roads.Rebuild(sm.Heightmap)
	sm.EventBus.Emit(string(core.EventRoadRemoved), segID)
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

func (sm *SimulationManager) UpgradeSegment(idx int, newType road.RoadType) bool {
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
	oldCost := road.RoadConstructionCost(old.RoadType)
	newCost := road.RoadConstructionCost(newType)
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
	sm.EventBus.Emit(string(core.EventRoadUpgraded), idx)
	return true
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
			sm.EventBus.Emit(string(core.EventParkingGaragePlaced), nil)
		} else {
			sm.EventBus.Emit(string(core.EventParkingLotPlaced), nil)
		}
	}
	return ok
}

func (sm *SimulationManager) PlaceBusDepot(x, z float32) bool {
	if sm.Money < 5000 {
		return false
	}
	slot := sm.Parking.PlaceBusDepot(x, z)
	if slot >= 0 {
		sm.Money -= 5000
		return true
	}
	return false
}

func (sm *SimulationManager) PlaceTramDepot(x, z float32) bool {
	if sm.Money < 5000 {
		return false
	}
	slot := sm.Parking.PlaceTramDepot(x, z)
	if slot >= 0 {
		sm.Money -= 5000
		return true
	}
	return false
}

func (sm *SimulationManager) PlaceMetroDepot(x, z float32) bool {
	if sm.Money < 5000 {
		return false
	}
	slot := sm.Parking.PlaceMetroDepot(x, z)
	if slot >= 0 {
		sm.Money -= 5000
		return true
	}
	return false
}

func (sm *SimulationManager) PlaceFerryDepot(x, z float32) bool {
	if sm.Money < 5000 {
		return false
	}
	slot := sm.Parking.PlaceFerryDepot(x, z)
	if slot >= 0 {
		sm.Money -= 5000
		return true
	}
	return false
}

func (sm *SimulationManager) PlaceMonorailDepot(x, z float32) bool {
	if sm.Money < 5000 {
		return false
	}
	slot := sm.Parking.PlaceMonorailDepot(x, z)
	if slot >= 0 {
		sm.Money -= 5000
		return true
	}
	return false
}

func (sm *SimulationManager) PlaceCableCarDepot(x, z float32) bool {
	if sm.Money < 5000 {
		return false
	}
	slot := sm.Parking.PlaceCableCarDepot(x, z)
	if slot >= 0 {
		sm.Money -= 5000
		return true
	}
	return false
}

func (sm *SimulationManager) PlaceTaxiDepot(x, z float32) bool {
	if sm.Money < 5000 {
		return false
	}
	slot := sm.Parking.PlaceTaxiDepot(x, z)
	if slot >= 0 {
		sm.Money -= 5000
		return true
	}
	return false
}

func (sm *SimulationManager) PlaceAirportDepot(x, z float32) bool {
	if sm.Money < 10000 {
		return false
	}
	slot := sm.Parking.PlaceAirportDepot(x, z)
	if slot >= 0 {
		sm.Money -= 10000
		sm.Transport.AddStop(x, z, transport.TransAir)
		return true
	}
	return false
}

func (sm *SimulationManager) PlacePortDepot(x, z float32) bool {
	if sm.Money < 8000 {
		return false
	}
	slot := sm.Parking.PlacePortDepot(x, z)
	if slot >= 0 {
		sm.Money -= 8000
		sm.Transport.AddStop(x, z, transport.TransShip)
		return true
	}
	return false
}

func (sm *SimulationManager) RemoveParkingLot(x, z float32) bool {
	bestSlot := int32(-1)
	bestDist := float32(100.0)
	sm.Parking.ForEachLot(func(lot *road.ParkingLot, slot int32) {
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
		sm.EventBus.Emit(string(core.EventParkingLotRemoved), bestSlot)
		return true
	}
	return false
}

func (sm *SimulationManager) InitTerraform(chunks []*terrain.Chunk, rebuildChunk func(idx int)) {
	sm.Terraform = terrain.NewTerraformSystem(sm.Heightmap, sm.Water, chunks, rebuildChunk)
}

func (sm *SimulationManager) SetNight(night bool) {
	sm.Night = night
	sm.EventBus.Emit(string(core.EventDayNightCycle), sm.Night)
}

func (sm *SimulationManager) ToggleDayNight() {
	sm.SetNight(!sm.Night)
}
