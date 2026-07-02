package ui

// InspectorPanel shows details for the selected entity (24.8).
type InspectorPanel struct {
	visible bool
}

func NewInspectorPanel() *InspectorPanel { return &InspectorPanel{} }

func (p *InspectorPanel) Draw() {}

func (p *InspectorPanel) Visible() bool { return p.visible }
