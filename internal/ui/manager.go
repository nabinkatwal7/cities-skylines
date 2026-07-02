package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/sim"
)

// UIManager coordinates every interface subsystem. It reads simulation state
// and never writes gameplay data.
type UIManager struct {
	ToolSystem
	HUD           *HUD
	Toolbar       *MainToolbar
	BuildMenus    *BuildMenus
	InfoViews     *InfoViews
	Inspector     *InspectorPanel
	Selection     *SelectionSystem
	Statistics    *StatisticsPanel
	Notifications *Notifications
	Advisors      *Advisors
	Search        *SearchSystem
	Overlays      *OverlayManager
	Dialogs       *DialogManager
	Input         *InputManager
	Camera        *CameraController
	Settings      *PlayerSettings
	TimeControls  *TimeControls
	Gamepad       *GamepadInput
	Options       *OptionsPanel
	Shortcuts     *GlobalShortcuts
	refresh       *UIRefresh
	unlocks       *UnlockRegistry

	Money         float32
	TimeStr       string
	MouseWorldX   float32
	MouseWorldZ   float32
	MouseOnGround bool
	lastView      ViewState
	worldCtx      WorldContext
	preview       PlacementPreview
	lastSim       *sim.SimulationManager

	unsubs []func()
}

func NewManager() *UIManager {
	unlocks := NewUnlockRegistry()
	settings := NewPlayerSettings()
	SetLocale(settings.Locale)
	m := &UIManager{
		ToolSystem:    *NewToolSystem(),
		HUD:           NewHUD(),
		Toolbar:       NewMainToolbar(unlocks),
		BuildMenus:    NewBuildMenus(unlocks),
		InfoViews:     NewInfoViews(),
		Inspector:     NewInspectorPanel(),
		Selection:     NewSelectionSystem(),
		Statistics:    NewStatisticsPanel(),
		Notifications: NewNotifications(),
		Advisors:      NewAdvisors(),
		Search:        NewSearchSystem(),
		Overlays:      NewOverlayManager(),
		Dialogs:       NewDialogManager(),
		Camera:        NewCameraController(),
		Settings:      settings,
		TimeControls:  NewTimeControls(),
		Gamepad:       NewGamepadInput(),
		Options:       NewOptionsPanel(),
		Shortcuts:     NewGlobalShortcuts(),
		refresh:       NewUIRefresh(),
		unlocks:       unlocks,
	}
	m.Input = NewInputManager(m)
	return m
}

func NewGameUI() *UIManager { return NewManager() }

type GameUI = UIManager

func (m *UIManager) Attach(bus *core.EventBus) {
	m.Detach()
	if bus == nil {
		return
	}
	m.unsubs = append(m.unsubs, m.HUD.Subscribe(bus)...)
	m.unsubs = append(m.unsubs, m.Notifications.Subscribe(bus)...)
	m.unsubs = append(m.unsubs,
		bus.On(string(core.EventZonePlaced), func(any) { m.HUD.MarkDirty(); m.refresh.MarkSimDirty() }),
		bus.On(string(core.EventZoneRemoved), func(any) { m.HUD.MarkDirty(); m.refresh.MarkSimDirty() }),
		bus.On(string(core.EventRoadPlaced), func(any) { m.HUD.MarkDirty(); m.refresh.MarkSimDirty() }),
		bus.On(string(core.EventRoadRemoved), func(any) { m.HUD.MarkDirty(); m.refresh.MarkSimDirty() }),
		bus.On(string(core.EventTimeDay), func(any) {
			m.Statistics.RecordDay(m.lastView.Population)
			m.refresh.MarkSimDirty()
		}),
		bus.On(string(core.EventFloodStarted), func(any) { m.Notifications.MarkDirty() }),
		bus.On(string(core.EventFloodReceded), func(any) { m.Notifications.MarkDirty() }),
	)
}

func (m *UIManager) Detach() {
	for _, off := range m.unsubs {
		off()
	}
	m.unsubs = nil
}

