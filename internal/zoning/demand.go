package zoning

import "math"

const demandThreshold = float32(0.05)

// DemandEngine drives RCI growth; office demand is derived from industrial × education.
type DemandEngine struct {
	Residential float32
	Commercial  float32
	Industrial  float32
	Office      float32

	Education  float32 // 0..1 city education level
	Population float32 // ponytail: zone proxy until household sim exists
	Jobs       float32 // ponytail: zone proxy until business sim exists
	Factors    CityFactors

	prevPopulation float32
}

// CityFactors tune residential demand; ponytail: derived fields refresh in Update until dedicated sims exist.
type CityFactors struct {
	Happiness      float32 // 0..1
	ResidentialTax float32 // 0..1
	LandValue      float32 // 0..1
	ServiceScore   float32 // 0..1 electricity/water/sewage/police etc.
	Immigration    float32 // 0..1 inflow
	Pollution      float32 // 0..1
	Crime          float32 // 0..1
	DeathWave      float32 // 0..1 active wave strength
	Abandonment    float32 // 0..1 abandoned residential share

	HouseholdIncome   float32 // 0..1
	GoodsAvailability float32 // 0..1
	Tourism           float32 // 0..1
	ShoppingDemand    float32 // 0..1

	IndustrialTax      float32 // 0..1
	PopulationGrowth   float32 // change rate proxy
	GoodsShortage      float32 // 0..1
	ExportOpportunity  float32 // 0..1
	WorkerShortage     float32 // 0..1
	FreightCongestion  float32 // 0..1 ponytail: until freight sim
	ResourceShortage   float32 // 0..1 ponytail: until resource economy hooks in
}

func DefaultCityFactors() CityFactors {
	return CityFactors{
		Happiness:      0.55,
		ResidentialTax: 0.09,
		LandValue:         0.5,
		ServiceScore:      1.0,
		HouseholdIncome:   0.5,
		GoodsAvailability: 0.6,
		Tourism:           0.2,
		ShoppingDemand:    0.5,
		IndustrialTax:     0.10,
	}
}

func NewDemandEngine() *DemandEngine {
	return &DemandEngine{
		Residential: 0.5,
		Commercial:  0.5,
		Industrial:  0.5,
		Factors:     DefaultCityFactors(),
	}
}

func (d *DemandEngine) Value(cat ZoneCategory) float32 {
	if d == nil {
		return 0
	}
	switch cat {
	case CategoryResidential:
		return d.Residential
	case CategoryCommercial:
		return d.Commercial
	case CategoryIndustrial:
		return d.Industrial
	case CategoryOffice:
		return d.Office
	default:
		return 0
	}
}

func (d *DemandEngine) HasDemand(cat ZoneCategory) bool {
	return d.Value(cat) > demandThreshold
}

// Update continuously adjusts RCI demand from zoning balance and derives office demand.
func (d *DemandEngine) Update(dt float64, zm *ZoneManager) {
	if d == nil || zm == nil || dt <= 0 {
		return
	}

	resCells, comCells, indCells, offCells := zm.CategoryCounts()
	resCap := float32(resCells) * 0.25
	comCap := float32(comCells) * 0.25
	indCap := float32(indCells) * 0.3
	jobCap := comCap + indCap + float32(offCells)*0.2

	d.Population = resCap
	d.Jobs = jobCap
	d.refreshResidentialFactors(zm, resCells, comCells, indCells, offCells, float32(dt))
	d.refreshCommercialFactors(resCells, comCells, indCells, float32(dt))
	d.refreshIndustrialFactors(indCap, float32(dt))

	resTarget := clampf((jobCap-resCap)*0.15+d.residentialModifier(), -1, 1)
	comTarget := clampf((resCap*0.8-comCap)*0.15+d.commercialModifier(comCap), -1, 1)
	indTarget := clampf((resCap*0.4-indCap)*0.12+d.industrialModifier(indCap), -1, 1)

	rate := float32(dt) * 0.35
	d.Residential = lerp(d.Residential, resTarget, rate)
	d.Commercial = lerp(d.Commercial, comTarget, rate)
	d.Industrial = lerp(d.Industrial, indTarget, rate)

	eduGrowth := float32(dt) * 0.00005 * (1 + resCap*0.1)
	d.Education = clampf(d.Education+eduGrowth, 0, 1)
	d.Office = clampf(d.Industrial*d.Education, 0, 1)
}

