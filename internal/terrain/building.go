package terrain

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Building struct {
	X, Z         float32
	Type         ZoneType
	Seed         int32
	Width, Depth float32
	Height       float32
}

type BuildingManager struct {
	Buildings []Building
	nextSeed  int32
	models    map[ZoneType]rl.Model
}

func NewBuildingManager() *BuildingManager {
	return &BuildingManager{models: make(map[ZoneType]rl.Model)}
}

func (bm *BuildingManager) LoadAssets() {
	m := rl.LoadModel("assets/building/my_panel16_14.obj")
	if rl.IsModelValid(m) {
		bm.models[ZoneResidentialHigh] = m
	}
	m = rl.LoadModel("assets/building/tiny/house.obj")
	if rl.IsModelValid(m) {
		bm.models[ZoneResidentialLow] = m
	}
}

func (bm *BuildingManager) Update(zm *ZoneManager, h *Heightmap, roads *RoadManager) {
	cellSize := WorldSize / float32(zm.width)
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			cell := &zm.Cells[z][x]
			if cell.Type == ZoneNone {
				continue
			}
			if cell.Density >= 0.5 {
				continue
			}
			cx := float32(x)*cellSize - WorldSize/2 + cellSize*0.5
			cz := float32(z)*cellSize - WorldSize/2 + cellSize*0.5
			if !roads.HasNearbyRoad(cx, cz, cellSize*2) {
				continue
			}
			cell.Density += 0.001
			if cell.Density >= 0.5 {
				w := 5 + float32(bm.nextSeed%3)*1.5
				d := 5 + float32((bm.nextSeed+1)%3)*1.5
				hgt := buildingHeight(cell.Type, bm.nextSeed)
				bm.Buildings = append(bm.Buildings, Building{
					X: cx, Z: cz,
					Type:   cell.Type,
					Seed:   bm.nextSeed,
					Width:  w,
					Depth:  d,
					Height: hgt,
				})
				bm.nextSeed++
			}
		}
	}
}

func buildingHeight(zt ZoneType, seed int32) float32 {
	h := float32(seed % 3)
	switch zt {
	case ZoneResidentialLow:
		return 3 + h*0.8
	case ZoneResidentialHigh:
		return 6 + h*2
	case ZoneCommercialLow:
		return 4 + h
	case ZoneCommercialHigh:
		return 8 + h*2.5
	case ZoneIndustrial:
		return 5 + h*1.5
	case ZoneOffice:
		return 6 + h*2
	default:
		return 3
	}
}

func (bm *BuildingManager) Draw(h *Heightmap, zm *ZoneManager) {
	for _, b := range bm.Buildings {
		hy := h.WorldHeight(b.X, b.Z)
		col := ZoneColor(b.Type)
		col.A = 255

		if m, ok := bm.models[b.Type]; ok && m.MeshCount > 0 {
			s := b.Width * 0.18
			if b.Type == ZoneResidentialLow {
				s = b.Width * 0.12
			}
			pos := rl.NewVector3(b.X, hy+0.5, b.Z)
			axis := rl.NewVector3(0, 1, 0)
			rl.DrawModelEx(m, pos, axis, float32(b.Seed)*60, rl.NewVector3(s, s, s), rl.White)
			continue
		}

		baseH := b.Height * 0.7
		roofH := b.Height * 0.3

		rl.DrawCube(rl.NewVector3(b.X, hy+baseH*0.5, b.Z), b.Width, baseH, b.Depth, col)

		roofCol := rl.NewColor(
			uint8(float32(col.R)*0.7),
			uint8(float32(col.G)*0.7),
			uint8(float32(col.B)*0.7),
			255,
		)
		roofTop := hy + baseH + roofH
		roofBase := hy + baseH

		if b.Type == ZoneIndustrial || b.Type == ZoneCommercialHigh {
			rl.DrawCube(rl.NewVector3(b.X, roofTop, b.Z), b.Width*0.85, roofH, b.Depth*0.85, roofCol)
		} else if b.Type == ZoneResidentialLow {
			hw := b.Width * 0.5
			hd := b.Depth * 0.5
			rl.DrawTriangle3D(rl.NewVector3(b.X-hw, roofBase, b.Z-hd), rl.NewVector3(b.X+hw, roofBase, b.Z-hd), rl.NewVector3(b.X, roofTop, b.Z), roofCol)
			rl.DrawTriangle3D(rl.NewVector3(b.X+hw, roofBase, b.Z-hd), rl.NewVector3(b.X+hw, roofBase, b.Z+hd), rl.NewVector3(b.X, roofTop, b.Z), roofCol)
			rl.DrawTriangle3D(rl.NewVector3(b.X+hw, roofBase, b.Z+hd), rl.NewVector3(b.X-hw, roofBase, b.Z+hd), rl.NewVector3(b.X, roofTop, b.Z), roofCol)
			rl.DrawTriangle3D(rl.NewVector3(b.X-hw, roofBase, b.Z+hd), rl.NewVector3(b.X-hw, roofBase, b.Z-hd), rl.NewVector3(b.X, roofTop, b.Z), roofCol)
		} else {
			roofY := roofBase + roofH*0.5
			rl.DrawCube(rl.NewVector3(b.X, roofY, b.Z), b.Width*0.9, roofH*0.9, b.Depth*0.9, roofCol)
		}

		wCol := rl.NewColor(200, 220, 240, 200)
		ww := b.Width * 0.12
		wd := b.Depth * 0.12
		nx := int(b.Width / 2.5)
		nz := int(b.Depth / 2.5)
		if nx < 1 {
			nx = 1
		}
		if nz < 1 {
			nz = 1
		}
		wSpacingX := b.Width / float32(nx+1)
		wSpacingZ := b.Depth / float32(nz+1)
		winY := hy + baseH*0.5
		for wi := 0; wi < nx; wi++ {
			wx := -b.Width*0.5 + wSpacingX*float32(wi+1)
			for wj := 0; wj < nz; wj++ {
				wz := -b.Depth*0.5 + wSpacingZ*float32(wj+1)
				for layer := 0; layer < int(baseH/2.5); layer++ {
					yoff := float32(layer)*2.5 + 1.5
					if yoff+1 < baseH {
						col2 := wCol
						if (b.Seed+int32(wi+wj+layer))%2 == 0 {
							col2 = rl.NewColor(120, 140, 160, 200)
						}
						rl.DrawCube(rl.NewVector3(b.X+wx, winY+yoff-baseH*0.5+1.25, b.Z+wz), ww, 1.5, wd, col2)
					}
				}
			}
		}
	}
}

func (bm *BuildingManager) Unload() {
	for _, m := range bm.models {
		if m.MeshCount > 0 {
			rl.UnloadModel(m)
		}
	}
}
