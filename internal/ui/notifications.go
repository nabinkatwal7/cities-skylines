package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Notifications shows batched simulation alerts (24.13).
type Notifications struct {
	queue []string
}

func NewNotifications() *Notifications { return &Notifications{} }

func (n *Notifications) Push(msg string) {
	n.queue = append(n.queue, msg)
	if len(n.queue) > 5 {
		n.queue = n.queue[len(n.queue)-5:]
	}
}

func (n *Notifications) Draw() {
	if len(n.queue) == 0 {
		return
	}
	y := int32(TopBarH + 4)
	for i, msg := range n.queue {
		DrawUIText(msg, ScreenW-320, y+int32(i*18), 14, rl.NewColor(255, 220, 120, 220))
	}
}
