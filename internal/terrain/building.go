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
	Type          ZoneType
	Seed          int32
	Width, Depth  float32
	Height        float32
	Level         int32
	UpgradeTimer  int32
	Workers       int32
	Residents     int32
	AbandonTimer  int32
	CellX, CellZ  int
	ConstructTimer int32
	Household     *HouseholdInfo
	Business      *BusinessInfo
	Consumption   ServiceConsumption
}

type BuildingStats struct {
	TotalPowerUsed   float32
	TotalWaterUsed   float32
	TotalGarbage     float32
	TotalWealth      int32
	TotalHappiness   int32
}

type BuildingCmdType uint8

const (
	CmdBuildCreate  BuildingCmdType = iota
	CmdBuildDestroy
	CmdBuildUpgrade
	CmdBuildDowngrade
	CmdBuildFireDamage
	CmdBuildCollapse
)

type BuildingCommand struct {
	Type     BuildingCmdType
	Slot     int32
	ZoneCellX int
	ZoneCellZ int
	ZoneType ZoneType
}

const BuildingPoolSize = 20000

type BuildingManager struct {
	Pool     [BuildingPoolSize]Building
	FreeList []int32
	Count    int32

	CmdQueue []BuildingCommand

	CellBuildings map[int][]int32

	Stats BuildingStats

	nextSeed  int32
	models    map[ZoneType]rl.Model
	resDemand int
	comDemand int
	indDemand int
	offDemand int
}

func NewBuildingManager() *BuildingManager {
	bm := &BuildingManager{
		models:   make(map[ZoneType]rl.Model),
		FreeList: make([]int32, BuildingPoolSize),
	}
	for i := 0; i < BuildingPoolSize; i++ {
		bm.Pool[i].Lifecycle = LifecycleUnallocated
		bm.FreeList[i] = int32(BuildingPoolSize - 1 - i)
	}
	return bm
}

func (bm *BuildingManager) Alloc() int32 {
	if len(bm.FreeList) == 0 {
		return -1
	}
	idx := bm.FreeList[len(bm.FreeList)-1]
	bm.FreeList = bm.FreeList[:len(bm.FreeList)-1]
	b := &bm.Pool[idx]
	b.Lifecycle = LifecycleInitializing
	bm.Count++
	return idx
}

func (bm *BuildingManager) Free(slot int32) {
	if slot < 0 || int(slot) >= BuildingPoolSize {
		return
	}
	bm.Pool[slot] = Building{}
	bm.Pool[slot].Lifecycle = LifecycleReturnedToPool
	bm.FreeList = append(bm.FreeList, slot)
	bm.Count--
}

func (bm *BuildingManager) ForEach(fn func(*Building, int32)) {
	for i := 0; i < BuildingPoolSize; i++ {
		if bm.Pool[i].Lifecycle == LifecycleActive {
			fn(&bm.Pool[i], int32(i))
		}
	}
}

func (bm *BuildingManager) PushCmd(cmd BuildingCommand) {
	bm.CmdQueue = append(bm.CmdQueue, cmd)
}

func (bm *BuildingManager) processCommands(zm *ZoneManager) {
	for _, cmd := range bm.CmdQueue {
		switch cmd.Type {
		case CmdBuildCreate:
			bm.createBuilding(zm, cmd.ZoneCellX, cmd.ZoneCellZ, cmd.ZoneType)
		case CmdBuildDestroy:
			if cmd.Slot >= 0 && int(cmd.Slot) < BuildingPoolSize {
				b := &bm.Pool[cmd.Slot]
				if zm != nil && b.CellX >= 0 && b.CellX < zm.width && b.CellZ >= 0 && b.CellZ < zm.height {
					zm.Cells[b.CellZ][b.CellX].Density = 0
				}
				bm.Free(cmd.Slot)
			}
		case CmdBuildUpgrade:
			bm.upgradeBuilding(cmd.Slot)
		case CmdBuildDowngrade:
			bm.downgradeBuilding(cmd.Slot)
		case CmdBuildFireDamage:
			bm.fireDamageBuilding(cmd.Slot)
		case CmdBuildCollapse:
			bm.collapseBuilding(cmd.Slot, zm)
		}
	}
	bm.CmdQueue = bm.CmdQueue[:0]
	bm.rebuildSpatialIndex(zm)
}