func (m *UIManager) SyncView(view ViewState) {
	m.refresh.Tick()
	m.lastView = view
	m.Money = view.Money
	m.TimeStr = view.TimeStr
	m.MouseWorldX = view.MouseWorldX
	m.MouseWorldZ = view.MouseWorldZ
	m.MouseOnGround = view.MouseOnGround
	m.HUD.Sync(view)
	m.unlocks.SyncPopulation(view.Population)
	m.Statistics.open = m.Toolbar.Selected == CatStatistics
	m.Options.open = m.Toolbar.Selected == CatOptions

	if m.Statistics.open {
		m.Statistics.Sync(view)
	}
	if m.lastSim != nil {
		if m.refresh.ShouldRefreshNotifications() {
			m.Notifications.Refresh(m.lastSim, view, true)
		}
		if m.refresh.ShouldRefreshAdvisors(m.Advisors.open) {
			m.Advisors.Refresh(m.lastSim, view)
		}
		if m.refresh.ShouldRefreshStats(m.Statistics.open) {
			m.Statistics.SyncSim(m.lastSim, view)
		}
		if m.Search.IsOpen() {
			m.Search.Update(m.lastSim)
		}
	}
	m.refresh.ClearSimDirty()
}

func (m *UIManager) UpdateWorld(ctx WorldContext) {
	ctx.Tools = &m.ToolSystem
	m.worldCtx = ctx
	m.lastSim = ctx.Sim
	m.preview = EvalPlacement(ctx)
	if ctx.Sim != nil {
		m.refresh.MarkNotificationsDirty()
	}
}

func (m *UIManager) UpdateCamera(dt float32) rl.Camera3D {
	if fx, fz, ok := m.FollowTarget(); ok {
		m.Camera.SetFollowTarget(fx, fz)
		m.Search.ClearCamera()
	} else {
		m.Camera.ClearFollow()
	}
	m.Gamepad.Update(m.Camera, m.Settings.A11y)
	m.Camera.Update(dt, m.Settings.Bindings, m.Settings.A11y, m.Camera.Position())
	return m.Camera.Camera3D()
}

func (m *UIManager) Preview() PlacementPreview { return m.preview }

func (m *UIManager) DrawWorldPreview() {
	if m.worldCtx.OnGround {
		DrawPreview3D(m.worldCtx, m.preview)
	}
}

func (m *UIManager) DrawWorldOverlays(sm *sim.SimulationManager) {
	m.Overlays.DrawWorld(sm, m.InfoViews, m.Selection)
}

func (m *UIManager) SetRoadChain(active bool, startNode uint32) {
	m.worldCtx.RoadActive = active
	m.worldCtx.RoadStartNode = startNode
}

func (m *UIManager) SnapPlacementCoords(sm *sim.SimulationManager, x, z float32) (float32, float32) {
	return SnapPlacement(m.snapContext(sm, x, z), x, z)
}

func (m *UIManager) snapContext(sm *sim.SimulationManager, previewX, previewZ float32) SnapContext {
	return SnapContext{
		Sim:           sm,
		Tool:          m.Selected,
		Mode:          m.Mode,
		RoadActive:    m.worldCtx.RoadActive,
		RoadStartNode: m.worldCtx.RoadStartNode,
		PreviewX:      previewX,
		PreviewZ:      previewZ,
	}
}

func (m *UIManager) DrawBuildGuides(camX, camZ float32) {
	if !m.worldCtx.OnGround {
		return
	}
	DrawBuildGrid(m.snapContext(m.lastSim, m.worldCtx.PreviewX, m.worldCtx.PreviewZ), camX, camZ)
}

func (m *UIManager) HandleWorldClick(sim *sim.SimulationManager, x, z float32) bool {
	switch m.Mode {
	case ModeInspect:
		if m.Selected == ToolPointer {
			m.Selection.Pick(sim, x, z, m.Inspector)
		}
		return m.Selected == ToolInspect || m.Selected == ToolPointer
	case ModeMeasure:
		m.MeasureClick(x, z)
		return true
	default:
		return false
	}
}

