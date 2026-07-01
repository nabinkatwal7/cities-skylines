package terrain

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Park struct {
	X, Z float32
	Size int32
}

type ServiceManager struct {
	Parks []Park
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{}
}

func (sm *ServiceManager) AddPark(wx, wz float32) {
	sm.Parks = append(sm.Parks, Park{X: wx, Z: wz, Size: 3})
}

func (sm *ServiceManager) Draw(h *Heightmap) {
	col := rl.NewColor(80, 200, 80, 160)
	for _, p := range sm.Parks {
		hy := h.WorldHeight(p.X, p.Z) + 0.3
		r := float32(p.Size)
		rl.DrawCube(rl.NewVector3(p.X, hy, p.Z), r, 0.2, r, col)

		treeCol := rl.NewColor(40, 140, 40, 200)
		for i := int32(0); i < p.Size*2; i++ {
			tx := p.X + float32(i%p.Size)*1.5 - float32(p.Size)*0.75
			tz := p.Z + float32(i/p.Size)*1.5 - float32(p.Size)*0.75
			th := h.WorldHeight(tx, tz)
			rl.DrawCube(rl.NewVector3(tx, th+0.5, tz), 0.3, 1.0, 0.3, treeCol)
			rl.DrawCube(rl.NewVector3(tx, th+0.2, tz), 0.5, 0.3, 0.5, rl.NewColor(100, 70, 40, 200))
		}
	}
}
