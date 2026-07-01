package terrain

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Building struct {
	X, Z           float32
	Type           ZoneType
	Seed           int32
	Width, Depth   float32
	Height         float32
	Level          int32
	UpgradeTimer   int32
	Workers        int32
	Residents      int32
	Abandoned      bool
	AbandonTimer   int32
	HasRoad        bool
	CellX, CellZ   int
	ConstructTimer int32
	Constructed    bool
}

type BuildingManager struct {
	Buildings []Building
	nextSeed  int32
	models    map[ZoneType]rl.Model
	resDemand int
	comDemand int
	indDemand int
}

func NewBuildingManager() *BuildingManager {
	return &BuildingManager{models: make(map[ZoneType]rl.Model)}
}

func (bm *BuildingManager) LoadAssets() {
}

func (bm *BuildingManager) Update(zm *ZoneManager, h *Heightmap, roads *RoadManager) {
	bm.calcDemand(zm)
	cellSize := WorldSize / float32(zm.width)
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			cell := &zm.Cells[z][x]
			if cell.Type == ZoneNone {
				continue
			}
			cx := float32(x)*cellSize - WorldSize/2 + cellSize*0.5
			cz := float32(z)*cellSize - WorldSize/2 + cellSize*0.5
			canDevelop := roads.HasNearbyRoad(cx, cz, cellSize*2)
			if !canDevelop {
				continue
			}
			if cell.Density >= 0.5 {
				continue
			}
			if bm.shouldDevelop(cell) {
				cell.Density += 0.005
			}
			if 			cell.Density >= 0.5 {
				w := 5 + float32(bm.nextSeed%3)*1.5
				d := 5 + float32((bm.nextSeed+1)%3)*1.5
				hgt := buildingHeight(cell.Type, bm.nextSeed)
				res, workers := calcPopulation(cell.Type, bm.nextSeed)
				bm.Buildings = append(bm.Buildings, Building{
					X: cx, Z: cz,
					Type:      cell.Type,
					Seed:      bm.nextSeed,
					Width:     w,
					Depth:     d,
					Height:    hgt,
					Level:     1,
					Residents: res,
					Workers:   workers,
					HasRoad:   true,
					CellX:     x,
					CellZ:     z,
					Constructed: false,
				})
				bm.nextSeed++
			}
		}
	}
	for i := range bm.Buildings {
		b := &bm.Buildings[i]
		if b.Abandoned {
			b.AbandonTimer++
			if b.AbandonTimer > 1800 {
				if b.CellX >= 0 && b.CellX < zm.width && b.CellZ >= 0 && b.CellZ < zm.height {
					zm.Cells[b.CellZ][b.CellX].Density = 0
				}
				b.Z = -99999
				b.X = -99999
				b.Residents = 0
				b.Workers = 0
			}
			continue
		}

		cellSize := WorldSize / float32(zm.width)
		hasRoad := roads.HasNearbyRoad(b.X, b.Z, cellSize*2)
		b.HasRoad = hasRoad
		if !hasRoad {
			b.Abandoned = true
			b.AbandonTimer = 0
			b.Residents = 0
			b.Workers = 0
			continue
		}

		if !b.Constructed {
			b.ConstructTimer++
			if b.ConstructTimer > 300 {
				b.Constructed = true
			}
			continue
		}

		if b.Level >= 5 {
			continue
		}
		b.UpgradeTimer++
		needed := int32(600-b.Level*100) + b.Seed%60
		if b.UpgradeTimer > needed {
			lv := landValue(b, h)
			if lv > b.Level*10 {
				b.Level++
				b.Height = buildingHeight(b.Type, b.Seed+b.Level*10)
				res, workers := calcPopulation(b.Type, b.Seed+b.Level*10)
				b.Residents = res
				b.Workers = workers
			}
		}
	}
}

func calcPopulation(zt ZoneType, seed int32) (res int32, workers int32) {
	switch zt {
	case ZoneResidentialLow:
		return 2 + seed%4, 0
	case ZoneResidentialHigh:
		return 6 + seed%8, 0
	case ZoneCommercialLow:
		return 0, 2 + seed%3
	case ZoneCommercialHigh:
		return 0, 6 + seed%6
	case ZoneIndustrial:
		return 0, 4 + seed%5
	case ZoneOffice:
		return 0, 5 + seed%5
	}
	return 0, 0
}

func (bm *BuildingManager) calcDemand(zm *ZoneManager) {
	resPop := int32(0)
	comJobs := int32(0)
	indJobs := int32(0)
	total := 0
	for _, b := range bm.Buildings {
		total++
		switch b.Type {
		case ZoneResidentialLow, ZoneResidentialHigh:
			resPop += b.Residents
		case ZoneCommercialLow, ZoneCommercialHigh:
			comJobs += b.Workers
		case ZoneIndustrial:
			indJobs += b.Workers
		case ZoneOffice:
		}
	}
	availableJobs := comJobs + indJobs
	bm.resDemand = int(availableJobs-resPop) / 2
	bm.comDemand = int(resPop/2-comJobs) / 2
	bm.indDemand = int(resPop/3-indJobs) / 3
	if total == 0 {
		bm.resDemand = 10
		bm.comDemand = 5
		bm.indDemand = 3
	}
}

