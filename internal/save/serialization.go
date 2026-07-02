package save

import (
	"bytes"

	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"

	"github.com/katwate/js-skylines/internal/building"
	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/transport"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SaveHeader struct {
	Checksum uint32
	_        [4]byte
}

type SaveData struct {
	Version  int32
	Seed     int64
	Money    float32
	TimeOfDay int32
	Night    bool

	TerrainHeight [terrain.HeightmapSize][terrain.HeightmapSize]float32
	Trees         []TreeData
	Resources     [terrain.HeightmapSize][terrain.HeightmapSize]ResourceCell

	RoadNodes    []RoadNodeData
	RoadSegments []RoadSegmentData

	TransportStops    []TransportStopData
	TransportLines    []TransportLineData
	TransportNetworks []TransportNetworkData
	TransportVehicles []TransportVehicleData

	ParkingSpots []ParkingSpotData
	ParkingLots  []ParkingLotData

	MetroTracks []MetroTrackData
	RailTracks  []RailTrackData

	BusDepots    []DepotData
	TramDepots   []DepotData
	MetroDepots  []DepotData
	FerryDepots  []DepotData
	MonorailDepots  []DepotData
	CableCarDepots  []DepotData
	TaxiDepots      []DepotData
	AirportDepots   []DepotData
	PortDepots      []DepotData

	CableConnections []CableConnectionData
	CargoStations    []CargoStationData
	CargoTrains      []TransportVehicleData
	TaxiFleet        []TransportVehicleData
	TaxiRequests     []TaxiRequestData
	TransferStations []TransferStationData

	Zoning building.SaveZoning
}

type RoadNodeData struct {
	X, Y, Z         float32
	TrafficLight    road.TrafficLightState
	JunctionType    uint8
	Flags           road.RoadFlags
	TrafficLightPhase int32
	HasTrafficLight bool
	IsOutsideConn   bool
}

type RoadSegmentData struct {
	NodeA            uint32
	NodeB            uint32
	RoadType         road.RoadType
	Length           float32
	SpeedLimit       float32
	LaneCount        int32
	Direction        int8
	Elevation        int32
	MaintenanceCost  float32
	ConstructionCost float32
	Damaged          bool
	RepairTimer      int32
	CurveP1x         float32
	CurveP1z         float32
	CurveP2x         float32
	CurveP2z         float32
	Lanes            []LaneData
}

type LaneData struct {
	Index      int32
	Direction  int8
	SpeedLimit float32
	Category   road.LaneCategory
	Width      float32
	Priority   int32
}

type TransportStopData struct {
	X, Z              float32
	TransType         transport.TransportType
	ConnectedNetworks []transport.TransportType
	Passengers        int32
	IsStation         bool
	Underground       bool
	DistrictID        int32
	Accessibility     float32
	Capacity          int32
	TransferStationID uint32
	Name              string
}

type TransportLineData struct {
	TransType      transport.TransportType
	StopIDs        []uint32
	Active         bool
	ColorR         uint8
	ColorG         uint8
	ColorB         uint8
	VehicleCount   int32
	PassengerCount int32
	Budget         float32
	TotalPassengers int64
	TotalIncome    float32
	IsCircular     bool
}

type TransportNetworkData struct {
	Type             transport.TransportType
	Active           bool
	VehicleCount     int32
	RouteCount       int32
	StopCount        int32
	StationCount     int32
	PassengersPerDay int32
	TotalIncome      float32
	MaintenanceCost  float32
	Capacity         int32
	Pollution        float32
	Noise            float32
	WeeklyPassengers  int32
	LifetimePassengers int64
	TotalExpenses    float32
	AvgWaitTime      float32
	CapacityUsage    float32
	VehicleUtilization float32
}

type TransferStationData struct {
	ID      uint32
	Name    string
	StopIDs []uint32
	X       float32
	Z       float32
}

type TransportVehicleData struct {
	ID         uint32
	LineID     uint32
	TransType  transport.TransportType
	X, Z       float32
	Speed      float32
	StopIdx    int32
	Passengers int32
	Capacity   int32
	Forward    bool
	Moving     bool
	TargetX    float32
	TargetZ    float32
	StandingCapacity int32
	HomeDepotType    uint8
	HomeDepotSlot    int32
	MaintenanceTimer int32
}

type ParkingSpotData struct {
	ID       uint32
	SpotType road.ParkingSpotType
	X, Z     float32
	RoadSeg  int32
	LaneIdx  int32
	Occupied bool
	OwnerSlot int32
}

type ParkingLotData struct {
	X, Z        float32
	LotType     road.ParkingSpotType
	Width, Depth float32
	Capacity    int32
	CellX, CellZ int
	SpotIndices []int32
}

type TreeData struct {
	X, Z    float32
	Species int
	Scale   float32
}

type MetroTrackData struct {
	ID         uint32
	StartX, StartZ float32
	EndX, EndZ float32
	Length     float32
	Elevated   bool
}

type RailTrackData struct {
	ID         uint32
	StartX, StartZ float32
	EndX, EndZ float32
	Length     float32
	SignalA    transport.SignalState
	SignalB    transport.SignalState
	BlockID    int32
	OutsideA   bool
	OutsideB   bool
	Occupied   bool
	OccupierID uint32
}

type DepotData struct {
	X, Z        float32
	CellX, CellZ int
}

type CableConnectionData struct {
	ID         uint32
	StopA, StopB uint32
	StartX, StartZ float32
	EndX, EndZ float32
}

