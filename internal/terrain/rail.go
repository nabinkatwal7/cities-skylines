package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const RailTrackPoolSize = 500
const CargoStationPoolSize = 50

type SignalState uint8

const (
	SignalGreen SignalState = iota
	SignalRed
)

type RailTrack struct {
	ID         uint32
	StartX, StartZ float32
	EndX, EndZ float32
	Length     float32
	Lifecycle  LifecycleState
	SignalA    SignalState
	SignalB    SignalState
	BlockID    int32
	OutsideA   bool
	OutsideB   bool
	Occupied   bool
	OccupierID uint32
}

type RailManager struct {
	Pool     [RailTrackPoolSize]RailTrack
	FreeList []int32
	Count    int32
	NextID   uint32
}

func NewRailManager() *RailManager {
	rm := &RailManager{
		FreeList: make([]int32, RailTrackPoolSize),
	}
	for i := 0; i < RailTrackPoolSize; i++ {
		rm.Pool[i].Lifecycle = LifecycleUnallocated
		rm.FreeList[i] = int32(RailTrackPoolSize - 1 - i)
	}
	return rm
}

func (rm *RailManager) allocTrack() int32 {
	if len(rm.FreeList) == 0 {
		return -1
	}
	idx := rm.FreeList[len(rm.FreeList)-1]
	rm.FreeList = rm.FreeList[:len(rm.FreeList)-1]
	rm.Pool[idx] = RailTrack{}
	rm.Pool[idx].Lifecycle = LifecycleAllocated
	rm.Count++
	return int32(idx)
}

func (rm *RailManager) freeTrack(slot int32) {
	if slot < 0 || int(slot) >= RailTrackPoolSize {
		return
	}
	rm.Pool[slot].Lifecycle = LifecycleReturnedToPool
	rm.FreeList = append(rm.FreeList, slot)
	rm.Count--
}

func (rm *RailManager) AddTrack(startX, startZ, endX, endZ float32) int32 {
	slot := rm.allocTrack()
	if slot < 0 {
		return -1
	}
	t := &rm.Pool[slot]
	t.ID = rm.NextID
	rm.NextID++
	t.StartX = startX
	t.StartZ = startZ
	t.EndX = endX
	t.EndZ = endZ
	dx := endX - startX
	dz := endZ - startZ
	t.Length = float32(math.Sqrt(float64(dx*dx + dz*dz)))
	t.SignalA = SignalGreen
	t.SignalB = SignalGreen
	t.BlockID = -1
	rm.assignBlock(slot)
	return slot
}

func (rm *RailManager) assignBlock(slot int32) {
	t := &rm.Pool[slot]
	if t.BlockID >= 0 {
		return
	}
	blockID := rm.NextID + 1000000 + uint32(slot)
	t.BlockID = int32(blockID)
	rm.propagateBlock(slot, t.BlockID)
}

func (rm *RailManager) propagateBlock(slot int32, blockID int32) {
	t := &rm.Pool[slot]
	t.BlockID = blockID
	x1, z1 := t.StartX, t.StartZ
	x2, z2 := t.EndX, t.EndZ

	for j := 0; j < RailTrackPoolSize; j++ {
		if j == int(slot) {
			continue
		}
		other := &rm.Pool[j]
		if other.Lifecycle != LifecycleAllocated || other.BlockID >= 0 {
			continue
		}
		if (other.StartX == x1 && other.StartZ == z1) ||
			(other.StartX == x2 && other.StartZ == z2) ||
			(other.EndX == x1 && other.EndZ == z1) ||
			(other.EndX == x2 && other.EndZ == z2) {
			rm.propagateBlock(int32(j), blockID)
		}
	}
}

func (rm *RailManager) ForEachTrack(fn func(*RailTrack, int32)) {
	for i := 0; i < RailTrackPoolSize; i++ {
		if rm.Pool[i].Lifecycle == LifecycleAllocated {
			fn(&rm.Pool[i], int32(i))
		}
	}
}

