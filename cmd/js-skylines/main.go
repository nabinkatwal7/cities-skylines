package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/save"
	simpkg "github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/terrain"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/ui"
	"github.com/katwate/js-skylines/internal/zoning"
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
	ui.SetScreenSize(screenWidth, screenHeight)
	ui.LoadUIFont()
	defer ui.UnloadUIFont()
	rl.SetTargetFPS(60)

	if err := t.LoadAssets(); err != nil {
		fmt.Printf("Warning: could not load assets: %v\n", err)
	}
	t.PrepareUpload()

	gameUI := ui.NewManager()
	gameUI.Attach(sim.EventBus)
	_ = gameUI.LoadPreferences(ui.DefaultUIStateFile)

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
			_ = gameUI.SavePreferences(ui.DefaultUIStateFile)
		}

		dt := float32(rl.GetFrameTime())
		cam := gameUI.UpdateCamera(dt)

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

		// Update UI state (presentation only)
		gameUI.SyncView(ui.ViewStateFromSim(sim, worldX, worldZ, mouseOnTerrain))

		// Snap to build grid / zone cells
		if roadActive && !sim.Roads.ValidNodeIndex(roadStartNode) {
			roadActive = false
			roadStartNode = 0
		}
		gameUI.SetRoadChain(roadActive, roadStartNode)
		previewX, previewZ := worldX, worldZ
		if mouseOnTerrain {
			previewX, previewZ = gameUI.SnapPlacementCoords(sim, worldX, worldZ)
		}
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

		gameUI.UpdateWorld(ui.WorldContext{
			Sim:                sim,
			WorldX:             worldX,
			WorldZ:             worldZ,
			PreviewX:           previewX,
			PreviewZ:           previewZ,
			OnGround:           mouseOnTerrain,
			RoadActive:         roadActive,
			RoadStartNode:      roadStartNode,
			RoadElevation:      roadElevation,
			TransportActive:    transportActive,
			TransportStartStop: transportStartStopID,
		})

		// Handle keyboard tool selection
		gameUI.HandleInput()

		// Handle mouse clicks for tools and toolbar
		placeClick := rl.IsMouseButtonPressed(rl.MouseButtonLeft)
		zoneDrag := gameUI.Selected == ui.ToolZone && rl.IsMouseButtonDown(rl.MouseButtonLeft)
		if placeClick || zoneDrag {
			mPos := rl.GetMousePosition()
			mx := int32(mPos.X)
			my := int32(mPos.Y)
			if gameUI.PointerOverUI(mx, my) {
				if placeClick {
					gameUI.HandleClick()
					if gameUI.ClickResetsRoadChain(mx, my) {
						roadActive = false
					}
				}
			} else if mouseOnTerrain {
				worldClick := placeClick && gameUI.HandleWorldClick(sim, worldX, worldZ)
				if !worldClick {
				pv := gameUI.Preview()
				switch gameUI.Selected {
				case ui.ToolRoad:
					px := clamp(previewX, -240, 240)
					pz := clamp(previewZ, -240, 240)
					if !roadActive {
						if !pv.TerrainOK || sim.Money < pv.Cost {
							break
						}
						roadStartNode = sim.PlaceRoadNode(px, pz)
						roadActive = true
					} else {
						if !pv.Valid {
							if len(pv.Messages) > 0 {
								gameUI.HelpText = pv.Messages[0]
							} else {
								gameUI.HelpText = "Cannot place road: check terrain slope, water, or obstacles"
							}
							break
						}
						moneyBefore := sim.Money
						newNode, segID, ok := sim.PlaceRoadSegment(roadStartNode, px, pz, gameUI.RoadType, roadElevation)
						if ok {
							roadStartNode = newNode
							sim.PushRoadPlace(segID, moneyBefore-sim.Money)
							gameUI.HelpText = ""
						}
					}
				case ui.ToolParking:
					if !pv.Valid {
						break
					}
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
				case ui.ToolTransport:
				if gameUI.CargoMode {
					if !pv.Valid {
						break
					}
					sim.Money -= pv.Cost
					sim.Transport.Cargo.AddStation(previewX, previewZ)
					break
				}
				if !pv.Valid {
					break
				}
				sim.Money -= pv.Cost
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
				case ui.ToolZone:
					zt := zoning.ZoneType(gameUI.ZoneType + 1)
					px, pz := previewX, previewZ
					if sim.Zones != nil {
						cx := sim.Zones.CellX(px)
						cz := sim.Zones.CellZ(pz)
						before := sim.Zones.Cells[cz][cx]
						if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
							sim.Zones.RemoveZone(px, pz)
							sim.EventBus.Emit(string(core.EventZoneRemoved), nil)
							sim.PushZoneChange(cx, cz, before)
						} else if pv.Valid {
							sim.Zones.SetZone(px, pz, zt)
							sim.EventBus.Emit(string(core.EventZonePlaced), zt)
							sim.PushZoneChange(cx, cz, before)
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
						sim.PushRoadRemove(idx)
					} else {
						sim.PushTreeRemove(worldX, worldZ, 10)
					}
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
		gameUI.DrawWorldOverlays(sim)

		// Draw placement preview (24.6)
		if mouseOnTerrain {
			gameUI.DrawWorldPreview()
		}
		gameUI.DrawBuildGuides(cam.Position.X, cam.Position.Z)
		rl.EndMode3D()

		gameUI.Draw()
		if gameUI.Shortcuts.ConsumeScreenshot() {
			rl.TakeScreenshot("screenshot.png")
		}
		rl.DrawFPS(10, 30)

		rl.EndDrawing()
	}

	t.Unload()
	_ = gameUI.SavePreferences(ui.DefaultUIStateFile)
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
