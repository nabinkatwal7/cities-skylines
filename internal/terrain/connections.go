package terrain

import "math/rand"

type ConnectionType int

const (
	ConnHighway ConnectionType = 0
	ConnRail    ConnectionType = 1
	ConnShip    ConnectionType = 2
	ConnAir     ConnectionType = 3
)

type OutsideConnection struct {
	Type     ConnectionType
	WorldX   float32
	WorldZ   float32
	Active   bool
}

type ConnectionSystem struct {
	Connections  []OutsideConnection
	Immigration  int32
	Tourism      int32
	Imports      int32
	Exports      int32
}

func NewConnectionSystem() *ConnectionSystem {
	cs := &ConnectionSystem{}

	half := float32(WorldSize / 2)

	cs.Connections = []OutsideConnection{
		{Type: ConnHighway, WorldX: -half + 10, WorldZ: -half, Active: true},
		{Type: ConnHighway, WorldX: half - 10, WorldZ: -half, Active: true},
		{Type: ConnHighway, WorldX: -half, WorldZ: -half + 10, Active: true},
		{Type: ConnHighway, WorldX: -half, WorldZ: half - 10, Active: true},
		{Type: ConnRail, WorldX: -half + 20, WorldZ: -half, Active: true},
		{Type: ConnRail, WorldX: half - 20, WorldZ: -half, Active: true},
		{Type: ConnShip, WorldX: 0, WorldZ: -half, Active: true},
		{Type: ConnAir, WorldX: half*0.6, WorldZ: -half*0.6, Active: true},
	}

	return cs
}

func (cs *ConnectionSystem) GetByType(t ConnectionType) []OutsideConnection {
	var result []OutsideConnection
	for _, c := range cs.Connections {
		if c.Type == t && c.Active {
			result = append(result, c)
		}
	}
	return result
}

func (cs *ConnectionSystem) Update() {
	if rand.Intn(120) == 0 {
		cs.Immigration += int32(rand.Intn(5))
	}
	if rand.Intn(180) == 0 {
		cs.Tourism += int32(rand.Intn(3))
	}
	if rand.Intn(240) == 0 {
		cs.Imports += int32(rand.Intn(10))
	}
	if rand.Intn(300) == 0 {
		cs.Exports += int32(rand.Intn(8))
	}
}
