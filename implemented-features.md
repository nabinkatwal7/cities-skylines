# 1. Core Engine Architecture

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

## 1.5 Simulation Scheduler ✅

Unlike frame-based games, Cities: Skylines separates rendering from simulation.

Rendering attempts to maintain the highest possible framerate.

Simulation advances independently according to scheduled update groups.

```text
Frame

↓

Input Processing

↓

Simulation Scheduler

↓

Fast Updates

↓

Medium Updates

↓

Slow Updates

↓

Very Slow Updates

↓

Renderer
```

Simulation time advances even if rendering slows down.

Likewise, rendering can exceed simulation frequency without affecting gameplay.

---

### Fast Update Group

Runs every simulation frame.

Responsible for:

- vehicle steering
- lane changing
- citizen walking
- traffic lights
- intersections
- train movement
- aircraft movement
- ship movement
- animations
- pedestrian crossings

Budget

```
1–3 ms
```

---

### Medium Update Group

Runs several times per second.

Responsible for:

- power propagation
- water distribution
- sewage flow
- garbage routing
- ambulance dispatch
- fire propagation
- police dispatch
- service requests
- district updates

Budget

```
2–5 ms
```

---

### Slow Update Group

Runs approximately once every in-game hour.

Responsible for:

- tax collection
- building upgrades
- employment calculations
- education progression
- happiness updates
- land value recalculation
- production chains
- resource balancing

Budget

```
5–10 ms
```

---

### Very Slow Update Group

Runs once every in-game day.

Responsible for:

- immigration
- emigration
- births
- deaths
- milestone evaluation
- tourism generation
- economic reports
- statistics generation
- achievement checks

---

## Scheduler Priorities

Each simulation task belongs to a priority class.

```text
CRITICAL

Road Connectivity
Vehicle Safety
Citizen State
Pathfinding

HIGH

Utilities
Emergency Services
Transport
Fire

MEDIUM

Economy
Education
Healthcare
Garbage
Crime

LOW

Statistics
Heatmaps
Achievements
UI Updates
Chirper
```

When the simulation exceeds its frame budget, lower-priority tasks are postponed to future updates while critical systems continue running uninterrupted.

---

## Deterministic Update Order

Every simulation frame follows the same execution order.

```text
1. Read Player Commands

2. Process Construction

3. Update Road Network

4. Update Utility Networks

5. Update Public Services

6. Update Buildings

7. Update Citizens

8. Update Vehicles

9. Update Public Transport

10. Update Economy

11. Update District Policies

12. Update Weather

13. Update Statistics

14. Queue Events

15. Render Frame
```

Maintaining a fixed execution order prevents race conditions and guarantees consistent simulation outcomes across save/load cycles.

---

## Simulation Performance Budget

To maintain responsiveness on large cities, each subsystem receives a fixed processing budget.

| System               |          Target Budget |
| -------------------- | ---------------------: |
| Traffic              |                   3 ms |
| Citizens             |                   2 ms |
| Buildings            |                   1 ms |
| Utilities            |                   1 ms |
| Economy              |                   1 ms |
| Public Services      |                   2 ms |
| Transport            |                   2 ms |
| Weather              |                 0.5 ms |
| Statistics           |                 0.5 ms |
| Rendering Submission | Remaining Frame Budget |

If a subsystem exceeds its allocation, remaining work is deferred using incremental processing queues rather than blocking the main thread.

---

## Core Design Goals

The engine architecture is designed to support:

- Cities containing hundreds of thousands of simulated citizens.
- Tens of thousands of buildings.
- Thousands of simultaneously active vehicles.
- Fully simulated utility networks.
- Deterministic save/load behavior.
- High-performance pathfinding.
- Incremental simulation updates.
- Extensible manager-based gameplay systems.
- Efficient multithreading.
- Large-scale modding support.
- Long-running simulations without memory fragmentation.
- Consistent gameplay regardless of rendering performance.

This architecture forms the foundation upon which every subsequent gameplay system—including zoning, traffic, public transport, economy, citizen AI, utilities, and disasters—is built.

## 1.6 Entity Lifecycle ✅

