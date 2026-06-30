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
	bld := buildState{roadType: terrain.RoadTwoLane}

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

		if rl.IsKeyPressed(rl.KeyR) {
			bld.roadType = (bld.roadType + 1) % 4
		}
		if rl.IsKeyPressed(rl.KeyEscape) {
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

		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mouseOnTerrain && !rl.IsKeyDown(rl.KeyLeftShift) {
			cx := clamp(worldX, -240, 240)
			cz := clamp(worldZ, -240, 240)
			if !bld.active {
				bld.active = true
				bld.startNode = t.Roads.AddNode(cx, cz)
			} else {
				endNode := t.Roads.AddNode(cx, cz)
				t.Roads.AddSegment(bld.startNode, endNode, bld.roadType)
				bld.startNode = endNode
			}
		}
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(135, 206, 235, 255))
		rl.BeginMode3D(cam)
		t.Draw(cam.Position.X, cam.Position.Z)
		if bld.active && mouseOnTerrain {
			h := t.Heightmap.WorldHeight(worldX, worldZ)
			rl.DrawSphere(rl.NewVector3(worldX, h+0.5, worldZ), 0.8, rl.Green)
			startNode := &t.Roads.Nodes[bld.startNode]
			sh := t.Heightmap.WorldHeight(startNode.X, startNode.Z)
			rl.DrawLine3D(rl.NewVector3(startNode.X, sh+0.2, startNode.Z), rl.NewVector3(worldX, h+0.2, worldZ), rl.Green)
		}
		if mouseOnTerrain {
			h := t.Heightmap.WorldHeight(worldX, worldZ)
			rl.DrawSphere(rl.NewVector3(worldX, h+0.3, worldZ), 0.4, rl.Red)
		}
		rl.DrawGrid(100, 4.0)
		rl.EndMode3D()
		rl.DrawFPS(10, 10)
		helpY := int32(30)
		if bld.active {
			rl.DrawText(fmt.Sprintf("BUILDING: place road (L-click) | R=change type | Esc/R-click=cancel"), 10, helpY, 15, rl.Green)
		} else {
			rl.DrawText(fmt.Sprintf("L-click on terrain to start building roads"), 10, helpY, 15, rl.White)
		}
		rl.DrawText(fmt.Sprintf("Road type: %s", roadTypeName(bld.roadType)), 10, helpY+20, 15, rl.Gray)
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
