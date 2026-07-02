package ui

// OverlayManager coordinates world-space and screen overlays (24.12).
type OverlayManager struct {
	visible bool
}

func NewOverlayManager() *OverlayManager { return &OverlayManager{} }

func (o *OverlayManager) Draw() {}

func (o *OverlayManager) Visible() bool { return o.visible }

func (o *OverlayManager) SetVisible(v bool) { o.visible = v }