Every object in the simulation follows a deterministic lifecycle managed by its owning subsystem. No entity is created or destroyed directly by gameplay systems. Instead, all requests are routed through the Simulation Manager to ensure thread safety, deterministic execution, and save-game consistency.

The lifecycle applies to all entity types:

- Buildings
- Citizens (Cims)
- Vehicles
- Trees
- Props
- Road Nodes
- Road Segments
- Utility Networks
- Districts
- Service Requests
- Public Transport Lines
- Disaster Objects

Every entity transitions through the following states:

```text
Unallocated
      │
      ▼
Allocated
      │
      ▼
Initializing
      │
      ▼
Active
      │
      ├──────────────┐
      ▼              │
Suspended            │
      │              │
      ▼              │
Reactivated          │
      │              │
      ▼              │
Marked For Removal ◄─┘
      │
      ▼
Destroyed
      │
      ▼
Returned To Pool
```

---

### Allocation

When the game requires a new entity, the owning manager retrieves an unused slot from its memory pool.

Example:

```cpp
CitizenID id = CitizenManager::Allocate();
```

The entity receives:

- Unique Entity ID
- Creation Timestamp
- Default Flags
- Owner Manager
- Initial Position
- Simulation Version
- Dirty State

No expensive heap allocations occur during gameplay.

---

### Initialization

Initialization validates the entity before it becomes active.

For example, a newly zoned building checks:

- Road access
- Power availability
- Water availability
- Valid terrain
- Lot size
- Collision
- District assignment

If validation fails, initialization is cancelled.

---

### Active State

Once active, the entity participates in the simulation.

Example:

Citizen

Daily schedule

Health

Employment

Pathfinding

Vehicle ownership

Household

Building

Occupancy

Tax generation

Land value

Pollution

Power consumption

Water consumption

Vehicle

Current path

Lane position

Speed

Destination

Fuel type

Service state

---

### Suspension

Inactive entities can enter a suspended state.

Examples include:

- Vehicles outside the active simulation radius
- Decorative props
- Buildings awaiting construction
- Tourist groups waiting to spawn

Suspended entities retain data but skip expensive updates.

---

### Destruction

Entities are never immediately deleted.

Instead they enter a pending destruction queue.

```text
Destroy Request

↓

Cleanup

↓

Notify Managers

↓

Release References

↓

Return Memory Slot
```

This prevents invalid references during update loops.

---

## 1.7 Game Time & Simulation Speed

Game time is independent from real time.

The engine maintains two clocks.

```text
Real Time

↓

Frame Time

↓

Simulation Time

↓

Calendar
```

Simulation time drives:

- Citizen schedules
- Building production
- Taxes
- Utility consumption
- Traffic demand
- Day/Night cycle
- Weather
- Policies
- Economy

---

### Calendar

The simulation tracks:

```text
Minute

Hour

Day

Week

Month

Year
```

Each unit advances deterministically.

Example

```text
08:00

Morning Rush Begins

↓

12:00

Lunch Activity

↓

17:00

Evening Rush

↓

23:00

Night Economy

↓

02:00

Service Vehicles Peak
```

---

### Simulation Speeds

Players may select:

```text
Paused

1×

2×

3×
```

Changing speed modifies simulation frequency instead of animation playback.

Rendering remains independent.

---

### Paused State

When paused:

Simulation stops.

Rendering continues.

Camera continues.

UI continues.

Menus remain interactive.

Save operations remain available.

---

### Time-Based Events

Every system subscribes to time events.

Examples:

Every Minute

- Vehicle updates
- Citizen schedule checks

Every Hour

- Taxes
- Production
- Happiness

Every Day

- Births
- Deaths
- Immigration
- Tourism

Every Week

- Loan payments
- Budget summaries

---

## 1.8 Event Bus & Messaging System ✅

Subsystems never communicate directly.

Instead they publish and subscribe to events.

```text
Fire Starts

↓

Event Bus

↓

Fire Manager

↓

Building Manager

↓

Citizen Manager

↓

Vehicle Manager

↓

UI
```

This prevents tight coupling.

---

### Event Structure

Every event contains:

```cpp
EventID

Timestamp

Priority

Source

Target

Payload

Flags
```

