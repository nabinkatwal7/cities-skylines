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

		if rl.IsKeyPressed(rl.KeyTab) {
			bld.zoneMode = !bld.zoneMode
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

		if bld.zoneMode {
			if rl.IsKeyPressed(rl.KeyR) {
				bld.zoneType = (bld.zoneType + 1) % 7
				if bld.zoneType == terrain.ZoneNone {
					bld.zoneType = 1
				}
			}
			if rl.IsMouseButtonDown(rl.MouseButtonLeft) && mouseOnTerrain && !rl.IsKeyDown(rl.KeyLeftShift) {
				t.Zones.SetZone(worldX, worldZ, bld.zoneType)
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
					bld.active = true
					bld.startNode = t.Roads.AddNode(cx, cz)
				} else {
					endNode := t.Roads.AddNode(cx, cz)
					t.Roads.AddSegment(bld.startNode, endNode, bld.roadType)
					t.Roads.Rebuild(t.Heightmap)
					bld.startNode = endNode
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
		helpY := int32(30)
		if bld.zoneMode {
			rl.DrawText(fmt.Sprintf("ZONE MODE: paint zones (L-click drag) | R=change type | TAB=back to roads"), 10, helpY, 15, rl.Blue)
			rl.DrawText(fmt.Sprintf("Zone: %s", zoneTypeName(bld.zoneType)), 10, helpY+20, 15, rl.Gray)
		} else if bld.active {
			rl.DrawText(fmt.Sprintf("ROAD BUILDING: place road (L-click) | R=change type | Esc=cancel | TAB=zones"), 10, helpY, 15, rl.Green)
			rl.DrawText(fmt.Sprintf("Road type: %s", roadTypeName(bld.roadType)), 10, helpY+20, 15, rl.Gray)
		} else {
			rl.DrawText(fmt.Sprintf("L-click=build road | TAB=zone mode | WASD=pan | Scroll=zoom | R-drag=orbit"), 10, helpY, 15, rl.White)
		}
		if mouseOnTerrain {
			rl.DrawText(fmt.Sprintf("(%.1f, %.1f)", worldX, worldZ), screenWidth-200, 50, 15, rl.Gray)
		}
		rl.EndDrawing()
	}

	t.Unload()
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
