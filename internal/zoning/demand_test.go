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
