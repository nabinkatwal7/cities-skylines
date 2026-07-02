package building

import (
	"testing"

	"github.com/katwate/js-skylines/internal/services"
	"github.com/katwate/js-skylines/internal/zoning"
)

func TestSpawnRequiresDevelopment(t *testing.T) {
	zm := zoning.NewZoneManager(16, 16, nil, nil)
	zm.SetDevelopmentDeps(services.NewStarter(), zoning.NewDemandEngine(), zoning.Catalog{})
	for x := 2; x < 6; x++ {
		for z := 2; z < 4; z++ {
			zm.SetZoneCell(x, z, zoning.ZoneResidentialLow)
		}
	}
	bm := NewManager(zm, zoning.NewDemandEngine(), services.NewStarter())
	bm.trySpawn()
	if len(bm.Buildings) != 0 {
		t.Fatal("should not spawn lot without road connection")
	}
}

func TestConstructionCompletes(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	zm.SetDevelopmentDeps(services.NewStarter(), zoning.NewDemandEngine(), zoning.Catalog{})
	lot := zoning.ZoneLot{ID: 0, X: 1, Z: 1, Width: 2, Height: 2, Type: zoning.ZoneResidentialLow, Cells: 4}
	bm := NewManager(zm, zoning.NewDemandEngine(), services.NewStarter())
	bm.startConstruction(&lot)
	b := &bm.Buildings[0]
	if b.State != StateConstructing || b.Stage != StageFoundation {
		t.Fatal("expected foundation stage")
	}
	for i := 0; i < 200; i++ {
		if len(bm.Buildings) == 0 {
			break
		}
		bm.tickBuilding(0, 1)
	}
	b = &bm.Buildings[0]
	if len(bm.Buildings) == 0 {
		t.Fatal("building removed during construction")
	}
	if b.State != StateOccupied {
		t.Fatalf("expected occupied after construction, got %v", b.State)
	}
	if b.Stage != StageCompleted {
		t.Fatalf("expected completed stage, got %v", b.Stage)
	}
	if b.Household.Members < 1 {
		t.Fatal("expected household initialized")
	}
}

func TestConstructionStages(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	bm := NewManager(zm, zoning.NewDemandEngine(), services.NewStarter())
	lot := zoning.ZoneLot{ID: 2, X: 0, Z: 0, Width: 2, Height: 2, Type: zoning.ZoneCommercialLow, Cells: 4}
	bm.startConstruction(&lot)
	b := &bm.Buildings[0]
	b.BuildTime = 100
	bm.tickBuilding(0, 40)
	if b.Stage != StageFramework {
		t.Fatalf("expected framework at 40%%, got %v", b.Stage)
	}
	bm.tickBuilding(0, 30)
	if b.Stage != StageCompleted {
		t.Fatalf("expected completed stage past 70%%, got %v", b.Stage)
	}
}

func TestUpgradeAndAbandon(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	demand := zoning.NewDemandEngine()
	demand.Education = 0.8
	demand.Factors.Happiness = 0.8
	demand.Factors.ServiceScore = 1
	svc := services.NewStarter()
	bm := NewManager(zm, demand, svc)
	lot := zoning.ZoneLot{ID: 1, X: 0, Z: 0, Width: 2, Height: 2, Type: zoning.ZoneResidentialLow, Cells: 4}
	bm.startConstruction(&lot)
	b := &bm.Buildings[0]
	b.State = StateOccupied
	b.Progress = 1
	b.LandValue = 0.7
	for i := 0; i < 500; i++ {
		bm.updateLandValueIncremental()
		bm.tickBuilding(0, 1)
	}
	if b.Level < 2 {
		t.Fatalf("expected level upgrade with good conditions, level=%d", b.Level)
	}

	bad := services.CityServices{}
	bm.services = &bad
	for i := 0; i < 20; i++ {
		bm.tickBuilding(0, 1)
	}
	if b.State != StateAbandoned {
		t.Fatalf("expected abandonment without power, state=%v", b.State)
	}
}

func TestDemolitionFreesLot(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 2; x++ {
		for z := 0; z < 2; z++ {
			zm.SetZoneCell(x, z, zoning.ZoneResidentialLow)
		}
	}
	bm := NewManager(zm, zoning.NewDemandEngine(), &services.CityServices{})
	lotID := 0
	bm.byLot[lotID] = 0
	bm.Buildings = append(bm.Buildings, Building{
		ID: 1, LotID: lotID, Type: zoning.ZoneResidentialLow,
		State: StateAbandoned, CellX: 0, CellZ: 0, Width: 2, Height: 2,
	})
	for i := 0; i < 20; i++ {
		if len(bm.Buildings) == 0 {
			break
		}
		bm.tickBuilding(0, 1)
	}
	if len(bm.Buildings) != 0 {
		t.Fatal("abandoned building should be demolished")
	}
	if _, used := bm.byLot[lotID]; used {
		t.Fatal("lot should be freed after demolition")
	}
	if zm.Cells[0][0].Type != zoning.ZoneResidentialLow {
		t.Fatal("zone should remain after demolition")
	}
}

