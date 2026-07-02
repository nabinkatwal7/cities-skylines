package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// InputManager routes keyboard and pointer input to UI subsystems before world tools.
type InputManager struct {
	mgr *UIManager
}

func NewInputManager(mgr *UIManager) *InputManager {
	return &InputManager{mgr: mgr}
}

func (im *InputManager) HandleKeyboard() GameTool {
	if im.mgr.Dialogs.CapturesInput() {
		return im.mgr.Selected
	}
	return im.mgr.ToolSystem.HandleKeyboard()
}

func (im *InputManager) HandleClick() GameTool {
	if im.mgr.Dialogs.CapturesInput() {
		return im.mgr.Selected
	}
	mPos := rl.GetMousePosition()
	my := int32(mPos.Y)
	if my >= ToolbarY {
		return im.mgr.Toolbar.HandleClick(&im.mgr.ToolSystem)
	}
	return im.mgr.Selected
}

func (im *InputManager) PointerOverChrome() bool {
	mPos := rl.GetMousePosition()
	my := int32(mPos.Y)
	if my >= ToolbarY {
		return true
	}
	if my < TopBarH && im.mgr.Overlays.Visible() {
		return true
	}
	return im.mgr.HasOptionsBar() && my >= ToolbarY-OptionsBarH
}
