package ui

import (
	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SelectKind is the unified selection target (24.17).
type SelectKind int

const (
	SelectNone SelectKind = iota
	SelectBuilding
	SelectRoad
	SelectCitizen
	SelectVehicle
	SelectDistrict
	SelectTree
	SelectProp
)

// SelectionSystem highlights targets and exposes contextual actions (24.17).
type SelectionSystem struct {
	kind   SelectKind
	x, z   float32
	active bool
}

func NewSelectionSystem() *SelectionSystem { return &SelectionSystem{} }

func (s *SelectionSystem) Clear() {
	s.kind = SelectNone
	s.active = false
}

func (s *SelectionSystem) FromInspector(sel Selection) {
	if sel.Kind == InspNone {
		s.Clear()
		return
	}
	s.active = true
	s.x, s.z = sel.followX, sel.followZ
	switch sel.Kind {
	case InspBuilding, InspIndustry:
		s.kind = SelectBuilding
	case InspRoad:
		s.kind = SelectRoad
	case InspCitizen:
		s.kind = SelectCitizen
	case InspVehicle:
		s.kind = SelectVehicle
	case InspZone, InspDistrict:
		s.kind = SelectDistrict
	default:
		s.kind = SelectProp
	}
}

func (s *SelectionSystem) Pick(sm *sim.SimulationManager, x, z float32, inspector *InspectorPanel) {
	inspector.Pick(sm, x, z)
	s.FromInspector(inspector.Selection())
}

func (s *SelectionSystem) Active() bool { return s.active }

func (s *SelectionSystem) Target() (x, z float32, ok bool) {
	if !s.active {
		return 0, 0, false
	}
	return s.x, s.z, true
}

func (s *SelectionSystem) DrawHighlight(sm *sim.SimulationManager) {
	if !s.active || sm == nil {
		return
	}
	h := sm.Heightmap.WorldHeight(s.x, s.z)
	col := rl.NewColor(255, 255, 120, 180)
	switch s.kind {
	case SelectBuilding:
		rl.DrawCubeWires(rl.NewVector3(s.x, h+1.5, s.z), 4, 3, 4, col)
	case SelectRoad:
		rl.DrawSphere(rl.NewVector3(s.x, h+0.5, s.z), 1.2, rl.NewColor(255, 200, 80, 100))
	case SelectVehicle:
		rl.DrawSphere(rl.NewVector3(s.x, h+0.8, s.z), 1.0, col)
	case SelectCitizen:
		rl.DrawSphere(rl.NewVector3(s.x, h+1, s.z), 0.6, rl.SkyBlue)
	case SelectTree:
		rl.DrawCylinder(rl.NewVector3(s.x, h, s.z), 0.4, 0.4, 2, 6, rl.Green)
	default:
		rl.DrawSphere(rl.NewVector3(s.x, h+0.5, s.z), 0.8, col)
	}
}
