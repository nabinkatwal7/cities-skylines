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
	if b.State != StateConstructing {
		t.Fatal("expected constructing")
	}
	for i := 0; i < 200; i++ {
		bm.tickBuilding(b, 1)
	}
	if b.State != StateOccupied {
		t.Fatalf("expected occupied after construction, got %v", b.State)
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
		bm.updateLandValue()
		bm.tickBuilding(b, 1)
	}
	if b.Level < 2 {
		t.Fatalf("expected level upgrade with good conditions, level=%d", b.Level)
	}

	bad := services.CityServices{}
	bm.services = &bad
	for i := 0; i < 20; i++ {
		bm.tickBuilding(b, 1)
	}
	if b.State != StateAbandoned {
		t.Fatalf("expected abandonment without power, state=%v", b.State)
	}
}

func TestLandValue(t *testing.T) {
	zm := zoning.NewZoneManager(8, 8, nil, nil)
	demand := zoning.NewDemandEngine()
	bm := NewManager(zm, demand, services.NewStarter())

	demand.Factors.Pollution = 0
	demand.Factors.Crime = 0
	bm.updateLandValue()
	clean := bm.landValueAt(0, 0)

	demand.Factors.Pollution = 0.9
	demand.Factors.Crime = 0.8
	bm.updateLandValue()
	polluted := bm.landValueAt(0, 0)
	if polluted >= clean {
		t.Fatalf("pollution/crime should lower land value: clean=%v polluted=%v", clean, polluted)
	}
}
