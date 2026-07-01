package terrain

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type HouseholdInfo struct {
	FamilyMembers int32
	Wealth        int32
	Education     int32
	Happiness     int32
}

type BusinessInfo struct {
	Production    int32
	GoodsStored   int32
	Profitability int32
}

type ServiceConsumption struct {
	Power   float32
	Water   float32
	Garbage float32
}

type Building struct {
	Entity
	Type         ZoneType
	Seed         int32
	Width, Depth float32
	Height       float32
	Level        int32
	UpgradeTimer int32
	Workers      int32
	Residents    int32
	AbandonTimer int32
	CellX, CellZ int
	ConstructTimer int32
	Household    *HouseholdInfo
	Business     *BusinessInfo
	Consumption  ServiceConsumption
}

type BuildingManager struct {
	Buildings []Building
	nextSeed  int32
	models    map[ZoneType]rl.Model
	resDemand int
	comDemand int
	indDemand int
	offDemand int
	TotalPowerUsed   float32
	TotalWaterUsed   float32
	TotalGarbage     float32
	TotalWealth      int32
	TotalHappiness   int32
}

func NewBuildingManager() *BuildingManager {
	return &BuildingManager{models: make(map[ZoneType]rl.Model)}
}

func (bm *BuildingManager) LoadAssets() {
}

