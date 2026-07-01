package terrain

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"

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

	TerrainHeight [HeightmapSize][HeightmapSize]float32
	Trees         []TreeData
	Resources     [HeightmapSize][HeightmapSize]ResourceCell

	RoadNodes    []RoadNodeData
	RoadSegments []RoadSegmentData

	ZoneCells [][]ZoneCellData

	Buildings []BuildingData

	Districts []DistrictData

	TransportStops    []TransportStopData
	TransportLines    []TransportLineData
	TransportNetworks []TransportNetworkData
	TransportVehicles []TransportVehicleData

	ParkingSpots []ParkingSpotData
	ParkingLots  []ParkingLotData
}

type RoadNodeData struct {
	X, Y, Z         float32
	TrafficLight    TrafficLightState
	JunctionType    uint8
	Flags           RoadFlags
	TrafficLightPhase int32
	HasTrafficLight bool
	IsOutsideConn   bool
}

type RoadSegmentData struct {
	NodeA            uint32
	NodeB            uint32
	RoadType         RoadType
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
	Category   LaneCategory
	Width      float32
	Priority   int32
}

type ZoneCellData struct {
	Type    ZoneType
	Density float32
}

type BuildingData struct {
	X, Z            float32
	Type            ZoneType
	Seed            int32
	Width, Depth    float32
	Height          float32
	Level           int32
	Residents       int32
	Workers         int32
	Abandoned       bool
	ConstructTimer  int32
	Constructed     bool
	HasRoad         bool
	UpgradeTimer    int32
	AbandonTimer    int32
	HouseholdWealth    int32
	HouseholdEducation int32
	HouseholdHappiness int32
	BusinessGoodsStored    int32
	BusinessProfitability  int32
	ConsPower  float32
	ConsWater  float32
	ConsGarbage float32
}

type DistrictData struct {
	Name     string
	CenterX  float32
	CenterZ  float32
	Radius   float32
	Policies []DistrictPolicy
}

type TransportStopData struct {
	X, Z              float32
	TransType         TransportType
	ConnectedNetworks []TransportType
	Passengers        int32
	IsStation         bool
	Underground       bool
	DistrictID        int32
	Accessibility     float32
	Capacity          int32
}

type TransportLineData struct {
	TransType      TransportType
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
	Type             TransportType
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
}

type TransportVehicleData struct {
	ID         uint32
	LineID     uint32
	TransType  TransportType
	X, Z       float32
	Speed      float32
	StopIdx    int32
	Passengers int32
	Capacity   int32
	Forward    bool
	Moving     bool
	TargetX    float32
	TargetZ    float32
}

type ParkingSpotData struct {
	ID       uint32
	SpotType ParkingSpotType
	X, Z     float32
	RoadSeg  int32
	LaneIdx  int32
	Occupied bool
	OwnerSlot int32
}

