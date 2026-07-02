package zoning

import "testing"

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
