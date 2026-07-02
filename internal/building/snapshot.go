package building

import "github.com/katwate/js-skylines/internal/zoning"

type SaveZoning struct {
	Width            int
	Height           int
	ZoneTypes        []uint8
	LandValue        []float32
	CellDistrict     []uint8
	DistrictPolicies []uint32
	Buildings        []BuildingSave
	Demand           [4]float32
	Education        float32
	OfficePolicy     float32
	Stats            Statistics
}

type BuildingSave struct {
	ID              uint32
	LotID           int
	Type            uint8
	CellX, CellZ    int
	Width, Height   int
	Level           int
	State           uint8
	Stage           uint8
	Progress        float32
	BuildTime       float32
	UpgradeProgress float32
	Residents       int
	Vacancies       int
	Workers         int
	Customers       int
	Employees       int
	Production      float32
	Profitability   float32
	HouseholdMembers int
}

func ExportZoning(zm *zoning.ZoneManager, bm *Manager, demand *zoning.DemandEngine) SaveZoning {
	sz := SaveZoning{
		Width:  zm.Width(),
		Height: zm.Height(),
	}
	if zm != nil {
		sz.ZoneTypes = zm.ExportZoneTypes()
		sz.CellDistrict = zm.ExportCellDistrict()
		sz.DistrictPolicies = zm.DistrictPolicies()
	}
	if bm != nil {
		sz.LandValue = bm.ExportLandValue()
		sz.Buildings = bm.ExportBuildings()
		sz.Stats = bm.Stats
	}
	if demand != nil {
		sz.Demand = [4]float32{demand.Residential, demand.Commercial, demand.Industrial, demand.Office}
		sz.Education = demand.Education
		sz.OfficePolicy = demand.Factors.OfficePolicy
	}
	return sz
}

func ImportZoning(sz SaveZoning, zm *zoning.ZoneManager, bm *Manager, demand *zoning.DemandEngine) {
	if zm != nil {
		zm.ImportZoneTypes(sz.ZoneTypes)
		zm.ImportCellDistrict(sz.CellDistrict)
		zm.ImportDistrictPolicies(sz.DistrictPolicies)
	}
	if demand != nil {
		demand.Residential = sz.Demand[0]
		demand.Commercial = sz.Demand[1]
		demand.Industrial = sz.Demand[2]
		demand.Office = sz.Demand[3]
		demand.Education = sz.Education
		demand.Factors.OfficePolicy = sz.OfficePolicy
	}
	if bm != nil {
		bm.ImportLandValue(sz.LandValue)
		bm.ImportBuildings(sz.Buildings)
		bm.Stats = sz.Stats
	}
}

func (m *Manager) ExportLandValue() []float32 {
	out := make([]float32, m.width*m.height)
	for z := 0; z < m.height; z++ {
		for x := 0; x < m.width; x++ {
			out[z*m.width+x] = m.landValue[z][x]
		}
	}
	return out
}

func (m *Manager) ImportLandValue(flat []float32) {
	if len(flat) != m.width*m.height {
		return
	}
	for z := 0; z < m.height; z++ {
		for x := 0; x < m.width; x++ {
			m.landValue[z][x] = flat[z*m.width+x]
		}
	}
}

func (m *Manager) ExportBuildings() []BuildingSave {
	out := make([]BuildingSave, len(m.Buildings))
	for i, b := range m.Buildings {
		out[i] = BuildingSave{
			ID: b.ID, LotID: b.LotID, Type: uint8(b.Type),
			CellX: b.CellX, CellZ: b.CellZ, Width: b.Width, Height: b.Height,
			Level: b.Level, State: uint8(b.State), Stage: uint8(b.Stage),
			Progress: b.Progress, BuildTime: b.BuildTime, UpgradeProgress: b.UpgradeProgress,
			Residents: b.Occupancy.Residents, Vacancies: b.Occupancy.Vacancies,
			Workers: b.Occupancy.Workers, Customers: b.Occupancy.Customers,
			Employees: b.Occupancy.Employees, Production: b.Occupancy.Production,
			Profitability: b.Business.Profitability, HouseholdMembers: b.Household.Members,
		}
	}
	return out
}

func (m *Manager) ImportBuildings(saved []BuildingSave) {
	m.Buildings = m.Buildings[:0]
	m.byLot = make(map[int]int)
	for _, s := range saved {
		wx, wz := m.zones.CellCenter(s.CellX+s.Width/2, s.CellZ+s.Height/2)
		b := Building{
			ID: s.ID, LotID: s.LotID, Type: zoning.ZoneType(s.Type),
			CellX: s.CellX, CellZ: s.CellZ, Width: s.Width, Height: s.Height,
			WorldX: wx, WorldZ: wz, Level: s.Level,
			State: State(s.State), Stage: ConstructStage(s.Stage),
			Progress: s.Progress, BuildTime: s.BuildTime, UpgradeProgress: s.UpgradeProgress,
			Business: Business{Profitability: s.Profitability},
			Household: Household{Members: s.HouseholdMembers},
			Occupancy: Occupancy{
				Residents: s.Residents, Vacancies: s.Vacancies,
				Workers: s.Workers, Customers: s.Customers,
				Employees: s.Employees, Production: s.Production,
			},
		}
		if s.ID >= m.nextID {
			m.nextID = s.ID + 1
		}
		idx := len(m.Buildings)
		m.Buildings = append(m.Buildings, b)
		m.byLot[s.LotID] = idx
	}
}
