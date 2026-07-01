package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type BusStop struct {
	ID        uint32
	X, Z      float32
	Name      string
	Passengers int32
}

type BusLine struct {
	ID     uint32
	Name   string
	Stops  []uint32
	Active bool
	Color  rl.Color
}

type BusVehicle struct {
	ID       uint32
	LineID   uint32
	X, Z     float32
	Speed    float32
	StopIdx  int
	Passengers int32
	Capacity int32
	Forward  bool
	Timer    int32
	Moving   bool
}

type TransportManager struct {
	Stops    []BusStop
	Lines    []BusLine
	Buses    []BusVehicle
	NextID   uint32
}

func NewTransportManager() *TransportManager {
	return &TransportManager{}
}

func (tm *TransportManager) AddStop(x, z float32) uint32 {
	id := tm.NextID
	tm.NextID++
	tm.Stops = append(tm.Stops, BusStop{
		ID: id,
		X:  x, Z: z,
		Name:      "Bus Stop",
	})
	return id
}

func (tm *TransportManager) AddLine(name string, stopIDs []uint32, col rl.Color) uint32 {
	id := tm.NextID
	tm.NextID++
	tm.Lines = append(tm.Lines, BusLine{
		ID:     id,
		Name:   name,
		Stops:  stopIDs,
		Active: true,
		Color:  col,
	})
	return id
}

func (tm *TransportManager) SpawnBus(lineIdx int) {
	if lineIdx < 0 || lineIdx >= len(tm.Lines) {
		return
	}
	line := &tm.Lines[lineIdx]
	if len(line.Stops) < 2 {
		return
	}
	stop0 := &tm.Stops[line.Stops[0]]
	tm.Buses = append(tm.Buses, BusVehicle{
		ID:       tm.NextID,
		LineID:   line.ID,
		X:        stop0.X,
		Z:        stop0.Z,
		Capacity: 30,
		Forward:  true,
		Moving:   true,
	})
	tm.NextID++
}

func (tm *TransportManager) Update(rm *RoadManager, h *Heightmap) {
	for li := range tm.Lines {
		line := &tm.Lines[li]
		if !line.Active {
			continue
		}
		busCount := 0
		for _, b := range tm.Buses {
			if b.LineID == line.ID {
				busCount++
			}
		}
		if busCount < 2 {
			tm.SpawnBus(li)
		}
	}

	for i := range tm.Buses {
		b := &tm.Buses[i]
		lineIdx := -1
		for li, line := range tm.Lines {
			if line.ID == b.LineID {
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

		currentStop := &tm.Stops[line.Stops[b.StopIdx]]
		nextIdx := (b.StopIdx + 1) % len(line.Stops)
		if !b.Forward {
			nextIdx = (b.StopIdx - 1 + len(line.Stops)) % len(line.Stops)
		}
		nextStop := &tm.Stops[line.Stops[nextIdx]]

		dx := nextStop.X - b.X
		dz := nextStop.Z - b.Z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if dist < 3 && b.Moving {
			b.Moving = false
			b.Timer = 0
			b.StopIdx = nextIdx
			if b.StopIdx == 0 || b.StopIdx == len(line.Stops)-1 {
				b.Forward = !b.Forward
			}
			boarded := int32(3)
			if b.Passengers+boarded > b.Capacity {
				boarded = b.Capacity - b.Passengers
			}
			b.Passengers += boarded
			currentStop.Passengers -= boarded
			if currentStop.Passengers < 0 {
				currentStop.Passengers = 0
			}
		}

		if !b.Moving {
			b.Timer++
			if b.Timer > 120 {
				b.Moving = true
				currentStop.Passengers += int32(math.Max(float64(currentStop.Passengers), 2))
			}
			continue
		}

		b.Speed = 20
		if dist > 0.5 {
			b.X += (dx / dist) * b.Speed * 0.02
			b.Z += (dz / dist) * b.Speed * 0.02
		}
	}

	for si := range tm.Stops {
		s := &tm.Stops[si]
		if s.Passengers < 3 {
			s.Passengers++
		}
	}
}

func (tm *TransportManager) Draw(h *Heightmap) {
	for _, s := range tm.Stops {
		hy := h.WorldHeight(s.X, s.Z) + 0.5
		rl.DrawCube(rl.NewVector3(s.X, hy, s.Z), 1, 1, 1, rl.NewColor(0, 150, 200, 200))
		rl.DrawCube(rl.NewVector3(s.X, hy+0.8, s.Z), 0.2, 0.8, 0.2, rl.NewColor(200, 200, 200, 200))
	}

	for _, b := range tm.Buses {
		hy := h.WorldHeight(b.X, b.Z) + 0.6
		rl.DrawCube(rl.NewVector3(b.X, hy, b.Z), 2.5, 0.8, 1.2, rl.NewColor(0, 100, 200, 255))
		rl.DrawCube(rl.NewVector3(b.X, hy+0.5, b.Z+0.8), 1.5, 0.3, 0.1, rl.NewColor(200, 200, 100, 255))
	}
}

func (tm *TransportManager) Unload() {
	tm.Buses = nil
	tm.Lines = nil
	tm.Stops = nil
}
