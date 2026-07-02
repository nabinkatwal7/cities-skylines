package ui

import (
	"testing"

	simpkg "github.com/katwate/js-skylines/internal/sim"
)

func TestCameraReset(t *testing.T) {
	c := NewCameraController()
	c.goalTarget.X = 100
	c.goalTarget.Z = 50
	c.goalDist = 20
	c.Reset()
	if c.goalTarget.X != 0 || c.goalTarget.Z != 0 || c.goalDist != 140 {
		t.Fatalf("reset target=%v dist=%v", c.goalTarget, c.goalDist)
	}
}

func TestBindingsUndoRequiresCtrl(t *testing.T) {
	b := DefaultBindings()
	if b.Get(ActionUndo) != 0 && b.Get(ActionUndo) == b.Get(ActionCamForward) {
		t.Fatal("unexpected binding overlap")
	}
}

func TestLocalizationSwitch(t *testing.T) {
	SetLocale("es")
	if T("hud.speed") != "Velocidad" {
		t.Fatalf("spanish speed=%q", T("hud.speed"))
	}
	SetLocale("en")
	if T("hud.speed") != "Speed" {
		t.Fatalf("english speed=%q", T("hud.speed"))
	}
}

func TestGlobalShortcutsStatisticsToggle(t *testing.T) {
	m := NewManager()
	m.lastSim = simpkg.NewSimulationManager(1)
	m.Shortcuts.Handle(m)
	if m.Statistics.open {
		t.Fatal("stats should start closed")
	}
}

func TestAccessibilityDefaults(t *testing.T) {
	a := DefaultAccessibility()
	if a.UIScale != 1 || a.ReducedMotion {
		t.Fatalf("defaults=%+v", a)
	}
}
