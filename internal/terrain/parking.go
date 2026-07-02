package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ParkingSpotType uint8

const (
	ParkingSpotRoadside ParkingSpotType = iota
	ParkingSpotLot
	ParkingSpotGarage
)

type ParkingSpot struct {
	ID       uint32
	SpotType ParkingSpotType
	X, Z     float32
	RoadSeg  int32
	LaneIdx  int32
	Occupied bool
	OwnerSlot int32
}

type ParkingLot struct {
	Entity
	LotType      ParkingSpotType
	Width, Depth float32
	Capacity     int32
	CellX, CellZ int
	Spots        []int32
}

const ParkingLotPoolSize = 500
const BusDepotPoolSize = 100
const TramDepotPoolSize = 100
const MetroDepotPoolSize = 100

type BusDepot struct {
	Entity
	X, Z   float32
	CellX, CellZ int
}

type TramDepot struct {
	Entity
	X, Z   float32
	CellX, CellZ int
}

type MetroDepot struct {
	Entity
	X, Z   float32
	CellX, CellZ int
}

type ParkingManager struct {
	Spots      []ParkingSpot
	Lots       [ParkingLotPoolSize]ParkingLot
	LotFreeList []int32
	LotCount   int32
	NextID     uint32
	Timer      int32

	BusDepots   [BusDepotPoolSize]BusDepot
	DepotFreeList []int32
	DepotCount  int32

	TramDepots   [TramDepotPoolSize]TramDepot
	TramDepotFreeList []int32
	TramDepotCount  int32

	MetroDepots   [MetroDepotPoolSize]MetroDepot
	MetroDepotFreeList []int32
	MetroDepotCount  int32
}

func NewParkingManager() *ParkingManager {
	pm := &ParkingManager{
		LotFreeList:  make([]int32, ParkingLotPoolSize),
		DepotFreeList: make([]int32, BusDepotPoolSize),
		TramDepotFreeList: make([]int32, TramDepotPoolSize),
		MetroDepotFreeList: make([]int32, MetroDepotPoolSize),
	}
	for i := 0; i < ParkingLotPoolSize; i++ {
		pm.Lots[i].Lifecycle = LifecycleUnallocated
		pm.LotFreeList[i] = int32(ParkingLotPoolSize - 1 - i)
	}
	for i := 0; i < BusDepotPoolSize; i++ {
		pm.BusDepots[i].Lifecycle = LifecycleUnallocated
		pm.DepotFreeList[i] = int32(BusDepotPoolSize - 1 - i)
	}
	for i := 0; i < TramDepotPoolSize; i++ {
		pm.TramDepots[i].Lifecycle = LifecycleUnallocated
		pm.TramDepotFreeList[i] = int32(TramDepotPoolSize - 1 - i)
	}
	for i := 0; i < MetroDepotPoolSize; i++ {
		pm.MetroDepots[i].Lifecycle = LifecycleUnallocated
		pm.MetroDepotFreeList[i] = int32(MetroDepotPoolSize - 1 - i)
	}
	return pm
}

func (pm *ParkingManager) allocBusDepot() int32 {
	if len(pm.DepotFreeList) == 0 {
		return -1
	}
	idx := pm.DepotFreeList[len(pm.DepotFreeList)-1]
	pm.DepotFreeList = pm.DepotFreeList[:len(pm.DepotFreeList)-1]
	pm.BusDepots[idx] = BusDepot{}
	pm.BusDepots[idx].Lifecycle = LifecycleAllocated
	pm.DepotCount++
	return int32(idx)
}

func (pm *ParkingManager) allocLot() int32 {
	if len(pm.LotFreeList) == 0 {
		return -1
	}
	idx := pm.LotFreeList[len(pm.LotFreeList)-1]
	pm.LotFreeList = pm.LotFreeList[:len(pm.LotFreeList)-1]
	pm.Lots[idx].Lifecycle = LifecycleInitializing
	pm.LotCount++
	return idx
}

