package terrain

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
	Connections []OutsideConnection
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
