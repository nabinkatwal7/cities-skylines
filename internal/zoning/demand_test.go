package zoning

import "testing"

func TestDemandUpdate_rciBalance(t *testing.T) {
	zm := NewZoneManager(16, 16, nil, nil)
	for x := 0; x < 10; x++ {
		for z := 0; z < 10; z++ {
			zm.SetZoneCell(x, z, ZoneIndustrial)
		}
	}

	d := NewDemandEngine()
	startRes := d.Residential
	for i := 0; i < 60; i++ {
		d.Update(1, zm)
	}
	if d.Residential <= startRes {
		t.Fatalf("residential demand should rise with job-heavy zoning: start=%v now=%v", startRes, d.Residential)
	}
	if !d.HasDemand(CategoryResidential) {
		t.Fatal("expected positive residential demand")
	}
}

func TestDemandUpdate_officeDerived(t *testing.T) {
	zm := NewZoneManager(8, 8, nil, nil)
	zm.SetZoneCell(0, 0, ZoneResidentialLow)
	zm.SetZoneCell(1, 0, ZoneIndustrial)
	d := NewDemandEngine()
	for i := 0; i < 300; i++ {
		d.Update(1, zm)
	}
	if d.Office < 0 || d.Office > 1 {
		t.Fatalf("office demand out of range: %v", d.Office)
	}
	if d.Education <= 0 {
		t.Fatal("education should rise over time")
	}
	if d.Industrial > 0 && d.Education > 0 && d.Office <= 0 {
		t.Fatal("office demand should be positive with industrial demand and education")
	}
}

func TestDemandUpdate_continuous(t *testing.T) {
	d := NewDemandEngine()
	zm := NewZoneManager(4, 4, nil, nil)
	zm.SetZoneCell(0, 0, ZoneResidentialLow)
	before := d.Residential
	for i := 0; i < 10; i++ {
		d.Update(0.5, zm)
	}
	if d.Residential == before {
		t.Fatal("demand should change continuously")
	}
}

func TestResidentialDemand_taxEffect(t *testing.T) {
	zm := NewZoneManager(4, 4, nil, nil)
	for x := 0; x < 4; x++ {
		zm.SetZoneCell(x, 0, ZoneIndustrial)
	}

	low := NewDemandEngine()
	low.Factors.ResidentialTax = 0.03
	for i := 0; i < 40; i++ {
		low.Update(1, zm)
	}

	high := NewDemandEngine()
	high.Factors.ResidentialTax = 0.22
	for i := 0; i < 40; i++ {
		high.Update(1, zm)
	}
	if high.Residential >= low.Residential {
		t.Fatalf("high tax should lower demand: low=%v high=%v", low.Residential, high.Residential)
	}
}

func TestResidentialDemand_pollutionAndCrime(t *testing.T) {
	zm := NewZoneManager(4, 4, nil, nil)
	for x := 0; x < 4; x++ {
		zm.SetZoneCell(x, 0, ZoneIndustrial)
	}

	clean := NewDemandEngine()
	clean.Factors.Pollution = 0
	clean.Factors.Crime = 0
	for i := 0; i < 40; i++ {
		clean.Update(1, zm)
	}

	dirty := NewDemandEngine()
	dirty.Factors.Pollution = 0.9
	dirty.Factors.Crime = 0.8
	for i := 0; i < 40; i++ {
		dirty.Update(1, zm)
	}
	if dirty.Residential >= clean.Residential {
		t.Fatalf("pollution/crime should lower demand: clean=%v dirty=%v", clean.Residential, dirty.Residential)
	}
}

func TestResidentialDemand_jobsAndUnemployment(t *testing.T) {
	jobsHeavy := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 6; x++ {
		for z := 0; z < 6; z++ {
			jobsHeavy.SetZoneCell(x, z, ZoneIndustrial)
		}
	}
	d := NewDemandEngine()
	for i := 0; i < 50; i++ {
		d.Update(1, jobsHeavy)
	}
	if d.Residential <= 0.5 {
		t.Fatalf("job surplus should lift residential demand, got %v", d.Residential)
	}

	resHeavy := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 6; x++ {
		for z := 0; z < 6; z++ {
			resHeavy.SetZoneCell(x, z, ZoneResidentialLow)
		}
	}
	d2 := NewDemandEngine()
	for i := 0; i < 50; i++ {
		d2.Update(1, resHeavy)
	}
	if d2.Residential >= d.Residential {
		t.Fatalf("housing surplus should lower demand vs job surplus: resHeavy=%v jobsHeavy=%v", d2.Residential, d.Residential)
	}
}