func (m *UIManager) HandleInput() GameTool {
	if m.Dialogs.CapturesInput() {
		return m.Selected
	}
	m.Shortcuts.Handle(m)
	m.TimeControls.HandleInput(m.lastSim, m.Settings.Bindings)
	m.Advisors.HandleInput()
	m.Search.HandleInput()
	if m.Search.IsOpen() {
		if m.lastSim != nil {
			m.Search.Update(m.lastSim)
		}
		return m.Selected
	}
	m.BuildMenus.HandleInput()
	m.Toolbar.HandleKeyboard(&m.ToolSystem, m.BuildMenus)
	return m.ToolSystem.HandleKeyboard()
}

func (m *UIManager) HandleInspectorClick(mx, my int32, sim *sim.SimulationManager) bool {
	return m.Inspector.HandleClick(mx, my, sim, &m.ToolSystem)
}

func (m *UIManager) FollowTarget() (x, z float32, ok bool) {
	if fx, fz, ok := m.Search.CameraTarget(); ok {
		return fx, fz, true
	}
	if m.Inspector.Following() {
		return m.Inspector.Selection().FollowTarget()
	}
	return 0, 0, false
}

func (m *UIManager) HandleClick() GameTool {
	if m.Dialogs.CapturesInput() {
		return m.Selected
	}
	mPos := rl.GetMousePosition()
	mx, my := int32(mPos.X), int32(mPos.Y)
	if m.Options.HandleClick(mx, my, m.Settings) {
		return m.Selected
	}
	if m.Notifications.HandleClick(mx, my) {
		return m.Selected
	}
	if m.Statistics.HandleClick(mx, my) {
		return m.Selected
	}
	if m.Inspector.HandleClick(mx, my, m.worldCtx.Sim, &m.ToolSystem) {
		m.Selection.FromInspector(m.Inspector.Selection())
		return m.Selected
	}
	return m.Input.HandleClick()
}

func (m *UIManager) HasOptionsBar() bool {
	return m.ToolSystem.HasOptionsBar() || m.BuildMenus.Visible()
}

func (m *UIManager) ChromeTopY() int32 {
	return m.Toolbar.ChromeTopY(&m.ToolSystem, m.BuildMenus)
}

func (m *UIManager) Unlocks() *UnlockRegistry { return m.unlocks }

func (m *UIManager) Draw() {
	m.HUD.Draw(m.Notifications, m.Settings)
	if len(m.Notifications.Active()) > 0 {
		m.Notifications.Draw()
	}
	if m.Statistics.open {
		m.Statistics.Draw()
	}
	if m.Options.open {
		m.Options.Draw(m.Settings)
	}
	if m.Search.IsOpen() {
		m.Search.Draw()
	}
	m.Overlays.Draw()
	if m.BuildMenus.Visible() {
		m.BuildMenus.Draw(&m.ToolSystem)
	}
	m.Toolbar.DrawOptions(&m.ToolSystem)
	if m.InfoViews.Active() != ViewNone {
		m.InfoViews.Draw()
	}
	if m.Inspector.Visible() {
		m.Inspector.Draw()
	}
	if m.Advisors.open {
		m.Advisors.Draw()
	}
	m.Toolbar.Draw(&m.ToolSystem)
	m.ToolSystem.DrawHelp()
	m.Gamepad.DrawHint(TopBarH+32, m.Settings.A11y)
	m.Shortcuts.Draw(m.Settings)
	if m.preview.Cost > 0 || len(m.preview.Messages) > 0 || m.preview.Elevation != 0 {
		DrawPreviewHUD(m.preview, ToolbarY-int32(BuildMenuH)-int32(OptionsBarH)-50)
	}
	m.Dialogs.Draw()
}
