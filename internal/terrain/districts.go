package terrain

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type DistrictPolicy uint8

const (
	PolicyNone DistrictPolicy = iota
	PolicyHighRiseBan
	PolicyHeavyTrafficBan
	PolicySelfSufficient
	PolicyITCluster
	PolicyOrganicProduce
	PolicyBigBusiness
)

type District struct {
	ID       uint32
	Name     string
	CenterX  float32
	CenterZ  float32
	Radius   float32
	Policies []DistrictPolicy
	Color    rl.Color
}

type DistrictManager struct {
	Districts []District
	NextID    uint32
}

func NewDistrictManager() *DistrictManager {
	return &DistrictManager{}
}

func (dm *DistrictManager) AddDistrict(name string, cx, cz, radius float32) uint32 {
	id := dm.NextID
	dm.NextID++
	col := rl.NewColor(
		uint8(100+id*50%100),
		uint8(80+id*30%100),
		uint8(120+id*70%100),
		100,
	)
	dm.Districts = append(dm.Districts, District{
		ID:      id,
		Name:    name,
		CenterX: cx,
		CenterZ: cz,
		Radius:  radius,
		Color:   col,
	})
	return id
}

func (dm *DistrictManager) SetPolicy(distIdx int, policy DistrictPolicy) {
	if distIdx < 0 || distIdx >= len(dm.Districts) {
		return
	}
	d := &dm.Districts[distIdx]
	for _, p := range d.Policies {
		if p == policy {
			return
		}
	}
	d.Policies = append(d.Policies, policy)
}

func (dm *DistrictManager) RemovePolicy(distIdx int, policy DistrictPolicy) {
	if distIdx < 0 || distIdx >= len(dm.Districts) {
		return
	}
	d := &dm.Districts[distIdx]
	filtered := d.Policies[:0]
	for _, p := range d.Policies {
		if p != policy {
			filtered = append(filtered, p)
		}
	}
	d.Policies = filtered
}

func (dm *DistrictManager) DistrictAt(x, z float32) int {
	for i, d := range dm.Districts {
		dx := d.CenterX - x
		dz := d.CenterZ - z
		if dx*dx+dz*dz < d.Radius*d.Radius {
			return i
		}
	}
	return -1
}

func (dm *DistrictManager) HasPolicy(x, z float32, policy DistrictPolicy) bool {
	idx := dm.DistrictAt(x, z)
	if idx < 0 {
		return false
	}
	for _, p := range dm.Districts[idx].Policies {
		if p == policy {
			return true
		}
	}
	return false
}

func (dm *DistrictManager) ApplyPolicies(b *Building) {
	if b.Household == nil && b.Business == nil {
		return
	}
	idx := dm.DistrictAt(b.X, b.Z)
	if idx < 0 {
		return
	}
	d := &dm.Districts[idx]
	for _, p := range d.Policies {
		switch p {
		case PolicyHighRiseBan:
			if b.Level > 3 {
				b.Level = 3
				b.Height = buildingHeight(b.Type, b.Seed)
			}
		case PolicySelfSufficient:
			if b.Household != nil {
				b.Consumption.Power *= 0.6
				b.Consumption.Water *= 0.6
			}
		case PolicyITCluster:
			if b.Type == ZoneOffice && b.Business != nil {
				b.Business.Profitability += 5
			}
		case PolicyOrganicProduce:
			if b.Type == ZoneCommercialLow && b.Business != nil {
				b.Business.Profitability += 3
			}
		case PolicyBigBusiness:
			if b.Type == ZoneCommercialHigh && b.Business != nil {
				b.Business.Profitability += 4
			}
		}
	}
}

func (dm *DistrictManager) Draw(h *Heightmap) {
	for _, d := range dm.Districts {
		hy := h.WorldHeight(d.CenterX, d.CenterZ) + 0.1
		steps := 20
		for ai := 0; ai < steps; ai++ {
			a1 := float32(ai) / float32(steps) * 6.2832
			a2 := float32(ai+1) / float32(steps) * 6.2832
			x1 := d.CenterX + float32(math.Cos(float64(a1)))*d.Radius
			z1 := d.CenterZ + float32(math.Sin(float64(a1)))*d.Radius
			x2 := d.CenterX + float32(math.Cos(float64(a2)))*d.Radius
			z2 := d.CenterZ + float32(math.Sin(float64(a2)))*d.Radius
			h1 := h.WorldHeight(x1, z1)
			h2 := h.WorldHeight(x2, z2)
			rl.DrawTriangle3D(
				rl.NewVector3(x1, h1+0.5, z1),
				rl.NewVector3(x2, h2+0.5, z2),
				rl.NewVector3(d.CenterX, hy+0.5, d.CenterZ),
				d.Color,
			)
		}
	}
}

func (dm *DistrictManager) Unload() {
	dm.Districts = nil
}