---

### Example

Building catches fire.

```text
Building

↓

Fire Event

↓

Event Queue

↓

Fire Department

↓

Dispatch Engine

↓

Citizen Notification

↓

UI Alert

↓

Statistics Update
```

---

### Priority Levels

```text
Critical

High

Normal

Low
```

Critical events:

- Building collapse
- Fire
- Flood
- Death
- Road disconnect

High:

- Vehicle accidents
- Crime
- Garbage overflow

Normal:

- Taxes
- Education
- Happiness

Low:

- Statistics
- Heatmaps
- Achievements

---

### Event Queue

Events are processed in FIFO order within their priority class.

```text
Critical Queue

↓

High Queue

↓

Normal Queue

↓

Low Queue
```

This guarantees emergency services always react before cosmetic systems update.

---

## 1.9 Serialization & Save Format

Every piece of simulation data is serializable.

Rendering assets are never stored.

Instead only gameplay state is written.

```text
Terrain

Roads

Buildings

Citizens

Vehicles

Economy

Policies

Weather

Statistics

Mods
```

---

### Save Layout

```text
Header

Version

Checksum

Mods

Terrain

Road Network

Zones

Buildings

Citizens

Vehicles

Districts

Utilities

Transport

Economy

Weather

Statistics

Achievements
```

---

### Entity Serialization

Each entity stores:

```cpp
ID

Position

Rotation

Flags

Owner

Simulation Data

Version
```

Pointers are never serialized.

Instead references are stored as IDs.

Example

Instead of

```cpp
Citizen*

Vehicle*
```

the save stores

```cpp
CitizenID

VehicleID
```

Upon loading, managers rebuild references.

---

### Incremental Saves

To reduce save times the engine supports dirty tracking.

Only modified entities are rewritten.

```text
Building Modified

↓

Dirty Flag

↓

Save Queue

↓

Serialize

↓

Clean Flag
```

---

### Version Compatibility

Each save contains:

```text
Game Version

Asset Version

Simulation Version

Mod Version
```

Older saves migrate through upgrade pipelines.

---

## 1.10 Job System & Background Processing ✅

Cities contain millions of calculations each simulation hour.

The engine distributes work across worker threads.

```text
Main Thread

↓

Job Scheduler

↓

Worker Pool

↓

Completed Jobs

↓

Simulation
```

---

### Typical Background Jobs

Traffic pathfinding

Electricity propagation

Water simulation

Land value updates

Noise calculations

Pollution diffusion

District analysis

Statistics generation

Heatmap generation

Tree growth

Vehicle routing

Citizen destination search

---

### Job Lifecycle

```text
Create Job

↓

Schedule

↓

Execute

↓

Validate

↓

Commit Result
```

Workers never modify simulation state directly.

Instead they produce immutable results.

The Simulation Manager applies completed jobs during synchronization points.

---

### Dependency Graph

Some jobs depend on others.

Example

```text
Road Update

↓

Lane Graph

↓

Pathfinding

↓

Vehicle Routing

↓

Traffic AI
```

A dependent job cannot execute until its prerequisite finishes.

---

### Job Priorities

```text
Critical

Pathfinding

Utilities

Vehicle Routing

High

Economy

Land Value

Services

Medium

Statistics

District Analysis

Low

Heatmaps

Achievements

Analytics
```

---

### Thread Safety

Workers cannot:

- Spawn entities
- Destroy entities
- Modify managers
- Allocate IDs
- Change game time

Workers may only:

- Read immutable snapshots
- Execute calculations
- Produce result buffers

The Simulation Manager validates and commits all results on the main simulation thread, ensuring deterministic behavior and eliminating race conditions.

---

### Performance Scaling

The scheduler dynamically adjusts job batch sizes based on available CPU cores.

Example:

| CPU Threads | Worker Threads |
| ----------- | -------------: |
| 4           |              3 |
| 8           |              7 |
| 12          |             11 |
| 16          |             15 |

The main thread remains dedicated to input processing, simulation coordination, rendering submission, and synchronization while computationally intensive tasks execute in parallel.

---