func (pm *ParkingManager) freeBusDepot(slot int32) {
	if slot < 0 || int(slot) >= BusDepotPoolSize {
		return
	}
	pm.BusDepots[slot].Lifecycle = LifecycleReturnedToPool
	pm.DepotFreeList = append(pm.DepotFreeList, slot)
	pm.DepotCount--
}

func (pm *ParkingManager) PlaceBusDepot(x, z float32) int32 {
	slot := pm.allocBusDepot()
	if slot < 0 {
		return -1
	}
	d := &pm.BusDepots[slot]
	d.X = x
	d.Z = z
	d.CellX = int(x) / 8
	d.CellZ = int(z) / 8
	return slot
}

func (pm *ParkingManager) PlaceTramDepot(x, z float32) int32 {
	slot := pm.allocTramDepot()
	if slot < 0 {
		return -1
	}
	d := &pm.TramDepots[slot]
	d.X = x
	d.Z = z
	d.CellX = int(x) / 8
	d.CellZ = int(z) / 8
	return slot
}

func (pm *ParkingManager) NearestBusDepot(x, z float32, maxDist float32) (int32, float32) {
	best := maxDist
	bestIdx := int32(-1)
	for i := 0; i < BusDepotPoolSize; i++ {
		if pm.BusDepots[i].Lifecycle != LifecycleAllocated {
			continue
		}
		d := &pm.BusDepots[i]
		dx := d.X - x
		dz := d.Z - z
		dist := dx*dx + dz*dz
		if dist < best {
			best = dist
			bestIdx = int32(i)
		}
	}
	return bestIdx, best
}

func (pm *ParkingManager) allocTramDepot() int32 {
	if len(pm.TramDepotFreeList) == 0 {
		return -1
	}
	idx := pm.TramDepotFreeList[len(pm.TramDepotFreeList)-1]
	pm.TramDepotFreeList = pm.TramDepotFreeList[:len(pm.TramDepotFreeList)-1]
	pm.TramDepots[idx] = TramDepot{}
	pm.TramDepots[idx].Lifecycle = LifecycleAllocated
	pm.TramDepotCount++
	return int32(idx)
}

func (pm *ParkingManager) freeTramDepot(slot int32) {
	if slot < 0 || int(slot) >= TramDepotPoolSize {
		return
	}
	pm.TramDepots[slot].Lifecycle = LifecycleReturnedToPool
	pm.TramDepotFreeList = append(pm.TramDepotFreeList, slot)
	pm.TramDepotCount--
}

func (pm *ParkingManager) NearestTramDepot(x, z float32, maxDist float32) (int32, float32) {
	best := maxDist
	bestIdx := int32(-1)
	for i := 0; i < TramDepotPoolSize; i++ {
		if pm.TramDepots[i].Lifecycle != LifecycleAllocated {
			continue
		}
		d := &pm.TramDepots[i]
		dx := d.X - x
		dz := d.Z - z
		dist := dx*dx + dz*dz
		if dist < best {
			best = dist
			bestIdx = int32(i)
		}
	}
	return bestIdx, best
}

func (pm *ParkingManager) allocMetroDepot() int32 {
	if len(pm.MetroDepotFreeList) == 0 {
		return -1
	}
	idx := pm.MetroDepotFreeList[len(pm.MetroDepotFreeList)-1]
	pm.MetroDepotFreeList = pm.MetroDepotFreeList[:len(pm.MetroDepotFreeList)-1]
	pm.MetroDepots[idx] = MetroDepot{}
	pm.MetroDepots[idx].Lifecycle = LifecycleAllocated
	pm.MetroDepotCount++
	return int32(idx)
}

func (pm *ParkingManager) freeMetroDepot(slot int32) {
	if slot < 0 || int(slot) >= MetroDepotPoolSize {
		return
	}
	pm.MetroDepots[slot].Lifecycle = LifecycleReturnedToPool
	pm.MetroDepotFreeList = append(pm.MetroDepotFreeList, slot)
	pm.MetroDepotCount--
}

