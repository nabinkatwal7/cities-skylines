package ui

import (
	"fmt"
	"strings"

	"github.com/katwate/js-skylines/internal/building"
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InspectorPanel shows entity-specific information panels (24.7–24.10).
type InspectorPanel struct {
	visible   bool
	selection Selection
	following bool
}

func NewInspectorPanel() *InspectorPanel { return &InspectorPanel{} }

func (p *InspectorPanel) Visible() bool { return p.visible }

func (p *InspectorPanel) Selection() Selection { return p.selection }

func (p *InspectorPanel) Following() bool { return p.following }

func (p *InspectorPanel) Clear() {
	p.visible = false
	p.following = false
	p.selection = Selection{}
}

func (p *InspectorPanel) Show(sel Selection) {
	p.selection = sel
	p.visible = sel.Kind != InspNone
}

func (p *InspectorPanel) ShowCitizen(sm *sim.SimulationManager, buildingIdx int) {
	p.Show(buildCitizenSelection(sm, buildingIdx))
}

func (p *InspectorPanel) Pick(sm *sim.SimulationManager, x, z float32) {
	if sm == nil {
		p.Clear()
		return
	}
	if sel, ok := pickTransport(sm, x, z); ok {
		p.Show(sel)
		return
	}
	if idx := sm.Buildings.NearestAt(x, z, 8); idx >= 0 {
		p.Show(buildBuildingSelection(sm, idx))
		return
	}
	if sel, ok := pickVehicle(sm, x, z); ok {
		p.Show(sel)
		return
	}
	if sel, ok := pickRoad(sm, x, z); ok {
		p.Show(sel)
		return
	}
	if sel, ok := pickZone(sm, x, z); ok {
		p.Show(sel)
		return
	}
	p.Clear()
}

func (p *InspectorPanel) HandleAction(id string, sm *sim.SimulationManager, tools *ToolSystem) {
	switch id {
	case "follow":
		p.following = true
	case "citizen":
		if p.selection.buildingIdx >= 0 {
			p.ShowCitizen(sm, p.selection.buildingIdx)
		}
	case "building":
		if p.selection.buildingIdx >= 0 {
			p.Show(buildBuildingSelection(sm, p.selection.buildingIdx))
		}
	case "upgrade":
		tools.Activate(ToolUpgrade)
	case "bulldoze":
		tools.Activate(ToolRemove)
	case "inspect":
		tools.Activate(ToolInspect)
	}
}

func (p *InspectorPanel) HandleClick(mx, my int32, sm *sim.SimulationManager, tools *ToolSystem) bool {
	if !p.visible {
		return false
	}
	w := int32(300)
	x := int32(ScreenW - w - 8)
	y := int32(TopBarH + 8)
	h := p.panelHeight()
	if mx < x || mx >= x+w || my < y || my >= y+h {
		return false
	}
	btnY := y + h - 28
	for i, act := range p.selection.Actions {
		bx := x + 8 + int32(i*72)
		if mx >= bx && mx < bx+68 && my >= btnY && my < btnY+22 {
			p.HandleAction(act.ID, sm, tools)
			return true
		}
	}
	return true
}

func (p *InspectorPanel) panelHeight() int32 {
	n := len(p.selection.Lines)
	h := int32(36 + n*16)
	if len(p.selection.Actions) > 0 {
		h += 32
	}
	if h < 100 {
		h = 100
	}
	if h > 340 {
		h = 340
	}
	return h
}

func (p *InspectorPanel) Draw() {
	if !p.visible || p.selection.Kind == InspNone {
		return
	}
	w := int32(300)
	h := p.panelHeight()
	x := int32(ScreenW - w - 8)
	y := int32(TopBarH + 8)
	rl.DrawRectangle(x, y, w, h, rl.NewColor(0, 0, 0, 220))
	border := rl.NewColor(120, 160, 220, 200)
	switch p.selection.Kind {
	case InspBuilding:
		border = rl.NewColor(120, 200, 140, 200)
	case InspCitizen:
		border = rl.NewColor(200, 180, 120, 200)
	case InspVehicle:
		border = rl.NewColor(120, 180, 220, 200)
	}
	rl.DrawRectangleLines(x, y, w, h, border)
	DrawUIText(p.selection.Title, x+10, y+8, 15, rl.White)
	for i, line := range p.selection.Lines {
		col := rl.LightGray
		if strings.HasPrefix(line, "⚠") {
			col = rl.NewColor(255, 180, 100, 230)
		}
		DrawUIText(line, x+10, y+28+int32(i*16), 13, col)
	}
	if len(p.selection.Actions) > 0 {
		btnY := y + h - 28
		for i, act := range p.selection.Actions {
			bx := x + 8 + int32(i*72)
			uiBtn(bx, btnY, 68, 22, act.Label, rl.NewColor(50, 55, 65, 220), rl.White, false)
		}
	}
	if p.following {
		DrawUIText("Following", x+10, y+h-44, 12, rl.NewColor(255, 220, 120, 220))
	}
}

func pickTransport(sm *sim.SimulationManager, x, z float32) (Selection, bool) {
	if sm.Transport == nil {
		return Selection{}, false
	}
	if stop := sm.Transport.NearestStop(x, z, 6); stop != nil {
		return Selection{
			Kind:  InspTransportLine,
			Title: transport.TypeName(stop.TransType) + " Stop",
			Lines: []string{
				fmt.Sprintf("Passengers: %d", stop.Passengers),
				fmt.Sprintf("Capacity: %d", stop.Capacity),
				fmt.Sprintf("Position: (%.0f, %.0f)", stop.X, stop.Z),
			},
			Actions: []InspectorAction{{ID: "inspect", Label: "Inspect"}},
			followX: stop.X,
			followZ: stop.Z,
		}, true
	}
	if line := sm.Transport.NearestLine(x, z, 10); line != nil {
		return Selection{
			Kind:  InspTransportLine,
			Title: line.Name,
			Lines: []string{
				fmt.Sprintf("Mode: %s", transport.TypeName(line.TransType)),
				fmt.Sprintf("Stops: %d", len(line.Stops)),
				fmt.Sprintf("Passengers: %d", line.PassengerCount),
				fmt.Sprintf("Income: $%.0f", line.TotalIncome),
			},
		}, true
	}
	return Selection{}, false
}

func pickVehicle(sm *sim.SimulationManager, x, z float32) (Selection, bool) {
	if sm.Vehicles == nil {
		return Selection{}, false
	}
	var best *road.Vehicle
	var bestSlot int32 = -1
	var bestD float32 = 36
	sm.Vehicles.ForEach(func(v *road.Vehicle, slot int32) {
		dx := v.Position.X - x
		dz := v.Position.Z - z
		if d := dx*dx + dz*dz; d < bestD {
			bestD = d
			best = v
			bestSlot = slot
		}
	})
	if best == nil {
		return Selection{}, false
	}
	return buildVehicleSelection(sm, bestSlot, best), true
}

func pickRoad(sm *sim.SimulationManager, x, z float32) (Selection, bool) {
	idx := sm.Roads.NearestSegment(x, z)
	if idx < 0 {
		return Selection{}, false
	}
	seg := sm.Roads.Segments[idx]
	na := &sm.Roads.Nodes[seg.NodeA]
	nb := &sm.Roads.Nodes[seg.NodeB]
	dx := (na.X+nb.X)*0.5 - x
	dz := (na.Z+nb.Z)*0.5 - z
	if dx*dx+dz*dz > 64 {
		return Selection{}, false
	}
	return Selection{
		Kind:  InspRoad,
		Title: "Road Segment",
		Lines: []string{
			fmt.Sprintf("Type: %s", road.RoadTypeName(seg.RoadType)),
			fmt.Sprintf("Length: %.0fm", seg.Length),
			fmt.Sprintf("Elevation: %d", seg.Elevation),
			fmt.Sprintf("Maint: $%.0f/wk", seg.MaintenanceCost),
		},
	}, true
}

func pickZone(sm *sim.SimulationManager, x, z float32) (Selection, bool) {
	if sm.Zones == nil {
		return Selection{}, false
	}
	cx := sm.Zones.CellX(x)
	cz := sm.Zones.CellZ(z)
	if cx < 0 || cz < 0 {
		return Selection{}, false
	}
	zt := sm.Zones.CellTypeAt(x, z)
	if zt == zoning.ZoneNone {
		return Selection{}, false
	}
	lines := []string{fmt.Sprintf("Cell: (%d, %d)", cx, cz)}
	p := sm.Zones.PoliciesAt(cx, cz)
	if p != 0 {
		lines = append(lines, fmt.Sprintf("District policy: 0x%x", p))
	}
	kind := InspZone
	title := zoneTypeName(zt) + " Zone"
	if zt == zoning.ZoneIndustrial {
		kind = InspIndustry
		title = "Industrial Zone"
	}
	return Selection{Kind: kind, Title: title, Lines: lines}, true
}

func buildingStateName(s uint8) string {
	switch building.State(s) {
	case building.StateConstructing:
		return "Constructing"
	case building.StateOccupied:
		return "Occupied"
	case building.StateAbandoned:
		return "Abandoned"
	default:
		return "Unknown"
	}
}
