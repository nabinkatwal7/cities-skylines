package sim

import (
	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/zoning"
)

const undoLimit = 32

// UndoStack stores limited construction undo history (24.20).
type UndoStack struct {
	entries []func(*SimulationManager)
}

func NewUndoStack() *UndoStack { return &UndoStack{} }

func (u *UndoStack) Push(fn func(*SimulationManager)) {
	if fn == nil {
		return
	}
	u.entries = append(u.entries, fn)
	if len(u.entries) > undoLimit {
		u.entries = u.entries[len(u.entries)-undoLimit:]
	}
}

func (u *UndoStack) Undo(sm *SimulationManager) bool {
	if sm == nil || len(u.entries) == 0 {
		return false
	}
	i := len(u.entries) - 1
	fn := u.entries[i]
	u.entries = u.entries[:i]
	fn(sm)
	return true
}

func (u *UndoStack) Clear() { u.entries = nil }

func (u *UndoStack) CanUndo() bool { return len(u.entries) > 0 }

func (sm *SimulationManager) PushRoadPlace(segID uint32, cost float32) {
	if sm == nil || sm.Undo == nil || segID == 0 {
		return
	}
	id := segID
	spent := cost
	sm.Undo.Push(func(s *SimulationManager) {
		idx := segmentIndexByID(s.Roads, id)
		if idx < 0 {
			return
		}
		s.RemoveSegment(idx)
		s.Money += spent
	})
}

type roadSegmentSnap struct {
	seg   road.RoadSegment
	nodeA road.RoadNode
	nodeB road.RoadNode
}

func (sm *SimulationManager) PushRoadRemove(idx int) {
	if sm == nil || sm.Undo == nil || idx < 0 || idx >= len(sm.Roads.Segments) {
		return
	}
	snap := roadSegmentSnap{
		seg:   sm.Roads.Segments[idx],
		nodeA: sm.Roads.Nodes[sm.Roads.Segments[idx].NodeA],
		nodeB: sm.Roads.Nodes[sm.Roads.Segments[idx].NodeB],
	}
	sm.RemoveSegment(idx)
	sm.Undo.Push(func(s *SimulationManager) {
		restoreRoadSegment(s, snap)
	})
}

func restoreRoadSegment(sm *SimulationManager, snap roadSegmentSnap) {
	if sm == nil || sm.Roads == nil {
		return
	}
	a := ensureNode(sm.Roads, snap.nodeA)
	b := ensureNode(sm.Roads, snap.nodeB)
	segID := sm.Roads.AddSegment(a, b, snap.seg.RoadType)
	for i := range sm.Roads.Segments {
		if sm.Roads.Segments[i].ID != segID {
			continue
		}
		sm.Roads.Segments[i].Elevation = snap.seg.Elevation
		sm.Roads.Segments[i].MaintenanceCost = snap.seg.MaintenanceCost
		sm.Roads.Segments[i].Direction = snap.seg.Direction
		break
	}
	sm.Roads.Rebuild(sm.Heightmap)
	sm.EventBus.Emit(string(core.EventRoadPlaced), segID)
}

func ensureNode(rm *road.RoadManager, want road.RoadNode) uint32 {
	if idx := nodeIndexByID(rm, want.ID); idx >= 0 {
		return uint32(idx)
	}
	idx := rm.AddNode(want.X, want.Y, want.Z)
	rm.Nodes[idx].Flags = want.Flags
	return idx
}

func segmentIndexByID(rm *road.RoadManager, id uint32) int {
	for i := range rm.Segments {
		if rm.Segments[i].ID == id {
			return i
		}
	}
	return -1
}

func nodeIndexByID(rm *road.RoadManager, id uint32) int {
	for i := range rm.Nodes {
		if rm.Nodes[i].ID == id {
			return i
		}
	}
	return -1
}

func (sm *SimulationManager) PushZoneChange(x, z int, before zoning.ZoneCell) {
	if sm == nil || sm.Undo == nil || sm.Zones == nil {
		return
	}
	prev := before
	sm.Undo.Push(func(s *SimulationManager) {
		if s.Zones != nil {
			s.Zones.RestoreCell(x, z, prev)
			s.EventBus.Emit(string(core.EventZonePlaced), prev.Type)
		}
	})
}

type treeSnap struct {
	idx int
	t   terrain.Tree
}

func (sm *SimulationManager) PushTreeRemove(worldX, worldZ, radius float32) int {
	if sm == nil || sm.Trees == nil {
		return 0
	}
	snaps := captureTrees(sm.Trees, worldX, worldZ, radius)
	if len(snaps) == 0 {
		return 0
	}
	moneyBefore := sm.Money
	removed := sm.RemoveTrees(worldX, worldZ, radius)
	if removed == 0 || sm.Undo == nil {
		return removed
	}
	cost := moneyBefore - sm.Money
	copySnaps := append([]treeSnap(nil), snaps...)
	sm.Undo.Push(func(s *SimulationManager) {
		for _, snap := range copySnaps {
			if snap.idx >= 0 && snap.idx < terrain.TreePoolSize {
				s.Trees.Pool[snap.idx] = snap.t
			}
		}
		s.Money += cost
	})
	return removed
}

func captureTrees(ts *terrain.TreeSystem, x, z, radius float32) []treeSnap {
	if ts == nil {
		return nil
	}
	r2 := radius * radius
	out := make([]treeSnap, 0, 4)
	for i := 0; i < terrain.TreePoolSize; i++ {
		t := &ts.Pool[i]
		if t.Lifecycle != core.LifecycleActive {
			continue
		}
		dx := t.X - x
		dz := t.Z - z
		if dx*dx+dz*dz <= r2 {
			out = append(out, treeSnap{idx: i, t: *t})
		}
	}
	return out
}