func TestCommercialDemand_populationBoost(t *testing.T) {
	sparse := NewZoneManager(8, 8, nil, nil)
	sparse.SetZoneCell(0, 0, ZoneResidentialLow)

	dense := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 6; x++ {
		for z := 0; z < 6; z++ {
			dense.SetZoneCell(x, z, ZoneResidentialLow)
		}
	}

	dSparse := NewDemandEngine()
	dDense := NewDemandEngine()
	for i := 0; i < 50; i++ {
		dSparse.Update(1, sparse)
		dDense.Update(1, dense)
	}
	if dDense.Commercial <= dSparse.Commercial {
		t.Fatalf("more population should raise commercial demand: sparse=%v dense=%v", dSparse.Commercial, dDense.Commercial)
	}
}

func TestCommercialDemand_oversupplyPenalty(t *testing.T) {
	balanced := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 4; x++ {
		balanced.SetZoneCell(x, 0, ZoneResidentialLow)
	}
	balanced.SetZoneCell(0, 1, ZoneCommercialLow)

	oversupplied := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 2; x++ {
		oversupplied.SetZoneCell(x, 0, ZoneResidentialLow)
	}
	for x := 0; x < 8; x++ {
		for z := 0; z < 6; z++ {
			oversupplied.SetZoneCell(x, z, ZoneCommercialLow)
		}
	}

	dBalanced := NewDemandEngine()
	dOver := NewDemandEngine()
	for i := 0; i < 60; i++ {
		dBalanced.Update(1, balanced)
		dOver.Update(1, oversupplied)
	}
	if dOver.Commercial >= dBalanced.Commercial {
		t.Fatalf("commercial oversupply should lower demand: balanced=%v over=%v", dBalanced.Commercial, dOver.Commercial)
	}
}

func TestCommercialDemand_incomeAndTourism(t *testing.T) {
	zm := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 4; x++ {
		zm.SetZoneCell(x, 0, ZoneResidentialLow)
	}
	for x := 0; x < 3; x++ {
		zm.SetZoneCell(x, 1, ZoneIndustrial)
	}

	low := NewDemandEngine()
	low.Factors.HouseholdIncome = 0.3
	low.Factors.Tourism = 0.05
	for i := 0; i < 40; i++ {
		low.Update(1, zm)
	}

	high := NewDemandEngine()
	high.Factors.HouseholdIncome = 0.85
	high.Factors.Tourism = 0.7
	for i := 0; i < 40; i++ {
		high.Update(1, zm)
	}
	if high.Commercial <= low.Commercial {
		t.Fatalf("income/tourism should raise commercial demand: low=%v high=%v", low.Commercial, high.Commercial)
	}
}

func TestIndustrialDemand_populationGrowth(t *testing.T) {
	growing := NewDemandEngine()
	zm := NewZoneManager(8, 8, nil, nil)
	for step := 0; step < 25; step++ {
		x, z := step%6, step/6
		zm.SetZoneCell(x, z, ZoneResidentialLow)
		growing.Update(1, zm)
	}

	static := NewDemandEngine()
	staticZM := NewZoneManager(8, 8, nil, nil)
	staticZM.SetZoneCell(0, 0, ZoneResidentialLow)
	for i := 0; i < 25; i++ {
		static.Update(1, staticZM)
	}
	if growing.Industrial <= static.Industrial {
		t.Fatalf("population growth should lift industrial demand: growing=%v static=%v", growing.Industrial, static.Industrial)
	}
}

func TestIndustrialDemand_goodsShortage(t *testing.T) {
	shortageZM := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 6; x++ {
		for z := 0; z < 6; z++ {
			shortageZM.SetZoneCell(x, z, ZoneResidentialLow)
		}
	}
	shortageZM.SetZoneCell(0, 6, ZoneIndustrial)

	surplusZM := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 6; x++ {
		for z := 0; z < 6; z++ {
			surplusZM.SetZoneCell(x, z, ZoneResidentialLow)
		}
	}
	for x := 0; x < 6; x++ {
		surplusZM.SetZoneCell(x, 6, ZoneIndustrial)
	}

	shortage := NewDemandEngine()
	for i := 0; i < 50; i++ {
		shortage.Update(1, shortageZM)
	}

	surplus := NewDemandEngine()
	for i := 0; i < 50; i++ {
		surplus.Update(1, surplusZM)
	}
	if shortage.Industrial <= surplus.Industrial {
		t.Fatalf("goods shortage should raise industrial demand: shortage=%v surplus=%v", shortage.Industrial, surplus.Industrial)
	}
}

