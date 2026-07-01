package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/katwate/js-skylines/internal/terrain"
)

type economy struct {
	money    float32
	taxTimer int32
}

var timeOfDay = 1 // 0=dawn, 1=day, 2=dusk, 3=night

func skyColor() rl.Color {
	switch timeOfDay {
	case 0:
		return rl.NewColor(200, 140, 100, 255)
	case 1:
		return rl.NewColor(135, 206, 235, 255)
	case 2:
		return rl.NewColor(180, 100, 80, 255)
	case 3:
		return rl.NewColor(20, 20, 40, 255)
	}
	return rl.RayWhite
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

	buildingAssetsLoaded := false
	ui := terrain.NewGameUI()

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
	eco := economy{money: 100000}
	roadActive := false
	roadStartNode := uint32(0)
	roadElevation := int32(0)

	for !rl.WindowShouldClose() {
		if !uploaded {
			rl.BeginDrawing()
			rl.ClearBackground(rl.DarkGray)
			pct := t.UploadProgress()
			text := fmt.Sprintf("Loading terrain... %d / %d", pct.Done, pct.Total)
			if !buildingAssetsLoaded {
				text = "Loading building models..."
				rl.DrawText(text, screenWidth/2-120, screenHeight/2-10, 20, rl.White)
				rl.EndDrawing()
				t.LoadBuildingAssets()
				buildingAssetsLoaded = true
				continue
			}
			done := t.UploadNextBatch(16)
			text = fmt.Sprintf("Loading terrain... %d / %d", pct.Done, pct.Total)
			if done {
				uploaded = true
			}
			rl.DrawText(text, screenWidth/2-120, screenHeight/2-10, 20, rl.White)
			rl.EndDrawing()
			continue
		}

		t.Update(float64(rl.GetFrameTime()))

		// Camera
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

		// Time cycle
		if rl.IsKeyPressed(rl.KeyT) {
			timeOfDay = (timeOfDay + 1) % 4
		}
		t.Night = timeOfDay == 3

		// Road elevation control
		if rl.IsKeyPressed(rl.KeyPageUp) {
			if ui.Selected == terrain.ToolRoad {
				currentEle := &roadElevation
				*currentEle++
				if *currentEle > 2 {
					*currentEle = 2
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyPageDown) {
			if ui.Selected == terrain.ToolRoad {
				currentEle := &roadElevation
				*currentEle--
				if *currentEle < 0 {
					*currentEle = 0
				}
			}
		}

		// Economy
		eco.taxTimer++
		if eco.taxTimer > 60 {
			eco.taxTimer = 0
			eco.money += float32(len(t.Buildings.Buildings)) * 0.5
		}

		// Mouse ray
		worldX := float32(0)
		worldZ := float32(0)
		mouseOnTerrain := false
		ray := rl.GetScreenToWorldRay(rl.GetMousePosition(), cam)
		if ray.Direction.Y != 0 {
			tPlane := -ray.Position.Y / ray.Direction.Y
			worldX = ray.Position.X + ray.Direction.X*tPlane
			worldZ = ray.Position.Z + ray.Direction.Z*tPlane
			mouseOnTerrain = true
		}

		// Update UI state
		ui.Money = eco.money
		ui.Population = t.Buildings.Population()
		rDem, cDem, iDem := t.Buildings.Demand()
		ui.ResDemand = rDem
		ui.ComDemand = cDem
		ui.IndDemand = iDem
		timeNames := []string{"Dawn", "Day", "Dusk", "Night"}
		ui.TimeStr = timeNames[timeOfDay]
		ui.MouseWorldX = worldX
		ui.MouseWorldZ = worldZ
		ui.MouseOnGround = mouseOnTerrain
		bInfo := t.Buildings.NearestInfo(worldX, worldZ, 8)
		ui.BuildingInfo = bInfo

		// Handle keyboard tool selection
		ui.HandleInput()

		// Handle mouse clicks for tools and toolbar
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			// Check toolbar click first
			mPos := rl.GetMousePosition()
			my := int32(mPos.Y)
			if my >= 660 {
				ui.HandleClick()
				roadActive = false
			} else if mouseOnTerrain {
				switch ui.Selected {
				case terrain.ToolRoad:
					cx := clamp(worldX, -240, 240)
					cz := clamp(worldZ, -240, 240)
					if t.Heightmap.IsUnderwater(cx, cz) {
						break
					}
					if eco.money >= 100 {
						if !roadActive {
							roadActive = true
							roadStartNode = t.Roads.AddNode(cx, cz)
							eco.money -= 100
						} else {
							sn := &t.Roads.Nodes[roadStartNode]
							if t.Heightmap.IsUnderwater(sn.X, sn.Z) {
								roadActive = false
								break
							}
							endNode := t.Roads.AddNode(cx, cz)
							segID := t.Roads.AddSegment(roadStartNode, endNode, ui.RoadType)
							if roadElevation > 0 {
								for i := range t.Roads.Segments {
									if t.Roads.Segments[i].ID == segID {
										t.Roads.Segments[i].Elevation = roadElevation
									}
								}
							}
							t.Roads.Rebuild(t.Heightmap)
							roadStartNode = endNode
							eco.money -= 100
						}
					}
				case terrain.ToolZone:
					if eco.money >= 20 && !t.Heightmap.IsUnderwater(worldX, worldZ) {
						t.Zones.SetZone(worldX, worldZ, ui.ZoneType, t.Roads)
						if t.Zones.CellTypeAt(worldX, worldZ) == ui.ZoneType {
							eco.money -= 20
						}
					}
				case terrain.ToolPark:
					if eco.money >= 500 && !t.Heightmap.IsUnderwater(worldX, worldZ) {
						t.Services.AddPark(worldX, worldZ)
						eco.money -= 500
					}
				case terrain.ToolRemove:
					idx := t.Roads.NearestSegment(worldX, worldZ)
					if idx >= 0 {
						t.Roads.RemoveSegment(idx)
						t.Roads.Rebuild(t.Heightmap)
					}
				case terrain.ToolUpgrade:
					idx := t.Roads.NearestSegment(worldX, worldZ)
					if idx >= 0 {
						t.Roads.UpgradeSegment(idx, ui.RoadType)
						t.Roads.Rebuild(t.Heightmap)
					}
				}
			}
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			roadActive = false
		}

		// --- Draw ---
		rl.BeginDrawing()
		rl.ClearBackground(skyColor())
		rl.BeginMode3D(cam)
		t.Draw(cam.Position.X, cam.Position.Z)

		// Draw previews
		if mouseOnTerrain {
			h := t.Heightmap.WorldHeight(worldX, worldZ)

			switch ui.Selected {
			case terrain.ToolRoad:
				rl.DrawSphere(rl.NewVector3(worldX, h+0.5, worldZ), 0.8, rl.Green)
				if roadActive {
					sn := &t.Roads.Nodes[roadStartNode]
					sh := t.Heightmap.WorldHeight(sn.X, sn.Z)
					rl.DrawLine3D(rl.NewVector3(sn.X, sh+0.2, sn.Z), rl.NewVector3(worldX, h+0.2, worldZ), rl.Green)
				}
			case terrain.ToolZone:
				rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ), 8, 0.3, 8, terrain.ZoneColor(ui.ZoneType))
			case terrain.ToolPark:
				rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 3, 0.2, 3, rl.NewColor(80, 200, 80, 120))
			case terrain.ToolRemove:
				idx := t.Roads.NearestSegment(worldX, worldZ)
				if idx >= 0 {
					seg := t.Roads.Segments[idx]
					na := &t.Roads.Nodes[seg.NodeA]
					nb := &t.Roads.Nodes[seg.NodeB]
					ha := t.Heightmap.WorldHeight(na.X, na.Z) + 0.5
					hb := t.Heightmap.WorldHeight(nb.X, nb.Z) + 0.5
					rl.DrawLine3D(rl.NewVector3(na.X, ha, na.Z), rl.NewVector3(nb.X, hb, nb.Z), rl.Red)
				}
			case terrain.ToolUpgrade:
				idx := t.Roads.NearestSegment(worldX, worldZ)
				if idx >= 0 {
					seg := t.Roads.Segments[idx]
					na := &t.Roads.Nodes[seg.NodeA]
					nb := &t.Roads.Nodes[seg.NodeB]
					ha := t.Heightmap.WorldHeight(na.X, na.Z) + 0.5
					hb := t.Heightmap.WorldHeight(nb.X, nb.Z) + 0.5
					rl.DrawLine3D(rl.NewVector3(na.X, ha, na.Z), rl.NewVector3(nb.X, hb, nb.Z), rl.Yellow)
				}
			}

			rl.DrawSphere(rl.NewVector3(worldX, h+0.3, worldZ), 0.4, rl.Red)
		}
		rl.DrawGrid(100, 4.0)
		rl.EndMode3D()

		ui.Draw()
		rl.DrawFPS(10, 30)

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