type CargoStationData struct {
	ID          uint32
	X, Z        float32
	Name        string
	GoodsStored int32
	Capacity    int32
	Active      bool
	TrainSlot   int32
}

type TaxiRequestData struct {
	ID        uint32
	CitizenX, CitizenZ float32
	DstX, DstZ float32
	Active    bool
	Assigned  bool
}

type ResourceCell struct {
	Fertility float32
	Ore       float32
	Oil       float32
}

type SaveStats struct {
	DirtyCount int32
	TotalCount int32
}

const currentSaveVersion int32 = 12

func calcChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

func migrateSaveData(data *SaveData) bool {
	if data.Version >= currentSaveVersion {
		return false
	}
	if data.Version < 1 {
		return false
	}
	for v := data.Version; v < currentSaveVersion; v++ {
		switch v {
		case 1:
			data.Version = 2
		case 2:
			for i := range data.RoadNodes {
				nd := &data.RoadNodes[i]
				if nd.Flags == 0 {
					if nd.HasTrafficLight {
						nd.TrafficLight = road.TrafficLightRed
					}
					if nd.IsOutsideConn {
						nd.Flags |= road.RoadFlagOutsideConn
					}
				}
			}
			for i := range data.RoadSegments {
				sd := &data.RoadSegments[i]
				if sd.SpeedLimit == 0 {
					sd.SpeedLimit = road.DefaultSpeed(sd.RoadType)
				}
				if sd.LaneCount == 0 {
					sd.LaneCount = int32(road.RoadLanes(sd.RoadType))
				}
			}
			data.Version = 3
		case 3:
			for i := range data.RoadSegments {
				sd := &data.RoadSegments[i]
				if len(sd.Lanes) == 0 {
					categories := road.LaneCategoriesForRoad(sd.RoadType)
					sd.Lanes = make([]LaneData, sd.LaneCount)
					half := sd.LaneCount / 2
					for li := int32(0); li < sd.LaneCount; li++ {
						var dir int8
						if sd.Direction == 1 {
							dir = 0
						} else if sd.Direction == -1 {
							dir = 1
						} else {
							if li < half {
								dir = 0
							} else {
								dir = 1
							}
						}
						cat := road.LaneDriving
						if int(li) < len(categories) {
							cat = categories[li]
						}
						sd.Lanes[li] = LaneData{
							Index:       li,
							Direction:   dir,
							SpeedLimit:  sd.SpeedLimit,
							Category:    cat,
							Width:       3.0,
							Priority:    li,
						}
					}
				}
			}
			data.Version = 4
		case 4:
			data.Version = 5
		case 5:
			for i := range data.RoadNodes {
				if data.RoadNodes[i].TrafficLight != road.TrafficLightNone && data.RoadNodes[i].TrafficLightPhase == 0 {
					maxLanes := int32(2)
					data.RoadNodes[i].TrafficLightPhase = maxLanes * 20
				}
			}
			data.Version = 6
		case 6:
			for i := range data.TransportLines {
				ld := &data.TransportLines[i]
				if ld.VehicleCount == 0 && ld.TotalPassengers == 0 {
					ld.IsCircular = ld.TransType != transport.TransBus && ld.TransType != transport.TransTram && ld.TransType != transport.TransTaxi
				}
			}
			data.Version = 7
		case 7:
			for i := range data.TransportStops {
				sd := &data.TransportStops[i]
				if len(sd.ConnectedNetworks) == 0 {
					sd.ConnectedNetworks = []transport.TransportType{sd.TransType}
				}
				if sd.Capacity <= 0 {
					sd.Capacity = 50
				}
				sd.DistrictID = -1
				sd.Accessibility = 0.5
			}
			data.Version = 8
		case 8:
			data.Version = 9
		case 9:
			data.Version = 10
		case 10:
			data.Version = 11
		case 11:
			data.Version = 12
		}
	}
	return true
}

func writeSaveFile(filename string, data *SaveData) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	payload := buf.Bytes()
	csum := calcChecksum(payload)

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer f.Close()

	header := SaveHeader{Checksum: csum}
	if err := binary.Write(f, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := f.Write(payload); err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	return nil
}

func readSaveFile(filename string) (*SaveData, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var header SaveHeader
	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	payload, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	if calcChecksum(payload) != header.Checksum {
		return nil, fmt.Errorf("checksum mismatch: file may be corrupted")
	}

	var data SaveData
	dec := gob.NewDecoder(bytes.NewReader(payload))
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &data, nil
}