func (bm *BuildingManager) rebuildSpatialIndex(zm *ZoneManager) {
	if zm == nil {
		return
	}
	bm.CellBuildings = make(map[int][]int32)
	bm.ForEach(func(b *Building, slot int32) {
		key := b.CellZ*zm.width + b.CellX
		bm.CellBuildings[key] = append(bm.CellBuildings[key], slot)
	})
}

func (bm *BuildingManager) createBuilding(zm *ZoneManager, cellX, cellZ int, zt ZoneType) {
	slot := bm.Alloc()
	if slot < 0 {
		return
	}
	cellSize := WorldSize / float32(zm.width)
	cx := float32(cellX)*cellSize - WorldSize/2 + cellSize*0.5
	cz := float32(cellZ)*cellSize - WorldSize/2 + cellSize*0.5

	w := 5 + float32(bm.nextSeed%3)*1.5
	d := 5 + float32((bm.nextSeed+1)%3)*1.5
	hgt := buildingHeight(zt, bm.nextSeed)
	res, workers := calcPopulation(zt, bm.nextSeed)

	b := &bm.Pool[slot]
	b.Entity = NewEntity(uint32(bm.nextSeed), cx, 0, cz, OwnerBuilding)
	b.Type = zt
	b.Seed = bm.nextSeed
	b.Width = w
	b.Depth = d
	b.Height = hgt
	b.Level = 1
	b.Residents = res
	b.Workers = workers
	b.CellX = cellX
	b.CellZ = cellZ
	b.SetFlag(FlagHasRoad)

	if zt == ZoneResidentialLow || zt == ZoneResidentialHigh {
		b.Household = &HouseholdInfo{
			FamilyMembers: res,
			Wealth:        30 + bm.nextSeed%30,
			Education:     10 + bm.nextSeed%20,
			Happiness:     50,
		}
		b.Consumption.Power = 1.0 + float32(b.Level)*0.5
		b.Consumption.Water = 0.8 + float32(b.Level)*0.3
		b.Consumption.Garbage = 0.5 + float32(b.Level)*0.2
	}
	if zt == ZoneCommercialLow || zt == ZoneCommercialHigh || zt == ZoneIndustrial || zt == ZoneOffice {
		b.Business = &BusinessInfo{
			GoodsStored:   10,
			Profitability: 50,
		}
		b.Consumption.Power = 2.0 + float32(b.Level)
		b.Consumption.Water = 1.0 + float32(b.Level)*0.5
		b.Consumption.Garbage = 1.0 + float32(b.Level)*0.5
	}
	bm.nextSeed++
}

func (bm *BuildingManager) upgradeBuilding(slot int32) {
	if slot < 0 || int(slot) >= BuildingPoolSize {
		return
	}
	b := &bm.Pool[slot]
	if b.Level >= 5 {
		return
	}
	b.Level++
	b.Height = buildingHeight(b.Type, b.Seed+b.Level*10)
	res, workers := calcPopulation(b.Type, b.Seed+b.Level*10)
	b.Residents = res
	b.Workers = workers
}

func (bm *BuildingManager) downgradeBuilding(slot int32) {
	if slot < 0 || int(slot) >= BuildingPoolSize {
		return
	}
	b := &bm.Pool[slot]
	if b.Level <= 1 {
		return
	}
	b.Level--
	b.Height = buildingHeight(b.Type, b.Seed+b.Level*10)
	res, workers := calcPopulation(b.Type, b.Seed+b.Level*10)
	b.Residents = res
	b.Workers = workers
}

func (bm *BuildingManager) fireDamageBuilding(slot int32) {
}

func (bm *BuildingManager) collapseBuilding(slot int32, zm *ZoneManager) {
	if slot < 0 || int(slot) >= BuildingPoolSize {
		return
	}
	b := &bm.Pool[slot]
	if zm != nil && b.CellX >= 0 && b.CellX < zm.width && b.CellZ >= 0 && b.CellZ < zm.height {
		zm.Cells[b.CellZ][b.CellX].Density = 0
	}
	if roadsForBuildings != nil {
		roadsForBuildings.DamageNearby(b.Position.X, b.Position.Z, 15)
	}
	bm.Free(slot)
}