func (pm *ParkingManager) PlaceMetroDepot(x, z float32) int32 {
	slot := pm.allocMetroDepot()
	if slot < 0 {
		return -1
	}
	d := &pm.MetroDepots[slot]
	d.X = x
	d.Z = z
	d.CellX = int(x) / 8
	d.CellZ = int(z) / 8
	return slot
}

func (pm *ParkingManager) NearestMetroDepot(x, z float32, maxDist float32) (int32, float32) {
	best := maxDist
	bestIdx := int32(-1)
	for i := 0; i < MetroDepotPoolSize; i++ {
		if pm.MetroDepots[i].Lifecycle != LifecycleAllocated {
			continue
		}
		d := &pm.MetroDepots[i]
		dx := d.X - x
		dz := d.Z - z
		dist := dx*dx + dz*dz
		if dist < best {
			best = dist
			bestIdx = int32(i)
		}
	}
	return bestIdx, best
}

func (pm *ParkingManager) freeLot(slot int32) {
	if slot < 0 || int(slot) >= ParkingLotPoolSize {
		return
	}
	lot := &pm.Lots[slot]
	for _, spotIdx := range lot.Spots {
		if int(spotIdx) < len(pm.Spots) {
			pm.Spots[spotIdx].Occupied = false
		}
	}
	pm.Lots[slot] = ParkingLot{}
	pm.Lots[slot].Lifecycle = LifecycleReturnedToPool
	pm.LotFreeList = append(pm.LotFreeList, slot)
	pm.LotCount--
}

func (pm *ParkingManager) addSpot(x, z float32, spotType ParkingSpotType, roadSeg, laneIdx int32, ownerSlot int32) int {
	pm.NextID++
	idx := len(pm.Spots)
	pm.Spots = append(pm.Spots, ParkingSpot{
		ID:        pm.NextID,
		SpotType:  spotType,
		X:         x,
		Z:         z,
		RoadSeg:   roadSeg,
		LaneIdx:   laneIdx,
		Occupied:  false,
		OwnerSlot: ownerSlot,
	})
	return idx
}

func (pm *ParkingManager) GenerateRoadsideSpots(rm *RoadManager) {
	targetSegs := make(map[int]bool)
	for i, seg := range rm.Segments {
		for _, lane := range seg.Lanes {
			if lane.Category == LaneParking {
				targetSegs[i] = true
				break
			}
		}
	}

	keep := make([]ParkingSpot, 0, len(pm.Spots))
	for _, sp := range pm.Spots {
		if sp.SpotType != ParkingSpotRoadside {
			keep = append(keep, sp)
		}
	}
	pm.Spots = keep

	for segIdx := range targetSegs {
		seg := rm.Segments[segIdx]
		for li, lane := range seg.Lanes {
			if lane.Category != LaneParking {
				continue
			}
			xs, zs, ds := rm.SampleLane(seg, int32(li), int(seg.Length/2))
			if len(xs) < 2 {
				continue
			}
			totalLen := ds[len(ds)-1]
			if totalLen < 0.01 {
				continue
			}
			spotSpacing := float32(8.0)
			spots := int(totalLen / spotSpacing)
			if spots < 1 {
				spots = 1
			}
			for si := 0; si < spots; si++ {
				t := (float32(si) + 0.5) * spotSpacing / totalLen
				if t > 1 {
					t = 1
				}
				idx := int(t * float32(len(xs)-1))
				if idx >= len(xs)-1 {
					idx = len(xs) - 2
				}
				frac := t*float32(len(xs)-1) - float32(idx)
				sx := xs[idx] + (xs[idx+1]-xs[idx])*frac
				sz := zs[idx] + (zs[idx+1]-zs[idx])*frac
				pm.addSpot(sx, sz, ParkingSpotRoadside, int32(segIdx), int32(li), -1)
			}
		}
	}
}

