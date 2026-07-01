package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type TransportType uint8

const (
	TransBus TransportType = iota
	TransTram
	TransMetro
	TransTrain
	TransFerry
)

type TransportStop struct {
	ID          uint32
	X, Z        float32
	Name        string
	TransType   TransportType
	Passengers  int32
	IsStation   bool
	Underground bool
}

type TransportLine struct {
	ID        uint32
	Name      string
	TransType TransportType
	Stops     []uint32
	Active    bool
	Color     rl.Color
}

type TransportVehicle struct {
	ID         uint32
	LineID     uint32
	TransType  TransportType
	X, Z       float32
	Speed      float32
	StopIdx    int
	Passengers int32
	Capacity   int32
	Forward    bool
	Timer      int32
	Moving     bool
	TargetX    float32
	TargetZ    float32
}

type TransportManager struct {
	Stops  []TransportStop
	Lines  []TransportLine
	Vehicles []TransportVehicle
	NextID uint32
}

func NewTransportManager() *TransportManager {
	return &TransportManager{}
}

func (tm *TransportManager) AddStop(x, z float32, tt TransportType) uint32 {
	id := tm.NextID
	tm.NextID++
	isStation := tt == TransMetro || tt == TransTrain
	tm.Stops = append(tm.Stops, TransportStop{
		ID:        id,
		X:         x, Z: z,
		Name:      "Stop",
		TransType: tt,
		IsStation: isStation,
		Underground: tt == TransMetro,
	})
	return id
}

func (tm *TransportManager) AddLine(name string, tt TransportType, stopIDs []uint32, col rl.Color) uint32 {
	id := tm.NextID
	tm.NextID++
	tm.Lines = append(tm.Lines, TransportLine{
		ID:        id,
		Name:      name,
		TransType: tt,
		Stops:     stopIDs,
		Active:    true,
		Color:     col,
	})
	return id
}

func (tm *TransportManager) SpawnVehicle(lineIdx int) {
	if lineIdx < 0 || lineIdx >= len(tm.Lines) {
		return
	}
	line := &tm.Lines[lineIdx]
	if len(line.Stops) < 2 {
		return
	}
	stop0 := &tm.Stops[line.Stops[0]]
	cap := int32(30)
	spd := float32(20)
	switch line.TransType {
	case TransTram:
		cap = 60
		spd = 25
	case TransMetro:
		cap = 120
		spd = 40
	case TransTrain:
		cap = 200
		spd = 60
	}
	tm.Vehicles = append(tm.Vehicles, TransportVehicle{
		ID:        tm.NextID,
		LineID:    line.ID,
		TransType: line.TransType,
		X:         stop0.X,
		Z:         stop0.Z,
		Capacity:  cap,
		Speed:     spd,
		Forward:   true,
		Moving:    true,
	})
	tm.NextID++
}

func (tm *TransportManager) Update(rm *RoadManager, h *Heightmap) {
	for li := range tm.Lines {
		line := &tm.Lines[li]
		if !line.Active {
			continue
		}
		count := 0
		for _, v := range tm.Vehicles {
			if v.LineID == line.ID {
				count++
			}
		}
		desired := 1
		switch line.TransType {
		case TransBus:
			desired = 2
		case TransMetro:
			desired = 3
		case TransTrain:
			desired = 2
		}
		if count < desired {
			tm.SpawnVehicle(li)
		}
	}

	for i := range tm.Vehicles {
		v := &tm.Vehicles[i]
		lineIdx := -1
		for li, line := range tm.Lines {
			if line.ID == v.LineID {
				lineIdx = li
				break
			}
		}
		if lineIdx < 0 {
			continue
		}
		line := &tm.Lines[lineIdx]
		if len(line.Stops) < 2 {
			continue
		}

		currentStop := &tm.Stops[line.Stops[v.StopIdx]]
		nextIdx := (v.StopIdx + 1) % len(line.Stops)
		if !v.Forward {
			nextIdx = (v.StopIdx - 1 + len(line.Stops)) % len(line.Stops)
		}
		nextStop := &tm.Stops[line.Stops[nextIdx]]

		if v.Moving {
			v.TargetX = nextStop.X
			v.TargetZ = nextStop.Z
		}

		dx := nextStop.X - v.X
		dz := nextStop.Z - v.Z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if dist < 3 && v.Moving {
			v.Moving = false
			v.Timer = 0
			v.StopIdx = nextIdx
			if v.StopIdx == 0 || v.StopIdx == len(line.Stops)-1 {
				v.Forward = !v.Forward
			}
			boarded := v.Capacity / 10
			if v.Passengers+boarded > v.Capacity {
				boarded = v.Capacity - v.Passengers
			}
			v.Passengers += boarded
			currentStop.Passengers -= boarded
			if currentStop.Passengers < 0 {
				currentStop.Passengers = 0
			}
		}

		if !v.Moving {
			v.Timer++
			wait := int32(60)
			if v.TransType == TransMetro || v.TransType == TransTrain {
				wait = 90
			}
			if v.Timer > wait {
				v.Moving = true
				currentStop.Passengers += 2
			}
			continue
		}

		if dist > 0.5 {
			v.X += (dx / dist) * v.Speed * 0.02
			v.Z += (dz / dist) * v.Speed * 0.02
		}
	}

	for si := range tm.Stops {
		s := &tm.Stops[si]
		if s.Passengers < 5 {
			s.Passengers++
		}
	}
}

func (tm *TransportManager) Draw(h *Heightmap) {
	for _, s := range tm.Stops {
		hy := h.WorldHeight(s.X, s.Z) + 0.5
		if s.TransType == TransMetro {
			if s.Underground {
				rl.DrawCube(rl.NewVector3(s.X, 0, s.Z), 3, 0.5, 3, rl.NewColor(100, 100, 200, 200))
				rl.DrawCube(rl.NewVector3(s.X, 1.5, s.Z), 1, 3, 1, rl.NewColor(150, 150, 220, 200))
			}
		} else if s.TransType == TransTram {
			rl.DrawCube(rl.NewVector3(s.X, hy, s.Z), 1.5, 0.5, 1.5, rl.NewColor(200, 100, 200, 200))
		} else {
			rl.DrawCube(rl.NewVector3(s.X, hy, s.Z), 1, 1, 1, rl.NewColor(0, 150, 200, 200))
		}
	}

	for _, v := range tm.Vehicles {
		hy := h.WorldHeight(v.X, v.Z) + 0.6
		switch v.TransType {
		case TransBus:
			rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 2.5, 0.8, 1.2, rl.NewColor(0, 100, 200, 255))
			rl.DrawCube(rl.NewVector3(v.X, hy+0.5, v.Z+0.8), 1.5, 0.3, 0.1, rl.NewColor(200, 200, 100, 255))
		case TransTram:
			rl.DrawCube(rl.NewVector3(v.X, hy, v.Z), 4, 0.6, 1, rl.NewColor(200, 50, 200, 255))
			rl.DrawCube(rl.NewVector3(v.X, hy+0.4, v.Z+0.6), 3, 0.2, 0.1, rl.NewColor(255, 255, 100, 255))
		case TransMetro:
			rl.DrawCube(rl.NewVector3(v.X, 0.5, v.Z), 5, 1, 1.5, rl.NewColor(100, 100, 220, 255))
			rl.DrawCube(rl.NewVector3(v.X, 1, v.Z), 4.5, 0.3, 1.3, rl.NewColor(200, 200, 255, 255))
		}
	}
}

func (tm *TransportManager) Unload() {
	tm.Vehicles = nil
	tm.Lines = nil
	tm.Stops = nil
}
