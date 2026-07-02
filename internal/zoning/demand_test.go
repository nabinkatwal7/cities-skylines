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
	d := NewDemandEngine()
	for i := 0; i < 300; i++ {
		d.Update(1, zm)
	}
	want := d.Industrial * d.Education
	if want < 0 {
		want = 0
	}
	if want > 1 {
		want = 1
	}
	if d.Office != want {
		t.Fatalf("office=%v want industrial×education=%v", d.Office, want)
	}
	if d.Education <= 0 {
		t.Fatal("education should rise over time")
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
