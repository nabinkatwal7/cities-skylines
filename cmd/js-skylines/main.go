package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/katwate/js-skylines/internal/terrain"
)

type buildState struct {
	active    bool
	startNode uint32
	roadType  terrain.RoadType
	zoneType  terrain.ZoneType
	zoneMode  bool
	parkMode  bool
}

type economy struct {
	money      float32
	taxTimer   int32
}

const (
	screenWidth  = 1280
	screenHeight = 720
)

func main() {
	t := terrain.NewManager(42)
	t.GenerateData()

	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(screenWidth, screenHeight, "JS Skylines - Go Edition")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	if err := t.LoadAssets(); err != nil {
		fmt.Printf("Warning: could not load assets: %v\n", err)
	}
	t.PrepareUpload()

	cam := rl.Camera3D{
		Position:   rl.NewVector3(100, 80, 100),
		Target:     rl.NewVector3(0, 0, 0),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       60,
		Projection: rl.CameraPerspective,
	}
	yaw := float32(-135 * math.Pi / 180)
	pitch := float32(25 * math.Pi / 180)
	dist := float32(140)
	target := rl.NewVector3(0, 0, 0)

	uploaded := false
	bld := buildState{roadType: terrain.RoadTwoLane, zoneType: terrain.ZoneResidentialLow}
	eco := economy{money: 100000}

	for !rl.WindowShouldClose() {
		if !uploaded {
			done := t.UploadNextBatch(16)
			rl.BeginDrawing()
			rl.ClearBackground(rl.DarkGray)
			pct := t.UploadProgress()
			text := fmt.Sprintf("Loading terrain... %d / %d", pct.Done, pct.Total)
			if done {
				uploaded = true
			}
			rl.DrawText(text, screenWidth/2-120, screenHeight/2-10, 20, rl.White)
			rl.EndDrawing()
			continue
		}

		t.Update(float64(rl.GetFrameTime()))

		wheel := rl.GetMouseWheelMove()
		dist -= wheel * 5
		if dist < 5 {
			dist = 5
		}
		if dist > 500 {
			dist = 500
		}

		if rl.IsMouseButtonDown(rl.MouseButtonRight) {
			delta := rl.GetMouseDelta()
			yaw -= delta.X * 0.005
			pitch -= delta.Y * 0.005
			if pitch > 1.4 {
				pitch = 1.4
			}
			if pitch < -0.3 {
				pitch = -0.3
			}
		}

		speed := float32(80.0) * float32(rl.GetFrameTime())
		if rl.IsKeyDown(rl.KeyW) {
			forward := rl.Vector3Subtract(cam.Target, cam.Position)
			forward.Y = 0
			forward = rl.Vector3Normalize(forward)
			target = rl.Vector3Add(target, rl.Vector3Scale(forward, speed))
		}
		if rl.IsKeyDown(rl.KeyS) {
			forward := rl.Vector3Subtract(cam.Target, cam.Position)
			forward.Y = 0
			forward = rl.Vector3Normalize(forward)
			target = rl.Vector3Add(target, rl.Vector3Scale(forward, -speed))
		}
		if rl.IsKeyDown(rl.KeyA) {
			forward := rl.Vector3Subtract(cam.Target, cam.Position)
			forward.Y = 0
			forward = rl.Vector3Normalize(forward)
			right := rl.Vector3CrossProduct(forward, rl.NewVector3(0, 1, 0))
			target = rl.Vector3Add(target, rl.Vector3Scale(right, -speed))
		}
		if rl.IsKeyDown(rl.KeyD) {
			forward := rl.Vector3Subtract(cam.Target, cam.Position)
			forward.Y = 0
			forward = rl.Vector3Normalize(forward)
			right := rl.Vector3CrossProduct(forward, rl.NewVector3(0, 1, 0))
			target = rl.Vector3Add(target, rl.Vector3Scale(right, speed))
		}
		if rl.IsKeyDown(rl.KeyQ) {
			target.Y -= speed
		}
		if rl.IsKeyDown(rl.KeyE) {
			target.Y += speed
		}

		cam.Target = target
		cam.Position.X = target.X + dist*float32(math.Cos(float64(pitch)))*float32(math.Sin(float64(yaw)))
		cam.Position.Y = target.Y + dist*float32(math.Sin(float64(pitch)))
		cam.Position.Z = target.Z + dist*float32(math.Cos(float64(pitch)))*float32(math.Cos(float64(yaw)))

		eco.taxTimer++
		if eco.taxTimer > 60 {
			eco.taxTimer = 0
			eco.money += float32(len(t.Buildings.Buildings)) * 0.5
		}

		if rl.IsKeyPressed(rl.KeyTab) {
			bld.zoneMode = !bld.zoneMode
			bld.parkMode = false
			bld.active = false
		}
		if rl.IsKeyPressed(rl.KeyF) {
			bld.parkMode = !bld.parkMode
			bld.zoneMode = false
			bld.active = false
		}

		worldX := float32(0)
		worldZ := float32(0)
		mouseOnTerrain := false
		ray := rl.GetScreenToWorldRay(rl.GetMousePosition(), cam)
		tPlane := float32(0)
		if ray.Direction.Y != 0 {
			tPlane = -ray.Position.Y / ray.Direction.Y
			worldX = ray.Position.X + ray.Direction.X*tPlane
			worldZ = ray.Position.Z + ray.Direction.Z*tPlane
			mouseOnTerrain = true
		}

		if bld.parkMode {
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mouseOnTerrain && eco.money >= 500 {
				t.Services.AddPark(worldX, worldZ)
				eco.money -= 500
			}
		} else if bld.zoneMode {
			if rl.IsKeyPressed(rl.KeyR) {
				bld.zoneType = (bld.zoneType + 1) % 7
				if bld.zoneType == terrain.ZoneNone {
					bld.zoneType = 1
				}
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mouseOnTerrain && eco.money >= 20 {
				t.Zones.SetZone(worldX, worldZ, bld.zoneType)
				eco.money -= 20
			}
		} else {
			if rl.IsKeyPressed(rl.KeyR) {
				bld.roadType = (bld.roadType + 1) % 4
			}
			if rl.IsKeyPressed(rl.KeyEscape) {
				bld.active = false
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mouseOnTerrain && !rl.IsKeyDown(rl.KeyLeftShift) {
				cx := clamp(worldX, -240, 240)
				cz := clamp(worldZ, -240, 240)
				if !bld.active {
					if eco.money >= 100 {
						bld.active = true
						bld.startNode = t.Roads.AddNode(cx, cz)
						eco.money -= 100
					}
				} else {
					if eco.money >= 100 {
						endNode := t.Roads.AddNode(cx, cz)
						t.Roads.AddSegment(bld.startNode, endNode, bld.roadType)
						t.Roads.Rebuild(t.Heightmap)
						bld.startNode = endNode
						eco.money -= 100
					}
				}
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(135, 206, 235, 255))
		rl.BeginMode3D(cam)
		t.Draw(cam.Position.X, cam.Position.Z)
		if !bld.zoneMode && bld.active && mouseOnTerrain {
			h := t.Heightmap.WorldHeight(worldX, worldZ)
			rl.DrawSphere(rl.NewVector3(worldX, h+0.5, worldZ), 0.8, rl.Green)
			startNode := &t.Roads.Nodes[bld.startNode]
			sh := t.Heightmap.WorldHeight(startNode.X, startNode.Z)
			rl.DrawLine3D(rl.NewVector3(startNode.X, sh+0.2, startNode.Z), rl.NewVector3(worldX, h+0.2, worldZ), rl.Green)
		}
		if bld.parkMode && mouseOnTerrain {
			h := t.Heightmap.WorldHeight(worldX, worldZ)
			rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 3, 0.2, 3, rl.NewColor(80, 200, 80, 120))
		}
		if bld.zoneMode && mouseOnTerrain {
			h := t.Heightmap.WorldHeight(worldX, worldZ)
			rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ), 8, 0.3, 8, terrain.ZoneColor(bld.zoneType))
		}
		if mouseOnTerrain {
			h := t.Heightmap.WorldHeight(worldX, worldZ)
			rl.DrawSphere(rl.NewVector3(worldX, h+0.3, worldZ), 0.4, rl.Red)
		}
		rl.DrawGrid(100, 4.0)
		rl.EndMode3D()
		rl.DrawFPS(10, 10)
		rl.DrawText(fmt.Sprintf("$%.0f", eco.money), 10, 10, 18, rl.NewColor(100, 220, 100, 220))
		pop := t.Buildings.Population()
		rl.DrawText(fmt.Sprintf("Pop: %d", pop), 100, 10, 18, rl.White)
		helpY := int32(35)
		if bld.parkMode {
			rl.DrawText(fmt.Sprintf("PARK MODE: place parks (L-click) | F=toggle off | TAB=roads | $500 each"), 10, helpY, 15, rl.Green)
		} else if bld.zoneMode {
			rl.DrawText(fmt.Sprintf("ZONE MODE: paint zones (L-click drag) | R=change type | TAB=back to roads"), 10, helpY, 15, rl.Blue)
			rl.DrawText(fmt.Sprintf("Zone: %s", zoneTypeName(bld.zoneType)), 10, helpY+20, 15, rl.Gray)
		} else if bld.active {
			rl.DrawText(fmt.Sprintf("ROAD BUILDING: place road (L-click) | R=change type | Esc=cancel | TAB=zones"), 10, helpY, 15, rl.Green)
			rl.DrawText(fmt.Sprintf("Road type: %s", roadTypeName(bld.roadType)), 10, helpY+20, 15, rl.Gray)
		} else {
			rl.DrawText(fmt.Sprintf("L-click=build road | TAB=zone mode | F=parks | WASD=pan | Scroll=zoom | R-drag=orbit"), 10, helpY, 15, rl.White)
		}
		if mouseOnTerrain {
			rl.DrawText(fmt.Sprintf("(%.1f, %.1f)", worldX, worldZ), screenWidth-200, 50, 15, rl.Gray)
			drawBuildingInfo(t, worldX, worldZ)
		}
		drawDemandBars(t)
		rl.EndDrawing()
	}

	t.Unload()
}