func SaveGame(filename string, m *sim.SimulationManager, money float32, timeOfDay int32) error {
	data := SaveData{
		Version:  currentSaveVersion,
		Seed:     m.Seed,
		Money:    money,
		TimeOfDay: timeOfDay,
		Night:    m.Night,
	}

	data.TerrainHeight = m.Heightmap.Data

	if m.Trees != nil {
		m.Trees.ForEach(func(t *terrain.Tree, _ int32) {
			data.Trees = append(data.Trees, TreeData{X: t.X, Z: t.Z, Species: int(t.Species), Scale: t.Scale})
		})
	}

	if m.Resources != nil {
		for z := 0; z < terrain.HeightmapSize; z++ {
			for x := 0; x < terrain.HeightmapSize; x++ {
				data.Resources[z][x] = ResourceCell{
					Fertility: m.Resources.Map.Fertility[z][x],
					Ore:       m.Resources.Map.Ore[z][x],
					Oil:       m.Resources.Map.Oil[z][x],
				}
			}
		}
	}

	if m.Roads != nil {
		for _, n := range m.Roads.Nodes {
			data.RoadNodes = append(data.RoadNodes, RoadNodeData{
				X: n.X, Y: n.Y, Z: n.Z,
				TrafficLight:    n.TrafficLight,
				JunctionType:    n.JunctionType,
				Flags:           n.Flags,
				TrafficLightPhase: n.TrafficLightPhase,
			})
		}
		for _, s := range m.Roads.Segments {
			lanesData := make([]LaneData, len(s.Lanes))
			for li, l := range s.Lanes {
				lanesData[li] = LaneData{
					Index:       l.Index,
					Direction:   l.Direction,
					SpeedLimit:  l.SpeedLimit,
					Category:    l.Category,
					Width:       l.Width,
					Priority:    l.Priority,
				}
			}
			data.RoadSegments = append(data.RoadSegments, RoadSegmentData{
				NodeA:            s.NodeA,
				NodeB:            s.NodeB,
				RoadType:         s.RoadType,
				Length:           s.Length,
				SpeedLimit:       s.SpeedLimit,
				LaneCount:        s.LaneCount,
				Direction:        s.Direction,
				Elevation:        s.Elevation,
				MaintenanceCost:  s.MaintenanceCost,
				ConstructionCost: s.ConstructionCost,
				Damaged:          s.Damaged,
				RepairTimer:      s.RepairTimer,
				CurveP1x:         s.Curve.P1x,
				CurveP1z:         s.Curve.P1z,
				CurveP2x:         s.Curve.P2x,
				CurveP2z:         s.Curve.P2z,
				Lanes:            lanesData,
			})
		}
	}

	if m.Transport != nil {
		for _, s := range m.Transport.Stops {
			copyNet := make([]transport.TransportType, len(s.ConnectedNetworks))
			copy(copyNet, s.ConnectedNetworks)
			data.TransportStops = append(data.TransportStops, TransportStopData{
				X:                 s.X,
				Z:                 s.Z,
				TransType:         s.TransType,
				ConnectedNetworks: copyNet,
				Passengers:        s.Passengers,
				IsStation:         s.IsStation,
				Underground:       s.Underground,
				DistrictID:        s.DistrictID,
				Accessibility:     s.Accessibility,
				Capacity:          s.Capacity,
				TransferStationID: s.TransferStationID,
				Name:              s.Name,
			})
		}
		for _, l := range m.Transport.Lines {
			col := l.Color
			data.TransportLines = append(data.TransportLines, TransportLineData{
				TransType:       l.TransType,
				StopIDs:         l.Stops,
				Active:          l.Active,
				ColorR:          col.R,
				ColorG:          col.G,
				ColorB:          col.B,
				VehicleCount:    l.VehicleCount,
				PassengerCount:  l.PassengerCount,
				Budget:          l.Budget,
				TotalPassengers: l.TotalPassengers,
				TotalIncome:     l.TotalIncome,
				IsCircular:      l.IsCircular,
			})
		}
		for _, n := range m.Transport.Networks {
			data.TransportNetworks = append(data.TransportNetworks, TransportNetworkData{
				Type:             n.Type,
				Active:           n.Active,
				VehicleCount:     n.VehicleCount,
				RouteCount:       n.RouteCount,
				StopCount:        n.StopCount,
				StationCount:     n.StationCount,
				PassengersPerDay: n.PassengersPerDay,
				TotalIncome:      n.TotalIncome,
				MaintenanceCost:  n.MaintenanceCost,
				Capacity:         n.Capacity,
				Pollution:        n.Pollution,
				Noise:            n.Noise,
				WeeklyPassengers:  n.WeeklyPassengers,
				LifetimePassengers: n.LifetimePassengers,
				TotalExpenses:    n.TotalExpenses,
				AvgWaitTime:      n.AvgWaitTime,
				CapacityUsage:    n.CapacityUsage,
				VehicleUtilization: n.VehicleUtilization,
			})
		}
		for _, v := range m.Transport.Vehicles {
			data.TransportVehicles = append(data.TransportVehicles, TransportVehicleData{
				ID:               v.ID,
				LineID:           v.LineID,
				TransType:        v.TransType,
				X:                v.X,
				Z:                v.Z,
				Speed:            v.Speed,
				StopIdx:          int32(v.StopIdx),
				Passengers:       v.Passengers,
				Capacity:         v.Capacity,
				Forward:          v.Forward,
				Moving:           v.Moving,
				TargetX:          v.TargetX,
				TargetZ:          v.TargetZ,
				StandingCapacity: v.StandingCapacity,
				HomeDepotType:    v.HomeDepotType,
				HomeDepotSlot:    v.HomeDepotSlot,
				MaintenanceTimer: v.MaintenanceTimer,
			})
		}
		for _, v := range m.Transport.Pool {
			if v.ID == math.MaxUint32 {
				continue
			}
			data.TransportVehicles = append(data.TransportVehicles, TransportVehicleData{
				ID:               v.ID,
				LineID:           v.LineID,
				TransType:        v.TransType,
				X:                v.X,
				Z:                v.Z,
				Speed:            v.Speed,
				StopIdx:          int32(v.StopIdx),
				Passengers:       v.Passengers,
				Capacity:         v.Capacity,
				Forward:          v.Forward,
				Moving:           v.Moving,
				TargetX:          v.TargetX,
				TargetZ:          v.TargetZ,
				StandingCapacity: v.StandingCapacity,
				HomeDepotType:    v.HomeDepotType,
				HomeDepotSlot:    v.HomeDepotSlot,
				MaintenanceTimer: v.MaintenanceTimer,
			})
		}
	}

	if m.Parking != nil {
		for _, sp := range m.Parking.Spots {
			data.ParkingSpots = append(data.ParkingSpots, ParkingSpotData{
				ID:        sp.ID,
				SpotType:  sp.SpotType,
				X:         sp.X,
				Z:         sp.Z,
				RoadSeg:   sp.RoadSeg,
				LaneIdx:   sp.LaneIdx,
				Occupied:  sp.Occupied,
				OwnerSlot: sp.OwnerSlot,
			})
		}
		m.Parking.ForEachLot(func(lot *road.ParkingLot, slot int32) {
			spotIndices := make([]int32, len(lot.Spots))
			copy(spotIndices, lot.Spots)
			data.ParkingLots = append(data.ParkingLots, ParkingLotData{
				X:           lot.Position.X,
				Z:           lot.Position.Z,
				LotType:     lot.LotType,
				Width:       lot.Width,
				Depth:       lot.Depth,
				Capacity:    lot.Capacity,
				CellX:       lot.CellX,
				CellZ:       lot.CellZ,
				SpotIndices: spotIndices,
			})
		})
	}

	if m.Transport != nil {
		if m.Transport.Tracks != nil {
			m.Transport.Tracks.ForEachTrack(func(t *transport.MetroTrack, slot int32) {
				data.MetroTracks = append(data.MetroTracks, MetroTrackData{
					ID: t.ID, StartX: t.StartX, StartZ: t.StartZ,
					EndX: t.EndX, EndZ: t.EndZ, Length: t.Length, Elevated: t.Elevated,
				})
			})
		}
		if m.Transport.Rails != nil {
			m.Transport.Rails.ForEachTrack(func(t *transport.RailTrack, slot int32) {
				data.RailTracks = append(data.RailTracks, RailTrackData{
					ID: t.ID, StartX: t.StartX, StartZ: t.StartZ,
					EndX: t.EndX, EndZ: t.EndZ, Length: t.Length,
					SignalA: t.SignalA, SignalB: t.SignalB, BlockID: t.BlockID,
					OutsideA: t.OutsideA, OutsideB: t.OutsideB,
					Occupied: t.Occupied, OccupierID: t.OccupierID,
				})
			})
		}
		if m.Transport.Cargo != nil {
			for _, cs := range m.Transport.Cargo.Stations {
				data.CargoStations = append(data.CargoStations, CargoStationData{
					ID: cs.ID, X: cs.X, Z: cs.Z, Name: cs.Name,
					GoodsStored: cs.GoodsStored, Capacity: cs.Capacity,
					Active: cs.Active, TrainSlot: cs.TrainSlot,
				})
			}
			for _, tv := range m.Transport.Cargo.Trains {
				data.CargoTrains = append(data.CargoTrains, TransportVehicleData{
					ID: tv.ID, TransType: tv.TransType,
					X: tv.X, Z: tv.Z, Speed: tv.Speed,
					Moving: tv.Moving,
					TargetX: tv.TargetX, TargetZ: tv.TargetZ,
				})
			}
		}
		for _, cc := range m.Transport.CableConnections {
			data.CableConnections = append(data.CableConnections, CableConnectionData{
				ID: cc.ID, StopA: cc.StopA, StopB: cc.StopB,
				StartX: cc.StartX, StartZ: cc.StartZ,
				EndX: cc.EndX, EndZ: cc.EndZ,
			})
		}
		for _, ts := range m.Transport.TransferStations {
			stopIDs := make([]uint32, len(ts.StopIDs))
			copy(stopIDs, ts.StopIDs)
			data.TransferStations = append(data.TransferStations, TransferStationData{
				ID: ts.ID, Name: ts.Name,
				StopIDs: stopIDs,
				X: ts.X, Z: ts.Z,
			})
		}
	}

	if m.Parking != nil {
		for i := 0; i < road.BusDepotPoolSize; i++ {
			if m.Parking.BusDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.BusDepots[i]
				data.BusDepots = append(data.BusDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.TramDepotPoolSize; i++ {
			if m.Parking.TramDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.TramDepots[i]
				data.TramDepots = append(data.TramDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.MetroDepotPoolSize; i++ {
			if m.Parking.MetroDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.MetroDepots[i]
				data.MetroDepots = append(data.MetroDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.FerryDepotPoolSize; i++ {
			if m.Parking.FerryDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.FerryDepots[i]
				data.FerryDepots = append(data.FerryDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.MonorailDepotPoolSize; i++ {
			if m.Parking.MonorailDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.MonorailDepots[i]
				data.MonorailDepots = append(data.MonorailDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.CableCarDepotPoolSize; i++ {
			if m.Parking.CableCarDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.CableCarDepots[i]
				data.CableCarDepots = append(data.CableCarDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.TaxiDepotPoolSize; i++ {
			if m.Parking.TaxiDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.TaxiDepots[i]
				data.TaxiDepots = append(data.TaxiDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.AirportDepotPoolSize; i++ {
			if m.Parking.AirportDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.AirportDepots[i]
				data.AirportDepots = append(data.AirportDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for i := 0; i < road.PortDepotPoolSize; i++ {
			if m.Parking.PortDepots[i].Lifecycle == core.LifecycleAllocated {
				d := &m.Parking.PortDepots[i]
				data.PortDepots = append(data.PortDepots, DepotData{X: d.X, Z: d.Z, CellX: d.CellX, CellZ: d.CellZ})
			}
		}
		for _, tv := range m.Parking.TaxiFleet {
			data.TaxiFleet = append(data.TaxiFleet, TransportVehicleData{
				X: tv.X, Z: tv.Z, Speed: tv.Speed,
				Moving: tv.Moving,
				TargetX: tv.TargetX, TargetZ: tv.TargetZ,
			})
		}
		for _, r := range m.Parking.TaxiRequests {
			data.TaxiRequests = append(data.TaxiRequests, TaxiRequestData{
				ID: r.ID, CitizenX: r.CitizenX, CitizenZ: r.CitizenZ,
				DstX: r.DstX, DstZ: r.DstZ, Active: r.Active, Assigned: r.Assigned,
			})
		}
	}

	if m.Zones != nil {
		data.Zoning = building.ExportZoning(m.Zones, m.Buildings, m.Demand)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create save: %w", err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode save: %w", err)
	}
	return nil
}

func LoadGame(filename string, m *sim.SimulationManager) (money float32, timeOfDay int32, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, 0, fmt.Errorf("open save: %w", err)
	}
	defer f.Close()

	var data SaveData
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&data); err != nil {
		return 0, 0, fmt.Errorf("decode save: %w", err)
	}

	if data.Version < 1 || data.Version > currentSaveVersion {
		return 0, 0, fmt.Errorf("unsupported save version: %d", data.Version)
	}
	if m.Undo != nil {
		m.Undo.Clear()
	}

	m.Seed = data.Seed
	m.Night = data.Night

	m.Heightmap.Data = data.TerrainHeight

	if m.Trees != nil {
		for _, td := range data.Trees {
			slot := m.Trees.Alloc()
			if slot < 0 {
				continue
			}
			t := &m.Trees.Pool[slot]
			t.X = td.X
			t.Z = td.Z
			t.Species = terrain.TreeSpecies(td.Species)
			t.Scale = td.Scale
			t.Age = 80
			t.Health = 1.0
			t.Lifecycle = core.LifecycleActive
		}
	}

	if m.Resources != nil {
		for z := 0; z < terrain.HeightmapSize; z++ {
			for x := 0; x < terrain.HeightmapSize; x++ {
				m.Resources.Map.Fertility[z][x] = data.Resources[z][x].Fertility
				m.Resources.Map.Ore[z][x] = data.Resources[z][x].Ore
				m.Resources.Map.Oil[z][x] = data.Resources[z][x].Oil
			}
		}
	}

	if m.Roads != nil {
		m.Roads.ClearModels()
		m.Roads.Nodes = nil
		m.Roads.Segments = nil
		m.Roads.NextID = 0
		for _, nd := range data.RoadNodes {
			idx := m.Roads.AddNode(nd.X, nd.Y, nd.Z)
			m.Roads.Nodes[idx].TrafficLight = nd.TrafficLight
			m.Roads.Nodes[idx].TrafficLightPhase = nd.TrafficLightPhase
			m.Roads.Nodes[idx].JunctionType = nd.JunctionType
			m.Roads.Nodes[idx].Flags = nd.Flags
		}
		for _, sd := range data.RoadSegments {
			segID := m.Roads.AddSegment(sd.NodeA, sd.NodeB, sd.RoadType)
			for i := range m.Roads.Segments {
				if m.Roads.Segments[i].ID == segID {
					m.Roads.Segments[i].SpeedLimit = sd.SpeedLimit
					m.Roads.Segments[i].LaneCount = sd.LaneCount
					m.Roads.Segments[i].Direction = sd.Direction
					m.Roads.Segments[i].Elevation = sd.Elevation
					m.Roads.Segments[i].MaintenanceCost = sd.MaintenanceCost
					m.Roads.Segments[i].ConstructionCost = sd.ConstructionCost
					m.Roads.Segments[i].Damaged = sd.Damaged
					m.Roads.Segments[i].RepairTimer = sd.RepairTimer
					m.Roads.Segments[i].Curve = road.CurveData{P1x: sd.CurveP1x, P1z: sd.CurveP1z, P2x: sd.CurveP2x, P2z: sd.CurveP2z}
					if len(sd.Lanes) > 0 {
						m.Roads.Segments[i].Lanes = make([]road.Lane, len(sd.Lanes))
						for li, ld := range sd.Lanes {
							m.Roads.Segments[i].Lanes[li] = road.Lane{
								Index:       ld.Index,
								Direction:   ld.Direction,
								SpeedLimit:  ld.SpeedLimit,
								Category:    ld.Category,
								Width:       ld.Width,
								Priority:    ld.Priority,
							}
						}
					}
					break
				}
			}
		}
	}

	if m.Transport != nil {
		m.Transport.Stops = nil
		m.Transport.Lines = nil
		m.Transport.Vehicles = nil
		m.Transport.NextID = 0
		for _, sd := range data.TransportStops {
			m.Transport.AddStop(sd.X, sd.Z, sd.TransType)
			if len(m.Transport.Stops) > 0 {
				s := &m.Transport.Stops[len(m.Transport.Stops)-1]
				s.ConnectedNetworks = sd.ConnectedNetworks
				if len(s.ConnectedNetworks) == 0 {
					s.ConnectedNetworks = []transport.TransportType{s.TransType}
				}
				s.Passengers = sd.Passengers
				s.DistrictID = sd.DistrictID
				s.Accessibility = sd.Accessibility
				s.Capacity = sd.Capacity
				if s.Capacity <= 0 {
					s.Capacity = 50
				}
				s.TransferStationID = sd.TransferStationID
				s.Name = sd.Name
			}
		}
		for _, ld := range data.TransportLines {
			col := rl.NewColor(ld.ColorR, ld.ColorG, ld.ColorB, 255)
			m.Transport.AddLine("Line", ld.TransType, ld.StopIDs, col)
			if len(m.Transport.Lines) > 0 {
				line := &m.Transport.Lines[len(m.Transport.Lines)-1]
				line.VehicleCount = ld.VehicleCount
				line.PassengerCount = ld.PassengerCount
				line.Budget = ld.Budget
				line.TotalPassengers = ld.TotalPassengers
				line.TotalIncome = ld.TotalIncome
				line.IsCircular = ld.IsCircular
			}
		}
		for i, nd := range data.TransportNetworks {
			if i < len(m.Transport.Networks) {
				m.Transport.Networks[i].Active = nd.Active
				m.Transport.Networks[i].VehicleCount = nd.VehicleCount
				m.Transport.Networks[i].RouteCount = nd.RouteCount
				m.Transport.Networks[i].StopCount = nd.StopCount
				m.Transport.Networks[i].StationCount = nd.StationCount
				m.Transport.Networks[i].PassengersPerDay = nd.PassengersPerDay
				m.Transport.Networks[i].TotalIncome = nd.TotalIncome
				m.Transport.Networks[i].MaintenanceCost = nd.MaintenanceCost
				m.Transport.Networks[i].Capacity = nd.Capacity
				m.Transport.Networks[i].Pollution = nd.Pollution
				m.Transport.Networks[i].Noise = nd.Noise
				m.Transport.Networks[i].WeeklyPassengers = nd.WeeklyPassengers
				m.Transport.Networks[i].LifetimePassengers = nd.LifetimePassengers
				m.Transport.Networks[i].TotalExpenses = nd.TotalExpenses
				m.Transport.Networks[i].AvgWaitTime = nd.AvgWaitTime
				m.Transport.Networks[i].CapacityUsage = nd.CapacityUsage
				m.Transport.Networks[i].VehicleUtilization = nd.VehicleUtilization
			}
		}
		m.Transport.Vehicles = make([]transport.TransportVehicle, 0, len(data.TransportVehicles))
		m.Transport.PoolNext = 0
		for i := range m.Transport.Pool {
			m.Transport.Pool[i].ID = math.MaxUint32
		}
		m.Transport.FreeList = make([]int32, transport.TransportVehiclePoolSize)
		for i := 0; i < transport.TransportVehiclePoolSize; i++ {
			m.Transport.FreeList[i] = int32(transport.TransportVehiclePoolSize - 1 - i)
		}
		for _, vd := range data.TransportVehicles {
			slot := m.Transport.AllocVehicle()
			if slot >= 0 {
				v := &m.Transport.Pool[slot]
				v.ID = vd.ID
				v.LineID = vd.LineID
				v.TransType = vd.TransType
				v.X = vd.X
				v.Z = vd.Z
				v.Speed = vd.Speed
				v.StopIdx = int(vd.StopIdx)
				v.Passengers = vd.Passengers
				v.Capacity = vd.Capacity
				v.Forward = vd.Forward
				v.Moving = vd.Moving
				v.TargetX = vd.TargetX
				v.TargetZ = vd.TargetZ
				v.StandingCapacity = vd.StandingCapacity
				v.HomeDepotType = vd.HomeDepotType
				v.HomeDepotSlot = vd.HomeDepotSlot
				v.MaintenanceTimer = vd.MaintenanceTimer
				if vd.ID >= m.Transport.PoolNext {
					m.Transport.PoolNext = vd.ID + 1
				}
			}
		}
	}

	if m.Parking != nil && len(data.ParkingSpots) > 0 {
		m.Parking.Spots = make([]road.ParkingSpot, len(data.ParkingSpots))
		for i, pd := range data.ParkingSpots {
			m.Parking.Spots[i] = road.ParkingSpot{
				ID:        pd.ID,
				SpotType:  pd.SpotType,
				X:         pd.X,
				Z:         pd.Z,
				RoadSeg:   pd.RoadSeg,
				LaneIdx:   pd.LaneIdx,
				Occupied:  pd.Occupied,
				OwnerSlot: pd.OwnerSlot,
			}
			if pd.ID >= m.Parking.NextID {
				m.Parking.NextID = pd.ID + 1
			}
		}
		for _, ld := range data.ParkingLots {
			slot := m.Parking.AllocLot()
			if slot < 0 {
				continue
			}
			lot := &m.Parking.Lots[slot]
			lot.Entity = core.NewEntity(m.Parking.NextID, ld.X, 0, ld.Z, core.OwnerBuilding)
			lot.Lifecycle = core.LifecycleActive
			lot.LotType = ld.LotType
			lot.Width = ld.Width
			lot.Depth = ld.Depth
			lot.Capacity = ld.Capacity
			lot.CellX = ld.CellX
			lot.CellZ = ld.CellZ
			lot.Spots = make([]int32, len(ld.SpotIndices))
			copy(lot.Spots, ld.SpotIndices)
			m.Parking.NextID++
		}
	}

	if m.Transport != nil {
		if m.Transport.Tracks != nil && len(data.MetroTracks) > 0 {
			for i := range m.Transport.Tracks.Pool {
				m.Transport.Tracks.Pool[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Transport.Tracks.FreeList = make([]int32, transport.MetroTrackPoolSize)
			for i := 0; i < transport.MetroTrackPoolSize; i++ {
				m.Transport.Tracks.FreeList[i] = int32(transport.MetroTrackPoolSize - 1 - i)
			}
			m.Transport.Tracks.Count = 0
			m.Transport.Tracks.NextID = 0
			for _, td := range data.MetroTracks {
				slot := m.Transport.Tracks.AddTrack(td.StartX, td.StartZ, td.EndX, td.EndZ)
				if slot >= 0 {
					t := &m.Transport.Tracks.Pool[slot]
					t.ID = td.ID
					t.Length = td.Length
					t.Elevated = td.Elevated
					if td.ID >= m.Transport.Tracks.NextID {
						m.Transport.Tracks.NextID = td.ID + 1
					}
				}
			}
		}
		if m.Transport.Rails != nil && len(data.RailTracks) > 0 {
			for i := range m.Transport.Rails.Pool {
				m.Transport.Rails.Pool[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Transport.Rails.FreeList = make([]int32, transport.RailTrackPoolSize)
			for i := 0; i < transport.RailTrackPoolSize; i++ {
				m.Transport.Rails.FreeList[i] = int32(transport.RailTrackPoolSize - 1 - i)
			}
			m.Transport.Rails.Count = 0
			m.Transport.Rails.NextID = 0
			for _, td := range data.RailTracks {
				slot := m.Transport.Rails.AddTrack(td.StartX, td.StartZ, td.EndX, td.EndZ)
				if slot >= 0 {
					t := &m.Transport.Rails.Pool[slot]
					t.ID = td.ID
					t.Length = td.Length
					t.SignalA = td.SignalA
					t.SignalB = td.SignalB
					t.BlockID = td.BlockID
					t.OutsideA = td.OutsideA
					t.OutsideB = td.OutsideB
					t.Occupied = td.Occupied
					t.OccupierID = td.OccupierID
					if td.ID >= m.Transport.Rails.NextID {
						m.Transport.Rails.NextID = td.ID + 1
					}
				}
			}
		}
		if m.Transport.Cargo != nil && len(data.CargoStations) > 0 {
			m.Transport.Cargo.Stations = nil
			m.Transport.Cargo.NextID = 0
			for _, cd := range data.CargoStations {
				cm := m.Transport.Cargo
				cm.AddStation(cd.X, cd.Z)
				if len(cm.Stations) > 0 {
					s := &cm.Stations[len(cm.Stations)-1]
					s.ID = cd.ID
					s.Name = cd.Name
					s.GoodsStored = cd.GoodsStored
					s.Capacity = cd.Capacity
					s.Active = cd.Active
					s.TrainSlot = cd.TrainSlot
					if cd.ID >= cm.NextID {
						cm.NextID = cd.ID + 1
					}
				}
			}
		}
		if m.Transport.Cargo != nil && len(data.CargoTrains) > 0 {
			m.Transport.Cargo.Trains = nil
			for _, vd := range data.CargoTrains {
				m.Transport.Cargo.Trains = append(m.Transport.Cargo.Trains, transport.TransportVehicle{
					ID: vd.ID, TransType: vd.TransType,
					X: vd.X, Z: vd.Z, Speed: vd.Speed,
					Moving: vd.Moving,
					TargetX: vd.TargetX, TargetZ: vd.TargetZ,
				})
			}
		}
		if len(data.CableConnections) > 0 {
			m.Transport.CableConnections = nil
		for _, cd := range data.CableConnections {
			m.Transport.CableConnections = append(m.Transport.CableConnections, transport.CableConnection{
				ID: cd.ID, StopA: cd.StopA, StopB: cd.StopB,
				StartX: cd.StartX, StartZ: cd.StartZ,
				EndX: cd.EndX, EndZ: cd.EndZ,
			})
			if cd.ID >= m.Transport.NextID {
				m.Transport.NextID = cd.ID + 1
			}
		}
		m.Transport.TransferStations = nil
		m.Transport.TransferNextID = 0
		for _, td := range data.TransferStations {
			m.Transport.TransferStations = append(m.Transport.TransferStations, transport.TransferStation{
				ID: td.ID, Name: td.Name, StopIDs: td.StopIDs,
				X: td.X, Z: td.Z,
			})
			if td.ID >= m.Transport.TransferNextID {
				m.Transport.TransferNextID = td.ID + 1
			}
		}
	}
	}

	if m.Parking != nil {
		if len(data.BusDepots) > 0 {
			for i := range m.Parking.BusDepots {
				m.Parking.BusDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.DepotFreeList = make([]int32, road.BusDepotPoolSize)
			for i := 0; i < road.BusDepotPoolSize; i++ {
				m.Parking.DepotFreeList[i] = int32(road.BusDepotPoolSize - 1 - i)
			}
			m.Parking.DepotCount = 0
			for _, dd := range data.BusDepots {
				slot := m.Parking.PlaceBusDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.BusDepots[slot].CellX = dd.CellX
					m.Parking.BusDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.TramDepots) > 0 {
			for i := range m.Parking.TramDepots {
				m.Parking.TramDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.TramDepotFreeList = make([]int32, road.TramDepotPoolSize)
			for i := 0; i < road.TramDepotPoolSize; i++ {
				m.Parking.TramDepotFreeList[i] = int32(road.TramDepotPoolSize - 1 - i)
			}
			m.Parking.TramDepotCount = 0
			for _, dd := range data.TramDepots {
				slot := m.Parking.PlaceTramDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.TramDepots[slot].CellX = dd.CellX
					m.Parking.TramDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.MetroDepots) > 0 {
			for i := range m.Parking.MetroDepots {
				m.Parking.MetroDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.MetroDepotFreeList = make([]int32, road.MetroDepotPoolSize)
			for i := 0; i < road.MetroDepotPoolSize; i++ {
				m.Parking.MetroDepotFreeList[i] = int32(road.MetroDepotPoolSize - 1 - i)
			}
			m.Parking.MetroDepotCount = 0
			for _, dd := range data.MetroDepots {
				slot := m.Parking.PlaceMetroDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.MetroDepots[slot].CellX = dd.CellX
					m.Parking.MetroDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.FerryDepots) > 0 {
			for i := range m.Parking.FerryDepots {
				m.Parking.FerryDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.FerryDepotFreeList = make([]int32, road.FerryDepotPoolSize)
			for i := 0; i < road.FerryDepotPoolSize; i++ {
				m.Parking.FerryDepotFreeList[i] = int32(road.FerryDepotPoolSize - 1 - i)
			}
			m.Parking.FerryDepotCount = 0
			for _, dd := range data.FerryDepots {
				slot := m.Parking.PlaceFerryDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.FerryDepots[slot].CellX = dd.CellX
					m.Parking.FerryDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.MonorailDepots) > 0 {
			for i := range m.Parking.MonorailDepots {
				m.Parking.MonorailDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.MonorailDepotFreeList = make([]int32, road.MonorailDepotPoolSize)
			for i := 0; i < road.MonorailDepotPoolSize; i++ {
				m.Parking.MonorailDepotFreeList[i] = int32(road.MonorailDepotPoolSize - 1 - i)
			}
			m.Parking.MonorailDepotCount = 0
			for _, dd := range data.MonorailDepots {
				slot := m.Parking.PlaceMonorailDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.MonorailDepots[slot].CellX = dd.CellX
					m.Parking.MonorailDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.CableCarDepots) > 0 {
			for i := range m.Parking.CableCarDepots {
				m.Parking.CableCarDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.CableCarDepotFreeList = make([]int32, road.CableCarDepotPoolSize)
			for i := 0; i < road.CableCarDepotPoolSize; i++ {
				m.Parking.CableCarDepotFreeList[i] = int32(road.CableCarDepotPoolSize - 1 - i)
			}
			m.Parking.CableCarDepotCount = 0
			for _, dd := range data.CableCarDepots {
				slot := m.Parking.PlaceCableCarDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.CableCarDepots[slot].CellX = dd.CellX
					m.Parking.CableCarDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.TaxiDepots) > 0 {
			for i := range m.Parking.TaxiDepots {
				m.Parking.TaxiDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.TaxiDepotFreeList = make([]int32, road.TaxiDepotPoolSize)
			for i := 0; i < road.TaxiDepotPoolSize; i++ {
				m.Parking.TaxiDepotFreeList[i] = int32(road.TaxiDepotPoolSize - 1 - i)
			}
			m.Parking.TaxiDepotCount = 0
			for _, dd := range data.TaxiDepots {
				slot := m.Parking.PlaceTaxiDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.TaxiDepots[slot].CellX = dd.CellX
					m.Parking.TaxiDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.AirportDepots) > 0 {
			for i := range m.Parking.AirportDepots {
				m.Parking.AirportDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.AirportDepotFreeList = make([]int32, road.AirportDepotPoolSize)
			for i := 0; i < road.AirportDepotPoolSize; i++ {
				m.Parking.AirportDepotFreeList[i] = int32(road.AirportDepotPoolSize - 1 - i)
			}
			m.Parking.AirportDepotCount = 0
			for _, dd := range data.AirportDepots {
				slot := m.Parking.PlaceAirportDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.AirportDepots[slot].CellX = dd.CellX
					m.Parking.AirportDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.PortDepots) > 0 {
			for i := range m.Parking.PortDepots {
				m.Parking.PortDepots[i].Lifecycle = core.LifecycleUnallocated
			}
			m.Parking.PortDepotFreeList = make([]int32, road.PortDepotPoolSize)
			for i := 0; i < road.PortDepotPoolSize; i++ {
				m.Parking.PortDepotFreeList[i] = int32(road.PortDepotPoolSize - 1 - i)
			}
			m.Parking.PortDepotCount = 0
			for _, dd := range data.PortDepots {
				slot := m.Parking.PlacePortDepot(dd.X, dd.Z)
				if slot >= 0 {
					m.Parking.PortDepots[slot].CellX = dd.CellX
					m.Parking.PortDepots[slot].CellZ = dd.CellZ
				}
			}
		}
		if len(data.TaxiFleet) > 0 {
			m.Parking.TaxiFleet = nil
			for _, vd := range data.TaxiFleet {
				m.Parking.TaxiFleet = append(m.Parking.TaxiFleet, road.TaxiFleetVehicle{
					X: vd.X, Z: vd.Z, Speed: vd.Speed,
					Moving: vd.Moving,
					TargetX: vd.TargetX, TargetZ: vd.TargetZ,
				})
			}
		}
		if len(data.TaxiRequests) > 0 {
			m.Parking.TaxiRequests = nil
			for _, rd := range data.TaxiRequests {
				m.Parking.TaxiRequests = append(m.Parking.TaxiRequests, road.TaxiRequest{
					ID: rd.ID, CitizenX: rd.CitizenX, CitizenZ: rd.CitizenZ,
					DstX: rd.DstX, DstZ: rd.DstZ, Active: rd.Active, Assigned: rd.Assigned,
				})
			}
		}
	}

	if m.Zones != nil && data.Zoning.Width > 0 {
		building.ImportZoning(data.Zoning, m.Zones, m.Buildings, m.Demand)
	}

	return data.Money, data.TimeOfDay, nil
}
