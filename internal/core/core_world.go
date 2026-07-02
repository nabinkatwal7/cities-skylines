package core

type World struct {
	SpatialIndex *QuadTree
	EventBus     *EventBus
	TimeSystem   *TimeSystem
	DeltaTime    float64
}

func NewWorld() *World {
	return &World{
		SpatialIndex: NewQuadTree(SpatialBounds{X: 0, Y: 0, W: 4096, H: 4096}, 4, 10),
		EventBus:     NewEventBus(),
		TimeSystem:   NewTimeSystem(),
	}
}

func (w *World) Update(delta float64) {
	w.DeltaTime = delta
	w.EventBus.Emit("world:update", delta)
}
