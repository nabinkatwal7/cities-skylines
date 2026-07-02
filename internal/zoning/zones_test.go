package zoning

import (
	"testing"

	"github.com/katwate/js-skylines/internal/road"
)

func TestZoneTraitsOf(t *testing.T) {
	office := ZoneTraitsOf(ZoneOffice)
	if !office.HighEducationJobs || !office.RequiresEducation || office.Pollution > 0.1 {
		t.Fatalf("office traits: %+v", office)
	}
	ind := ZoneTraitsOf(ZoneIndustrial)
	if !ind.Goods || !ind.Freight || ind.Pollution < 0.5 {
		t.Fatalf("industrial traits: %+v", ind)
	}
	resLow := ZoneTraitsOf(ZoneResidentialLow)
	if !resLow.Housing || resLow.Density != 0.5 {
		t.Fatalf("res low traits: %+v", resLow)
	}
	if ZoneCategoryOf(ZoneCommercialHigh) != CategoryCommercial {
		t.Fatal("commercial high category mismatch")
	}
}

func TestZoneDepthForRoad(t *testing.T) {
	if ZoneDepthForRoad(road.RoadHighway) != 8 {
		t.Fatal("highway depth")
	}
	if ZoneDepthForRoad(road.RoadFourLane) != 6 {
		t.Fatal("arterial depth")
	}
	if ZoneDepthForRoad(road.RoadTwoLane) != 4 {
		t.Fatal("collector depth")
	}
	if ZoneDepthForRoad(road.RoadGravel) != 2 {
		t.Fatal("local depth")
	}
}

func TestRebuildLots(t *testing.T) {
	zm := NewZoneManager(16, 16, nil, nil)
	for x := 2; x < 6; x++ {
		for z := 2; z < 4; z++ {
			zm.SetZoneCell(x, z, ZoneResidentialLow)
		}
	}
	lots := zm.Lots()
	if len(lots) != 1 {
		t.Fatalf("expected 1 lot, got %d", len(lots))
	}
	lot := lots[0]
	if lot.Width != 4 || lot.Height != 2 || lot.Cells != 8 {
		t.Fatalf("lot size: %+v", lot)
	}
	if zm.LotAtCell(3, 3) == nil {
		t.Fatal("cell should belong to lot")
	}

	zm.SetZoneCell(6, 2, ZoneResidentialLow)
	if len(zm.Lots()) != 0 {
		t.Fatal("L-shaped zone should not form a rectangular lot")
	}
}

func TestCheckDevelopment(t *testing.T) {
	zm := NewZoneManager(16, 16, nil, nil)
	zm.SetDevelopmentDeps(
		stubServices{elec: true, water: true, sewage: true},
		NewDemandEngine(),
		Catalog{},
	)
	for x := 2; x < 6; x++ {
		for z := 2; z < 4; z++ {
			zm.SetZoneCell(x, z, ZoneResidentialLow)
		}
	}
	lot := zm.Lots()[0]
	check := zm.CheckDevelopment(&lot)
	if check.Ready() {
		t.Fatal("should not develop without road connection")
	}
	if !check.Missing.Has(DevRoad) {
		t.Fatalf("expected missing road, got %+v", check)
	}
	if check.Missing&(DevElectricity|DevWater|DevSewage|DevDemand|DevTerrain|DevAsset) != 0 {
		t.Fatalf("expected utilities/demand/terrain/asset met, got %+v", check)
	}
}

func TestCheckDevelopment_missingServices(t *testing.T) {
	zm := NewZoneManager(16, 16, nil, nil)
	zm.SetDevelopmentDeps(stubServices{}, NewDemandEngine(), Catalog{})
	for x := 2; x < 4; x++ {
		zm.SetZoneCell(x, 2, ZoneIndustrial)
	}
	lot := zm.Lots()[0]
	check := zm.CheckDevelopment(&lot)
	if check.Ready() {
		t.Fatal("lot should not be ready without services")
	}
	if check.Missing&(DevElectricity|DevWater|DevSewage) == 0 {
		t.Fatalf("expected missing utilities, got %+v", check)
	}
}

func TestCheckDevelopment_noDemand(t *testing.T) {
	zm := NewZoneManager(16, 16, nil, nil)
	demand := NewDemandEngine()
	demand.Office = 0
	zm.SetDevelopmentDeps(stubServices{elec: true, water: true, sewage: true}, demand, Catalog{})
	for x := 2; x < 4; x++ {
		zm.SetZoneCell(x, 2, ZoneOffice)
	}
	lot := zm.Lots()[0]
	check := zm.CheckDevelopment(&lot)
	if check.Missing&DevDemand == 0 {
		t.Fatalf("office demand should block development, got %+v", check)
	}
}

type stubServices struct {
	elec, water, sewage bool
}

func (s stubServices) HasElectricity(_, _ float32) bool { return s.elec }
func (s stubServices) HasWater(_, _ float32) bool       { return s.water }
func (s stubServices) HasSewage(_, _ float32) bool    { return s.sewage }
