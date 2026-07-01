## 1.1 Design Philosophy

The simulation engine is the authoritative source of truth for the entire game. Rendering, user interface, audio, animations, visual effects, camera movement, and editor tools never modify gameplay data directly. Every gameplay system operates on centralized simulation data managed by dedicated subsystem managers.

The engine follows a deterministic simulation model where every update produces the same results when given identical input and game state. This ensures reliable save/load behavior, reproducible simulations, and consistent interactions between all gameplay systems.

The architecture follows four fundamental principles:

- Simulation First
- Data-Oriented Design
- Manager-Based Systems
- Event-Driven Communication

Unlike traditional object-oriented architectures where each object contains significant behavior, the majority of gameplay logic is centralized into managers responsible for processing large batches of entities efficiently.

```text
SimulationManager

├── TerrainManager
├── RoadManager
├── ZoneManager
├── BuildingManager
├── CitizenManager
├── VehicleManager
├── TransportManager
├── UtilityManager
├── EconomyManager
├── DistrictManager
├── DisasterManager
├── WeatherManager
├── StatisticsManager
├── UnlockManager
└── EventManager
```

Each manager owns all simulation logic for its corresponding entity type and exposes controlled interfaces to the rest of the engine.

---

## 1.2 Simulation Layers

The engine is divided into independent layers. Each layer has a single responsibility and communicates with adjacent layers through well-defined interfaces.

```text
+-------------------------------------------------------+
| Presentation Layer                                    |
|-------------------------------------------------------|
| UI                                                    |
| Audio                                                 |
| Camera                                                |
| Effects                                               |
| Animation                                             |
+-------------------------------------------------------+

+-------------------------------------------------------+
| Rendering Layer                                       |
|-------------------------------------------------------|
| Terrain                                               |
| Buildings                                             |
| Roads                                                 |
| Vehicles                                              |
| Citizens                                              |
| Trees                                                 |
| Props                                                  |
| Water                                                 |
| Sky                                                   |
+-------------------------------------------------------+

+-------------------------------------------------------+
| Simulation Layer                                      |
|-------------------------------------------------------|
| Economy                                               |
| Utilities                                             |
| Traffic                                               |
| Citizens                                              |
| Buildings                                             |
| Services                                              |
| Districts                                             |
| Weather                                               |
+-------------------------------------------------------+

+-------------------------------------------------------+
| Core Systems                                          |
|-------------------------------------------------------|
| Scheduler                                             |
| Entity Registry                                       |
| Serialization                                         |
| Save System                                           |
| Pathfinding                                           |
| Job System                                            |
| Resource Loader                                       |
| Event Bus                                             |
+-------------------------------------------------------+
```

Rendering never owns gameplay state.

Simulation never depends on rendering.

The UI reads simulation data but cannot modify simulation state without issuing validated commands through the simulation layer.

---

## 1.3 World Representation

The world exists as a continuous coordinate system rather than a strict tile grid.

Roads, buildings, trees, props, utility networks, and moving agents occupy precise floating-point world positions.

Each simulation entity contains a lightweight data record.

```cpp
struct Entity
{
    uint32 id;
    Vector3 position;
    Quaternion rotation;
    BoundingBox bounds;
    uint32 flags;
    uint16 owner;
    uint16 lodLevel;
}
```

Example Building

```text
Entity ID
42151

Type
Commercial Level 3

World Position
(1042.48, 28.34, 938.16)

Rotation
180°

Flags
Powered
Watered
Connected
Occupied

Owner
BuildingManager
```

Road geometry is defined independently from terrain geometry.

Road nodes represent intersections.

Road segments connect nodes using spline curves.

Buildings align to generated roadside lots instead of rigid tiles.

Terrain height is sampled independently of road elevation, allowing bridges, tunnels, embankments, elevated highways, and cuttings.

---

## 1.4 Manager Architecture ✅

Each gameplay system owns a dedicated manager responsible for lifecycle management, updates, serialization, statistics, and communication.

Example:

```text
BuildingManager

Responsibilities

• Create buildings
• Destroy buildings
• Upgrade buildings
• Downgrade buildings
• Fire damage
• Collapse
• Occupancy
• Tax calculation
• Serialization
```

Every manager contains:

```text
Entity Pool

Free List

Update Queue

Dirty Queue

Spatial Index

Statistics Cache

Command Queue
```

Managers allocate entities using contiguous memory pools rather than scattered heap allocations.

Instead of

```cpp
new Citizen();
new Citizen();
new Citizen();
```

the engine allocates

```cpp
Citizen citizens[MAX_CITIZENS];
```

Advantages include:

- predictable memory usage
- cache-friendly iteration
- minimal fragmentation
- constant-time lookup
- SIMD optimization
- easier serialization
- reduced garbage collection pressure
- efficient multithreading

Destroyed entities return to a free list for immediate reuse without reallocating memory.

---
