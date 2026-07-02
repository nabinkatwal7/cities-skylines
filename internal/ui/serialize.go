package ui

import (
	"encoding/json"
	"os"
)

const DefaultUIStateFile = "ui.prefs"

// SaveUIState writes persistent UI preferences to disk (24.25).
func SaveUIState(path string, state PersistentUIState) error {
	if path == "" {
		path = DefaultUIStateFile
	}
	state.Version = uiStateVersion
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadUIState reads persistent UI preferences (24.25).
func LoadUIState(path string) (PersistentUIState, error) {
	if path == "" {
		path = DefaultUIStateFile
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return PersistentUIState{}, err
	}
	var state PersistentUIState
	if err := json.Unmarshal(data, &state); err != nil {
		return PersistentUIState{}, err
	}
	if state.Panels == nil {
		state.Panels = DefaultUIState().Panels
	}
	return state, nil
}

// CaptureState gathers current UI into a serializable snapshot (24.25).
func (m *UIManager) CaptureState() PersistentUIState {
	s := DefaultUIState()
	if m == nil {
		return s
	}
	s.ShowFavFilter = m.BuildMenus.showFav
	s.ShowRecentFilter = m.BuildMenus.showRecent
	for id := range m.BuildMenus.favorites {
		s.Favorites = append(s.Favorites, id)
	}
	s.Recent = append(s.Recent, m.BuildMenus.recent...)
	s.CameraYaw = m.Camera.Yaw
	s.CameraPitch = m.Camera.Pitch
	s.CameraDist = m.Camera.Dist
	s.CameraTargetX = m.Camera.Target.X
	s.CameraTargetY = m.Camera.Target.Y
	s.CameraTargetZ = m.Camera.Target.Z
	s.InfoView = int(m.InfoViews.Active())
	s.UIScale = m.Settings.A11y.UIScale
	s.Locale = m.Settings.Locale
	s.StatsTab = m.Statistics.tab
	s.BuildMenuScroll = m.BuildMenus.scrollRow

	s.Panels[string(PanelInspector)] = PanelLayout{
		X: m.Inspector.posX, Y: m.Inspector.posY,
		Visible: m.Inspector.visible, Persist: true,
	}
	s.Panels[string(PanelStats)] = PanelLayout{
		Visible: m.Statistics.open, Persist: true,
	}
	s.Panels[string(PanelSearch)] = PanelLayout{
		Visible: m.Search.IsOpen(), Persist: m.Search.persist,
	}
	s.Panels[string(PanelAdvisors)] = PanelLayout{
		Visible: m.Advisors.open, Persist: m.Advisors.persist,
	}
	s.Panels[string(PanelOptions)] = PanelLayout{
		Visible: m.Options.open, Persist: true,
	}
	return s
}

// ApplyState restores UI from a saved snapshot (24.25).
func (m *UIManager) ApplyState(state PersistentUIState) {
	if m == nil {
		return
	}
	m.BuildMenus.showFav = state.ShowFavFilter
	m.BuildMenus.showRecent = state.ShowRecentFilter
	m.BuildMenus.favorites = make(map[string]bool, len(state.Favorites))
	for _, id := range state.Favorites {
		m.BuildMenus.favorites[id] = true
	}
	m.BuildMenus.recent = append(m.BuildMenus.recent[:0], state.Recent...)
	m.BuildMenus.scrollRow = state.BuildMenuScroll
	m.BuildMenus.refilter()

	if state.CameraDist > 0 {
		m.Camera.Yaw = state.CameraYaw
		m.Camera.Pitch = state.CameraPitch
		m.Camera.Dist = state.CameraDist
		m.Camera.goalYaw = state.CameraYaw
		m.Camera.goalPitch = state.CameraPitch
		m.Camera.goalDist = state.CameraDist
		m.Camera.Target.X = state.CameraTargetX
		m.Camera.Target.Y = state.CameraTargetY
		m.Camera.Target.Z = state.CameraTargetZ
		m.Camera.goalTarget = m.Camera.Target
	}
	m.InfoViews.Set(InfoViewKind(state.InfoView))
	if state.UIScale > 0 {
		m.Settings.A11y.UIScale = state.UIScale
	}
	if state.Locale != "" {
		m.Settings.Locale = state.Locale
		SetLocale(state.Locale)
	}
	m.Statistics.tab = state.StatsTab

	if lay, ok := state.Panels[string(PanelInspector)]; ok && lay.Persist {
		m.Inspector.posX, m.Inspector.posY = panelXY(lay, PanelInspector)
	}
	if lay, ok := state.Panels[string(PanelStats)]; ok && lay.Visible && lay.Persist {
		m.Statistics.open = true
		m.Toolbar.Selected = CatStatistics
	}
	if lay, ok := state.Panels[string(PanelOptions)]; ok && lay.Visible && lay.Persist {
		m.Options.open = true
		m.Toolbar.Selected = CatOptions
	}
	if lay, ok := state.Panels[string(PanelSearch)]; ok && lay.Visible && lay.Persist {
		m.Search.open = true
	}
	if lay, ok := state.Panels[string(PanelAdvisors)]; ok && lay.Visible && lay.Persist {
		m.Advisors.open = true
	}
}

func (m *UIManager) SavePreferences(path string) error {
	return SaveUIState(path, m.CaptureState())
}

func (m *UIManager) LoadPreferences(path string) error {
	state, err := LoadUIState(path)
	if err != nil {
		return err
	}
	m.ApplyState(state)
	return nil
}
