package ui

// Advisors surfaces contextual city guidance (24.14).
type Advisors struct {
	message string
}

func NewAdvisors() *Advisors { return &Advisors{} }

func (a *Advisors) Draw() {}

func (a *Advisors) SetMessage(msg string) { a.message = msg }
