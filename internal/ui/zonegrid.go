package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawZonePlacementGrid(ctx SnapContext, camX, camZ float32) {
	sm := ctx.Sim
	if sm == nil || sm.Zones == nil || sm.Heightmap == nil {
		return
	}
	zm := sm.Zones
	cs := zm.CellSize()
	if cs <= 0 {
		return
	}
	const maxDist float32 = 120
	const hoverRadius = 5 // cells around cursor
	maxDistSq := maxDist * maxDist
	hoverCX := zm.CellX(ctx.PreviewX)
	hoverCZ := zm.CellZ(ctx.PreviewZ)

	validCol := rl.NewColor(240, 250, 255, 36)
	invalidCol := rl.NewColor(80, 80, 90, 18)
	hoverCol := rl.NewColor(120, 220, 255, 100)
	lineCol := rl.NewColor(255, 255, 255, 45)

	for cz := hoverCZ - hoverRadius; cz <= hoverCZ+hoverRadius; cz++ {
		for cx := hoverCX - hoverRadius; cx <= hoverCX+hoverRadius; cx++ {
			if cx < 0 || cz < 0 || cx >= zm.Width() || cz >= zm.Height() {
				continue
			}
			wx, wz := zm.CellCenter(cx, cz)
			dx := wx - camX
			dz := wz - camZ
			if dx*dx+dz*dz > maxDistSq {
				continue
			}
			if !zm.HasRoadInfluence(wx, wz) {
				continue
			}
			hy := sm.Heightmap.WorldHeight(wx, wz) + 0.08
			half := cs * 0.5
			x0, z0 := wx-half, wz-half
			x1, z1 := wx+half, wz+half

			if zm.CanZoneCell(cx, cz) {
				rl.DrawCube(rl.NewVector3(wx, hy, wz), cs*0.96, 0.04, cs*0.96, validCol)
			} else {
				rl.DrawCube(rl.NewVector3(wx, hy, wz), cs*0.96, 0.03, cs*0.96, invalidCol)
			}
			if cx == hoverCX && cz == hoverCZ {
				rl.DrawCube(rl.NewVector3(wx, hy+0.05, wz), cs*0.98, 0.06, cs*0.98, hoverCol)
			}
			rl.DrawLine3D(rl.NewVector3(x0, hy, z0), rl.NewVector3(x1, hy, z0), lineCol)
			rl.DrawLine3D(rl.NewVector3(x1, hy, z0), rl.NewVector3(x1, hy, z1), lineCol)
			rl.DrawLine3D(rl.NewVector3(x1, hy, z1), rl.NewVector3(x0, hy, z1), lineCol)
			rl.DrawLine3D(rl.NewVector3(x0, hy, z1), rl.NewVector3(x0, hy, z0), lineCol)
		}
	}
}
