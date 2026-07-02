package ui

// UIRefresh batches expensive UI work (24.26).
type UIRefresh struct {
	frame         int
	notifEvery    int
	notifDirty    bool
	statsDirty    bool
	advisorsDirty bool
	simDirty      bool
}

func NewUIRefresh() *UIRefresh {
	return &UIRefresh{notifEvery: 30}
}

func (r *UIRefresh) MarkSimDirty() {
	r.simDirty = true
	r.notifDirty = true
	r.statsDirty = true
	r.advisorsDirty = true
}

func (r *UIRefresh) MarkNotificationsDirty() { r.notifDirty = true }

func (r *UIRefresh) Tick() { r.frame++ }

func (r *UIRefresh) ShouldRefreshNotifications() bool {
	if r.notifDirty || r.frame%r.notifEvery == 0 {
		r.notifDirty = false
		return true
	}
	return false
}

func (r *UIRefresh) ShouldRefreshStats(open bool) bool {
	if !open {
		return false
	}
	if r.statsDirty || r.simDirty {
		r.statsDirty = false
		return true
	}
	return false
}

func (r *UIRefresh) ShouldRefreshAdvisors(open bool) bool {
	if !open {
		return false
	}
	if r.advisorsDirty || r.simDirty {
		r.advisorsDirty = false
		return true
	}
	return false
}

func (r *UIRefresh) ClearSimDirty() { r.simDirty = false }

// visiblePanel gates update/draw to open surfaces (24.26).
func visiblePanel(open bool) bool { return open }