func (bm *BuildingManager) shouldDevelop(cell *ZoneCell) bool {
	switch cell.Type {
	case ZoneResidentialLow, ZoneResidentialHigh:
		return bm.resDemand > 0
	case ZoneCommercialLow, ZoneCommercialHigh:
		return bm.comDemand > 0
	case ZoneIndustrial:
		return bm.indDemand > 0
	case ZoneOffice:
		return true
	}
	return false
}

func buildingHeight(zt ZoneType, seed int32) float32 {
	base := float32(3)
	switch zt {
	case ZoneResidentialHigh:
		base = 6
	case ZoneCommercialLow:
		base = 4
	case ZoneCommercialHigh:
		base = 8
	case ZoneIndustrial:
		base = 5
	case ZoneOffice:
		base = 6
	}
	return base + float32(seed%3)*0.8
}

func landValue(b *Building, h *Heightmap) int32 {
	val := int32(30)
	if roadsNearby(b.X, b.Z, 30, h) {
		val += 20
	}
	switch b.Type {
	case ZoneResidentialLow, ZoneResidentialHigh:
		val += 10
	case ZoneIndustrial:
		val -= 10
	case ZoneOffice:
		val += 15
	}
	hy := h.WorldHeight(b.X, b.Z)
	if hy < 12 {
		val += 15
	}
	return val
}

func roadsNearby(x, z float32, radius float32, h *Heightmap) bool {
	return true
}

func (bm *BuildingManager) Demand() (res, com, ind int) {
	return bm.resDemand, bm.comDemand, bm.indDemand
}

func (bm *BuildingManager) Population() int32 {
	total := int32(0)
	for _, b := range bm.Buildings {
		if b.X < -99998 {
			continue
		}
		total += b.Residents
	}
	return total
}

func (bm *BuildingManager) NearestInfo(wx, wz float32, radius float32) string {
	best := float32(radius * radius)
	idx := -1
	for i, b := range bm.Buildings {
		if b.X < -99998 {
			continue
		}
		dx := b.X - wx
		dz := b.Z - wz
		d := dx*dx + dz*dz
		if d < best {
			best = d
			idx = i
		}
	}
	if idx < 0 {
		return ""
	}
	b := bm.Buildings[idx]
	name := ZoneTypeName(b.Type)
	lvl := "I"
	for b.Level > 1 {
		lvl += "I"
	}
	extra := ""
	if b.Abandoned {
		extra = " | ABANDONED"
	} else if b.Residents > 0 {
		extra = fmt.Sprintf(" | Pop: %d", b.Residents)
	} else if b.Workers > 0 {
		extra = fmt.Sprintf(" | Jobs: %d", b.Workers)
	}
	return fmt.Sprintf("%s Lvl %s%s", name, lvl, extra)
}

func (bm *BuildingManager) Draw(h *Heightmap, zm *ZoneManager, isNight bool) {
	for _, b := range bm.Buildings {
		if b.X < -99998 {
			continue
		}
		hy := h.WorldHeight(b.X, b.Z)
		col := ZoneColor(b.Type)
		col.A = 255

		lvlScale := 1.0 + float32(b.Level-1)*0.15

		if m, ok := bm.models[b.Type]; ok && m.MeshCount > 0 {
			s := b.Width * 0.18 * lvlScale
			if b.Type == ZoneResidentialLow {
				s = b.Width * 0.12 * lvlScale
			}
			pos := rl.NewVector3(b.X, hy+0.5, b.Z)
			axis := rl.NewVector3(0, 1, 0)
			rl.DrawModelEx(m, pos, axis, float32(b.Seed)*60, rl.NewVector3(s, s, s), rl.White)
			continue
		}

		if b.Abandoned {
			grey := rl.NewColor(80, 80, 80, 255)
			rl.DrawCube(rl.NewVector3(b.X, hy+b.Height*0.5*lvlScale, b.Z), b.Width, b.Height*lvlScale, b.Depth, grey)
			continue
		}

		if !b.Constructed {
			progress := float32(b.ConstructTimer) / 300.0
			if progress > 1 {
				progress = 1
			}
			currH := b.Height * progress * lvlScale
			constructCol := rl.NewColor(
				uint8(float32(col.R)*0.6),
				uint8(float32(col.G)*0.6),
				uint8(float32(col.B)*0.6),
				255,
			)
			if progress < 0.3 {
				rl.DrawCube(rl.NewVector3(b.X, hy+currH*0.3, b.Z), b.Width*0.5, currH*0.6, b.Depth*0.5, constructCol)
			} else {
				rl.DrawCube(rl.NewVector3(b.X, hy+currH*0.5, b.Z), b.Width, currH, b.Depth, constructCol)
			}
			continue
		}

		baseH := b.Height * 0.7 * lvlScale
		roofH := b.Height * 0.3 * lvlScale

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

		isLit := isNight
		wCol := rl.NewColor(200, 220, 240, 200)
		if isLit {
			wCol = rl.NewColor(255, 220, 120, 220)
		}
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
						if isLit && (b.Seed+int32(wi+wj+layer))%3 != 0 {
							col2 = rl.NewColor(255, 200, 80, 200)
						} else if !isLit && (b.Seed+int32(wi+wj+layer))%2 == 0 {
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