func (rm *RailManager) NearestTrack(x, z float32, maxDist float32) (int32, float32) {
	best := maxDist
	bestIdx := int32(-1)
	for i := 0; i < RailTrackPoolSize; i++ {
		t := &rm.Pool[i]
		if t.Lifecycle != LifecycleAllocated {
			continue
		}
		midX := (t.StartX + t.EndX) / 2
		midZ := (t.StartZ + t.EndZ) / 2
		dx := midX - x
		dz := midZ - z
		d := dx*dx + dz*dz
		if d < best {
			best = d
			bestIdx = int32(i)
		}
	}
	return bestIdx, best
}

func (rm *RailManager) FindTrackPath(startX, startZ, endX, endZ float32) []int32 {
	type nodeDist struct {
		prev    int32
		dist    float32
		visited bool
		idx     int32
	}

	startTrack, _ := rm.NearestTrack(startX, startZ, 2500)
	endTrack, _ := rm.NearestTrack(endX, endZ, 2500)
	if startTrack < 0 || endTrack < 0 {
		return nil
	}

	nodes := make([]nodeDist, RailTrackPoolSize)
	for i := range nodes {
		nodes[i].prev = -1
		nodes[i].dist = math.MaxFloat32
		nodes[i].idx = -1
	}

	for i := 0; i < RailTrackPoolSize; i++ {
		if rm.Pool[i].Lifecycle == LifecycleAllocated {
			nodes[i].idx = int32(i)
		}
	}
	nodes[startTrack].dist = 0

	for {
		best := -1
		bestD := float32(math.MaxFloat32)
		for i := 0; i < RailTrackPoolSize; i++ {
			if rm.Pool[i].Lifecycle != LifecycleAllocated {
				continue
			}
			if !nodes[i].visited && nodes[i].dist < bestD {
				best = i
				bestD = nodes[i].dist
			}
		}
		if best < 0 || int32(best) == endTrack {
			break
		}
		nodes[best].visited = true
		t := &rm.Pool[best]

		for j := 0; j < RailTrackPoolSize; j++ {
			if rm.Pool[j].Lifecycle != LifecycleAllocated || j == best {
				continue
			}
			other := &rm.Pool[j]
			shared := (t.StartX == other.StartX && t.StartZ == other.StartZ) ||
				(t.StartX == other.EndX && t.StartZ == other.EndZ) ||
				(t.EndX == other.StartX && t.EndZ == other.StartZ) ||
				(t.EndX == other.EndX && t.EndZ == other.EndZ)
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

	if nodes[endTrack].prev < 0 {
		path := []int32{startTrack, endTrack}
		return path
	}

	path := make([]int32, 0)
	cur := int32(endTrack)
	for cur >= 0 {
		path = append(path, cur)
		if cur == startTrack {
			break
		}
		cur = nodes[cur].prev
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func (rm *RailManager) InitOutsideConnections(cs *ConnectionSystem) {
	for _, conn := range cs.GetByType(ConnRail) {
		inwardX := conn.WorldX
		inwardZ := conn.WorldZ
		if conn.WorldX < -WorldSize/2+50 {
			inwardX = conn.WorldX + 80
		} else if conn.WorldX > WorldSize/2-50 {
			inwardX = conn.WorldX - 80
		}
		if conn.WorldZ < -WorldSize/2+50 {
			inwardZ = conn.WorldZ + 80
		} else if conn.WorldZ > WorldSize/2-50 {
			inwardZ = conn.WorldZ - 80
		}
		slot := rm.AddTrack(conn.WorldX, conn.WorldZ, inwardX, inwardZ)
		if slot >= 0 {
			rm.Pool[slot].OutsideA = true
		}
	}
}

func (rm *RailManager) Draw(h *Heightmap) {
	for i := 0; i < RailTrackPoolSize; i++ {
		t := &rm.Pool[i]
		if t.Lifecycle != LifecycleAllocated {
			continue
		}
		hy1 := h.WorldHeight(t.StartX, t.StartZ) + 0.3
		hy2 := h.WorldHeight(t.EndX, t.EndZ) + 0.3
		col := rl.NewColor(200, 150, 50, 180)
		if t.Occupied {
			col = rl.NewColor(255, 50, 50, 200)
		}
		rl.DrawLine3D(rl.NewVector3(t.StartX, hy1, t.StartZ), rl.NewVector3(t.EndX, hy2, t.EndZ), col)

		if t.OutsideA || t.OutsideB {
			x, z := t.StartX, t.StartZ
			if t.OutsideB {
				x, z = t.EndX, t.EndZ
			}
			hy := h.WorldHeight(x, z) + 1
			rl.DrawCube(rl.NewVector3(x, hy, z), 6, 2, 6, rl.NewColor(150, 100, 50, 200))
			rl.DrawCubeWires(rl.NewVector3(x, hy, z), 6, 2, 6, rl.NewColor(200, 150, 50, 255))
		}
	}
}

type CargoStation struct {
	ID          uint32
	X, Z        float32
	Name        string
	GoodsStored int32
	Capacity    int32
	Active      bool
	TrainSlot   int32
}

type CargoManager struct {
	Stations []CargoStation
	NextID   uint32
	Trains   []TransportVehicle
}

func NewCargoManager() *CargoManager {
	return &CargoManager{}
}

func (cm *CargoManager) AddStation(x, z float32) uint32 {
	id := cm.NextID
	cm.NextID++
	cm.Stations = append(cm.Stations, CargoStation{
		ID:          id,
		X:           x,
		Z:           z,
		Name:        "Cargo Station",
		GoodsStored: 0,
		Capacity:    100,
		Active:      true,
		TrainSlot:   -1,
	})
	return id
}

func (cm *CargoManager) RemoveStation(id uint32) {
	for i := range cm.Stations {
		if cm.Stations[i].ID == id {
			cm.Stations = append(cm.Stations[:i], cm.Stations[i+1:]...)
			return
		}
	}
}

func (cm *CargoManager) NearestStation(x, z float32, maxDist float32) *CargoStation {
	best := maxDist
	var found *CargoStation
	for i := range cm.Stations {
		if !cm.Stations[i].Active {
			continue
		}
		s := &cm.Stations[i]
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

func (cm *CargoManager) FindBestPair() (int, int) {
	bestSrc, bestDst := -1, -1
	bestNeed := int32(0)
	for si := range cm.Stations {
		src := &cm.Stations[si]
		if !src.Active || src.GoodsStored < 10 {
			continue
		}
		for di := range cm.Stations {
			if si == di {
				continue
			}
			dst := &cm.Stations[di]
			if !dst.Active || dst.GoodsStored >= dst.Capacity-10 {
				continue
			}
			need := src.GoodsStored / 2
			if dst.Capacity-dst.GoodsStored < need {
				need = dst.Capacity - dst.GoodsStored
			}
			if need > bestNeed {
				bestNeed = need
				bestSrc = si
				bestDst = di
			}
		}
	}
	return bestSrc, bestDst
}

func (cm *CargoManager) Update(rm *RailManager) {
	for vi := range cm.Trains {
		train := &cm.Trains[vi]
		if train.Moving {
			cm.moveCargoTrain(train, rm)
		} else {
			train.Timer++
			if train.Timer > 120 {
				train.Moving = true
			}
		}
	}

	for si := range cm.Stations {
		s := &cm.Stations[si]
		if s.GoodsStored < s.Capacity {
			s.GoodsStored++
		}
	}

	for len(cm.Trains) < 2 {
		src, dst := cm.FindBestPair()
		if src < 0 || dst < 0 {
			break
		}
		srcS := &cm.Stations[src]
		dstS := &cm.Stations[dst]
		amount := srcS.GoodsStored / 2
		if amount > 20 {
			amount = 20
		}
		if dstS.Capacity-dstS.GoodsStored < amount {
			amount = dstS.Capacity - dstS.GoodsStored
		}
		if amount <= 0 {
			break
		}
		srcS.GoodsStored -= amount
		cm.Trains = append(cm.Trains, TransportVehicle{
			ID:         cm.NextID,
			TransType:  TransTrain,
			X:          srcS.X,
			Z:          srcS.Z,
			Speed:      50,
			Capacity:   amount,
			Passengers: amount,
			Forward:    true,
			Moving:     true,
			StopIdx:    src,
			TargetX:    dstS.X,
			TargetZ:    dstS.Z,
		})
		cm.NextID++
	}

	keep := make([]TransportVehicle, 0, len(cm.Trains))
	for _, t := range cm.Trains {
		if t.Moving || t.Timer < 200 {
			keep = append(keep, t)
		}
	}
	cm.Trains = keep
}

func (cm *CargoManager) moveCargoTrain(train *TransportVehicle, rm *RailManager) {
	if rm == nil || rm.Count == 0 {
		dx := train.TargetX - train.X
		dz := train.TargetZ - train.Z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
		if dist < 5 {
			train.Moving = false
			train.Timer = 0
			return
		}
		train.X += (dx / dist) * train.Speed * 0.02
		train.Z += (dz / dist) * train.Speed * 0.02
		return
	}

	if len(train.Path) == 0 {
		path := rm.FindTrackPath(train.X, train.Z, train.TargetX, train.TargetZ)
		train.Path = make([]uint32, len(path))
		for i, idx := range path {
			train.Path[i] = uint32(idx)
		}
		train.PathIdx = 0
	}

	for train.PathIdx < len(train.Path)-1 {
		trackIdx := int(train.Path[train.PathIdx])
		nextTrackIdx := int(train.Path[train.PathIdx+1])
		if trackIdx >= RailTrackPoolSize || nextTrackIdx >= RailTrackPoolSize {
			train.Path = nil
			return
		}
		track := &rm.Pool[trackIdx]
		nextTrack := &rm.Pool[nextTrackIdx]

		tx, tz := track.EndX, track.EndZ
		dx := nextTrack.StartX - track.EndX
		dz := nextTrack.StartZ - track.EndZ
		if dx*dx+dz*dz > (nextTrack.StartX-track.StartX)*(nextTrack.StartX-track.StartX)+(nextTrack.StartZ-track.StartZ)*(nextTrack.StartZ-track.StartZ) {
			tx, tz = track.StartX, track.StartZ
		}

		if !track.Occupied {
			track.Occupied = true
			track.OccupierID = train.ID
		} else if track.OccupierID != train.ID {
			stopped := train.Passengers > 0
			_ = stopped
			return
		}

		dx = tx - train.X
		dz = tz - train.Z
		targetDist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if targetDist < 2 {
			track.Occupied = false
			track.OccupierID = 0
			train.PathIdx++
			continue
		}

		train.X += (dx / targetDist) * train.Speed * 0.02
		train.Z += (dz / targetDist) * train.Speed * 0.02
		return
	}

	for _, idx := range train.Path {
		if int(idx) < RailTrackPoolSize {
			rm.Pool[idx].Occupied = false
			rm.Pool[idx].OccupierID = 0
		}
	}
	train.Path = nil
	train.Moving = false
	train.Timer = 0

	for si := range cm.Stations {
		s := &cm.Stations[si]
		dx := s.X - train.X
		dz := s.Z - train.Z
		if dx*dx+dz*dz < 100 && train.Passengers > 0 {
			s.GoodsStored += train.Passengers
			if s.GoodsStored > s.Capacity {
				s.GoodsStored = s.Capacity
			}
			train.Passengers = 0
			break
		}
	}
}

func (cm *CargoManager) Draw(h *Heightmap) {
	for _, s := range cm.Stations {
		hy := h.WorldHeight(s.X, s.Z) + 0.5
		rl.DrawCube(rl.NewVector3(s.X, hy+1, s.Z), 5, 2, 5, rl.NewColor(200, 150, 50, 200))
		rl.DrawCubeWires(rl.NewVector3(s.X, hy+1, s.Z), 5, 2, 5, rl.NewColor(200, 200, 100, 255))
		barW := float32(4)
		barH := float32(0.3)
		fill := float32(s.GoodsStored) / float32(s.Capacity)
		if fill > 1 {
			fill = 1
		}
		rl.DrawCube(rl.NewVector3(s.X, hy+2.5, s.Z), barW*fill, barH, 0.2, rl.NewColor(255, 200, 50, 220))
		rl.DrawCubeWires(rl.NewVector3(s.X, hy+2.5, s.Z), barW, barH, 0.2, rl.NewColor(60, 60, 60, 150))
	}

	for _, v := range cm.Trains {
		hy := h.WorldHeight(v.X, v.Z) + 0.6
		rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 4, 1, 1.5, rl.NewColor(200, 150, 50, 200))
	}
}