func (bm *BuildingManager) Update(zm *ZoneManager, h *Heightmap, roads *RoadManager, dm *DistrictManager) {
	bm.calcDemand(zm)
	bm.TotalPowerUsed = 0
	bm.TotalWaterUsed = 0
	bm.TotalGarbage = 0
	bm.TotalWealth = 0
	bm.TotalHappiness = 0
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
			if cell.Density >= 0.5 {
				w := 5 + float32(bm.nextSeed%3)*1.5
				d := 5 + float32((bm.nextSeed+1)%3)*1.5
				hgt := buildingHeight(cell.Type, bm.nextSeed)
				res, workers := calcPopulation(cell.Type, bm.nextSeed)
				bld := Building{
					Entity:    NewEntity(uint32(bm.nextSeed), cx, 0, cz, OwnerBuilding),
					Type:      cell.Type,
					Seed:      bm.nextSeed,
					Width:     w,
					Depth:     d,
					Height:    hgt,
					Level:     1,
					Residents: res,
					Workers:   workers,
					CellX:     x,
					CellZ:     z,
				}
				bld.SetFlag(FlagHasRoad)
				if cell.Type == ZoneResidentialLow || cell.Type == ZoneResidentialHigh {
					bld.Household = &HouseholdInfo{
						FamilyMembers: res,
						Wealth:        30 + bm.nextSeed%30,
						Education:     10 + bm.nextSeed%20,
						Happiness:     50,
					}
					bld.Consumption.Power = 1.0 + float32(bld.Level)*0.5
					bld.Consumption.Water = 0.8 + float32(bld.Level)*0.3
					bld.Consumption.Garbage = 0.5 + float32(bld.Level)*0.2
				}
				if cell.Type == ZoneCommercialLow || cell.Type == ZoneCommercialHigh || cell.Type == ZoneIndustrial || cell.Type == ZoneOffice {
					bld.Business = &BusinessInfo{
						GoodsStored:   10,
						Profitability: 50,
					}
					bld.Consumption.Power = 2.0 + float32(bld.Level)
					bld.Consumption.Water = 1.0 + float32(bld.Level)*0.5
					bld.Consumption.Garbage = 1.0 + float32(bld.Level)*0.5
				}
				bm.Buildings = append(bm.Buildings, bld)
				bm.nextSeed++
			}
		}
	}
	for i := range bm.Buildings {
		b := &bm.Buildings[i]
		if b.HasFlag(FlagAbandoned) {
			b.AbandonTimer++
			if b.AbandonTimer > 1800 {
				if b.CellX >= 0 && b.CellX < zm.width && b.CellZ >= 0 && b.CellZ < zm.height {
					zm.Cells[b.CellZ][b.CellX].Density = 0
				}
				b.SetFlag(FlagRemoved)
				b.Residents = 0
				b.Workers = 0
			}
			continue
		}

		cellSize := WorldSize / float32(zm.width)
		hasRoad := roads.HasNearbyRoad(b.Position.X, b.Position.Z, cellSize*2)
		if hasRoad {
			b.SetFlag(FlagHasRoad)
		} else {
			b.ClearFlag(FlagHasRoad)
		}
		if !hasRoad {
			b.SetFlag(FlagAbandoned)
			b.AbandonTimer = 0
			b.Residents = 0
			b.Workers = 0
			continue
		}

		if !b.HasFlag(FlagConstructed) {
			b.ConstructTimer++
			if b.ConstructTimer > 300 {
				b.SetFlag(FlagConstructed)
			}
			continue
		}

		if b.Household != nil {
			bm.TotalWealth += b.Household.Wealth
			bm.TotalHappiness += b.Household.Happiness
			bm.TotalPowerUsed += b.Consumption.Power
			bm.TotalWaterUsed += b.Consumption.Water
			bm.TotalGarbage += b.Consumption.Garbage
			if b.Household.Happiness < 80 {
				b.Household.Happiness++
			}
			if b.Household.Wealth < 70 {
				b.Household.Wealth++
			}
		}
		if b.Business != nil {
			bm.TotalPowerUsed += b.Consumption.Power
			bm.TotalWaterUsed += b.Consumption.Water
			if b.Business.GoodsStored < 50 {
				b.Business.GoodsStored++
			}
			if b.Business.Profitability < 60 {
				b.Business.Profitability++
			}
		}
		if dm != nil {
			dm.ApplyPolicies(b)
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
	offJobs := int32(0)
	total := 0
	for _, b := range bm.Buildings {
		if b.HasFlag(FlagRemoved) {
			continue
		}
		total++
		switch b.Type {
		case ZoneResidentialLow, ZoneResidentialHigh:
			resPop += b.Residents
		case ZoneCommercialLow, ZoneCommercialHigh:
			comJobs += b.Workers
		case ZoneIndustrial:
			indJobs += b.Workers
		case ZoneOffice:
			offJobs += b.Workers
		}
	}
	availableJobs := comJobs + indJobs + offJobs
	bm.resDemand = int(availableJobs-resPop) / 2
	bm.comDemand = int(resPop/2-comJobs) / 2
	bm.indDemand = int(resPop/3-indJobs) / 3
	bm.offDemand = int(resPop/4-offJobs) / 2
	if total == 0 {
		bm.resDemand = 10
		bm.comDemand = 5
		bm.indDemand = 3
		bm.offDemand = 2
	}
	if bm.resDemand < 1 {
		bm.resDemand = 1
	}
	if bm.comDemand < 1 {
		bm.comDemand = 1
	}
	if bm.indDemand < 1 {
		bm.indDemand = 1
	}
	if bm.offDemand < 1 {
		bm.offDemand = 1
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
		return bm.offDemand > 0
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
	if b.HasFlag(FlagHasRoad) {
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
	hy := h.WorldHeight(b.Position.X, b.Position.Z)
	if hy < 12 {
		val += 15
	}
	return val
}

func (bm *BuildingManager) Demand() (res, com, ind int) {
	return bm.resDemand, bm.comDemand, bm.indDemand
}

func (bm *BuildingManager) Population() int32 {
	total := int32(0)
	for _, b := range bm.Buildings {
		if b.HasFlag(FlagRemoved) {
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
		if b.HasFlag(FlagRemoved) {
			continue
		}
		dx := b.Position.X - wx
		dz := b.Position.Z - wz
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
	levelCount := b.Level
	lvl := ""
	for i := int32(0); i < levelCount; i++ {
		lvl += "I"
	}
	extra := " | "
	if b.HasFlag(FlagAbandoned) {
		extra += "ABANDONED"
	} else if b.Residents > 0 {
		extra += fmt.Sprintf("Pop: %d", b.Residents)
		if b.Household != nil {
			extra += fmt.Sprintf(" W:%.0f H:%d", b.Consumption.Power, b.Household.Happiness)
		}
	} else if b.Workers > 0 {
		extra += fmt.Sprintf("Jobs: %d", b.Workers)
		if b.Business != nil {
			extra += fmt.Sprintf(" P:%d%%", b.Business.Profitability)
		}
	}
	if !b.HasFlag(FlagConstructed) {
		pct := int(float32(b.ConstructTimer) / 300.0 * 100)
		extra += fmt.Sprintf(" Building %d%%", pct)
	}
	return fmt.Sprintf("%s Lvl %s%s", name, lvl, extra)
}

func (bm *BuildingManager) Draw(h *Heightmap, zm *ZoneManager, isNight bool) {
	for _, b := range bm.Buildings {
		if b.HasFlag(FlagRemoved) {
			continue
		}
		hy := h.WorldHeight(b.Position.X, b.Position.Z)
		col := ZoneColor(b.Type)
		col.A = 255

		lvlScale := 1.0 + float32(b.Level-1)*0.15

		if m, ok := bm.models[b.Type]; ok && m.MeshCount > 0 {
			s := b.Width * 0.18 * lvlScale
			if b.Type == ZoneResidentialLow {
				s = b.Width * 0.12 * lvlScale
			}
			axis := rl.NewVector3(0, 1, 0)
			rl.DrawModelEx(m, b.Position, axis, b.Rotation.W*rl.Rad2deg, rl.NewVector3(s, s, s), rl.White)
			continue
		}

		if b.HasFlag(FlagAbandoned) {
			grey := rl.NewColor(80, 80, 80, 255)
			rl.DrawCube(rl.NewVector3(b.Position.X, hy+b.Height*0.5*lvlScale, b.Position.Z), b.Width, b.Height*lvlScale, b.Depth, grey)
			continue
		}

		if !b.HasFlag(FlagConstructed) {
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
				rl.DrawCube(rl.NewVector3(b.Position.X, hy+currH*0.3, b.Position.Z), b.Width*0.5, currH*0.6, b.Depth*0.5, constructCol)
			} else {
				rl.DrawCube(rl.NewVector3(b.Position.X, hy+currH*0.5, b.Position.Z), b.Width, currH, b.Depth, constructCol)
			}
			continue
		}

		baseH := b.Height * 0.7 * lvlScale
		roofH := b.Height * 0.3 * lvlScale

		rl.DrawCube(rl.NewVector3(b.Position.X, hy+baseH*0.5, b.Position.Z), b.Width, baseH, b.Depth, col)

		roofCol := rl.NewColor(
			uint8(float32(col.R)*0.7),
			uint8(float32(col.G)*0.7),
			uint8(float32(col.B)*0.7),
			255,
		)
		roofTop := hy + baseH + roofH
		roofBase := hy + baseH

		if b.Type == ZoneIndustrial || b.Type == ZoneCommercialHigh {
			rl.DrawCube(rl.NewVector3(b.Position.X, roofTop, b.Position.Z), b.Width*0.85, roofH, b.Depth*0.85, roofCol)
		} else if b.Type == ZoneResidentialLow {
			hw := b.Width * 0.5
			hd := b.Depth * 0.5
			rl.DrawTriangle3D(rl.NewVector3(b.Position.X-hw, roofBase, b.Position.Z-hd), rl.NewVector3(b.Position.X+hw, roofBase, b.Position.Z-hd), rl.NewVector3(b.Position.X, roofTop, b.Position.Z), roofCol)
			rl.DrawTriangle3D(rl.NewVector3(b.Position.X+hw, roofBase, b.Position.Z-hd), rl.NewVector3(b.Position.X+hw, roofBase, b.Position.Z+hd), rl.NewVector3(b.Position.X, roofTop, b.Position.Z), roofCol)
			rl.DrawTriangle3D(rl.NewVector3(b.Position.X+hw, roofBase, b.Position.Z+hd), rl.NewVector3(b.Position.X-hw, roofBase, b.Position.Z+hd), rl.NewVector3(b.Position.X, roofTop, b.Position.Z), roofCol)
			rl.DrawTriangle3D(rl.NewVector3(b.Position.X-hw, roofBase, b.Position.Z+hd), rl.NewVector3(b.Position.X-hw, roofBase, b.Position.Z-hd), rl.NewVector3(b.Position.X, roofTop, b.Position.Z), roofCol)
		} else {
			roofY := roofBase + roofH*0.5
			rl.DrawCube(rl.NewVector3(b.Position.X, roofY, b.Position.Z), b.Width*0.9, roofH*0.9, b.Depth*0.9, roofCol)
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
						rl.DrawCube(rl.NewVector3(b.Position.X+wx, winY+yoff-baseH*0.5+1.25, b.Position.Z+wz), ww, 1.5, wd, col2)
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