These systems establish the backbone of the simulation engine. Together with Sections **1.1–1.5**, they define how every entity is created, updated, communicated, serialized, and processed efficiently, providing a scalable foundation for large cities with hundreds of thousands of citizens and thousands of active vehicles.

# 2. Terrain & Map Generation System

The terrain system is the foundation of every city. Unlike traditional tile-based city builders, the world exists as a continuous heightfield that supports terrain deformation, water simulation, road adaptation, zoning, and environmental simulation.

The terrain engine is responsible for:

- Terrain generation
- Heightmap storage
- Water simulation
- Tree distribution
- Natural resources
- Buildability
- Terraforming
- Terrain rendering
- Terrain serialization
- Collision queries

---

## 2.1 World Structure ✅

The map is represented as a square terrain composed of evenly spaced vertices.

```text
World

├── Terrain
├── Water
├── Resources
├── Trees
├── Props
├── Outside Connections
└── Simulation Grid
```

Each terrain vertex stores:

```cpp
struct TerrainVertex
{
    float height;
    float waterLevel;
    float pollution;
    float fertility;
    float ore;
    float oil;
    float forestry;
    float buildability;
}
```

Unlike roads and buildings, terrain is immutable in topology.

Only vertex heights change during terraforming.

---

## 2.2 Coordinate System ✅

The terrain uses world-space coordinates.

```text
Origin (0,0)

↓

Positive X

↓

Positive Z

↓

Height (Y)
```

Everything references terrain coordinates.

Roads sample terrain.

Buildings sample terrain.

Trees sample terrain.

Water samples terrain.

---

## 2.3 Heightmap ✅

Terrain elevation is stored as a heightmap.

```text
1025 x 1025 vertices

↓

Interpolated

↓

Continuous Surface
```

Each vertex stores only elevation.

Normals are calculated automatically.

Slope is calculated dynamically.

---

## 2.4 Terrain Chunks ✅

The terrain is divided into chunks.

```text
Terrain

├── Chunk
├── Chunk
├── Chunk
└── Chunk
```

Each chunk contains:

- Height Data
- Trees
- Props
- Collision
- LOD Mesh
- Dirty Flag

Chunks update independently.

Only modified chunks rebuild.

---

## 2.5 Terrain LOD ✅

Different terrain meshes exist depending on camera distance.

```text
LOD0

Highest Detail

↓

LOD1

↓

LOD2

↓

LOD3

Lowest Detail
```

Only nearby terrain uses maximum resolution.

Far terrain is heavily simplified.

---

## 2.6 Terrain Generation ✅

New maps are generated procedurally.

Generation stages:

```text
Random Seed

↓

Base Noise

↓

Mountain Pass

↓

River Generation

↓

Erosion

↓

Resource Placement

↓

Tree Distribution

↓

Outside Connections

↓

Spawn Tile
```

Generation is deterministic.

The same seed always produces the same terrain.

---

## 2.7 Noise Layers ✅

Terrain uses multiple blended noise maps.

Example:

```text
Continents

+

Mountains

+

Hills

+

Fine Detail

=

Final Heightmap
```

Each layer contributes different frequencies.

---

## 2.8 River Generation ✅

Rivers follow terrain gradients.

Generation:

```text
High Elevation

↓

Flow Direction

↓

Channel Carving

↓

River Width

↓

Water Simulation
```

River paths affect:

- Water flow
- Fertility
- Pollution
- Shipping
- Terrain erosion

---

## 2.9 Water Bodies ✅

Water is simulated independently of terrain.

Supported water bodies:

- Rivers
- Lakes
- Ocean
- Reservoirs
- Canals

Each water body stores:

```cpp
Level

Velocity

Direction

Pollution

Current
```

---

## 2.10 Water Simulation ✅

Water behaves as a dynamic fluid.

Every simulation step calculates:

- Height
- Pressure
- Flow
- Velocity

Example

```text
Higher Water

↓

Pressure

↓

Lower Water

↓

Flow
```

Water continuously seeks equilibrium.

---

## 2.12 Terraforming ✅

Players may alter terrain at runtime.

Supported tools:

- Raise
- Lower
- Level
- Smooth
- Flatten
- Slope
- Soften