func (bm *BuildingManager) LoadAssets() {
}

func (bm *BuildingManager) Update(zm *ZoneManager, h *Heightmap, roads *RoadManager, dm *DistrictManager, transport *TransportManager) {
	bm.processCommands(zm)
	bm.calcDemand()
	bm.developZones(zm, roads)
	bm.updateBuildings(zm, h, roads, dm, transport)
}

func (bm *BuildingManager) calcDemand() {
	resPop := int32(0)
	comJobs := int32(0)
	indJobs := int32(0)
	offJobs := int32(0)
	total := 0
	bm.ForEach(func(b *Building, _ int32) {
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
	})
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

func (bm *BuildingManager) developZones(zm *ZoneManager, roads *RoadManager) {
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
			if buildabilityChecker != nil {
				info := buildabilityChecker.GetBuildability(cx, cz)
				if info.Score < 0.5 {
					continue
				}
				if info.IsUnderwater {
					continue
				}
			}
			if cell.Density >= 0.5 {
				continue
			}
			if bm.shouldDevelop(cell) {
				cell.Density += 0.005
			}
			if cell.Density >= 0.5 {
				bm.PushCmd(BuildingCommand{
					Type:      CmdBuildCreate,
					ZoneCellX: x,
					ZoneCellZ: z,
					ZoneType:  cell.Type,
				})
			}
		}
	}
}

var buildabilityChecker *BuildabilityChecker
var resourceForBuildings *ResourceSystem

func SetBuildabilityChecker(bc *BuildabilityChecker) {
	buildabilityChecker = bc
}

func SetResourceForBuildings(rs *ResourceSystem) {
	resourceForBuildings = rs
}

