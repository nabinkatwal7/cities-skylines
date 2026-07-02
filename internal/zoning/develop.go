package zoning

type DevFlag uint8

const (
	DevRoad DevFlag = 1 << iota
	DevElectricity
	DevWater
	DevSewage
	DevDemand
	DevTerrain
	DevAsset
)

const DevAll = DevRoad | DevElectricity | DevWater | DevSewage | DevDemand | DevTerrain | DevAsset

type DevelopmentCheck struct {
	Met     DevFlag
	Missing DevFlag
}

func (d DevelopmentCheck) Ready() bool {
	return d.Missing == 0
}

func (f DevFlag) Has(flag DevFlag) bool {
	return f&flag != 0
}

type ServiceCoverage interface {
	HasElectricity(worldX, worldZ float32) bool
	HasWater(worldX, worldZ float32) bool
	HasSewage(worldX, worldZ float32) bool
}

type DemandProvider interface {
	HasDemand(cat ZoneCategory) bool
}

type BuildingCatalog interface {
	HasAsset(zt ZoneType, width, height int) bool
}

type DemandEngine struct {
	Residential float32
	Commercial  float32
	Industrial  float32
	Office      float32
}

func NewDemandEngine() *DemandEngine {
	return &DemandEngine{Residential: 1, Commercial: 1, Industrial: 1}
}

func (d *DemandEngine) HasDemand(cat ZoneCategory) bool {
	if d == nil {
		return false
	}
	switch cat {
	case CategoryResidential:
		return d.Residential > 0
	case CategoryCommercial:
		return d.Commercial > 0
	case CategoryIndustrial:
		return d.Industrial > 0
	case CategoryOffice:
		return d.Office > 0
	default:
		return false
	}
}

type Catalog struct{}

func (Catalog) HasAsset(zt ZoneType, width, height int) bool {
	if zt == ZoneNone || width < 1 || height < 1 {
		return false
	}
	// ponytail: placeholder assets for lots up to 4×4; chapter 12 replaces with real catalog.
	return width <= 4 && height <= 4
}

func (zm *ZoneManager) SetDevelopmentDeps(services ServiceCoverage, demand DemandProvider, buildings BuildingCatalog) {
	zm.services = services
	zm.demand = demand
	zm.buildings = buildings
}

func (zm *ZoneManager) CheckDevelopment(lot *ZoneLot) DevelopmentCheck {
	var met DevFlag
	if lot == nil || lot.Type == ZoneNone {
		return DevelopmentCheck{Missing: DevAll}
	}

	if zm.lotHasRoad(lot) {
		met |= DevRoad
	}
	if zm.lotBuildable(lot) {
		met |= DevTerrain
	}

	wx, wz := zm.CellCenter(lot.X+lot.Width/2, lot.Z+lot.Height/2)
	if zm.services != nil {
		if zm.services.HasElectricity(wx, wz) {
			met |= DevElectricity
		}
		if zm.services.HasWater(wx, wz) {
			met |= DevWater
		}
		if zm.services.HasSewage(wx, wz) {
			met |= DevSewage
		}
	}
	if zm.demand != nil && zm.demand.HasDemand(ZoneCategoryOf(lot.Type)) {
		met |= DevDemand
	}
	if zm.buildings != nil && zm.buildings.HasAsset(lot.Type, lot.Width, lot.Height) {
		met |= DevAsset
	}

	return DevelopmentCheck{Met: met, Missing: DevAll &^ met}
}

func (zm *ZoneManager) CanDevelop(lot *ZoneLot) bool {
	return zm.CheckDevelopment(lot).Ready()
}

func (zm *ZoneManager) lotHasRoad(lot *ZoneLot) bool {
	for z := lot.Z; z < lot.Z+lot.Height; z++ {
		for x := lot.X; x < lot.X+lot.Width; x++ {
			wx, wz := zm.CellCenter(x, z)
			if zm.roadConnected(wx, wz) {
				return true
			}
		}
	}
	return false
}

func (zm *ZoneManager) lotBuildable(lot *ZoneLot) bool {
	if zm.buildability == nil {
		return true
	}
	for z := lot.Z; z < lot.Z+lot.Height; z++ {
		for x := lot.X; x < lot.X+lot.Width; x++ {
			wx, wz := zm.CellCenter(x, z)
			info := zm.buildability.GetBuildability(wx, wz)
			if info.Score < 0.3 || info.IsUnderwater {
				return false
			}
		}
	}
	return true
}
