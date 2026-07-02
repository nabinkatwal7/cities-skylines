package ui

// InfoViews toggles analytical map overlays (24.11).
type InfoViews struct {
	active string
}

func NewInfoViews() *InfoViews { return &InfoViews{} }

func (v *InfoViews) Draw() {}

func (v *InfoViews) Active() string { return v.active }