func (pm *ParkingManager) PlaceParkingLot(x, z, w, d float32, isGarage bool) bool {
	slot := pm.allocLot()
	if slot < 0 {
		return false
	}
	lot := &pm.Lots[slot]
	lot.Entity = NewEntity(pm.NextID, x, 0, z, OwnerBuilding)
	lot.Lifecycle = LifecycleActive
	lot.LotType = ParkingSpotLot
	if isGarage {
		lot.LotType = ParkingSpotGarage
	}
	lot.Width = w
	lot.Depth = d
	spotArea := float32(25.0)
	capacity := int32(math.Max(1, float64(w*d/spotArea)))
	if isGarage {
		capacity *= 3
	}
	lot.Capacity = capacity

	spacing := float32(5.0)
	spotsPerRow := int(w / spacing)
	if spotsPerRow < 1 {
		spotsPerRow = 1
	}
	rows := int(d / spacing)
	if rows < 1 {
		rows = 1
	}
	halfW := w * 0.5
	halfD := d * 0.5
	count := 0
	for ri := 0; ri < rows && count < int(capacity); ri++ {
		for si := 0; si < spotsPerRow && count < int(capacity); si++ {
			sx := x - halfW + (float32(si)+0.5)*spacing
			sz := z - halfD + (float32(ri)+0.5)*spacing
			idx := pm.addSpot(sx, sz, lot.LotType, -1, -1, slot)
			lot.Spots = append(lot.Spots, int32(idx))
			count++
		}
	}
	pm.NextID++
	return true
}

func (pm *ParkingManager) RemoveParkingLot(slot int32) {
	if slot < 0 || int(slot) >= ParkingLotPoolSize {
		return
	}
	lot := &pm.Lots[slot]
	removeIDs := make(map[uint32]bool)
	for _, spotIdx := range lot.Spots {
		if int(spotIdx) < len(pm.Spots) {
			removeIDs[pm.Spots[spotIdx].ID] = true
		}
	}
	keep := make([]ParkingSpot, 0, len(pm.Spots))
	for _, sp := range pm.Spots {
		if !removeIDs[sp.ID] {
			keep = append(keep, sp)
		}
	}
	pm.Spots = keep
	pm.freeLot(slot)
}