type ParkingLotData struct {
	X, Z        float32
	LotType     ParkingSpotType
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

type ResourceCell struct {
	Fertility float32
	Ore       float32
	Oil       float32
}

type SaveStats struct {
	DirtyCount int32
	TotalCount int32
}

const currentSaveVersion int32 = 9

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
						nd.TrafficLight = TrafficLightRed
					}
					if nd.IsOutsideConn {
						nd.Flags |= RoadFlagOutsideConn
					}
				}
			}
			for i := range data.RoadSegments {
				sd := &data.RoadSegments[i]
				if sd.SpeedLimit == 0 {
					sd.SpeedLimit = roadSpeed(sd.RoadType)
				}
				if sd.LaneCount == 0 {
					sd.LaneCount = int32(roadLanes(sd.RoadType))
				}
			}
			data.Version = 3
		case 3:
			for i := range data.RoadSegments {
				sd := &data.RoadSegments[i]
				if len(sd.Lanes) == 0 {
					categories := laneCategoriesForRoad(sd.RoadType)
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
						cat := LaneDriving
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
				if data.RoadNodes[i].TrafficLight != TrafficLightNone && data.RoadNodes[i].TrafficLightPhase == 0 {
					maxLanes := int32(2)
					data.RoadNodes[i].TrafficLightPhase = maxLanes * 20
				}
			}
			data.Version = 6
		case 6:
			for i := range data.TransportLines {
				ld := &data.TransportLines[i]
				if ld.VehicleCount == 0 && ld.TotalPassengers == 0 {
					ld.IsCircular = ld.TransType != TransBus && ld.TransType != TransTram && ld.TransType != TransTaxi
				}
			}
			data.Version = 7
		case 7:
			for i := range data.TransportStops {
				sd := &data.TransportStops[i]
				if len(sd.ConnectedNetworks) == 0 {
					sd.ConnectedNetworks = []TransportType{sd.TransType}
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

func SaveGame(filename string, m *SimulationManager, money float32, timeOfDay int32) error {
	data := SaveData{
		Version:  currentSaveVersion,
		Seed:     m.Seed,
		Money:    money,
		TimeOfDay: timeOfDay,
		Night:    m.Night,
	}

	data.TerrainHeight = m.Heightmap.Data

	if m.Trees != nil {
		m.Trees.ForEach(func(t *Tree, _ int32) {
			data.Trees = append(data.Trees, TreeData{X: t.X, Z: t.Z, Species: int(t.Species), Scale: t.Scale})
		})
	}

	if m.Resources != nil {
		for z := 0; z < HeightmapSize; z++ {
			for x := 0; x < HeightmapSize; x++ {
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

	if m.Zones != nil {
		data.ZoneCells = make([][]ZoneCellData, m.Zones.height)
		for z := 0; z < m.Zones.height; z++ {
			data.ZoneCells[z] = make([]ZoneCellData, m.Zones.width)
			for x := 0; x < m.Zones.width; x++ {
				data.ZoneCells[z][x] = ZoneCellData{
					Type:    m.Zones.Cells[z][x].Type,
					Density: m.Zones.Cells[z][x].Density,
				}
			}
		}
	}

		if m.Buildings != nil {
		m.Buildings.ForEach(func(b *Building, _ int32) {
			bd := BuildingData{
				X: b.Position.X, Z: b.Position.Z,
				Type:          b.Type,
				Seed:          b.Seed,
				Width:         b.Width,
				Depth:         b.Depth,
				Height:        b.Height,
				Level:         b.Level,
				Residents:     b.Residents,
				Workers:       b.Workers,
				Abandoned:     b.HasFlag(FlagAbandoned),
				ConstructTimer: b.ConstructTimer,
				Constructed:   b.HasFlag(FlagConstructed),
				HasRoad:       b.HasFlag(FlagHasRoad),
				UpgradeTimer:  b.UpgradeTimer,
				AbandonTimer:  b.AbandonTimer,
				ConsPower:     b.Consumption.Power,
				ConsWater:     b.Consumption.Water,
				ConsGarbage:   b.Consumption.Garbage,
			}
			if b.Household != nil {
				bd.HouseholdWealth = b.Household.Wealth
				bd.HouseholdEducation = b.Household.Education
				bd.HouseholdHappiness = b.Household.Happiness
			}
			if b.Business != nil {
				bd.BusinessGoodsStored = b.Business.GoodsStored
				bd.BusinessProfitability = b.Business.Profitability
			}
			data.Buildings = append(data.Buildings, bd)
		})
	}

	if m.Districts != nil {
		for _, d := range m.Districts.Districts {
			data.Districts = append(data.Districts, DistrictData{
				Name:     d.Name,
				CenterX:  d.CenterX,
				CenterZ:  d.CenterZ,
				Radius:   d.Radius,
				Policies: d.Policies,
			})
		}
	}

	if m.Transport != nil {
		for _, s := range m.Transport.Stops {
			copyNet := make([]TransportType, len(s.ConnectedNetworks))
			copy(copyNet, s.ConnectedNetworks)
			data.TransportStops = append(data.TransportStops, TransportStopData{
				X: s.X, Z: s.Z,
				TransType:         s.TransType,
				ConnectedNetworks: copyNet,
				Passengers:        s.Passengers,
				IsStation:         s.IsStation,
				Underground:       s.Underground,
				DistrictID:        s.DistrictID,
				Accessibility:     s.Accessibility,
				Capacity:          s.Capacity,
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
			})
		}
		for _, v := range m.Transport.Vehicles {
			data.TransportVehicles = append(data.TransportVehicles, TransportVehicleData{
				ID:         v.ID,
				LineID:     v.LineID,
				TransType:  v.TransType,
				X:          v.X,
				Z:          v.Z,
				Speed:      v.Speed,
				StopIdx:    int32(v.StopIdx),
				Passengers: v.Passengers,
				Capacity:   v.Capacity,
				Forward:    v.Forward,
				Moving:     v.Moving,
				TargetX:    v.TargetX,
				TargetZ:    v.TargetZ,
			})
		}
		for _, v := range m.Transport.Pool {
			if v.ID == math.MaxUint32 {
				continue
			}
			data.TransportVehicles = append(data.TransportVehicles, TransportVehicleData{
				ID:         v.ID,
				LineID:     v.LineID,
				TransType:  v.TransType,
				X:          v.X,
				Z:          v.Z,
				Speed:      v.Speed,
				StopIdx:    int32(v.StopIdx),
				Passengers: v.Passengers,
				Capacity:   v.Capacity,
				Forward:    v.Forward,
				Moving:     v.Moving,
				TargetX:    v.TargetX,
				TargetZ:    v.TargetZ,
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
		m.Parking.ForEachLot(func(lot *ParkingLot, slot int32) {
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

func LoadGame(filename string, m *SimulationManager) (money float32, timeOfDay int32, err error) {
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
			t.Species = TreeSpecies(td.Species)
			t.Scale = td.Scale
			t.Age = 80
			t.Health = 1.0
			t.Lifecycle = LifecycleActive
		}
	}

	if m.Resources != nil {
		for z := 0; z < HeightmapSize; z++ {
			for x := 0; x < HeightmapSize; x++ {
				m.Resources.Map.Fertility[z][x] = data.Resources[z][x].Fertility
				m.Resources.Map.Ore[z][x] = data.Resources[z][x].Ore
				m.Resources.Map.Oil[z][x] = data.Resources[z][x].Oil
			}
		}
	}

	if m.Roads != nil {
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
					m.Roads.Segments[i].Curve = CurveData{P1x: sd.CurveP1x, P1z: sd.CurveP1z, P2x: sd.CurveP2x, P2z: sd.CurveP2z}
					if len(sd.Lanes) > 0 {
						m.Roads.Segments[i].Lanes = make([]Lane, len(sd.Lanes))
						for li, ld := range sd.Lanes {
							m.Roads.Segments[i].Lanes[li] = Lane{
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

	if m.Zones != nil {
		m.Zones.Cells = make([][]ZoneCell, len(data.ZoneCells))
		for z := 0; z < len(data.ZoneCells); z++ {
			m.Zones.Cells[z] = make([]ZoneCell, len(data.ZoneCells[z]))
			for x := 0; x < len(data.ZoneCells[z]); x++ {
				m.Zones.Cells[z][x] = ZoneCell{
					Type:    data.ZoneCells[z][x].Type,
					Density: data.ZoneCells[z][x].Density,
				}
			}
		}
	}

	if m.Buildings != nil {
		for _, bd := range data.Buildings {
			slot := m.Buildings.Alloc()
			if slot < 0 {
				continue
			}
			b := &m.Buildings.Pool[slot]
			b.Entity = NewEntity(uint32(bd.Seed), bd.X, 0, bd.Z, OwnerBuilding)
			b.Type = bd.Type
			b.Seed = bd.Seed
			b.Width = bd.Width
			b.Depth = bd.Depth
			b.Height = bd.Height
			b.Level = bd.Level
			b.Residents = bd.Residents
			b.Workers = bd.Workers
			b.ConstructTimer = bd.ConstructTimer
			b.UpgradeTimer = bd.UpgradeTimer
			b.AbandonTimer = bd.AbandonTimer
			if bd.Abandoned {
				b.SetFlag(FlagAbandoned)
			}
			if bd.Constructed {
				b.SetFlag(FlagConstructed)
			}
			if bd.HasRoad {
				b.SetFlag(FlagHasRoad)
			}
			b.Consumption.Power = bd.ConsPower
			b.Consumption.Water = bd.ConsWater
			b.Consumption.Garbage = bd.ConsGarbage
			if bd.Type == ZoneResidentialLow || bd.Type == ZoneResidentialHigh {
				b.Household = &HouseholdInfo{
					FamilyMembers: bd.Residents,
					Wealth:        bd.HouseholdWealth,
					Education:     bd.HouseholdEducation,
					Happiness:     bd.HouseholdHappiness,
				}
			}
			if bd.Type == ZoneCommercialLow || bd.Type == ZoneCommercialHigh || bd.Type == ZoneIndustrial || bd.Type == ZoneOffice {
				b.Business = &BusinessInfo{
					GoodsStored:   bd.BusinessGoodsStored,
					Profitability: bd.BusinessProfitability,
				}
			}
			if bd.Seed >= m.Buildings.nextSeed {
				m.Buildings.nextSeed = bd.Seed + 1
			}
		}
	}

	if m.Districts != nil {
		m.Districts.Districts = nil
		m.Districts.NextID = 0
		for _, dd := range data.Districts {
			m.Districts.AddDistrict(dd.Name, dd.CenterX, dd.CenterZ, dd.Radius)
			idx := len(m.Districts.Districts) - 1
			m.Districts.Districts[idx].Policies = dd.Policies
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
					s.ConnectedNetworks = []TransportType{s.TransType}
				}
				s.Passengers = sd.Passengers
				s.DistrictID = sd.DistrictID
				s.Accessibility = sd.Accessibility
				s.Capacity = sd.Capacity
				if s.Capacity <= 0 {
					s.Capacity = 50
				}
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
			}
		}
		m.Transport.Vehicles = make([]TransportVehicle, 0, len(data.TransportVehicles))
		m.Transport.PoolNext = 0
		for i := range m.Transport.Pool {
			m.Transport.Pool[i].ID = math.MaxUint32
		}
		m.Transport.FreeList = make([]int32, TransportVehiclePoolSize)
		for i := 0; i < TransportVehiclePoolSize; i++ {
			m.Transport.FreeList[i] = int32(TransportVehiclePoolSize - 1 - i)
		}
		for _, vd := range data.TransportVehicles {
			slot := m.Transport.allocVehicle()
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
				if vd.ID >= m.Transport.PoolNext {
					m.Transport.PoolNext = vd.ID + 1
				}
			}
		}
	}

	if m.Parking != nil && len(data.ParkingSpots) > 0 {
		m.Parking.Spots = make([]ParkingSpot, len(data.ParkingSpots))
		for i, pd := range data.ParkingSpots {
			m.Parking.Spots[i] = ParkingSpot{
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
			slot := m.Parking.allocLot()
			if slot < 0 {
				continue
			}
			lot := &m.Parking.Lots[slot]
			lot.Entity = NewEntity(m.Parking.NextID, ld.X, 0, ld.Z, OwnerBuilding)
			lot.Lifecycle = LifecycleActive
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

	return data.Money, data.TimeOfDay, nil
}
