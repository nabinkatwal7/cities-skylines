package ui

import (
	"github.com/katwate/js-skylines/internal/core"
)

// UIManager coordinates every interface subsystem. It reads simulation state
// and never writes gameplay data.
//
//	├── HUD
//	├── Main Toolbar
//	├── Build Menus
//	├── Information Views
//	├── Inspector Panels
//	├── Statistics
//	├── Notifications
//	├── Advisors
//	├── Tool System
//	├── Overlay Manager
//	├── Dialog Manager
//	└── Input Manager
type UIManager struct {
	ToolSystem
	HUD           *HUD
	Toolbar       *MainToolbar
	BuildMenus    *BuildMenus
	InfoViews     *InfoViews
	Inspector     *InspectorPanel
	Statistics    *StatisticsPanel
	Notifications *Notifications
	Advisors      *Advisors
	Overlays      *OverlayManager
	Dialogs       *DialogManager
	Input         *InputManager

	// Compatibility fields mirrored from ViewState for game loop code.
	Money         float32
	TimeStr       string
	MouseWorldX   float32
	MouseWorldZ   float32
	MouseOnGround bool

	unsubs []func()
}

func NewManager() *UIManager {
	m := &UIManager{
		ToolSystem:    *NewToolSystem(),
		HUD:           NewHUD(),
		Toolbar:       NewMainToolbar(),
		BuildMenus:    NewBuildMenus(),
		InfoViews:     NewInfoViews(),
		Inspector:     NewInspectorPanel(),
		Statistics:    NewStatisticsPanel(),
		Notifications: NewNotifications(),
		Advisors:      NewAdvisors(),
		Overlays:      NewOverlayManager(),
		Dialogs:       NewDialogManager(),
	}
	m.Input = NewInputManager(m)
	return m
}

// NewGameUI returns a UIManager. Kept for callers that still use the old name.
func NewGameUI() *UIManager { return NewManager() }

// GameUI is the legacy alias for UIManager.
type GameUI = UIManager

func (m *UIManager) Attach(bus *core.EventBus) {
	m.Detach()
	if bus == nil {
		return
	}
	m.unsubs = append(m.unsubs, m.HUD.Subscribe(bus)...)
	m.unsubs = append(m.unsubs,
		bus.On(string(core.EventZonePlaced), func(any) { m.HUD.MarkDirty() }),
		bus.On(string(core.EventZoneRemoved), func(any) { m.HUD.MarkDirty() }),
		bus.On(string(core.EventRoadPlaced), func(any) { m.HUD.MarkDirty() }),
		bus.On(string(core.EventRoadRemoved), func(any) { m.HUD.MarkDirty() }),
		bus.On(string(core.EventFloodStarted), func(data any) {
			m.HUD.MarkDirty()
			if msg, ok := data.(string); ok {
				m.Notifications.Push(msg)
			} else {
				m.Notifications.Push("Flood warning")
			}
		}),
		bus.On(string(core.EventFloodReceded), func(any) {
			m.HUD.MarkDirty()
			m.Notifications.Push("Flood receded")
		}),
	)
}

func (m *UIManager) Detach() {
	for _, off := range m.unsubs {
		off()
	}
	m.unsubs = nil
}

// SyncView updates presentation-only state from the simulation layer.
func (m *UIManager) SyncView(view ViewState) {
	m.Money = view.Money
	m.TimeStr = view.TimeStr
	m.MouseWorldX = view.MouseWorldX
	m.MouseWorldZ = view.MouseWorldZ
	m.MouseOnGround = view.MouseOnGround
	m.HUD.Sync(view)
}

func (m *UIManager) HandleInput() GameTool {
	return m.Input.HandleKeyboard()
}

func (m *UIManager) HandleClick() GameTool {
	return m.Input.HandleClick()
}

func (m *UIManager) HasOptionsBar() bool {
	return m.ToolSystem.HasOptionsBar()
}

func (m *UIManager) Draw() {
	m.HUD.Draw()
	m.Overlays.Draw()
	m.BuildMenus.Draw()
	m.InfoViews.Draw()
	m.Inspector.Draw()
	m.Statistics.Draw()
	m.Advisors.Draw()
	m.Notifications.Draw()
	m.Toolbar.Draw(&m.ToolSystem)
	m.ToolSystem.DrawHelp()
	m.Dialogs.Draw()
}
