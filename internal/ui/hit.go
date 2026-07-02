package ui

// PointerOverUI reports whether a screen point hits chrome (not the 3D world).
func (m *UIManager) PointerOverUI(mx, my int32) bool {
	if my < TopBarH {
		return true
	}
	if m.Inspector.visible && m.inspectorHit(mx, my) {
		return true
	}
	if m.Statistics.open && m.statisticsHit(mx, my) {
		return true
	}
	if m.Options.open && m.optionsHit(mx, my) {
		return true
	}
	if m.Search.IsOpen() && m.searchHit(mx, my) {
		return true
	}
	if m.Advisors.open && m.advisorsHit(mx, my) {
		return true
	}
	if len(m.Notifications.Active()) > 0 && m.notificationsHit(mx, my) {
		return true
	}
	if m.InfoViews.Active() != ViewNone && mx < 220 && my < TopBarH+64 {
		return true
	}
	if my >= m.ChromeTopY() {
		return true
	}
	return false
}

// ClickResetsRoadChain is true when a click cancels an in-progress road draw.
func (m *UIManager) ClickResetsRoadChain(mx, my int32) bool {
	if my >= ToolbarY {
		return true
	}
	if m.BuildMenus.Visible() {
		y := m.BuildMenus.Y()
		if my >= y && my < y+BuildMenuH {
			return true
		}
	}
	return false
}

func (m *UIManager) inspectorHit(mx, my int32) bool {
	w := int32(320)
	x, y := m.Inspector.panelPos()
	h := m.Inspector.panelHeight()
	return mx >= x && mx < x+w && my >= y && my < y+h
}

func (m *UIManager) statisticsHit(mx, my int32) bool {
	w, h := int32(440), int32(340)
	x := (ScreenW - w) / 2
	y := TopBarH + 12
	return mx >= x && mx < x+w && my >= y && my < y+h
}

func (m *UIManager) optionsHit(mx, my int32) bool {
	scale := m.Settings.UIScale()
	w := ScaleSize(440, scale)
	h := ScaleSize(320, scale)
	x := (ScreenW - w) / 2
	y := ScaleY(TopBarH+8, scale)
	return mx >= x && mx < x+w && my >= y && my < y+h
}

func (s *SearchSystem) hit(mx, my int32) bool {
	w, h := int32(440), int32(280)
	x := (ScreenW - w) / 2
	y := TopBarH + 48
	return mx >= x && mx < x+w && my >= y && my < y+h
}

func (m *UIManager) searchHit(mx, my int32) bool {
	return m.Search.hit(mx, my)
}

func (m *UIManager) advisorsHit(mx, my int32) bool {
	w := int32(400)
	h := int32(36 + int32(len(m.Advisors.tips))*40)
	if len(m.Advisors.tips) == 0 {
		h = 80
	}
	x := int32(10)
	y := TopBarH + 64
	return mx >= x && mx < x+w && my >= y && my < y+h
}

func (m *UIManager) notificationsHit(mx, my int32) bool {
	active := m.Notifications.Active()
	if len(active) == 0 {
		return false
	}
	x := ScreenW - 320
	y := TopBarH + 8
	h := int32(len(active)*26 + 12)
	return mx >= x && mx < x+312 && my >= y && my < y+h
}
