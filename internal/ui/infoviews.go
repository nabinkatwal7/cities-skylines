package ui

import (
	"github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InfoViewKind selects a simulation overlay (24.11). Only one active at a time.
type InfoViewKind int

const (
	ViewNone InfoViewKind = iota
	ViewElectricity
	ViewWater
	ViewSewage
	ViewHeating
	ViewGarbage
	ViewHealthcare
	ViewFireSafety
	ViewCrime
	ViewEducation
	ViewHappiness
	ViewLandValue
	ViewPollution
	ViewNoise
	ViewTraffic
	ViewPublicTransport
	ViewWind
	ViewNaturalResources
	ViewDistricts
	viewCount
)

var infoViewNames = []string{
	"None", "Electricity", "Water", "Sewage", "Heating", "Garbage",
	"Healthcare", "Fire Safety", "Crime", "Education", "Happiness",
	"Land Value", "Pollution", "Noise", "Traffic", "Public Transport",
	"Wind", "Resources", "Districts",
}

// InfoViews toggles analytical map overlays (24.11).
type InfoViews struct {
	active InfoViewKind
}

func NewInfoViews() *InfoViews { return &InfoViews{} }

func (v *InfoViews) Active() InfoViewKind { return v.active }

func (v *InfoViews) Name() string {
	if int(v.active) < len(infoViewNames) {
		return infoViewNames[v.active]
	}
	return "None"
}

func (v *InfoViews) Set(kind InfoViewKind) {
	if kind >= 0 && kind < viewCount {
		v.active = kind
	}
}

func (v *InfoViews) Cycle() {
	v.active++
	if v.active >= viewCount {
		v.active = ViewNone
	}
}

func (v *InfoViews) Clear() { v.active = ViewNone }

func (v *InfoViews) HandleInput() {}

func (v *InfoViews) Draw() {
	if v.active == ViewNone {
		return
	}
	drawPanel(8, TopBarH+6, 200, 52)
	drawLabel("Info View", 16, TopBarH+12, FontSm, csTextDim)
	drawLabel(v.Name(), 16, TopBarH+30, FontMd, csBarLine)
}

// DrawWorld renders the active overlay on the terrain (24.11).
func (v *InfoViews) DrawWorld(sm *sim.SimulationManager) {
	if v.active == ViewNone || sm == nil || sm.Zones == nil {
		return
	}
	w := sm.Zones.Width()
	h := sm.Zones.Height()
	if w == 0 || h == 0 {
		return
	}
	cs := float32(terrain.WorldSize) / float32(w)
	half := float32(terrain.WorldSize) / 2
	for z := 0; z < h; z++ {
		for x := 0; x < w; x++ {
			val := v.cellValue(sm, x, z)
			if val < 0 {
				continue
			}
			wx := float32(x)*cs - half + cs*0.5
			wz := float32(z)*cs - half + cs*0.5
			hy := sm.Heightmap.WorldHeight(wx, wz) + 0.15
			col := heatColor(val)
			col.A = 100
			rl.DrawCube(rl.NewVector3(wx, hy, wz), cs*0.9, 0.2, cs*0.9, col)
		}
	}
}

func (v *InfoViews) cellValue(sm *sim.SimulationManager, x, z int) float32 {
	if sm.Demand == nil {
		return -1
	}
	f := sm.Demand.Factors
	switch v.active {
	case ViewElectricity:
		if sm.Services != nil && sm.Services.HasElectricity(0, 0) {
			return 1
		}
		return 0
	case ViewWater:
		if sm.Services != nil && sm.Services.HasWater(0, 0) {
			return 1
		}
		return 0
	case ViewSewage:
		if sm.Services != nil && sm.Services.HasSewage(0, 0) {
			return 1
		}
		return 0
	case ViewHeating:
		return 0.5 // ponytail: until heating network exists
	case ViewGarbage:
		return 1 - f.ServiceScore*0.5
	case ViewHealthcare, ViewFireSafety:
		return f.ServiceScore
	case ViewCrime:
		return f.Crime
	case ViewEducation:
		return sm.Demand.Education
	case ViewHappiness:
		return f.Happiness
	case ViewLandValue:
		if sm.Buildings != nil {
			if grid, gw, gh := sm.Buildings.LandValueGrid(); grid != nil && z < gh && x < gw {
				return grid[z][x]
			}
		}
		return f.LandValue
	case ViewPollution:
		return f.Pollution
	case ViewNoise:
		return f.Pollution * 0.7 // ponytail: proxy until noise sim exists
	case ViewTraffic:
		if sm.Roads != nil {
			idx := sm.Roads.NearestSegment(
				float32(x)*terrain.WorldSize/float32(sm.Zones.Width())-terrain.WorldSize/2,
				float32(z)*terrain.WorldSize/float32(sm.Zones.Height())-terrain.WorldSize/2,
			)
			if idx >= 0 {
				return 0.6
			}
		}
		return 0.1
	case ViewPublicTransport:
		if sm.Transport != nil && len(sm.Transport.Stops) > 0 {
			return 0.8
		}
		return 0
	case ViewWind:
		return 0.4
	case ViewNaturalResources:
		return 0.3
	case ViewDistricts:
		p := sm.Zones.PoliciesAt(x, z)
		if p == 0 {
			return -1
		}
		return 0.5 + float32(p%8)*0.06
	default:
		return -1
	}
}

func heatColor(t float32) rl.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	r := uint8(min(255, int(255*t)))
	g := uint8(min(255, int(255*(1-abs(t-0.5)*2))))
	b := uint8(min(255, int(255*(1-t))))
	return rl.NewColor(r, g, b, 255)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