Terraforming immediately updates:

- Collision
- Terrain mesh
- Water flow
- Buildability
- Tree placement

---

## 2.13 Buildability ✅

Every terrain cell maintains a buildability score.

Factors:

- Slope
- Water
- Existing objects
- Outside boundaries
- Roads

Example

```text
Flat

100%

↓

Gentle Hill

75%

↓

Steep Hill

25%

↓

Cliff

0%
```

---

## 2.14 Terrain Slope ✅

Slope determines construction limits.

Calculated from neighboring vertices.

Higher slopes increase:

- Construction cost
- Road difficulty
- Bridge requirements

Extreme slopes prohibit zoning.

---

## 2.15 Tree System ✅

Trees are independent simulation entities.

Each tree stores:

```cpp
Position

Species

Age

Health

Burn State
```

Trees influence:

- Forestry resources
- Land value
- Noise reduction
- Pollution absorption

---

## 2.16 Tree Growth ✅

Trees grow over time.

States:

```text
Seedling

↓

Young

↓

Mature

↓

Old

↓

Dead
```

Dead trees may decay naturally or burn.

---

## 2.17 Tree Removal ✅

Removing trees:

- Costs money
- Reduces forestry
- Decreases land value
- Reduces noise absorption

Large-scale deforestation affects the environment.

---

## 2.18 Natural Resources ✅

Resource maps exist independently from terrain.

Resources include:

- Ore
- Oil
- Fertile Land
- Forest

Each cell stores density.

```text
0

↓

255
```

Higher density yields higher production.

Resources are visible only in resource overlays.

---

## 2.20 Outside Connections ✅

The terrain contains fixed entry points.

Supported connections:

- Highway
- Rail
- Ship
- Air

Outside connections are permanent.

They define:

- Imports
- Exports
- Immigration
- Tourism

---

## 2.21 Terrain Serialization ✅

Saved terrain includes:

- Heightmap
- Water Levels
- Trees
- Resources
- Pollution
- Terraforming
- Chunk States

Rendering meshes are regenerated after loading.

Only simulation data is stored.

---

## 2.22 Performance Optimization ✅

Terrain updates are incremental.

Modified chunks are marked dirty.

```text
Terraform

↓

Dirty Chunk

↓

Rebuild Mesh

↓

Update Collision

↓

Refresh Water

↓

Render
```

Unchanged chunks remain untouched.

This allows very large maps without rebuilding the entire terrain.

---

## 2.23 Design Goals

The terrain system is designed to support:

- Continuous world coordinates
- Massive maps
- Dynamic terraforming
- Water simulation
- Resource extraction
- Road adaptation
- Environmental interaction
- High-performance rendering
- Efficient serialization
- Incremental updates
- Realistic terrain deformation
- Large-scale environmental simulation

The terrain engine serves as the foundation for roads, zoning, buildings, utilities, transportation, disasters, and every other simulation system within the city.

# 3. Road Network & Traffic Simulation

The road network is the backbone of the entire city simulation. Nearly every gameplay system—including zoning, utilities, public services, pathfinding, economy, public transportation, emergency response, and citizen AI—depends on a connected road graph.

Unlike tile-based systems, roads are represented as a graph of nodes and segments with procedurally generated geometry.

---

## 3.1 Road Network Architecture

The road system consists of two fundamental components:

```text
Road Node
    │
Road Segment
    │
Road Node
```

### Road Nodes

Nodes represent:

- Dead ends
- Intersections
- Roundabouts
- Highway junctions
- Bridge endpoints
- Tunnel entrances
- Network transitions

Each node stores:

```cpp
struct RoadNode
{
    uint32 id;
    Vector3 position;
    uint8 connectedSegments;
    TrafficLightState trafficLight;
    JunctionType junctionType;
    uint32 flags;
}
```

---

### Road Segments

Segments connect two nodes.

Each segment contains:

```cpp
Road Type

Length

Speed Limit

Lane Count

Elevation

Direction

Maintenance Cost

Construction Cost

Curve Data
```

---

## 3.2 Supported Road Types

The engine supports multiple predefined road templates.