func (pm *ParkingManager) FindSpot(x, z, radius float32) int {
	bestIdx := -1
	bestDist := radius * radius
	for i, sp := range pm.Spots {
		if sp.Occupied {
			continue
		}
		dx := sp.X - x
		dz := sp.Z - z
		d := dx*dx + dz*dz
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return bestIdx
}

func (pm *ParkingManager) OccupySpot(spotIdx int, vehicleID uint32) bool {
	if spotIdx < 0 || spotIdx >= len(pm.Spots) {
		return false
	}
	sp := &pm.Spots[spotIdx]
	if sp.Occupied {
		return false
	}
	sp.Occupied = true
	return true
}

func (pm *ParkingManager) FreeSpot(spotIdx int) {
	if spotIdx < 0 || spotIdx >= len(pm.Spots) {
		return
	}
	pm.Spots[spotIdx].Occupied = false
}

func (pm *ParkingManager) ForEachLot(fn func(*ParkingLot, int32)) {
	for i := 0; i < ParkingLotPoolSize; i++ {
		if pm.Lots[i].Lifecycle == LifecycleActive {
			fn(&pm.Lots[i], int32(i))
		}
	}
}

func (pm *ParkingManager) Draw(h *Heightmap) {
	for _, sp := range pm.Spots {
		if sp.Occupied {
			continue
		}
		hy := h.WorldHeight(sp.X, sp.Z) + 0.1
		var col rl.Color
		switch sp.SpotType {
		case ParkingSpotRoadside:
			col = rl.NewColor(100, 100, 200, 120)
		case ParkingSpotLot:
			col = rl.NewColor(80, 180, 80, 120)
		case ParkingSpotGarage:
			col = rl.NewColor(80, 80, 200, 120)
		}
		rl.DrawCube(rl.NewVector3(sp.X, hy, sp.Z), 1.8, 0.05, 1.0, col)
		if sp.SpotType != ParkingSpotRoadside {
			rl.DrawCubeWires(rl.NewVector3(sp.X, hy, sp.Z), 1.8, 0.05, 1.0, rl.NewColor(60, 60, 60, 100))
		}
	}

	pm.ForEachLot(func(lot *ParkingLot, slot int32) {
		hy := h.WorldHeight(lot.Position.X, lot.Position.Z) + 0.3
		var col rl.Color
		if lot.LotType == ParkingSpotGarage {
			col = rl.NewColor(100, 100, 200, 100)
			rl.DrawCube(rl.NewVector3(lot.Position.X, hy+1.5, lot.Position.Z), lot.Width, 3.0, lot.Depth, col)
			rl.DrawCubeWires(rl.NewVector3(lot.Position.X, hy+1.5, lot.Position.Z), lot.Width, 3.0, lot.Depth, rl.NewColor(60, 60, 100, 150))
		} else {
			col = rl.NewColor(80, 160, 80, 80)
			rl.DrawCube(rl.NewVector3(lot.Position.X, hy, lot.Position.Z), lot.Width, 0.3, lot.Depth, col)
			rl.DrawCubeWires(rl.NewVector3(lot.Position.X, hy, lot.Position.Z), lot.Width, 0.3, lot.Depth, rl.NewColor(60, 100, 60, 100))
		}
	})

	for i := 0; i < BusDepotPoolSize; i++ {
		d := &pm.BusDepots[i]
		if d.Lifecycle != LifecycleAllocated {
			continue
		}
		hy := h.WorldHeight(d.X, d.Z) + 0.5
		rl.DrawCube(rl.NewVector3(d.X, hy, d.Z), 6, 1, 4, rl.NewColor(200, 180, 50, 180))
		rl.DrawCubeWires(rl.NewVector3(d.X, hy, d.Z), 6, 1, 4, rl.NewColor(200, 180, 50, 255))
		rl.DrawCube(rl.NewVector3(d.X, hy+0.6, d.Z), 2, 0.3, 1, rl.NewColor(255, 200, 100, 200))
	}

	for i := 0; i < TramDepotPoolSize; i++ {
		d := &pm.TramDepots[i]
		if d.Lifecycle != LifecycleAllocated {
			continue
		}
		hy := h.WorldHeight(d.X, d.Z) + 0.5
		rl.DrawCube(rl.NewVector3(d.X, hy, d.Z), 6, 1, 4, rl.NewColor(180, 50, 180, 180))
		rl.DrawCubeWires(rl.NewVector3(d.X, hy, d.Z), 6, 1, 4, rl.NewColor(180, 50, 180, 255))
		rl.DrawCube(rl.NewVector3(d.X, hy+0.6, d.Z), 2, 0.3, 1, rl.NewColor(255, 100, 255, 200))
	}

	for i := 0; i < MetroDepotPoolSize; i++ {
		d := &pm.MetroDepots[i]
		if d.Lifecycle != LifecycleAllocated {
			continue
		}
		hy := h.WorldHeight(d.X, d.Z) + 0.5
		rl.DrawCube(rl.NewVector3(d.X, hy, d.Z), 6, 1, 4, rl.NewColor(80, 80, 200, 180))
		rl.DrawCubeWires(rl.NewVector3(d.X, hy, d.Z), 6, 1, 4, rl.NewColor(80, 80, 200, 255))
		rl.DrawCube(rl.NewVector3(d.X, hy+0.6, d.Z), 2, 0.3, 1, rl.NewColor(150, 150, 255, 200))
	}
}

func (pm *ParkingManager) Unload() {
	for i := 0; i < ParkingLotPoolSize; i++ {
		if pm.Lots[i].Lifecycle == LifecycleActive {
			pm.Lots[i].Lifecycle = LifecycleReturnedToPool
		}
	}
	pm.LotCount = 0
	pm.Spots = nil
}
