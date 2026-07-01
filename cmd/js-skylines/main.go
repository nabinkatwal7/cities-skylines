package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/katwate/js-skylines/internal/terrain"
)

func skyColor(isNight bool) rl.Color {
	if isNight {
		return rl.NewColor(20, 20, 40, 255)
	}
	return rl.NewColor(135, 206, 235, 255)
}

const (
	screenWidth  = 1280
	screenHeight = 720
)

func main() {
	sim := terrain.NewSimulationManager(42)
	sim.InitDefaultRoads()
	sim.Parking.GenerateRoadsideSpots(sim.Roads)

	t := terrain.NewManager(sim)
	t.InitChunks()
	sim.InitTerraform(t.Chunks, func(idx int) { t.RebuildChunk(idx) })

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

		sim.Update(float64(rl.GetFrameTime()))

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

		// Simulation speed / pause
		if rl.IsKeyPressed(rl.KeySpace) {
			sim.TogglePause()
		}
		if rl.IsKeyPressed(rl.KeyOne) {
			sim.SetSpeed(1)
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			sim.SetSpeed(2)
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			sim.SetSpeed(3)
		}

		// Road elevation control
		if rl.IsKeyPressed(rl.KeyPageUp) {
			if ui.Selected == terrain.ToolRoad {
				switch roadElevation {
				case 0:
					roadElevation = 1
				case 1:
					roadElevation = 2
				case 2:
					roadElevation = -1
				case -1:
					roadElevation = 0
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyPageDown) {
			if ui.Selected == terrain.ToolRoad {
				switch roadElevation {
				case 0:
					roadElevation = -1
				case -1:
					roadElevation = 2
				case 2:
					roadElevation = 1
				case 1:
					roadElevation = 0
				}
			}
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
		ui.Money = sim.Money
		ui.Population = sim.Buildings.Population()
		rDem, cDem, iDem := sim.Buildings.Demand()
		ui.ResDemand = rDem
		ui.ComDemand = cDem
		ui.IndDemand = iDem
		ui.TimeStr = sim.Time.TimeString()
		if sim.Time.IsPaused {
			ui.TimeStr += " ⏸"
		} else if sim.Time.Speed > 1 {
			speedStr := fmt.Sprintf(" ⏩x%.0f", sim.Time.Speed)
			ui.TimeStr += speedStr
		}
		ui.MouseWorldX = worldX
		ui.MouseWorldZ = worldZ
		ui.MouseOnGround = mouseOnTerrain
		bInfo := sim.Buildings.NearestInfo(worldX, worldZ, 8)
		ui.BuildingInfo = bInfo

		// Snap to outside connection for preview
		previewX, previewZ := worldX, worldZ
		if mouseOnTerrain && ui.Selected == terrain.ToolRoad {
			for _, c := range sim.Connections.GetByType(terrain.ConnHighway) {
				dx := c.WorldX - worldX
				dz := c.WorldZ - worldZ
				if dx*dx+dz*dz < 64 {
					for idx := range sim.Roads.Nodes {
						n := &sim.Roads.Nodes[idx]
						if n.Flags&terrain.RoadFlagOutsideConn != 0 {
							nx := n.X - c.WorldX
							nz := n.Z - c.WorldZ
							if nx*nx+nz*nz < 0.01 {
								previewX = n.X
								previewZ = n.Z
							}
						}
					}
				}
			}
		}

		// Handle keyboard tool selection
		ui.HandleInput()

		// Handle mouse clicks for tools and toolbar
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			mPos := rl.GetMousePosition()
			my := int32(mPos.Y)
			if my >= 660 {
				ui.HandleClick()
				roadActive = false
			} else if mouseOnTerrain {
				switch ui.Selected {
				case terrain.ToolRoad:
					px := clamp(previewX, -240, 240)
					pz := clamp(previewZ, -240, 240)
					if sim.Heightmap.IsUnderwater(px, pz) {
						break
					}
					if sim.Money >= 100 {
						if !roadActive {
							roadActive = true
							roadStartNode = sim.PlaceRoadNode(px, pz)
						} else {
							sn := &sim.Roads.Nodes[roadStartNode]
							if sim.Heightmap.IsUnderwater(sn.X, sn.Z) {
								roadActive = false
								break
							}
							newNode, _, ok := sim.PlaceRoadSegment(roadStartNode, px, pz, ui.RoadType, roadElevation)
							if ok {
								roadStartNode = newNode
							}
							if !ok {
								ui.HelpText = "Cannot place road: check terrain slope, water, or obstacles"
							}
						}
					}
				case terrain.ToolZone:
					if sim.Money >= 20 && !sim.Heightmap.IsUnderwater(worldX, worldZ) {
						sim.SetZone(worldX, worldZ, ui.ZoneType)
					}
			case terrain.ToolPark:
				if sim.Money >= 500 && !sim.Heightmap.IsUnderwater(worldX, worldZ) {
					sim.PlacePark(worldX, worldZ)
				}
			case terrain.ToolParking:
				cost := float32(1000)
				if ui.ParkingGarage {
					cost = 3000
				}
				if sim.Money >= cost && !sim.Heightmap.IsUnderwater(worldX, worldZ) {
					sim.PlaceParkingLot(worldX, worldZ, ui.ParkingGarage)
				}
				case terrain.ToolRemove:
					idx := sim.Roads.NearestSegment(worldX, worldZ)
					if idx >= 0 {
						sim.RemoveSegment(idx)
					}
					sim.RemoveTrees(worldX, worldZ, 10)
				case terrain.ToolUpgrade:
					idx := sim.Roads.NearestSegment(worldX, worldZ)
					if idx >= 0 {
						if !sim.UpgradeSegment(idx, ui.RoadType) {
							ui.HelpText = "Cannot upgrade: insufficient funds or already that type"
						}
					}
				}
			}
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			roadActive = false
		}

		// --- Draw ---
		rl.BeginDrawing()
		rl.ClearBackground(skyColor(sim.Night))
		rl.BeginMode3D(cam)
		t.Draw(cam.Position.X, cam.Position.Z)

		// Draw previews
		if mouseOnTerrain {
			h := sim.Heightmap.WorldHeight(worldX, worldZ)

			switch ui.Selected {
			case terrain.ToolRoad:
				rl.DrawSphere(rl.NewVector3(previewX, h+0.5, previewZ), 0.8, rl.Green)
				if roadActive {
					sn := &sim.Roads.Nodes[roadStartNode]
					sh := sim.Heightmap.WorldHeight(sn.X, sn.Z)
					rl.DrawLine3D(rl.NewVector3(sn.X, sh+0.2, sn.Z), rl.NewVector3(previewX, h+0.2, previewZ), rl.Green)
				}
			case terrain.ToolZone:
				rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ), 8, 0.3, 8, terrain.ZoneColor(ui.ZoneType))
			case terrain.ToolPark:
				rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 3, 0.2, 3, rl.NewColor(80, 200, 80, 120))
			case terrain.ToolParking:
				col := rl.NewColor(80, 80, 200, 100)
				if !ui.ParkingGarage {
					col = rl.NewColor(80, 160, 80, 80)
				}
				rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 20, 0.3, 15, col)
				rl.DrawCubeWires(rl.NewVector3(worldX, h+0.3, worldZ), 20, 0.3, 15, rl.NewColor(60, 60, 100, 150))
			case terrain.ToolRemove:
				idx := sim.Roads.NearestSegment(worldX, worldZ)
				if idx >= 0 {
					seg := sim.Roads.Segments[idx]
					na := &sim.Roads.Nodes[seg.NodeA]
					nb := &sim.Roads.Nodes[seg.NodeB]
					ha := sim.Heightmap.WorldHeight(na.X, na.Z) + 0.5
					hb := sim.Heightmap.WorldHeight(nb.X, nb.Z) + 0.5
					rl.DrawLine3D(rl.NewVector3(na.X, ha, na.Z), rl.NewVector3(nb.X, hb, nb.Z), rl.Red)
				}
			case terrain.ToolUpgrade:
				idx := sim.Roads.NearestSegment(worldX, worldZ)
				if idx >= 0 {
					seg := sim.Roads.Segments[idx]
					na := &sim.Roads.Nodes[seg.NodeA]
					nb := &sim.Roads.Nodes[seg.NodeB]
					ha := sim.Heightmap.WorldHeight(na.X, na.Z) + 0.5
					hb := sim.Heightmap.WorldHeight(nb.X, nb.Z) + 0.5
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