func (d *DemandEngine) refreshResidentialFactors(zm *ZoneManager, res, com, ind, off int, dt float32) {
	total := res + com + ind + off
	if total > 0 {
		d.Factors.Pollution = clampf(float32(ind)/float32(total)*0.9, 0, 1)
	}
	if d.Jobs > d.Population {
		d.Factors.Immigration = clampf(d.Factors.Immigration+dt*0.002, 0, 1)
	} else {
		d.Factors.Immigration = clampf(d.Factors.Immigration-dt*0.004, 0, 1)
	}
	_ = zm
}

func (d *DemandEngine) residentialModifier() float32 {
	f := d.Factors
	mod := float32(0)

	if d.Jobs > d.Population {
		mod += clampf((d.Jobs-d.Population)*0.04, 0, 0.25)
	} else if d.Population > 0 {
		mod -= clampf((d.Population-d.Jobs)/d.Population*0.35, 0, 0.35)
	}

	mod += (f.Happiness - 0.5) * 0.35
	mod += (0.12 - f.ResidentialTax) * 1.0
	mod += (f.ServiceScore - 0.5) * 0.3
	mod += f.Immigration * 0.25
	mod += (f.LandValue - 0.5) * 0.25
	mod -= f.Pollution * 0.35
	mod -= f.Crime * 0.3
	mod -= f.DeathWave * 0.45
	mod -= f.Abandonment * 0.35

	return clampf(mod, -0.6, 0.6)
}

func (d *DemandEngine) refreshCommercialFactors(res, com, ind int, dt float32) {
	if d.Population > 0 {
		goodsSupply := float32(ind) * 0.3
		d.Factors.GoodsAvailability = clampf(goodsSupply/d.Population, 0, 1)
		d.Factors.HouseholdIncome = clampf(0.35+(d.Jobs/d.Population)*0.25+d.Factors.Happiness*0.2, 0, 1)
	}
	d.Factors.ShoppingDemand = clampf(d.Population*0.06+d.Factors.HouseholdIncome*0.35, 0, 1)
	d.Factors.Tourism = clampf(d.Factors.Tourism+dt*0.00015*d.Population, 0, 1)
	_ = res
	_ = com
}

func (d *DemandEngine) commercialModifier(comCap float32) float32 {
	f := d.Factors
	mod := float32(0)

	mod += clampf(d.Population*0.08, 0, 0.3)
	mod += (f.HouseholdIncome - 0.5) * 0.35
	mod += (f.GoodsAvailability - 0.5) * 0.25
	mod += f.Tourism * 0.2
	mod += (f.ShoppingDemand - 0.5) * 0.3

	if d.Population > 0 {
		ratio := comCap / d.Population
		if ratio > 1.2 {
			mod -= clampf((ratio-1.2)*0.45, 0, 0.55)
		}
	} else if comCap > 0 {
		mod -= 0.35
	}

	return clampf(mod, -0.6, 0.6)
}

func (d *DemandEngine) refreshIndustrialFactors(indCap, dt float32) {
	growth := d.Population - d.prevPopulation
	d.Factors.PopulationGrowth = clampf(growth, -0.5, 0.5)
	d.prevPopulation = d.Population

	d.Factors.GoodsShortage = clampf(1-d.Factors.GoodsAvailability, 0, 1)

	workersNeeded := indCap * 0.85
	if workersNeeded > 0 && d.Population < workersNeeded {
		d.Factors.WorkerShortage = clampf((workersNeeded-d.Population)/workersNeeded, 0, 1)
	} else {
		d.Factors.WorkerShortage = 0
	}

	_ = dt
}

func (d *DemandEngine) industrialModifier(indCap float32) float32 {
	f := d.Factors
	mod := float32(0)

	if f.PopulationGrowth > 0 {
		mod += clampf(f.PopulationGrowth*0.5, 0, 0.25)
	}
	mod += f.GoodsShortage * 0.3
	mod += f.ExportOpportunity * 0.25
	mod -= f.WorkerShortage * 0.35
	mod -= f.FreightCongestion * 0.25
	mod -= f.ResourceShortage * 0.3
	mod += (0.12 - f.IndustrialTax) * 0.9

	if f.PopulationGrowth < 0 {
		mod += f.PopulationGrowth * 0.15
	}
	_ = indCap

	return clampf(mod, -0.6, 0.6)
}

func ServiceScore(s ServiceCoverage) float32 {
	if s == nil {
		return 0
	}
	score := float32(0)
	if s.HasElectricity(0, 0) {
		score += 0.34
	}
	if s.HasWater(0, 0) {
		score += 0.33
	}
	if s.HasSewage(0, 0) {
		score += 0.33
	}
	return score
}

func clampf(v, lo, hi float32) float32 {
	return float32(math.Max(float64(lo), math.Min(float64(hi), float64(v))))
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}