Examples include:

- Two-Lane Road
- One-Way Road
- Four-Lane Road
- Six-Lane Road
- Avenue
- Highway
- Gravel Road
- Bus Road
- Tram Road
- Bicycle Road
- Tree-Lined Road
- Asymmetric Road
- Pedestrian Street
- Quay Road
- Tunnel
- Elevated Road
- Bridge

Each road asset defines:

```text
Lane Layout

Speed

Width

Allowed Vehicles

Decoration

Lighting

Sidewalk Width

Noise

Maintenance

Construction Cost
```

---

## 3.3 Lane System

Every road segment owns one or more lane objects.

Example:

```text
Road

├── Lane 1
├── Lane 2
├── Lane 3
├── Sidewalk
└── Bicycle Lane
```

Each lane stores:

```cpp
Lane ID

Direction

Speed Limit

Vehicle Type

Bezier Curve

Width

Priority
```

Lane geometry is generated procedurally from the road spline.

---

## 3.4 Lane Types

Supported lane categories include:

- Driving
- Parking
- Bus
- Tram
- Bicycle
- Pedestrian
- Emergency
- Turning
- Highway Merge

Vehicle restrictions are enforced at the lane level.

---

## 3.5 Road Geometry

Road meshes are generated procedurally.

Construction pipeline:

```text
Road Nodes

↓

Spline Generation

↓

Lane Offset Calculation

↓

Road Surface Mesh

↓

Curbs

↓

Sidewalks

↓

Lane Markings

↓

Street Lights

↓

Trees

↓

Collision Mesh
```

No static road meshes exist.

Every road adapts dynamically to terrain.

---

## 3.6 Curves

Road segments are represented using spline interpolation.

Advantages:

- Smooth curves
- Natural highways
- Adaptive intersections
- Accurate lane generation

Each lane follows its own spline offset from the road centerline.

---

## 3.7 Elevation

Roads may exist at different elevations.

Supported:

- Ground
- Elevated
- Bridge
- Tunnel

Elevation affects:

- Construction cost
- Bridge pillars
- Tunnel portals
- Terrain deformation

---

## 3.8 Road Construction Validation

Before placement the engine validates:

- Terrain slope
- Collision
- Water
- Existing roads
- Buildings
- Map boundary
- Minimum curve radius
- Maximum slope

Invalid placement is rejected.

---

## 3.9 Road Upgrades

Existing roads may be upgraded.

Example:

```text
Two Lane

↓

Four Lane

↓

Six Lane

↓

Tree-Lined Avenue
```

Upgrade preserves:

- Connected buildings
- Zoning
- Intersections
- Utilities
- Public Transport Lines

---

## 3.10 Junction Generation

Whenever roads intersect, the engine generates a procedural junction.

Generation includes:

- Lane connections
- Turning lanes
- Crosswalks
- Stop lines
- Traffic lights
- Yield signs

The resulting junction depends on:

- Number of connected roads
- Road widths
- Angles
- Road hierarchy

---

## 3.11 Traffic Rules

Every intersection contains routing rules.

Supported:

- Stop
- Yield
- Traffic Light
- Priority Road
- Roundabout

Rules determine vehicle behavior.

---

## 3.12 Lane Connectivity

Every incoming lane connects to one or more outgoing lanes.

Example:

```text
Incoming Lane

↓

Left

↓

Straight

↓

Right
```

These connections form the navigation graph used by vehicles.

---

## 3.13 Road Hierarchy

Roads are classified by importance.

```text
Highway

↓

Arterial

↓

Collector

↓

Local Road
```

Traffic prefers higher hierarchy roads whenever possible.

---

## 3.14 Vehicle Pathfinding

Vehicles use graph-based pathfinding.

Each request evaluates:

- Distance
- Speed limits
- Road hierarchy
- Congestion
- Allowed vehicle types
- Lane restrictions

The resulting path consists of road segments rather than world positions.

---

## 3.15 Lane Selection

After a path is found, vehicles choose lanes.

Selection depends on:

- Upcoming turns
- Current congestion
- Lane availability
- Vehicle type

Lane changes occur before intersections whenever possible.

---
