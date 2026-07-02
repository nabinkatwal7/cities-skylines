package ui

// PanelID names a dockable UI surface (24.25).
type PanelID string

const (
	PanelInspector PanelID = "inspector"
	PanelStats     PanelID = "stats"
	PanelSearch    PanelID = "search"
	PanelAdvisors  PanelID = "advisors"
	PanelOptions   PanelID = "options"
)

// PanelLayout stores window position and visibility (24.25).
type PanelLayout struct {
	X       int32 `json:"x"`
	Y       int32 `json:"y"`
	Visible bool  `json:"visible"`
	Persist bool  `json:"persist"` // transient windows skip unless persist=true
}

// PersistentUIState is saved UI preferences (24.25).
type PersistentUIState struct {
	Version          int                    `json:"version"`
	Panels           map[string]PanelLayout `json:"panels"`
	ShowFavFilter    bool                   `json:"show_fav_filter"`
	ShowRecentFilter bool                   `json:"show_recent_filter"`
	Favorites        []string               `json:"favorites"`
	Recent           []string               `json:"recent"`
	CameraYaw        float32                `json:"camera_yaw"`
	CameraPitch      float32                `json:"camera_pitch"`
	CameraDist       float32                `json:"camera_dist"`
	CameraTargetX    float32                `json:"camera_tx"`
	CameraTargetY    float32                `json:"camera_ty"`
	CameraTargetZ    float32                `json:"camera_tz"`
	InfoView         int                    `json:"info_view"`
	UIScale          float32                `json:"ui_scale"`
	Locale           string                 `json:"locale"`
	StatsTab         int                    `json:"stats_tab"`
	BuildMenuScroll  int                    `json:"build_scroll"`
}

const uiStateVersion = 1

func DefaultUIState() PersistentUIState {
	return PersistentUIState{
		Version: uiStateVersion,
		Panels: map[string]PanelLayout{
			string(PanelInspector): {X: -1, Y: -1, Visible: false, Persist: true},
			string(PanelStats):     {X: -1, Y: -1, Visible: false, Persist: true},
			string(PanelSearch):    {Visible: false, Persist: false},
			string(PanelAdvisors):  {Visible: false, Persist: false},
			string(PanelOptions):   {Visible: false, Persist: true},
		},
		UIScale: 1,
		Locale:  "en",
	}
}

func defaultPanelPos(id PanelID) (x, y int32) {
	switch id {
	case PanelInspector:
		return ScreenW - 328, TopBarH + 8
	case PanelStats, PanelOptions:
		return (ScreenW - 420) / 2, TopBarH + 8
	case PanelSearch:
		return (ScreenW - 400) / 2, TopBarH + 40
	case PanelAdvisors:
		return 8, TopBarH + 60
	default:
		return 0, TopBarH + 8
	}
}

func panelXY(layout PanelLayout, id PanelID) (x, y int32) {
	x, y = defaultPanelPos(id)
	if layout.X >= 0 {
		x = layout.X
	}
	if layout.Y >= 0 {
		y = layout.Y
	}
	return x, y
}
