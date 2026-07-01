package terrain

import (
	"encoding/gob"
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

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
}

type RoadNodeData struct {
	X, Z            float32
	HasTrafficLight bool
	IsOutsideConn   bool
}

type RoadSegmentData struct {
	NodeA    uint32
	NodeB    uint32
	RoadType RoadType
	Length   float32
	Elevation int32
	Damaged  bool
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
}

type DistrictData struct {
	Name     string
	CenterX  float32
	CenterZ  float32
	Radius   float32
	Policies []DistrictPolicy
}

type TransportStopData struct {
	X, Z        float32
	TransType   TransportType
	Passengers  int32
	IsStation   bool
	Underground bool
}

type TransportLineData struct {
	TransType TransportType
	StopIDs   []uint32
	Active    bool
	ColorR    uint8
	ColorG    uint8
	ColorB    uint8
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

func SaveGame(filename string, m *Manager, money float32, timeOfDay int32) error {
	data := SaveData{
		Version:  1,
		Seed:     m.Seed,
		Money:    money,
		TimeOfDay: timeOfDay,
		Night:    m.Night,
	}

	data.TerrainHeight = m.Heightmap.Data

	if m.Trees != nil {
		for _, t := range m.Trees.Trees {
			data.Trees = append(data.Trees, TreeData{X: t.X, Z: t.Z, Species: int(t.Species), Scale: t.Scale})
		}
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
				X: n.X, Z: n.Z,
				HasTrafficLight: n.HasTrafficLight,
				IsOutsideConn:   n.IsOutsideConn,
			})
		}
		for _, s := range m.Roads.Segments {
			data.RoadSegments = append(data.RoadSegments, RoadSegmentData{
				NodeA: s.NodeA, NodeB: s.NodeB,
				RoadType:  s.RoadType,
				Length:    s.Length,
				Elevation: s.Elevation,
				Damaged:   s.Damaged,
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
		for _, b := range m.Buildings.Buildings {
			if b.X < -99998 {
				continue
			}
			data.Buildings = append(data.Buildings, BuildingData{
				X: b.X, Z: b.Z,
				Type:          b.Type,
				Seed:          b.Seed,
				Width:         b.Width,
				Depth:         b.Depth,
				Height:        b.Height,
				Level:         b.Level,
				Residents:     b.Residents,
				Workers:       b.Workers,
				Abandoned:     b.Abandoned,
				ConstructTimer: b.ConstructTimer,
				Constructed:   b.Constructed,
				HasRoad:       b.HasRoad,
			})
		}
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
			data.TransportStops = append(data.TransportStops, TransportStopData{
				X: s.X, Z: s.Z,
				TransType:   s.TransType,
				Passengers:  s.Passengers,
				IsStation:   s.IsStation,
				Underground: s.Underground,
			})
		}
		for _, l := range m.Transport.Lines {
			col := l.Color
			data.TransportLines = append(data.TransportLines, TransportLineData{
				TransType: l.TransType,
				StopIDs:   l.Stops,
				Active:    l.Active,
				ColorR:    col.R,
				ColorG:    col.G,
				ColorB:    col.B,
			})
		}
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

func LoadGame(filename string, m *Manager) (money float32, timeOfDay int32, err error) {
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

	if data.Version != 1 {
		return 0, 0, fmt.Errorf("unsupported save version: %d", data.Version)
	}

	m.Seed = data.Seed
	m.Night = data.Night

	m.Heightmap.Data = data.TerrainHeight

	if m.Trees != nil {
		m.Trees.Trees = nil
		for _, td := range data.Trees {
			m.Trees.Trees = append(m.Trees.Trees, Tree{
				X: td.X, Z: td.Z,
				Species: TreeSpecies(td.Species),
				Scale:   td.Scale,
			})
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
			idx := m.Roads.AddNode(nd.X, nd.Z)
			m.Roads.Nodes[idx].HasTrafficLight = nd.HasTrafficLight
			m.Roads.Nodes[idx].IsOutsideConn = nd.IsOutsideConn
		}
		for _, sd := range data.RoadSegments {
			m.Roads.AddSegment(sd.NodeA, sd.NodeB, sd.RoadType)
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
		m.Buildings.Buildings = nil
		m.Buildings.nextSeed = 0
		for _, bd := range data.Buildings {
			b := Building{
				X: bd.X, Z: bd.Z,
				Type:          bd.Type,
				Seed:          bd.Seed,
				Width:         bd.Width,
				Depth:         bd.Depth,
				Height:        bd.Height,
				Level:         bd.Level,
				Residents:     bd.Residents,
				Workers:       bd.Workers,
				Abandoned:     bd.Abandoned,
				ConstructTimer: bd.ConstructTimer,
				Constructed:   bd.Constructed,
				HasRoad:       bd.HasRoad,
			}
			if bd.Type == ZoneResidentialLow || bd.Type == ZoneResidentialHigh {
				b.Household = &HouseholdInfo{
					FamilyMembers: bd.Residents,
					Wealth:        30,
					Education:     10,
					Happiness:     50,
				}
			}
			if bd.Type == ZoneCommercialLow || bd.Type == ZoneCommercialHigh || bd.Type == ZoneIndustrial || bd.Type == ZoneOffice {
				b.Business = &BusinessInfo{
					GoodsStored:   10,
					Profitability: 50,
				}
			}
			m.Buildings.Buildings = append(m.Buildings.Buildings, b)
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
		}
		for _, ld := range data.TransportLines {
			col := rl.NewColor(ld.ColorR, ld.ColorG, ld.ColorB, 255)
			m.Transport.AddLine("Line", ld.TransType, ld.StopIDs, col)
		}
	}

	return data.Money, data.TimeOfDay, nil
}