func TestIndustrialDemand_workerShortage(t *testing.T) {
	workers := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 5; x++ {
		for z := 0; z < 5; z++ {
			workers.SetZoneCell(x, z, ZoneResidentialLow)
		}
	}
	for x := 0; x < 2; x++ {
		workers.SetZoneCell(x, 5, ZoneIndustrial)
	}

	noWorkers := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 8; x++ {
		for z := 0; z < 6; z++ {
			noWorkers.SetZoneCell(x, z, ZoneIndustrial)
		}
	}

	dWorkers := NewDemandEngine()
	dNoWorkers := NewDemandEngine()
	for i := 0; i < 50; i++ {
		dWorkers.Update(1, workers)
		dNoWorkers.Update(1, noWorkers)
	}
	if dNoWorkers.Industrial >= dWorkers.Industrial {
		t.Fatalf("worker shortage should lower industrial demand: workers=%v none=%v", dWorkers.Industrial, dNoWorkers.Industrial)
	}
}

func TestIndustrialDemand_highTax(t *testing.T) {
	zm := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 4; x++ {
		zm.SetZoneCell(x, 0, ZoneResidentialLow)
	}
	for x := 0; x < 2; x++ {
		zm.SetZoneCell(x, 1, ZoneIndustrial)
	}

	lowTax := NewDemandEngine()
	lowTax.Factors.IndustrialTax = 0.04
	for i := 0; i < 40; i++ {
		lowTax.Update(1, zm)
	}

	highTax := NewDemandEngine()
	highTax.Factors.IndustrialTax = 0.22
	for i := 0; i < 40; i++ {
		highTax.Update(1, zm)
	}
	if highTax.Industrial >= lowTax.Industrial {
		t.Fatalf("high industrial tax should lower demand: low=%v high=%v", lowTax.Industrial, highTax.Industrial)
	}
}

func TestOfficeDemand_educationAndPolicy(t *testing.T) {
	zm := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 4; x++ {
		zm.SetZoneCell(x, 0, ZoneResidentialLow)
	}
	zm.SetZoneCell(0, 1, ZoneIndustrial)

	low := NewDemandEngine()
	low.Education = 0.1
	low.Factors.OfficePolicy = 0.2
	for i := 0; i < 60; i++ {
		low.Update(1, zm)
	}

	high := NewDemandEngine()
	high.Education = 0.7
	high.Factors.OfficePolicy = 0.9
	for i := 0; i < 60; i++ {
		high.Update(1, zm)
	}
	if high.Office <= low.Office {
		t.Fatalf("education/policy should raise office demand: low=%v high=%v", low.Office, high.Office)
	}
}

func TestOfficeDemand_lowIndustrialBoost(t *testing.T) {
	d := NewDemandEngine()
	d.Education = 0.6
	d.Industrial = 0.15
	d.Factors.EducatedWorkers = 0.4
	d.Factors.HighTechEconomy = 0.45
	d.Factors.OfficePolicy = 0.5

	withBoost := d.officeTarget()
	baseOnly := clampf(d.Industrial*d.Education, 0, 1)
	if withBoost <= baseOnly {
		t.Fatalf("low industrial demand modifier should boost office: base=%v target=%v", baseOnly, withBoost)
	}
}

func TestOfficeDemand_educationConsumption(t *testing.T) {
	zm := NewZoneManager(8, 8, nil, nil)
	for x := 0; x < 4; x++ {
		for z := 0; z < 4; z++ {
			zm.SetZoneCell(x, z, ZoneOffice)
		}
	}

	d := NewDemandEngine()
	d.Education = 0.8
	for i := 0; i < 100; i++ {
		d.Update(1, zm)
	}
	if d.Education >= 0.79 {
		t.Fatalf("office zones should consume education: got %v", d.Education)
	}
}
