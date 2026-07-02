# Architecture

Feature-based packages matching `implemented-features.md` sections **1–4**.

## Package layout

```
cmd/js-skylines/          entry point, camera, input
internal/
  core/                   entity lifecycle, event bus, scheduler, jobs, game time
  terrain/                heightmap, generation, water, trees, resources, terraform, connections, buildability, chunks
  road/                   road graph, vehicles, parking (spots, lots, depots)
  transport/              public transport, rail, cargo
  sim/                    SimulationManager orchestration + world renderer (Manager)
  save/                   serialization / load-save
  ui/                     toolbar, HUD, font
```

## Import DAG (acyclic)

```
core
  ↑
terrain
  ↑
road ──────────────► transport
  ↑                      ↑
  └──────── sim ─────────┘
              ↑
         save, ui, cmd
```

- `road` does **not** import `transport`. Depots use `road.TransportCoordinator` implemented by `transport.Manager`.
- `terrain.BuildabilityChecker` uses `RoadAccess` interface implemented by `road.RoadManager`.

## ID conventions

| Field | Stores | Lookup |
|-------|--------|--------|
| `RoadNode.Connected` | segment **IDs** | `road.RoadManager.SegmentByID(id)` |
| `Vehicle.RoadSeg` | segment **ID** | `SegmentByID` |
| `TransportStop` / line stops | stop **IDs** | `transport.Manager.StopByID(id)` |
| `AddNode()` return value | node **index** | `Nodes[idx]` |

## Simulation flow

1. `sim.NewSimulationManager(seed)` wires terrain + road + transport + core infra
2. Each frame: `sim.Update(dt)` runs scheduler groups (water, trees, roads, vehicles, transport, parking, tax)
3. `sim.Manager.Draw()` renders chunks, terrain features, roads, vehicles, parking, transport

## Tests

```bash
go test ./...
```

- `internal/core` — event bus + event names
- `internal/sim` — smoke tests (heightmap pick, traffic rules, segment lookup)
