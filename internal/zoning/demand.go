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
}

func NewDemandEngine() *DemandEngine {
	return &DemandEngine{
		Residential: 0.5,
		Commercial:  0.5,
		Industrial:  0.5,
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

	resTarget := clampf((jobCap-resCap)*0.15, -1, 1)
	comTarget := clampf((resCap*0.8-comCap)*0.15, -1, 1)
	indTarget := clampf((resCap*0.4-indCap)*0.12, -1, 1)

	rate := float32(dt) * 0.35
	d.Residential = lerp(d.Residential, resTarget, rate)
	d.Commercial = lerp(d.Commercial, comTarget, rate)
	d.Industrial = lerp(d.Industrial, indTarget, rate)

	eduGrowth := float32(dt) * 0.00005 * (1 + resCap*0.1)
	d.Education = clampf(d.Education+eduGrowth, 0, 1)
	d.Office = clampf(d.Industrial*d.Education, 0, 1)
}

func clampf(v, lo, hi float32) float32 {
	return float32(math.Max(float64(lo), math.Min(float64(hi), float64(v))))
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}
