package terrain

import (
	"math"

	"github.com/katwate/js-skylines/internal/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type WaterBodyType uint8

const (
	WaterOcean     WaterBodyType = 0
	WaterLake      WaterBodyType = 1
	WaterRiver     WaterBodyType = 2
	WaterReservoir WaterBodyType = 3
	WaterCanal     WaterBodyType = 4
)

type WaterBody struct {
	ID                uint32
	Type              WaterBodyType
	Level             float32
	Velocity          float32
	Pollution         float32
	CenterX, CenterZ  float32
	Radius            float32
	FlowDirX, FlowDirZ float32
}

const (
	WaterGridSize  = 129
	SeaLevel       = 0.15
	LakeThreshold  = 0.25
	FloodThreshold = 0.08
)

type WaterCell struct {
	Height   float32
	Velocity float32
	FlowX    float32
	FlowZ    float32
	Base     float32
}

type WaterSystem struct {
	Grid          [WaterGridSize][WaterGridSize]WaterCell
	Bodies        []WaterBody
	nextBodyID    uint32
	FloodActive   bool
	FloodCells    int
	FloodTimer    int32
	EventBus      *core.EventBus
}

func NewWaterSystem() *WaterSystem {
	return &WaterSystem{}
}

func (ws *WaterSystem) SetEventBus(eb *core.EventBus) {
	ws.EventBus = eb
}

func (ws *WaterSystem) Init(h *Heightmap) {
	riverTop := ActiveSeaLevel()
	for z := 0; z < WaterGridSize; z++ {
		for x := 0; x < WaterGridSize; x++ {
			hx := float64(x) / float64(WaterGridSize-1) * float64(HeightmapSize-1)
			hz := float64(z) / float64(WaterGridSize-1) * float64(HeightmapSize-1)

			terrainH := h.Get(int(hx), int(hz))
			cell := &ws.Grid[z][x]
			cell.Base = terrainH

			if FlatTestMap {
				if terrainH < riverTop {
					cell.Height = riverTop - terrainH
				} else {
					cell.Height = 0
				}
				continue
			}

			if terrainH < ActiveSeaLevel() {
				cell.Height = ActiveSeaLevel() - terrainH
			} else if terrainH < LakeThreshold {
				cell.Height = float32(math.Max(0, float64(LakeThreshold-terrainH)*0.3))
			} else {
				cell.Height = 0
			}
		}
	}

	ws.initDefaultBodies(h)
	if !FlatTestMap {
		ws.carveLake(h)
	}
}

func (ws *WaterSystem) initDefaultBodies(h *Heightmap) {
	if FlatTestMap {
		return
	}
	ws.Bodies = append(ws.Bodies, WaterBody{
		ID:       ws.nextBodyID,
		Type:     WaterOcean,
		Level:    SeaLevel * MaxHeight,
		Velocity: 0,
	})

	centerVal := h.Get(HeightmapSize/2, HeightmapSize/2)
	if centerVal < LakeThreshold {
		ws.Bodies = append(ws.Bodies, WaterBody{
			ID:       ws.nextBodyID,
			Type:     WaterLake,
			Level:    LakeThreshold * MaxHeight * 0.3,
			Velocity: 0.1,
			CenterX:  0,
			CenterZ:  0,
			Radius:   80,
		})
	}
	ws.nextBodyID++
}

func (ws *WaterSystem) AddRiver(worldX, worldZ, targetX, targetZ, width float32) uint32 {
	id := ws.nextBodyID
	ws.nextBodyID++
	dx := targetX - worldX
	dz := targetZ - worldZ
	dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	if dist < 0.01 {
		dist = 1
	}
	body := WaterBody{
		ID:       id,
		Type:     WaterRiver,
		Level:    SeaLevel*MaxHeight + 0.5,
		Velocity: 2.0,
		CenterX:  (worldX + targetX) * 0.5,
		CenterZ:  (worldZ + targetZ) * 0.5,
		Radius:   width * 0.5,
		FlowDirX: dx / dist,
		FlowDirZ: dz / dist,
	}
	ws.Bodies = append(ws.Bodies, body)
	ws.applyBody(body)
	return id
}

func (ws *WaterSystem) AddReservoir(worldX, worldZ, radius float32) uint32 {
	id := ws.nextBodyID
	ws.nextBodyID++
	body := WaterBody{
		ID:       id,
		Type:     WaterReservoir,
		Level:    LakeThreshold*MaxHeight*0.3 + 0.5,
		Velocity: 0.05,
		CenterX:  worldX,
		CenterZ:  worldZ,
		Radius:   radius,
	}
	ws.Bodies = append(ws.Bodies, body)
	ws.applyBody(body)
	return id
}

func (ws *WaterSystem) AddCanal(worldX, worldZ, targetX, targetZ, width float32) uint32 {
	id := ws.nextBodyID
	ws.nextBodyID++
	dx := targetX - worldX
	dz := targetZ - worldZ
	dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
	if dist < 0.01 {
		dist = 1
	}
	body := WaterBody{
		ID:       id,
		Type:     WaterCanal,
		Level:    SeaLevel*MaxHeight + 0.3,
		Velocity: 1.0,
		CenterX:  (worldX + targetX) * 0.5,
		CenterZ:  (worldZ + targetZ) * 0.5,
		Radius:   width * 0.5,
		FlowDirX: dx / dist,
		FlowDirZ: dz / dist,
	}
	ws.Bodies = append(ws.Bodies, body)
	ws.applyBody(body)
	return id
}

func (ws *WaterSystem) applyBody(body WaterBody) {
	cx := body.CenterX / WorldSize * float32(WaterGridSize-1)
	cz := body.CenterZ / WorldSize * float32(WaterGridSize-1)
	gcx := int(cx + float32(WaterGridSize)/2)
	gcz := int(cz + float32(WaterGridSize)/2)
	gridRadius := int(body.Radius / WorldSize * float32(WaterGridSize))

	minZ := max(0, gcz-gridRadius)
	maxZ := min(WaterGridSize-1, gcz+gridRadius)
	minX := max(0, gcx-gridRadius)
	maxX := min(WaterGridSize-1, gcx+gridRadius)

	for z := minZ; z <= maxZ; z++ {
		for x := minX; x <= maxX; x++ {
			dx := float32(x - gcx)
			dz := float32(z - gcz)
			dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if float32(gridRadius) <= 0 {
				continue
			}
			if dist/float32(gridRadius) > 1 {
				continue
			}
			falloff := 1 - dist/float32(gridRadius)
			targetH := body.Level / MaxHeight
			cell := &ws.Grid[z][x]
			if cell.Base < targetH {
				cell.Height = targetH - cell.Base
				if cell.Height < 0 {
					cell.Height = 0
				}
			}
			cell.Velocity = body.Velocity * falloff
			cell.FlowX = body.FlowDirX * cell.Velocity
			cell.FlowZ = body.FlowDirZ * cell.Velocity
		}
	}
}

func (ws *WaterSystem) carveLake(h *Heightmap) {
	center := HeightmapSize / 2
	radius := 20.0
	for z := center - 30; z <= center+30; z++ {
		for x := center - 30; x <= center+30; x++ {
			if x < 0 || x >= HeightmapSize || z < 0 || z >= HeightmapSize {
				continue
			}
			dist := math.Sqrt(float64((x-center)*(x-center) + (z-center)*(z-center)))
			if dist < radius {
				val := float64(h.Get(x, z))
				lower := (1 - dist/radius) * 0.15
				h.Set(x, z, float32(math.Max(0, val-lower)))
			}
		}
	}
}

func (ws *WaterSystem) Update(dt float64) {
	iterations := 3
	for iter := 0; iter < iterations; iter++ {
		for z := 1; z < WaterGridSize-1; z++ {
			for x := 1; x < WaterGridSize-1; x++ {
				cell := &ws.Grid[z][x]
				if cell.Height <= 0.001 {
					continue
				}

				flow := cell.Height * 0.25 * float32(dt)
				neighbors := [4]*WaterCell{
					&ws.Grid[z-1][x],
					&ws.Grid[z+1][x],
					&ws.Grid[z][x-1],
					&ws.Grid[z][x+1],
				}
				dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

				for i, n := range neighbors {
					nTotal := n.Height + n.Base
					cTotal := cell.Height + cell.Base
					if nTotal < cTotal && cell.Height > 0.001 {
						diff := cTotal - nTotal
						amount := math.Min(float64(flow), float64(diff)*0.1)
						amount = math.Min(amount, float64(cell.Height)*0.5)

						cell.Height -= float32(amount)
						n.Height += float32(amount)
						cell.FlowX += float32(dirs[i][0]) * float32(amount)
						cell.FlowZ += float32(dirs[i][1]) * float32(amount)
					}
				}
			}
		}
	}

	ws.propagateFlood(dt)
}

func (ws *WaterSystem) propagateFlood(dt float64) {
	fcells := 0
	for z := 1; z < WaterGridSize-1; z++ {
		for x := 1; x < WaterGridSize-1; x++ {
			cell := &ws.Grid[z][x]
			if cell.Height <= FloodThreshold {
				continue
			}
			fdepth := cell.Height - FloodThreshold
			if fdepth <= 0.001 {
				continue
			}
			fcells++

			spread := fdepth * 0.1 * float32(dt)
			neighbors := [4]*WaterCell{
				&ws.Grid[z-1][x],
				&ws.Grid[z+1][x],
				&ws.Grid[z][x-1],
				&ws.Grid[z][x+1],
			}
			for _, n := range neighbors {
				if n.Height <= FloodThreshold && n.Base < cell.Base+fdepth*0.5 {
					amount := float32(math.Min(float64(spread), float64(fdepth)*0.3))
					cell.Height -= amount
					n.Height += amount
				}
			}
		}
	}

	prev := ws.FloodActive
	ws.FloodActive = fcells > 0
	ws.FloodCells = fcells
	if ws.FloodActive {
		ws.FloodTimer++
	} else {
		ws.FloodTimer = 0
	}

	if prev != ws.FloodActive && ws.FloodActive {
		ws.emitFloodEvent(true)
	} else if prev != ws.FloodActive && !ws.FloodActive {
		ws.emitFloodEvent(false)
	}
}

func (ws *WaterSystem) emitFloodEvent(started bool) {
	if ws.EventBus != nil {
		if started {
			ws.EventBus.Emit(string(core.EventFloodStarted), nil)
		} else {
			ws.EventBus.Emit(string(core.EventFloodReceded), nil)
		}
	}
}

func waterWorldToGrid(worldX, worldZ float32) (x, z int, ok bool) {
	tx := (worldX + WorldSize/2) / WorldSize * float32(WaterGridSize-1)
	tz := (worldZ + WorldSize/2) / WorldSize * float32(WaterGridSize-1)
	x = int(tx)
	z = int(tz)
	if x < 0 || x >= WaterGridSize || z < 0 || z >= WaterGridSize {
		return 0, 0, false
	}
	return x, z, true
}

func (ws *WaterSystem) IsWet(worldX, worldZ float32) bool {
	x, z, ok := waterWorldToGrid(worldX, worldZ)
	if !ok {
		return false
	}
	return ws.Grid[z][x].Height > 0.01
}

func (ws *WaterSystem) FloodDepthAt(worldX, worldZ float32) float32 {
	x, z, ok := waterWorldToGrid(worldX, worldZ)
	if !ok {
		return 0
	}
	cell := &ws.Grid[z][x]
	if cell.Height <= FloodThreshold {
		return 0
	}
	return cell.Height - FloodThreshold
}

func (ws *WaterSystem) IsFlooded(worldX, worldZ float32) bool {
	return ws.FloodDepthAt(worldX, worldZ) > 0.001
}

func (ws *WaterSystem) BodyAt(worldX, worldZ float32) *WaterBody {
	for i := range ws.Bodies {
		b := &ws.Bodies[i]
		dx := worldX - b.CenterX
		dz := worldZ - b.CenterZ
		if dx*dx+dz*dz <= b.Radius*b.Radius || b.Type == WaterOcean {
			return b
		}
	}
	return nil
}

func (ws *WaterSystem) Draw() {
	if FlatTestMap {
		ws.drawFlatRiverWater()
		return
	}
	h := ActiveWaterSurfaceY()

	oceanCol := rl.NewColor(30, 120, 210, 160)
	lakeCol := rl.NewColor(40, 150, 220, 160)
	riverCol := rl.NewColor(50, 160, 230, 170)
	reservoirCol := rl.NewColor(60, 140, 200, 160)
	canalCol := rl.NewColor(70, 170, 210, 160)

	oceanDrawn := false
	for _, b := range ws.Bodies {
		var col rl.Color
		switch b.Type {
		case WaterOcean:
			col = oceanCol
		case WaterLake:
			col = lakeCol
		case WaterRiver:
			col = riverCol
		case WaterReservoir:
			col = reservoirCol
		case WaterCanal:
			col = canalCol
		}
		if b.Type == WaterOcean {
			if oceanDrawn {
				continue
			}
			oceanDrawn = true
			rl.DrawPlane(rl.NewVector3(0, h, 0), rl.NewVector2(WorldSize, WorldSize), col)
		} else {
			rl.DrawCube(rl.NewVector3(b.CenterX, h, b.CenterZ), b.Radius*2, 0.2, b.Radius*2, col)
		}
	}
	if !oceanDrawn {
		rl.DrawPlane(rl.NewVector3(0, h, 0), rl.NewVector2(WorldSize, WorldSize), oceanCol)
	}
}

func (ws *WaterSystem) drawFlatRiverWater() {
	col := rl.NewColor(50, 160, 230, 180)
	surfY := ActiveWaterSurfaceY()
	step := WorldSize / float32(WaterGridSize-1)
	for z := 0; z < WaterGridSize; z++ {
		for x := 0; x < WaterGridSize; x++ {
			if ws.Grid[z][x].Height < 0.001 {
				continue
			}
			wx := float32(x)/float32(WaterGridSize-1)*WorldSize - WorldSize/2
			wz := float32(z)/float32(WaterGridSize-1)*WorldSize - WorldSize/2
			rl.DrawCube(rl.NewVector3(wx, surfY, wz), step*1.05, 0.12, step*1.05, col)
		}
	}
}

func (ws *WaterSystem) Unload() {
}