func TestHouseholdAndBusiness(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	demand := zoning.NewDemandEngine()
	demand.Factors.ShoppingDemand = 0.8
	demand.Factors.GoodsAvailability = 0.7
	bm := NewManager(zm, demand, services.NewStarter())

	resLot := zoning.ZoneLot{ID: 10, X: 0, Z: 0, Width: 2, Height: 1, Type: zoning.ZoneResidentialLow, Cells: 2}
	bm.startConstruction(&resLot)
	res := &bm.Buildings[0]
	res.State = StateOccupied
	res.Stage = StageCompleted
	bm.initOccupancy(res)
	for i := 0; i < 10; i++ {
		bm.tickOccupancy(res, 1)
	}
	if res.Occupancy.Residents < 1 || res.Household.Members < 1 {
		t.Fatalf("household should have members: %+v", res.Household)
	}

	comLot := zoning.ZoneLot{ID: 11, X: 3, Z: 0, Width: 2, Height: 1, Type: zoning.ZoneCommercialLow, Cells: 2}
	bm.startConstruction(&comLot)
	com := &bm.Buildings[1]
	com.State = StateOccupied
	bm.initOccupancy(com)
	for i := 0; i < 5; i++ {
		bm.tickOccupancy(com, 1)
	}
	if com.Business.Employees < 1 || com.Occupancy.Workers < 1 {
		t.Fatalf("business should have workers: %+v", com.Business)
	}
	if com.Business.Profitability == 0 && demand.Factors.ShoppingDemand > 0 {
		t.Fatal("business should evaluate profitability")
	}
}

func TestLandValue(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	zm.Init()
	demand := zoning.NewDemandEngine()
	bm := NewManager(zm, demand, services.NewStarter())

	demand.Factors.Pollution = 0
	demand.Factors.Crime = 0
	bm.calcLandValueAt(0, 0)
	clean := bm.landValueAt(0, 0)

	demand.Factors.Pollution = 0.9
	demand.Factors.Crime = 0.8
	bm.calcLandValueAt(0, 0)
	polluted := bm.landValueAt(0, 0)
	if polluted >= clean {
		t.Fatalf("pollution/crime should lower land value: clean=%v polluted=%v", clean, polluted)
	}
}

func TestServiceConsumption(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	zm.Init()
	demand := zoning.NewDemandEngine()
	bm := NewManager(zm, demand, services.NewStarter())
	lot := zoning.ZoneLot{ID: 3, X: 0, Z: 0, Width: 2, Height: 1, Type: zoning.ZoneResidentialLow, Cells: 2}
	bm.startConstruction(&lot)
	b := &bm.Buildings[0]
	b.State = StateOccupied
	bm.initOccupancy(b)
	bm.tickServices(b)
	if b.Consumption.Electricity <= 0 || b.Consumption.Water <= 0 {
		t.Fatalf("residential should consume utilities: %+v", b.Consumption)
	}
	if !b.ServiceOK {
		t.Fatal("starter services should satisfy building")
	}
}

func TestStatistics(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	zm.Init()
	demand := zoning.NewDemandEngine()
	bm := NewManager(zm, demand, services.NewStarter())
	lot := zoning.ZoneLot{ID: 4, X: 0, Z: 0, Width: 2, Height: 2, Type: zoning.ZoneResidentialLow, Cells: 4}
	bm.startConstruction(&lot)
	b := &bm.Buildings[0]
	b.State = StateOccupied
	bm.initOccupancy(b)
	bm.refreshStats(1)
	if bm.Stats.Population < 1 {
		t.Fatalf("stats should track population: %+v", bm.Stats)
	}
}

func TestZoningSnapshotRoundTrip(t *testing.T) {
	zm := zoning.NewZoneManager(4, 4, nil, nil)
	zm.Init()
	zm.SetZoneCell(1, 1, zoning.ZoneIndustrial)
	demand := zoning.NewDemandEngine()
	bm := NewManager(zm, demand, services.NewStarter())
	lot := zoning.ZoneLot{ID: 5, X: 1, Z: 1, Width: 1, Height: 1, Type: zoning.ZoneIndustrial, Cells: 1}
	bm.startConstruction(&lot)
	bm.Buildings[0].State = StateOccupied

	sz := ExportZoning(zm, bm, demand)
	zm2 := zoning.NewZoneManager(4, 4, nil, nil)
	zm2.Init()
	bm2 := NewManager(zm2, zoning.NewDemandEngine(), services.NewStarter())
	ImportZoning(sz, zm2, bm2, demand)
	if zm2.Cells[1][1].Type != zoning.ZoneIndustrial {
		t.Fatal("zone grid should round-trip")
	}
	if len(bm2.Buildings) != 1 {
		t.Fatalf("buildings should round-trip, got %d", len(bm2.Buildings))
	}
}

func TestDistrictPolicyHighRiseBan(t *testing.T) {
	zm := zoning.NewZoneManager(4, 4, nil, nil)
	zm.Init()
	zm.SetDistrictPolicy(0, zoning.PolicyHighRiseBan, true)
	bm := NewManager(zm, zoning.NewDemandEngine(), services.NewStarter())
	b := Building{CellX: 0, CellZ: 0, Level: 3}
	if bm.maxLevelFor(&b) != 3 {
		t.Fatal("high rise ban should cap at level 3")
	}
}
