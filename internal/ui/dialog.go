package ui

// DialogManager owns modal windows that capture input (24.20+).
type DialogManager struct {
	open bool
}

func NewDialogManager() *DialogManager { return &DialogManager{} }

func (d *DialogManager) Draw() {}

func (d *DialogManager) CapturesInput() bool { return d.open }

func (d *DialogManager) Open() { d.open = true }

func (d *DialogManager) Close() { d.open = false }