func drawBuildingInfo(t *terrain.Manager, wx, wz float32) {
	info := t.Buildings.NearestInfo(wx, wz, 8)
	if info != "" {
		rl.DrawRectangle(screenWidth/2-150, 10, 300, 50, rl.NewColor(0, 0, 0, 180))
		rl.DrawText(info, screenWidth/2-140, 18, 16, rl.White)
	}
}

func drawDemandBars(t *terrain.Manager) {
	r, c, i := t.Buildings.Demand()
	bx := int32(screenWidth - 300)
	by := int32(10)
	bw := int32(80)

	bar := func(label string, val int, y int32, col rl.Color) {
		rl.DrawText(label, bx, by+y, 14, rl.Gray)
		w := clampInt32(int32(val)*10, 0, bw)
		rl.DrawRectangle(bx+40, by+y, w, 12, col)
		rl.DrawRectangleLines(bx+40, by+y, bw, 12, rl.DarkGray)
	}
	bar("R", r, 0, rl.NewColor(100, 200, 100, 220))
	bar("C", c, 20, rl.NewColor(100, 150, 255, 220))
	bar("I", i, 40, rl.NewColor(255, 200, 80, 220))
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func roadTypeName(rt terrain.RoadType) string {
	switch rt {
	case terrain.RoadTwoLane:
		return "Two-Lane Road"
	case terrain.RoadOneWay:
		return "One-Way Road"
	case terrain.RoadFourLane:
		return "Four-Lane Road"
	case terrain.RoadGravel:
		return "Gravel Road"
	default:
		return "Unknown"
	}
}

func clampInt32(v, min, max int32) int32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func zoneTypeName(zt terrain.ZoneType) string {
	switch zt {
	case terrain.ZoneResidentialLow:
		return "Residential Low"
	case terrain.ZoneResidentialHigh:
		return "Residential High"
	case terrain.ZoneCommercialLow:
		return "Commercial Low"
	case terrain.ZoneCommercialHigh:
		return "Commercial High"
	case terrain.ZoneIndustrial:
		return "Industrial"
	case terrain.ZoneOffice:
		return "Office"
	default:
		return "None"
	}
}
