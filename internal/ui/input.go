package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// InputManager routes keyboard and pointer input to UI subsystems before world tools.
type InputManager struct {
	mgr *UIManager
}

func NewInputManager(mgr *UIManager) *InputManager {
	return &InputManager{mgr: mgr}
}

func (im *InputManager) HandleClick() GameTool {
	if im.mgr.Dialogs.CapturesInput() {
		return im.mgr.Selected
	}
	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)

	if im.mgr.BuildMenus.HandleClick(mx, my, &im.mgr.ToolSystem) {
		return im.mgr.Selected
	}
	if my >= ToolbarY {
		return im.mgr.Toolbar.HandleClick(&im.mgr.ToolSystem, im.mgr.BuildMenus)
	}
	if im.mgr.ToolSystem.HasOptionsBar() && my >= ToolbarY-OptionsBarH {
		return im.mgr.Selected
	}
	return im.mgr.Selected
}

func (im *InputManager) PointerOverChrome() bool {
	mPos := rl.GetMousePosition()
	my := int32(mPos.Y)
	if my >= im.mgr.ChromeTopY() {
		return true
	}
	if my < TopBarH && im.mgr.Overlays.Visible() {
		return true
	}
	return false
}
