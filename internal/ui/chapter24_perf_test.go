package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUIStateRoundTrip(t *testing.T) {
	m := NewManager()
	m.BuildMenus.favorites["road_Two Lane"] = true
	m.BuildMenus.showFav = true
	m.Camera.Dist = 200
	m.Settings.A11y.UIScale = 1.25
	m.Settings.Locale = "es"
	m.InfoViews.Set(ViewTraffic)

	path := filepath.Join(t.TempDir(), "ui.prefs")
	if err := m.SavePreferences(path); err != nil {
		t.Fatal(err)
	}
	m2 := NewManager()
	if err := m2.LoadPreferences(path); err != nil {
		t.Fatal(err)
	}
	if !m2.BuildMenus.favorites["road_Two Lane"] || !m2.BuildMenus.showFav {
		t.Fatal("favorites/filter not restored")
	}
	if m2.Camera.Dist != 200 {
		t.Fatalf("camera dist=%v", m2.Camera.Dist)
	}
	if m2.Settings.A11y.UIScale != 1.25 {
		t.Fatalf("ui scale=%v", m2.Settings.A11y.UIScale)
	}
	if m2.InfoViews.Active() != ViewTraffic {
		t.Fatalf("info view=%v", m2.InfoViews.Active())
	}
}

func TestTransientPanelsNotRestoredByDefault(t *testing.T) {
	state := DefaultUIState()
	state.Panels[string(PanelSearch)] = PanelLayout{Visible: true, Persist: false}
	m := NewManager()
	m.ApplyState(state)
	if m.Search.IsOpen() {
		t.Fatal("transient search should not restore without persist")
	}
	state.Panels[string(PanelSearch)] = PanelLayout{Visible: true, Persist: true}
	m.ApplyState(state)
	if !m.Search.IsOpen() {
		t.Fatal("persisted search should restore")
	}
}

func TestBuildMenuVirtualRange(t *testing.T) {
	b := NewBuildMenus(NewUnlockRegistry())
	b.open = true
	b.filtered = make([]int, 40)
	b.scrollRow = 2
	start, end := b.visibleAssetRange()
	if start != 12 || end > 34 {
		t.Fatalf("range=%d..%d", start, end)
	}
}

func TestNotificationsBatchSkip(t *testing.T) {
	n := NewNotifications()
	n.Refresh(nil, ViewState{}, false)
}

func TestSaveUIStateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	if err := SaveUIState(path, DefaultUIState()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
