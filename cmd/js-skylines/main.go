package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/save"
	simpkg "github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/ui"
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
	sim := simpkg.NewSimulationManager(42)
	sim.InitDefaultRoads()
	sim.Parking.GenerateRoadsideSpots(sim.Roads)

	t := simpkg.NewManager(sim)
	t.InitChunks()
	sim.InitTerraform(t.Chunks, func(idx int) { t.RebuildChunk(idx) })

	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(screenWidth, screenHeight, "JS Skylines - Go Edition")
	defer rl.CloseWindow()
	ui.LoadUIFont()
	defer ui.UnloadUIFont()
	rl.SetTargetFPS(60)

	if err := t.LoadAssets(); err != nil {
		fmt.Printf("Warning: could not load assets: %v\n", err)
	}
	t.PrepareUpload()

	gameUI := ui.NewGameUI()

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
	transportActive := false
	transportStartStopID := uint32(0)
	transportLineID := uint32(0)
	saveTimer := int32(0)
	saveFilename := "autosave.sav"

	for !rl.WindowShouldClose() {
		if !uploaded {
			rl.BeginDrawing()
			rl.ClearBackground(rl.DarkGray)
			pct := t.UploadProgress()
			text := fmt.Sprintf("Loading terrain... %d / %d", pct.Done, pct.Total)
			done := t.UploadNextBatch(16)
			if done {
				uploaded = true
			}
			ui.DrawUIText(text, screenWidth/2-120, screenHeight/2-10, 20, rl.White)
			rl.EndDrawing()
			continue
		}

		sim.Update(float64(rl.GetFrameTime()))

		saveTimer++
		if saveTimer > 3600 {
			saveTimer = 0
			save.SaveGame(saveFilename, sim, sim.Money, int32(sim.Time.TotalTime))
		}

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
		if rl.IsKeyPressed(rl.KeyF1) {
			sim.SetSpeed(1)
		}
		if rl.IsKeyPressed(rl.KeyF2) {
			sim.SetSpeed(2)
		}
		if rl.IsKeyPressed(rl.KeyF3) {
			sim.SetSpeed(3)
		}

		// Road elevation control
		if rl.IsKeyPressed(rl.KeyPageUp) {
			if gameUI.Selected == ui.ToolRoad {
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
			if gameUI.Selected == ui.ToolRoad {
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

		// Mouse ray → terrain surface
		worldX := float32(0)
		worldZ := float32(0)
		mouseOnTerrain := false
		ray := rl.GetScreenToWorldRay(rl.GetMousePosition(), cam)
		worldX, worldZ, mouseOnTerrain = sim.Heightmap.PickXZ(ray)

		// Update UI state
		gameUI.Money = sim.Money
		gameUI.TimeStr = sim.Time.TimeString()
		if sim.Time.IsPaused {
			gameUI.TimeStr += " ⏸"
		} else if sim.Time.Speed > 1 {
			speedStr := fmt.Sprintf(" ⏩x%.0f", sim.Time.Speed)
			gameUI.TimeStr += speedStr
		}
		gameUI.MouseWorldX = worldX
		gameUI.MouseWorldZ = worldZ
		gameUI.MouseOnGround = mouseOnTerrain

		// Snap to outside connection for preview
		previewX, previewZ := worldX, worldZ
		if mouseOnTerrain && gameUI.Selected == ui.ToolRoad {
			for _, c := range sim.Connections.GetByType(terrain.ConnHighway) {
				dx := c.WorldX - worldX
				dz := c.WorldZ - worldZ
				if dx*dx+dz*dz < 64 {
					for idx := range sim.Roads.Nodes {
						n := &sim.Roads.Nodes[idx]
						if n.Flags&road.RoadFlagOutsideConn != 0 {
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
		gameUI.HandleInput()

		// Handle mouse clicks for tools and toolbar
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			mPos := rl.GetMousePosition()
			my := int32(mPos.Y)
			if my >= ui.ToolbarY {
				gameUI.HandleClick()
				roadActive = false
			} else if my >= ui.ToolbarY-ui.OptionsBarH && gameUI.HasOptionsBar() {
				// option bar handles its own clicks during Draw()
			} else if mouseOnTerrain {
				switch gameUI.Selected {
				case ui.ToolRoad:
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
							newNode, _, ok := sim.PlaceRoadSegment(roadStartNode, px, pz, gameUI.RoadType, roadElevation)
							if ok {
								roadStartNode = newNode
							}
							if !ok {
								gameUI.HelpText = "Cannot place road: check terrain slope, water, or obstacles"
							}
						}
					}
				case ui.ToolParking:
				cost := float32(1000)
				if gameUI.ParkingGarage {
					cost = 3000
				}
				if gameUI.BusDepotMode || gameUI.TramDepotMode || gameUI.MetroDepotMode || gameUI.FerryDepotMode || gameUI.MonorailDepotMode || gameUI.CableCarDepotMode || gameUI.TaxiDepotMode {
					cost = 5000
				}
				if gameUI.AirportMode {
					cost = 10000
				}
				if gameUI.PortMode {
					cost = 8000
				}
				if sim.Money >= cost && !sim.Heightmap.IsUnderwater(worldX, worldZ) {
					switch {
					case gameUI.BusDepotMode:
						sim.PlaceBusDepot(worldX, worldZ)
					case gameUI.TramDepotMode:
						sim.PlaceTramDepot(worldX, worldZ)
					case gameUI.MetroDepotMode:
						sim.PlaceMetroDepot(worldX, worldZ)
					case gameUI.FerryDepotMode:
						sim.PlaceFerryDepot(worldX, worldZ)
					case gameUI.MonorailDepotMode:
						sim.PlaceMonorailDepot(worldX, worldZ)
					case gameUI.CableCarDepotMode:
						sim.PlaceCableCarDepot(worldX, worldZ)
					case gameUI.TaxiDepotMode:
						sim.PlaceTaxiDepot(worldX, worldZ)
					case gameUI.AirportMode:
						sim.PlaceAirportDepot(worldX, worldZ)
					case gameUI.PortMode:
						sim.PlacePortDepot(worldX, worldZ)
					default:
						sim.PlaceParkingLot(worldX, worldZ, gameUI.ParkingGarage)
					}
				}
			case ui.ToolTransport:
				if gameUI.CargoMode {
					if sim.Money >= 5000 {
						cost := float32(5000)
						if sim.Money >= cost {
							sim.Money -= cost
							sim.Transport.Cargo.AddStation(previewX, previewZ)
						}
					}
					break
				}
				if !sim.Heightmap.IsUnderwater(worldX, worldZ) {
					if !transportActive {
						stopID := sim.Transport.AddStop(previewX, previewZ, gameUI.TransportType)
						transportStartStopID = stopID
						transportActive = true
					} else {
						stopID := sim.Transport.AddStop(previewX, previewZ, gameUI.TransportType)
						if transportLineID == 0 {
							col := transport.TransportStopColor(gameUI.TransportType)
							transportLineID = sim.Transport.AddLine("Line", gameUI.TransportType, []uint32{transportStartStopID, stopID}, col)
						} else {
							sim.Transport.AddStopToLine(transportLineID, stopID)
						}
						transportStartStopID = stopID
					}
				}
				case ui.ToolRemove:
					if sim.Transport != nil {
						stop := sim.Transport.NearestStop(worldX, worldZ, 8)
						if stop != nil {
							sim.Transport.RemoveStop(stop.ID)
							break
						}
						line := sim.Transport.NearestLine(worldX, worldZ, 12)
						if line != nil {
							sim.Transport.RemoveLine(line.ID)
							break
						}
					}
					idx := sim.Roads.NearestSegment(worldX, worldZ)
					if idx >= 0 {
						sim.RemoveSegment(idx)
					}
					sim.RemoveTrees(worldX, worldZ, 10)
				case ui.ToolUpgrade:
					idx := sim.Roads.NearestSegment(worldX, worldZ)
					if idx >= 0 {
						if !sim.UpgradeSegment(idx, gameUI.RoadType) {
							gameUI.HelpText = "Cannot upgrade: insufficient funds or already that type"
						}
					}
				}
			}
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			roadActive = false
			transportActive = false
			transportLineID = 0
		}

		// --- Draw ---
		rl.BeginDrawing()
		rl.ClearBackground(skyColor(sim.Night))
		rl.BeginMode3D(cam)
		t.Draw(cam.Position.X, cam.Position.Z)

		// Draw previews
		if mouseOnTerrain {
			h := sim.Heightmap.WorldHeight(worldX, worldZ)

			switch gameUI.Selected {
			case ui.ToolRoad:
				rl.DrawSphere(rl.NewVector3(previewX, h+0.5, previewZ), 0.8, rl.Green)
				if roadActive {
					sn := &sim.Roads.Nodes[roadStartNode]
					sh := sim.Heightmap.WorldHeight(sn.X, sn.Z)
					rl.DrawLine3D(rl.NewVector3(sn.X, sh+0.2, sn.Z), rl.NewVector3(previewX, h+0.2, previewZ), rl.Green)
				}
			case ui.ToolParking:
				if gameUI.AirportMode {
					rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 12, 0.3, 8, rl.NewColor(180, 180, 180, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.3, worldZ), 12, 0.3, 8, rl.NewColor(180, 180, 180, 200))
					rl.DrawCube(rl.NewVector3(worldX+4, h+1.5, worldZ), 4, 3, 3, rl.NewColor(200, 200, 220, 120))
					rl.DrawCube(rl.NewVector3(worldX-5, h+0.2, worldZ), 20, 0.1, 2, rl.NewColor(160, 160, 160, 150))
					rl.DrawCube(rl.NewVector3(worldX-5, h+0.2, worldZ+3), 20, 0.1, 2, rl.NewColor(160, 160, 160, 150))
				} else if gameUI.PortMode {
					rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 8, 0.5, 10, rl.NewColor(139, 90, 43, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.3, worldZ), 8, 0.5, 10, rl.NewColor(139, 90, 43, 200))
					rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ+5), 6, 0.3, 4, rl.NewColor(100, 100, 100, 120))
				} else if gameUI.FerryDepotMode {
					rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 5, 0.5, 5, rl.NewColor(50, 150, 200, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.3, worldZ), 5, 0.5, 5, rl.NewColor(50, 150, 200, 200))
				} else if gameUI.MonorailDepotMode {
					rl.DrawCube(rl.NewVector3(worldX, h+2, worldZ), 8, 2, 4, rl.NewColor(100, 200, 200, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+2, worldZ), 8, 2, 4, rl.NewColor(100, 200, 200, 200))
				} else if gameUI.CableCarDepotMode {
					rl.DrawCube(rl.NewVector3(worldX, h+2, worldZ), 4, 3, 4, rl.NewColor(200, 200, 100, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+2, worldZ), 4, 3, 4, rl.NewColor(200, 200, 100, 200))
				} else if gameUI.TaxiDepotMode {
					rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ), 5, 1, 4, rl.NewColor(200, 200, 50, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.5, worldZ), 5, 1, 4, rl.NewColor(200, 200, 50, 200))
				} else if gameUI.MetroDepotMode {
					rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ), 6, 1, 4, rl.NewColor(80, 80, 200, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.5, worldZ), 6, 1, 4, rl.NewColor(80, 80, 200, 200))
				} else if gameUI.TramDepotMode {
					rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ), 6, 1, 4, rl.NewColor(180, 50, 180, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.5, worldZ), 6, 1, 4, rl.NewColor(180, 50, 180, 200))
				} else if gameUI.BusDepotMode {
					rl.DrawCube(rl.NewVector3(worldX, h+0.5, worldZ), 6, 1, 4, rl.NewColor(200, 180, 50, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.5, worldZ), 6, 1, 4, rl.NewColor(200, 180, 50, 200))
				} else {
					col := rl.NewColor(80, 80, 200, 100)
					if !gameUI.ParkingGarage {
						col = rl.NewColor(80, 160, 80, 80)
					}
					rl.DrawCube(rl.NewVector3(worldX, h+0.3, worldZ), 20, 0.3, 15, col)
					rl.DrawCubeWires(rl.NewVector3(worldX, h+0.3, worldZ), 20, 0.3, 15, rl.NewColor(60, 60, 100, 150))
				}
			case ui.ToolTransport:
				if gameUI.CargoMode {
					rl.DrawCube(rl.NewVector3(worldX, h+1, worldZ), 5, 2, 5, rl.NewColor(200, 150, 50, 120))
					rl.DrawCubeWires(rl.NewVector3(worldX, h+1, worldZ), 5, 2, 5, rl.NewColor(200, 200, 100, 200))
					break
				}
				stopCol := transport.TransportStopColor(gameUI.TransportType)
				if transportActive {
					if sn := sim.Transport.StopByID(transportStartStopID); sn != nil {
						sh := sim.Heightmap.WorldHeight(sn.X, sn.Z)
						rl.DrawLine3D(rl.NewVector3(sn.X, sh+0.5, sn.Z), rl.NewVector3(previewX, h+0.5, previewZ), stopCol)
					}
				}
				rl.DrawSphere(rl.NewVector3(previewX, h+0.5, previewZ), 0.6, stopCol)
			case ui.ToolRemove:
				if sim.Transport != nil {
					stop := sim.Transport.NearestStop(worldX, worldZ, 8)
					if stop != nil {
						sh := sim.Heightmap.WorldHeight(stop.X, stop.Z) + 0.5
						rl.DrawCube(rl.NewVector3(stop.X, sh, stop.Z), 2, 2, 2, rl.NewColor(200, 50, 50, 80))
						rl.DrawCubeWires(rl.NewVector3(stop.X, sh, stop.Z), 2, 2, 2, rl.Red)
						break
					}
					line := sim.Transport.NearestLine(worldX, worldZ, 12)
					if line != nil {
						col := transport.TransportStopColor(line.TransType)
						for _, sid := range line.Stops {
							s := sim.Transport.StopByID(sid)
							if s == nil {
								continue
							}
							sh := sim.Heightmap.WorldHeight(s.X, s.Z) + 0.5
							rl.DrawCubeWires(rl.NewVector3(s.X, sh, s.Z), 1.5, 1.5, 1.5, col)
						}
						break
					}
				}
				idx := sim.Roads.NearestSegment(worldX, worldZ)
				if idx >= 0 {
					seg := sim.Roads.Segments[idx]
					na := &sim.Roads.Nodes[seg.NodeA]
					nb := &sim.Roads.Nodes[seg.NodeB]
					ha := sim.Heightmap.WorldHeight(na.X, na.Z) + 0.5
					hb := sim.Heightmap.WorldHeight(nb.X, nb.Z) + 0.5
					rl.DrawLine3D(rl.NewVector3(na.X, ha, na.Z), rl.NewVector3(nb.X, hb, nb.Z), rl.Red)
				}
			case ui.ToolUpgrade:
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

		gameUI.Draw()
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
