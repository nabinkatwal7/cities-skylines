# Bullet points for anchored summary

## Entity system
- entity.go: Entity struct (ID, Position, Rotation, Bounds, Flags, Owner, LODLevel), helpers (SetPosition, SetRotationY, HasFlag, SetFlag, ClearFlag, UpdateBounds).
- Building embeds Entity → removed X,Z,Abandoned,HasRoad,Constructed fields. Uses Entity.Position and flag methods (FlagAbandoned, FlagHasRoad, FlagConstructed, FlagRemoved).
- Vehicle embeds Entity → removed X,Z,Parked fields. Uses Entity.Position and FlagParked.
- core_entitypool.go interface renamed Entity → EntityI to avoid conflict with Entity struct.

## Districts
- districts.go: DistrictManager (New, Add/Remove/Apply policies, DistrictAt). 6 policy types (HighRiseBan, HeavyTrafficBan, SelfSufficient, ITCluster, OrganicProduce, BigBusiness). Radial circular rendering.

## Simulation
- simulation.go: SimulationManager owns 13 sub-managers, controlled mutation methods (PlaceRoadNode, PlaceRoadSegment, RemoveSegment, UpgradeSegment, SetZone, PlacePark, SetNight), wires Scheduler (4 groups) + EventBus (12 event types inside terrain/).

## Refactored
- manager.go: stripped to pure renderer (Chunks, Models, terrainTex, uploadIdx), holds read-only *SimulationManager pointer for Draw() — no Update(), no GenerateData().
- terraforming.go: decoupled from Manager — takes Heightmap, WaterSystem, []Chunk, rebuildChunk func.
- serialization.go: SaveGame/LoadGame changed from *Manager to *SimulationManager. Building serialization updated for Entity-based fields.

## Deleted
- internal/engine/ package removed. All 8 core files (eventbus, scheduler, timesystem, entitypool, lifecycle, quadtree, world, jobqueue) moved to internal/terrain/ as core_*.go.

## Removed (main.go)
- Direct .Roads.AddNode/AddSegment calls → sim.PlaceRoadNode/PlaceRoadSegment.
- sim.Money direct writes → money mutations inside SimulationManager methods.
- sim.Roads.Rebuild calls → handled by RemoveSegment/UpgradeSegment.
- sim.Night direct set → uses sim.SetNight().
- engine/ import.

## Fixed bugs
- TotalPowerUsed/TotalWaterUsed/TotalGarbage/TotalWealth/TotalHappiness reset each cycle.
- roadsNearby stub → b.HasFlag(FlagHasRoad) in landValue.
- HouseholdInfo/BusinessInfo serialized to BuildingData.
- calcDemand: office demand tracked, minimum floor of 1, office no longer always-allowed.

## Road elevation
- Elevated roads (Elevation 1/2) use fixed absolute Y (5/10) independent of terrain height.
- Ground roads still follow terrain with 0.15 offset.
- Junction markers check connected segments for elevation.
- buildSurfaceMesh, drawFallback, drawMarkings all updated to use absolute Y for elevated roads.
