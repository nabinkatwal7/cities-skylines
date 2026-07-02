package ui

// BuildMenus lists placeable assets per build category (24.4).
type BuildMenus struct{}

func NewBuildMenus() *BuildMenus { return &BuildMenus{} }

func (b *BuildMenus) Draw() {}