func (bm *BuildingManager) updateBuildings(zm *ZoneManager, h *Heightmap, roads *RoadManager, dm *DistrictManager, transport *TransportManager) {
	bm.Stats = BuildingStats{}
	cellSize := WorldSize / float32(zm.width)

	for i := 0; i < BuildingPoolSize; i++ {
		b := &bm.Pool[i]
		slot := int32(i)

		switch b.Lifecycle {
		case LifecycleInitializing:
			hasRoad := roads.HasNearbyRoad(b.Position.X, b.Position.Z, cellSize*2)
			if hasRoad {
				b.SetFlag(FlagHasRoad)
				b.ConstructTimer = 0
				b.Lifecycle = LifecycleActive
			} else {
				bm.Free(slot)
			}

		case LifecycleActive:
			if b.HasFlag(FlagAbandoned) {
				b.AbandonTimer++
				if b.AbandonTimer > 1800 {
					b.Lifecycle = LifecycleMarkedForRemoval
					b.RemovalTimer = 0
				}
				continue
			}

			isFlooded := false
			if waterForBuildings != nil {
				isFlooded = waterForBuildings.IsFlooded(b.Position.X, b.Position.Z)
			}
			if isFlooded {
				b.SetFlag(FlagFlooded)
				b.ClearFlag(FlagPowered)
				b.ClearFlag(FlagWatered)
				b.SetFlag(FlagAbandoned)
				b.AbandonTimer = 0
				b.Residents = 0
				b.Workers = 0
				continue
			} else {
				b.ClearFlag(FlagFlooded)
			}

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
				bm.Stats.TotalWealth += b.Household.Wealth
				bm.Stats.TotalHappiness += b.Household.Happiness
				bm.Stats.TotalPowerUsed += b.Consumption.Power
				bm.Stats.TotalWaterUsed += b.Consumption.Water
				bm.Stats.TotalGarbage += b.Consumption.Garbage
				if b.Household.Happiness < 80 {
					b.Household.Happiness++
				}
				if b.Household.Wealth < 70 {
					b.Household.Wealth++
				}
			}
			if b.Business != nil {
				bm.Stats.TotalPowerUsed += b.Consumption.Power
				bm.Stats.TotalWaterUsed += b.Consumption.Water
				if b.Type == ZoneCommercialLow || b.Type == ZoneCommercialHigh {
					if b.Business.GoodsStored > 0 && transportForBuildings != nil && transportForBuildings.Cargo != nil {
						cs := transportForBuildings.Cargo.NearestStation(b.Position.X, b.Position.Z, 5000)
						if cs != nil && cs.GoodsStored > 0 {
							if b.Business.GoodsStored < 50 {
								b.Business.GoodsStored += 2
							}
							if b.Business.Profitability < 80 {
								b.Business.Profitability += 2
							}
						} else {
							if b.Business.GoodsStored > 0 {
								b.Business.GoodsStored--
							}
							if b.Business.Profitability > 20 {
								b.Business.Profitability--
							}
						}
					} else {
						if b.Business.GoodsStored < 50 {
							b.Business.GoodsStored++
						}
						if b.Business.Profitability < 60 {
							b.Business.Profitability++
						}
					}
				} else if b.Type == ZoneIndustrial {
					if b.Business.GoodsStored < 100 {
						b.Business.GoodsStored++
					}
					if b.Business.Profitability < 60 {
						b.Business.Profitability++
					}
					if transportForBuildings != nil && transportForBuildings.Cargo != nil {
						cs := transportForBuildings.Cargo.NearestStation(b.Position.X, b.Position.Z, 5000)
						if cs != nil && cs.GoodsStored < cs.Capacity-10 {
							transfer := int32(5)
							if b.Business.GoodsStored < transfer {
								transfer = b.Business.GoodsStored
							}
							if cs.Capacity-cs.GoodsStored < transfer {
								transfer = cs.Capacity - cs.GoodsStored
							}
							b.Business.GoodsStored -= transfer
							cs.GoodsStored += transfer
						}
					}
				} else {
					if b.Business.GoodsStored < 50 {
						b.Business.GoodsStored++
					}
					if b.Business.Profitability < 60 {
						b.Business.Profitability++
					}
				}
				if b.Type == ZoneIndustrial && resourceForBuildings != nil {
					rx := int((b.Position.X/WorldSize + 0.5) * float32(HeightmapSize-1))
					rz := int((b.Position.Z/WorldSize + 0.5) * float32(HeightmapSize-1))
					if rx >= 0 && rx < HeightmapSize && rz >= 0 && rz < HeightmapSize {
						resourceForBuildings.ExtractOre(rx, rz, 0.001)
						resourceForBuildings.ExtractOil(rx, rz, 0.001)
					}
				}
			}
			if dm != nil {
				dm.ApplyPolicies(b)
			}

			if b.Household != nil && resourceForBuildings != nil {
				noise := resourceForBuildings.NoiseAt(b.Position.X, b.Position.Z)
				if noise > 0.2 {
					b.Household.Happiness -= 1
				}
				if transport != nil {
					cov := transport.CoverageScore(b.Position.X, b.Position.Z)
					if cov > 0.3 {
						b.Household.Happiness++
					}
					if cov < 0.1 && noise < 0.1 {
						b.Household.Happiness--
					}
				}
			}

			if b.Level >= 5 {
				continue
			}
			b.UpgradeTimer++
			needed := int32(600-b.Level*100) + b.Seed%60
			if b.UpgradeTimer > needed {
				lv := landValue(b, h, transport)
				if lv > b.Level*10 {
					bm.PushCmd(BuildingCommand{
						Type: CmdBuildUpgrade,
						Slot: slot,
					})
				}
			}

		case LifecycleSuspended:
			continue

		case LifecycleMarkedForRemoval:
			b.RemovalTimer++
			if b.RemovalTimer > 60 {
				b.Lifecycle = LifecycleDestroyed
			}

		case LifecycleDestroyed:
			if zm != nil && b.CellX >= 0 && b.CellX < zm.width && b.CellZ >= 0 && b.CellZ < zm.height {
				zm.Cells[b.CellZ][b.CellX].Density = 0
			}
			bm.Free(slot)
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

func landValue(b *Building, h *Heightmap, transport *TransportManager) int32 {
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
	if landValueTrees != nil {
		trees := landValueTrees.TreeCountAt(b.Position.X, b.Position.Z, 20)
		val += int32(trees) * 2
		if trees > 5 {
			val += 10
		}
	}
	if transport != nil {
		cov := transport.CoverageScore(b.Position.X, b.Position.Z)
		val += int32(cov * 20)
	}
	return val
}

var landValueTrees *TreeSystem

func SetLandValueTrees(ts *TreeSystem) {
	landValueTrees = ts
}

func (bm *BuildingManager) FlushStats() {
	bm.Stats = BuildingStats{}
	bm.ForEach(func(b *Building, _ int32) {
		if b.Household != nil {
			bm.Stats.TotalWealth += b.Household.Wealth
			bm.Stats.TotalHappiness += b.Household.Happiness
			bm.Stats.TotalPowerUsed += b.Consumption.Power
			bm.Stats.TotalWaterUsed += b.Consumption.Water
			bm.Stats.TotalGarbage += b.Consumption.Garbage
		}
		if b.Business != nil {
			bm.Stats.TotalPowerUsed += b.Consumption.Power
			bm.Stats.TotalWaterUsed += b.Consumption.Water
		}
	})
	if landValueTrees != nil {
		defPct := landValueTrees.DeforestPct()
		if defPct > 0.3 {
			penalty := int32(defPct * 50)
			bm.Stats.TotalHappiness -= penalty
		}
	}
}

func (bm *BuildingManager) Demand() (res, com, ind int) {
	return bm.resDemand, bm.comDemand, bm.indDemand
}

func (bm *BuildingManager) Population() int32 {
	total := int32(0)
	bm.ForEach(func(b *Building, _ int32) {
		total += b.Residents
	})
	return total
}

func (bm *BuildingManager) NearestInfo(wx, wz float32, radius float32) string {
	best := float32(radius * radius)
	var nearest *Building
	bm.ForEach(func(b *Building, _ int32) {
		dx := b.Position.X - wx
		dz := b.Position.Z - wz
		d := dx*dx + dz*dz
		if d < best {
			best = d
			nearest = b
		}
	})
	if nearest == nil {
		return ""
	}
	name := ZoneTypeName(nearest.Type)
	levelCount := nearest.Level
	lvl := ""
	for i := int32(0); i < levelCount; i++ {
		lvl += "I"
	}
	extra := " | "
	if nearest.HasFlag(FlagAbandoned) {
		extra += "ABANDONED"
	} else if nearest.Residents > 0 {
		extra += fmt.Sprintf("Pop: %d", nearest.Residents)
		if nearest.Household != nil {
			extra += fmt.Sprintf(" W:%.0f H:%d", nearest.Consumption.Power, nearest.Household.Happiness)
		}
	} else if nearest.Workers > 0 {
		extra += fmt.Sprintf("Jobs: %d", nearest.Workers)
		if nearest.Business != nil {
			extra += fmt.Sprintf(" P:%d%%", nearest.Business.Profitability)
		}
	}
	if !nearest.HasFlag(FlagConstructed) {
		pct := int(float32(nearest.ConstructTimer) / 300.0 * 100)
		extra += fmt.Sprintf(" Building %d%%", pct)
	}
	return fmt.Sprintf("%s Lvl %s%s", name, lvl, extra)
}

func (bm *BuildingManager) Draw(h *Heightmap, zm *ZoneManager, isNight bool) {
	bm.ForEach(func(b *Building, _ int32) {
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
			return
		}

		if b.HasFlag(FlagAbandoned) {
			grey := rl.NewColor(80, 80, 80, 255)
			rl.DrawCube(rl.NewVector3(b.Position.X, hy+b.Height*0.5*lvlScale, b.Position.Z), b.Width, b.Height*lvlScale, b.Depth, grey)
			return
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
			return
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
	})
}

var waterForBuildings *WaterSystem

func SetWaterForBuildings(ws *WaterSystem) {
	waterForBuildings = ws
}

var roadsForBuildings *RoadManager

func SetRoadsForBuildings(rm *RoadManager) {
	roadsForBuildings = rm
}

var transportForBuildings *TransportManager

func SetTransportForBuildings(tm *TransportManager) {
	transportForBuildings = tm
}

func (bm *BuildingManager) Unload() {
	for _, m := range bm.models {
		if m.MeshCount > 0 {
			rl.UnloadModel(m)
		}
	}
}
