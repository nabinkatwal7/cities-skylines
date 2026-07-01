# 1. Core Engine Architecture

# 2. Terrain & Map Generation System

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

## 2.11 Flooding ✅

Flooding occurs when water exceeds terrain elevation.

Flooded buildings become:

- Inaccessible
- Unpowered
- Abandoned
- Damaged

Roads become unusable.

Citizens evacuate flooded areas.

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

## 2.19 Resource Depletion ✅

Ore and oil deplete over time.

Each extraction building reduces nearby reserves.

Eventually:

```text
Rich Deposit

↓

Medium

↓

Poor

↓

Empty
```

Forests regenerate.

Fertile land does not deplete.

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

## 3.16 Vehicle AI

Each vehicle continuously updates:

```text
Desired Speed

↓

Acceleration

↓

Lane Position

↓

Obstacle Detection

↓

Traffic Lights

↓

Intersection Rules

↓

Destination
```

Vehicles never teleport unless explicitly despawned.

---

## 3.17 Traffic Lights

Traffic lights operate using phases.

```text
North South

Green

↓

Yellow

↓

Red

↓

East West

Green
```

Signal timing adapts to junction configuration.

Pedestrian crossings synchronize with vehicle phases.

---

## 3.18 Roundabouts

Roundabouts are treated as specialized junctions.

Rules:

- Entering vehicles yield.
- Vehicles inside have priority.
- Lane selection occurs before entry.
- Exits follow predefined lane connectivity.

---

## 3.19 Parking

Citizens attempt to park near destinations.

Parking options:

- Roadside
- Parking Lots
- Parking Garages

If parking is unavailable:

- Search nearby
- Continue driving
- Select alternative parking
- Walk remaining distance

Parking availability influences traffic congestion.

---

## 3.20 Road Maintenance

Every road incurs maintenance.

Factors:

- Length
- Type
- Decorations
- Lighting
- Elevation
- Bridges
- Tunnels

Maintenance contributes to the city's operating expenses.

---

## 3.21 Road Damage

Roads may become temporarily unusable due to:

- Flooding
- Natural disasters
- Construction
- Collapsed buildings

Blocked roads trigger automatic path recalculation.

---

## 3.22 Outside Connections

Certain roads connect to external regions.

External highways support:

- Imports
- Exports
- Immigration
- Tourists
- Service Requests

Outside traffic follows the same road simulation as local traffic.

---

## 3.23 Road Serialization

Saved road data includes:

- Nodes
- Segments
- Lane Layout
- Traffic Lights
- Road Names
- Elevation
- Decorations
- Maintenance State

Procedural meshes are regenerated after loading.

---

## 3.24 Performance Optimization

Road simulation uses:

- Spatial partitioning
- Incremental updates
- Cached junctions
- Lane graph caching
- Object pooling
- Dirty segment rebuilding
- Hierarchical pathfinding

Only modified roads rebuild geometry.

Traffic graph updates occur only when connectivity changes.

---

## 3.25 Design Goals

The road network is designed to support:

- Unlimited intersections
- Dynamic road upgrades
- Multi-lane traffic
- Realistic vehicle routing
- Procedural geometry
- Large-scale transportation networks
- Efficient pathfinding
- Modular road assets
- Public transportation integration
- Utility network integration
- High-performance simulation

The road system serves as the central navigation graph for citizens, vehicles, emergency services, public transportation, zoning, utilities, and city growth.

# 4. Public Transportation & External Connections

The public transportation system provides citizens and tourists with alternatives to private vehicle travel. Every transport mode operates on top of the road and network infrastructure while integrating directly with citizen pathfinding, traffic simulation, land value, tourism, and the city's economy.

Unlike simple waypoint systems, every transport vehicle is simulated as an independent entity that follows schedules, capacities, routing rules, and traffic conditions.

---

## 4.1 Transport Network Architecture

The transport system consists of independent network types.

```text
TransportManager

├── Bus Network
├── Tram Network
├── Metro Network
├── Train Network
├── Ferry Network
├── Monorail Network
├── Cable Car Network
├── Taxi System
├── Air Transport
└── Ship Transport
```

Each transport mode maintains:

- Routes
- Stops
- Stations
- Vehicles
- Passenger Statistics
- Maintenance
- Ticket Income
- Capacity

---

## 4.2 Transport Entity Model

Every transport system consists of four core objects.

```text
Transport Line

↓

Stops

↓

Vehicles

↓

Passengers
```

Each object exists independently.

Deleting a route does not delete stations.

Removing a station automatically updates affected routes.

---

## 4.3 Transport Line

A transport line represents an ordered sequence of stops.

Each line stores:

```cpp
LineID

TransportType

Color

Name

Stop List

Vehicle Count

Passenger Count

Budget

Statistics
```

The stop order forms a closed or open route depending on transport type.

---

## 4.4 Stops & Stations

Every stop stores:

```cpp
StationID

Position

Connected Networks

Waiting Passengers

District

Accessibility

Capacity
```

Stations act as transfer hubs between transport networks.

Examples:

- Bus Stop
- Metro Station
- Train Station
- Ferry Pier
- Airport Terminal
- Taxi Stand

---

## 4.5 Supported Transport Types

The engine supports:

- Walking
- Bicycle
- Private Car
- Taxi
- Bus
- Tram
- Metro
- Train
- Ferry
- Monorail
- Cable Car
- Blimp
- Airplane
- Ship

Each mode defines:

- Speed
- Capacity
- Operating Cost
- Ticket Revenue
- Pollution
- Noise
- Vehicle Size
- Infrastructure Requirements

---

## 4.6 Bus System

Buses operate on road networks.

Requirements:

- Bus Depot
- Road Access
- Bus Stops
- Active Line

Bus vehicles are automatically spawned from depots.

Each bus stores:

```cpp
Current Stop

Passenger Count

Capacity

Current Path

Fuel Type

Maintenance

Delay State
```

Buses obey:

- Traffic lights
- Congestion
- Lane restrictions
- One-way roads
- Speed limits

---

## 4.7 Tram System

Trams operate on dedicated tram tracks.

Unlike buses:

- Ignore road congestion where segregated.
- Share intersections.
- Require tram-enabled roads.
- Use tram depots.

Passengers board only at tram stops.

---

## 4.8 Metro System

Metro networks operate underground or elevated.

Components:

```text
Station

↓

Track

↓

Tunnel

↓

Depot

↓

Train
```

Metro trains ignore road traffic.

Stations connect directly to pedestrian networks.

---

## 4.9 Train Network

The rail network supports:

- Passenger Rail
- Cargo Rail

Passenger trains transport:

- Citizens
- Tourists

Cargo trains transport:

- Goods
- Raw Materials
- Mail (if enabled)

Railways support:

- Junctions
- Signals
- Outside Connections

---

## 4.10 Ferry Network

Ferries operate on navigable water.

Requirements:

- Ferry Depot
- Ferry Stops
- Water Route

Water depth is validated before route construction.

---

## 4.11 Monorail

Monorails operate on elevated guideways.

Characteristics:

- High capacity
- Medium speed
- Independent network
- Elevated only

---

## 4.12 Cable Cars

Cable cars operate between stations connected by cables.

Best suited for:

- Mountains
- Valleys
- Rivers

Cable cars ignore terrain.

---

## 4.13 Taxi System

Unlike fixed-route transportation, taxis operate on demand.

Process:

```text
Citizen Requests Taxi

↓

Taxi Dispatch

↓

Pickup

↓

Destination

↓

Available Again
```

Taxi availability depends on:

- Taxi Depot
- Number of Vehicles
- Traffic
- Distance

---

## 4.14 Air Transport

Airports provide:

- Passenger Flights
- Cargo Flights

Airport components:

- Runway
- Taxiway
- Terminal
- Gate
- Control Tower

Aircraft operate independently of road traffic.

Passengers use pedestrian connections to enter terminals.

---

## 4.15 Ship Transport

Ships connect ports with outside regions.

Supported cargo:

- Goods
- Oil
- Ore
- Forestry
- Farming Products

Passenger ships transport tourists.

---

## 4.16 Passenger AI

Citizens evaluate every available transportation option.

The decision considers:

- Total travel time
- Walking distance
- Waiting time
- Ticket cost
- Traffic
- Transfers
- Personal preferences

Example:

```text
Home

↓

Walk

↓

Bus

↓

Metro

↓

Walk

↓

Office
```

Every leg is independently simulated.

---

## 4.17 Route Selection

Citizens compute a generalized travel cost.

Example factors:

- Walking Time
- Vehicle Time
- Waiting Time
- Transfer Penalty
- Monetary Cost

The route with the lowest total cost is selected.

---

## 4.18 Transfers

Stations may connect multiple transport systems.

Example:

```text
Bus Stop

↓

Metro Station

↓

Train Station

↓

Airport
```

Transfers increase travel time but expand reachable destinations.

---

## 4.19 Vehicle Capacity

Each vehicle maintains:

```cpp
Maximum Capacity

Current Passengers

Standing Capacity

Boarding Queue
```

If full:

Passengers remain waiting.

Alternative routes may be selected.

---

## 4.20 Depot System

Vehicles originate from depots.

Responsibilities:

- Spawn vehicles
- Store vehicles
- Maintenance
- Replacement

Destroying a depot removes associated vehicles over time.

---

## 4.21 Budgets

Each transport mode has its own operating budget.

Budget influences:

- Vehicle Count
- Maintenance
- Frequency
- Operating Cost

Reducing the budget decreases service quality.

---

## 4.22 Ticket Revenue

Revenue is generated from passenger trips.

Income depends on:

- Passenger Count
- Ticket Price
- Distance
- Transport Type

Tourists also contribute to transport income.

---

## 4.23 Public Transport Coverage

Coverage is calculated based on walking distance to stations.

Higher coverage:

- Reduces traffic
- Increases land value
- Encourages development
- Improves citizen happiness

Poor coverage increases dependence on private vehicles.

---

## 4.24 External Connections

The city connects to neighboring regions through:

- Highway
- Rail
- Sea
- Air

Outside connections support:

- Immigration
- Tourism
- Imports
- Exports

External vehicles enter the city using the same simulation rules as local traffic.

---

## 4.25 Traffic Integration

Public transport integrates directly with traffic simulation.

Examples:

- Buses share roads.
- Trams share intersections.
- Taxis obey traffic.
- Emergency vehicles may bypass congestion.
- Metro ignores road traffic.
- Trains require rail paths.

Transport congestion influences citizen route selection.

---

## 4.26 Statistics

Each transport network tracks:

- Daily Passengers
- Weekly Passengers
- Lifetime Passengers
- Revenue
- Expenses
- Capacity Usage
- Average Wait Time
- Vehicle Utilization

Statistics feed into city reports and transport overlays.

---

## 4.27 Serialization

Saved transport data includes:

- Lines
- Stops
- Stations
- Vehicles
- Passenger Counts
- Colors
- Names
- Budgets
- Statistics

Network geometry is regenerated after loading.

---

## 4.28 Performance Optimization

The transport system uses:

- Route caching
- Station indexing
- Passenger batching
- Vehicle pooling
- Incremental updates
- Shared pathfinding
- Spatial partitioning

Only modified routes recalculate passenger paths.

---

## 4.29 Design Goals

The transportation system is designed to provide:

- Realistic passenger movement
- Seamless multimodal travel
- Dynamic traffic reduction
- Efficient public service
- Scalable transport networks
- Accurate economic integration
- Support for tourism and external connections
- High-performance simulation
- Flexible route management
- Independent simulation of every transport mode

Together with the road system, public transportation forms the complete mobility network of the city, enabling efficient movement of citizens, tourists, goods, and services while directly influencing traffic congestion, land value, economic activity, and overall city development.

# 5. Zoning System & City Growth Simulation

The zoning system governs how land is developed over time. Unlike traditional city builders where zoning immediately creates buildings, zones represent **development opportunities**. Buildings are constructed only when demand, services, accessibility, and economic conditions satisfy growth requirements.

Every zoned cell continuously evaluates whether it should develop, upgrade, remain occupied, become abandoned, or be demolished.

---

## 5.1 Zone Architecture

The Zone Manager maintains all zoned land independently from buildings.

```text
ZoneManager

├── Residential
├── Commercial
├── Industrial
├── Office
├── District Policies
├── Demand Engine
├── Growth Simulation
└── Land Value
```

Zones are simulation objects rather than buildings.

Removing a building does not remove its zone.

---

## 5.2 Zone Types

Supported zones:

### Residential

- Low Density Residential
- High Density Residential

Purpose:

- Housing
- Population Growth
- Tax Income

---

### Commercial

- Low Density Commercial
- High Density Commercial

Purpose:

- Shopping
- Entertainment
- Employment

---

### Industrial

Generic Industry

Produces:

- Goods
- Freight
- Pollution
- Employment

---

### Office

Produces:

- High Education Jobs
- Low Pollution
- Tax Revenue

Requires educated workforce.

---

## 5.3 Zone Grid

Zones are placed adjacent to roads.

```text
Road

□□□□□□□□

□□□□□□□□

□□□□□□□□

□□□□□□□□
```

Maximum depth depends on road type.

Buildings automatically occupy contiguous zone cells.

---

## 5.4 Buildable Lots

Adjacent zone cells combine into lots.

Example:

```text
■■■■

■■■■

↓↓↓

4×2 Building Lot
```

Lot size determines which assets may spawn.

---

## 5.5 Development Requirements

A zone develops only when all conditions are met.

Required:

- Road Connection
- Electricity
- Water
- Sewage
- Demand
- Buildable Terrain
- Available Building Asset

Failure prevents development.

---

## 5.6 RCI Demand

City growth is driven by RCI demand.

```text
Residential

Commercial

Industrial
```

Office demand is derived from Industrial demand as education levels increase.

Demand changes continuously.

---

## 5.7 Residential Demand

Residential demand increases when:

- Jobs available
- High happiness
- Low taxes
- Good services
- Immigration
- High land value

Residential demand decreases when:

- High unemployment
- Pollution
- Crime
- High taxes
- Death waves
- Abandonment

---

## 5.8 Commercial Demand

Commercial demand depends on:

- Population
- Household income
- Goods availability
- Tourism
- Shopping demand

Too much commercial zoning causes business failure.

---

## 5.9 Industrial Demand

Industrial demand increases with:

- Population growth
- Goods shortages
- Export opportunities

Industrial demand decreases when:

- Workers unavailable
- Freight congestion
- Resource shortages
- High taxes

---

## 5.10 Office Demand

Office demand depends upon:

- Educated workers
- High-tech economy
- Low industrial demand
- Office policies

Offices consume education instead of raw materials.

---

## 5.11 Building Spawn Process

```text
Zone

↓

Demand

↓

Select Lot

↓

Choose Building

↓

Construction

↓

Occupied
```

Buildings never appear instantly.

Construction takes simulation time.

---

## 5.12 Building Levels

Buildings evolve over time.

Example

Residential

```text
Level 1

↓

Level 2

↓

Level 3

↓

Level 4

↓

Level 5
```

Higher levels provide:

- More residents
- Higher taxes
- Better appearance
- Greater service demand

---

## 5.13 Upgrade Requirements

Buildings evaluate:

- Land Value
- Education
- Fire Coverage
- Police Coverage
- Health
- Parks
- Transport
- Happiness

Failure delays upgrading.

---

## 5.14 Land Value

Land value is continuously calculated.

Positive factors:

- Parks
- Trees
- Waterfront
- Public Transport
- Education
- Healthcare

Negative factors:

- Pollution
- Noise
- Crime
- Heavy Industry
- Garbage

Land value directly influences building level.

---

## 5.15 Abandonment

Buildings become abandoned if requirements remain unmet.

Reasons:

Residential

- No Power
- No Water
- High Pollution
- High Crime
- High Taxes

Commercial

- No Customers
- No Goods
- No Workers

Industrial

- No Workers
- No Freight
- No Resources

Office

- No Educated Workers

---

## 5.16 Demolition

Abandoned buildings may eventually be demolished.

Demolition frees the lot.

The zone remains.

New development may occur later.

---

## 5.17 Household Simulation

Residential buildings contain households.

Households store:

- Family Members
- Wealth
- Education
- Vehicles
- Happiness

Citizens belong to households rather than buildings directly.

---

## 5.18 Business Simulation

Commercial and industrial buildings maintain:

- Employee Count
- Production
- Storage
- Freight
- Profitability

Businesses continuously evaluate profitability.

---

## 5.19 Building Occupancy

Every building tracks:

Residential

- Residents
- Vacancies

Commercial

- Workers
- Customers

Industrial

- Workers
- Production

Office

- Employees

---

## 5.20 Construction

Construction stages:

```text
Empty Lot

↓

Foundation

↓

Framework

↓

Completed

↓

Occupied
```

Construction consumes time.

---

## 5.21 Service Consumption

Buildings consume:

Residential

- Electricity
- Water
- Garbage
- Healthcare
- Education

Commercial

- Goods
- Workers
- Electricity
- Water

Industrial

- Workers
- Electricity
- Water
- Freight

Office

- Educated Workers
- Electricity
- Internet (Future DLC)
- Water

---

## 5.22 District Integration

Zones inherit district policies.

Examples:

- High Rise Ban
- Heavy Traffic Ban
- Self Sufficient Housing
- IT Cluster
- Organic Produce

Policies modify building behavior.

---

## 5.23 Building AI

Every simulation cycle buildings evaluate:

- Occupancy
- Resources
- Services
- Income
- Expenses
- Happiness
- Growth
- Pollution

Buildings react dynamically to city conditions.

---

## 5.24 Statistics

The zoning system records:

- Population
- Vacant Homes
- Vacant Jobs
- Demand
- Building Levels
- Land Value
- Development Rate
- Abandonment

---

## 5.25 Serialization

Saved zoning data includes:

- Zone Grid
- Building Levels
- Occupancy
- Land Value
- Demand
- Construction Progress
- Policies
- Statistics

---

## 5.26 Performance Optimization

Optimizations include:

- Incremental growth updates
- Dirty zone recalculation
- Cached land value
- Spatial indexing
- Batched building evaluation

Only changed districts are recalculated.

---

## 5.27 Design Goals

The zoning system is designed to support:

- Organic city growth
- Dynamic building evolution
- Realistic economic development
- Service-dependent expansion
- Policy-driven specialization
- Continuous land valuation
- Large-scale city simulation
- Efficient performance
- Deterministic building lifecycle

The zoning engine transforms player planning into a living city by continuously evaluating demand, services, land value, and economic conditions to determine where and how development occurs.

# 6. Citizen Simulation (Cim AI)

The Citizen Simulation System is the heart of the city. Every resident, tourist, student, worker, prisoner, patient, and senior is represented by a simulated **Citizen (Cim)**. Citizens drive demand, consume city services, travel across transportation networks, generate taxes, occupy buildings, and influence nearly every gameplay system.

Unlike simple population counters, citizens exist as individual simulation entities connected to households, buildings, workplaces, schools, and transportation systems.

---

## 6.1 Citizen Architecture

The Citizen Manager is responsible for creating, updating, and removing all citizens.

```text
CitizenManager

├── Citizens
├── Households
├── Tourists
├── Students
├── Workers
├── Passengers
├── Visitors
└── Dead Citizens
```

Each citizen belongs to exactly one household and maintains references to their home, workplace or school, and current travel destination.

---

## 6.2 Citizen Data Model

Each citizen stores only simulation data. Visual appearance is generated separately.

```cpp
struct Citizen
{
    CitizenID id;

    HouseholdID household;

    BuildingID home;

    BuildingID workplace;

    BuildingID school;

    VehicleID vehicle;

    AgeGroup age;

    EducationLevel education;

    HealthState health;

    Happiness happiness;

    Wealth wealth;

    CurrentState state;

    CurrentLocation location;

    Destination destination;
}
```

---

## 6.3 Age Groups

Every citizen belongs to one age category.

```text
Child

↓

Teenager

↓

Young Adult

↓

Adult

↓

Senior

↓

Death
```

Each group has unique behaviors.

### Children

- Attend elementary school
- Cannot work
- Live with parents

---

### Teenagers

- Attend high school
- Begin education progression

---

### Young Adults

- Seek university
- Begin employment
- Form new households

---

### Adults

- Work
- Pay taxes
- Raise children
- Own vehicles

---

### Seniors

- Retire
- Require healthcare
- Generate increased deathcare demand

---

## 6.4 Education Levels

Education determines employment eligibility.

Levels include:

```text
Uneducated

↓

Educated

↓

Well Educated

↓

Highly Educated
```

Higher education unlocks office jobs and advanced industries.

---

## 6.5 Household System

Citizens are grouped into households.

Each household stores:

- Members
- Home
- Income
- Wealth
- Vehicles
- Pets (future expansion)

Households, rather than individual citizens, occupy residential buildings.

---

## 6.6 Citizen States

Each citizen follows a finite state machine.

```text
At Home

↓

Leaving Home

↓

Walking

↓

Waiting

↓

Board Vehicle

↓

Travel

↓

Work

↓

Shopping

↓

Leisure

↓

Return Home
```

The current state determines behavior and destination.

---

## 6.7 Daily Schedule

Citizens follow daily routines based on simulation time.

Example weekday:

```text
06:30 Wake Up

↓

07:30 Commute

↓

08:00 Work / School

↓

12:00 Lunch

↓

17:00 Leave Work

↓

18:00 Shopping / Leisure

↓

21:00 Return Home

↓

23:00 Sleep
```

Schedules vary by age, occupation, and city policies.

---

## 6.8 Destination Selection

Citizens continuously evaluate possible destinations.

Potential destinations include:

- Workplace
- School
- Commercial buildings
- Parks
- Tourist attractions
- Hospitals
- Cemeteries
- Home

The destination is selected based on the current state and citizen needs.

---

## 6.9 Travel Planning

Before moving, a citizen computes a complete travel plan.

Possible modes include:

- Walking
- Bicycle
- Private Car
- Taxi
- Bus
- Tram
- Metro
- Train
- Ferry
- Monorail
- Cable Car
- Airplane
- Ship

Travel plans may combine multiple modes.

Example:

```text
Home

↓

Walk

↓

Bus

↓

Metro

↓

Walk

↓

Office
```

---

## 6.10 Pathfinding

Citizens use the global transportation graph to determine optimal routes.

Factors include:

- Travel Time
- Walking Distance
- Congestion
- Waiting Time
- Ticket Cost
- Transfers
- Parking Availability

The cheapest valid route is selected.

---

## 6.11 Employment

Adults search for available jobs.

Requirements depend on:

- Education
- Distance
- Available positions
- Transport accessibility

Failure to find work contributes to unemployment.

---

## 6.12 School Enrollment

Eligible citizens automatically enroll in nearby schools if capacity exists.

Priority is determined by:

- Distance
- Available seats
- Education budget

Students graduate automatically after completing the required education period.

---

## 6.13 Shopping Behavior

Citizens periodically visit commercial buildings.

Shopping frequency depends on:

- Household wealth
- Happiness
- Free time
- Distance
- Transport availability

Commercial buildings rely on customer visits for income.

---

## 6.14 Leisure Activities

Citizens may visit:

- Parks
- Plazas
- Beaches
- Zoos
- Stadiums
- Tourist Attractions

Leisure improves happiness and increases park income.

---

## 6.15 Vehicle Ownership

Households may own private vehicles.

Vehicles remain parked until needed.

If a household owns no vehicle, citizens prefer public transport or walking.

---

## 6.16 Happiness

Citizen happiness is continuously evaluated.

Positive influences:

- Parks
- Low taxes
- Good education
- Healthcare
- Public transport
- Safety

Negative influences:

- Crime
- Pollution
- Noise
- Traffic
- High taxes
- Poor services

Happiness influences immigration and land value.

---

## 6.17 Health

Health decreases due to:

- Pollution
- Dirty water
- Noise
- Lack of healthcare
- Aging

Healthcare buildings restore health when capacity permits.

---

## 6.18 Sickness

When health falls below a threshold, a citizen becomes sick.

The household requests medical assistance.

Possible outcomes:

- Recovery
- Hospitalization
- Death

Untreated sickness reduces workplace productivity.

---

## 6.19 Death

Citizens eventually die from:

- Old age
- Illness
- Disasters

When death occurs:

```text
Citizen Dies

↓

Building Flags Death

↓

Hearse Requested

↓

Body Collected

↓

Cemetery/Crematorium

↓

Building Cleared
```

Failure to collect deceased citizens reduces surrounding happiness and health.

---

## 6.20 Immigration

New citizens arrive when:

- Residential demand exists
- Housing is available
- Jobs are available
- Outside connections exist

Immigrants immediately join households and occupy vacant homes.

---

## 6.21 Emigration

Citizens leave the city when:

- Taxes are too high
- Unemployment persists
- Services fail
- Pollution becomes excessive
- Housing is abandoned

Emigration reduces population and demand.

---

## 6.22 Tourists

Tourists behave similarly to citizens but do not permanently settle.

They visit:

- Hotels
- Parks
- Commercial areas
- Attractions
- Public transport

Tourists generate commercial revenue.

---

## 6.23 Citizen AI Update Cycle

Every simulation update, each citizen evaluates:

```text
Current State

↓

Needs

↓

Destination

↓

Travel Plan

↓

Movement

↓

Activity

↓

Next State
```

This state machine drives the living city simulation.

---

## 6.24 Statistics

The Citizen Manager tracks:

- Total Population
- Age Distribution
- Employment Rate
- Education Levels
- Happiness
- Health
- Birth Rate
- Death Rate
- Immigration
- Emigration
- Tourist Count

These statistics feed city reports and demand calculations.

---

## 6.25 Serialization

Citizen save data includes:

- Household
- Home
- Workplace
- School
- Current State
- Destination
- Education
- Health
- Happiness
- Wealth
- Age
- Vehicle Ownership

Simulation resumes exactly where each citizen left off after loading.

---

## 6.26 Performance Optimization

To support cities with hundreds of thousands of citizens, the system employs:

- Citizen pooling
- Incremental AI updates
- Level-of-detail simulation
- Shared pathfinding
- Batched state evaluation
- Cached destination queries
- Spatial indexing

Only active citizens near the camera receive full simulation updates, while distant citizens transition to lower-detail behavioral models without affecting overall simulation accuracy.

---

## 6.27 Design Goals

The Citizen Simulation System is designed to provide:

- Individual citizen simulation
- Household-based population management
- Dynamic education progression
- Realistic employment
- Multimodal transportation
- Organic daily routines
- Service-driven happiness
- Healthcare and deathcare integration
- Tourism simulation
- Scalable performance
- Deterministic behavior

The citizen system forms the core of the city's living ecosystem, connecting zoning, economy, transportation, services, and population growth into a unified simulation.

# 7. Utilities & Resource Network Simulation

The Utility System provides essential services required for every occupied building to function. Utilities operate as interconnected infrastructure networks that continuously transport resources from production facilities to consumers.

Every occupied building requires access to one or more utility networks. Failure of any critical utility immediately affects citizen happiness, economic productivity, and building occupancy.

---

## 7.1 Utility Manager Architecture

The Utility Manager oversees all infrastructure networks.

```text
UtilityManager

├── Electricity
├── Water Supply
├── Sewage
├── Heating
├── Garbage
├── Resource Storage
├── Production
└── Consumption
```

Each network is simulated independently but interacts with buildings, roads, industries, and public services.

---

## 7.2 Utility Networks

Supported utility systems include:

- Electricity
- Water Supply
- Sewage
- Heating (Snowfall)
- Garbage Collection
- Mail (Future Expansion)
- Internet (Future Expansion)

Each network tracks:

- Production
- Consumption
- Capacity
- Bottlenecks
- Failures

---

## 7.3 Electricity Network

Electricity powers all developed buildings.

Power is generated by:

- Coal Power Plant
- Oil Power Plant
- Nuclear Power Plant
- Wind Turbine
- Solar Power Plant
- Hydroelectric Dam
- Geothermal Plant

Each power source defines:

```cpp
PowerOutput

MaintenanceCost

Noise

AirPollution

GroundPollution

WaterConsumption
```

---

## 7.4 Power Distribution

Unlike pipe-based systems, electricity propagates through connected buildings and power lines.

```text
Power Plant

↓

Power Line

↓

Industrial Area

↓

Commercial Area

↓

Residential Area
```

Disconnected buildings immediately lose power.

---

## 7.5 Power Consumption

Every building has an electrical demand.

Examples:

Residential

- Lighting
- Appliances

Commercial

- Lighting
- Refrigeration
- Equipment

Industrial

- Heavy Machinery
- Production

Power demand changes with:

- Building Level
- Occupancy
- Time of Day
- Policies

---

## 7.6 Water Supply

Fresh water is produced by:

- Water Pumping Station
- Water Tower

Water is distributed through underground pipes.

Each pipe connects to neighboring pipes, creating a continuous network.

---

## 7.7 Water Network

```text
Pump

↓

Pipe

↓

Pipe

↓

Building
```

Buildings require:

- Water Connection
- Sufficient Pressure
- Available Capacity

Disconnected buildings become abandoned over time.

---

## 7.8 Sewage System

Wastewater is collected through the same underground pipe network.

Treatment options include:

- Drain Pipe
- Sewage Treatment Plant
- Eco Water Treatment

Untreated sewage pollutes nearby water sources.

---

## 7.9 Water Pollution

Water pollution spreads through flowing water.

Sources:

- Sewage
- Industrial Waste
- Landfill Leakage

Consequences:

- Citizen illness
- Reduced land value
- Contaminated water pumps
- Fish death
- Tourism decline

---

## 7.10 Heating Network (Snowfall)

Cold-weather maps introduce heating demand.

Heat sources include:

- Boiler Stations
- Geothermal Heating

Buildings require both:

- Electricity
- Heating

Insufficient heating reduces happiness and increases illness.

---

## 7.11 Garbage System

Every occupied building generates garbage.

Garbage production depends on:

- Population
- Commercial Activity
- Industrial Production

Garbage accumulates until collected.

---

## 7.12 Garbage Collection

Garbage facilities include:

- Landfill
- Incineration Plant
- Recycling Center

Collection process:

```text
Garbage Generated

↓

Truck Dispatched

↓

Collection

↓

Facility

↓

Processing
```

Overflowing garbage lowers land value and health.

---

## 7.13 Resource Production

Industrial and utility buildings produce resources.

Examples:

Coal Plant

Produces:

- Electricity

Consumes:

- Coal

Water Pump

Produces:

- Water

Consumes:

- Electricity

Treatment Plant

Produces:

- Clean Water

Consumes:

- Electricity

---

## 7.14 Capacity Management

Every utility tracks:

```text
Production

↓

Available Capacity

↓

Current Consumption

↓

Remaining Capacity
```

If demand exceeds production, shortages occur.

---

## 7.15 Service Radius vs Network

Utilities fall into two categories:

### Network-Based

- Electricity
- Water
- Sewage
- Heating

### Vehicle-Based

- Garbage
- Healthcare
- Fire
- Police

Network utilities require physical connections.

Vehicle services require road access.

---

## 7.16 Utility Failure

Common failures include:

Electricity

- No generation
- Disconnected network

Water

- Broken pipes
- Pump failure

Heating

- Fuel shortage
- Plant shutdown

Garbage

- Facility full
- Traffic congestion

Failures propagate through dependent systems.

---

## 7.17 Utility Demand

Demand changes continuously.

Influenced by:

- Population
- Building Levels
- Industry
- Weather
- Policies
- Time of Day

Peak demand typically occurs during morning and evening hours.

---

## 7.18 Building Utility Simulation

Each simulation update evaluates:

```text
Power

↓

Water

↓

Sewage

↓

Heating

↓

Garbage

↓

Building Status
```

If any mandatory utility is unavailable, warning notifications are generated.

---

## 7.19 Utility Policies

Policies modify consumption.

Examples:

- Power Usage
- Water Usage
- Recycling
- Energy Conservation

Policies influence:

- Costs
- Pollution
- Resource Consumption

---

## 7.20 Statistics

The Utility Manager tracks:

- Power Production
- Power Consumption
- Water Production
- Water Consumption
- Sewage Capacity
- Garbage Production
- Garbage Processing
- Heating Production
- Utility Costs

Statistics appear in city information panels.

---

## 7.21 Serialization

Utility save data includes:

- Network Graphs
- Pipe Layout
- Power Lines
- Production
- Consumption
- Capacities
- Facility States
- Policies

Procedural network meshes are regenerated after loading.

---

## 7.22 Performance Optimization

The utility simulation uses:

- Cached network graphs
- Dirty network updates
- Incremental propagation
- Chunk-based recalculation
- Shared spatial indexing

Only modified utility networks are recalculated.

---

## 7.23 Design Goals

The Utility System is designed to provide:

- Realistic infrastructure management
- Network-based resource distribution
- Dynamic production and consumption
- Pollution interaction
- Weather integration
- Scalable city support
- Efficient simulation
- Seamless integration with zoning, economy, transportation, and public services

The Utility System forms the foundation of city infrastructure, ensuring that every building receives the essential resources required for a functioning and prosperous city.

# 8. Public Services & City Service Simulation

The Public Service System provides the essential governmental functions required to maintain a healthy, safe, educated, and prosperous city. Unlike utilities, which distribute resources through infrastructure networks, public services operate primarily through **coverage**, **service capacity**, and **vehicle dispatch simulation**.

Every service building generates service requests, dispatches vehicles, responds to emergencies, and continuously influences land value, citizen happiness, and city growth.

---

## 8.1 Public Service Architecture

The Service Manager coordinates every public service.

```text
ServiceManager

├── Fire Department
├── Police Department
├── Healthcare
├── Deathcare
├── Education
├── Garbage
├── Parks & Recreation
├── Libraries
├── Disaster Response
├── Mail Service
└── Childcare & Eldercare
```

Each subsystem operates independently while sharing common dispatch, routing, budgeting, and statistics systems.

---

# 8.2 Service Request System

Buildings generate service requests whenever required.

Example:

```text
Building

↓

Needs Fire Response

↓

Fire Request Created

↓

Dispatch Queue

↓

Nearest Available Fire Engine

↓

Response

↓

Request Closed
```

Every request stores:

```cpp
RequestID

ServiceType

Priority

BuildingID

Timestamp

Status

AssignedVehicle
```

---

# 8.3 Coverage vs Capacity

Every service operates using two independent concepts.

## Coverage

Represents how effectively an area is served.

Factors:

- Distance
- Road connectivity
- District policies
- Traffic

---

## Capacity

Represents the total number of citizens or buildings that can be served.

Examples

Hospital

```
100 Patients
```

School

```
500 Students
```

Fire Station

```
15 Fire Engines
```

Coverage without capacity provides poor service.

Capacity without coverage provides no service.

---

# 8.4 Fire Department

Fire stations reduce fire risk and respond to emergencies.

Fire stations contain:

- Fire Engines
- Personnel
- Maintenance Budget

Fire probability increases with:

- Industrial activity
- Dense development
- Poor fire coverage
- Certain disasters

---

## Fire Dispatch

Process:

```text
Fire Starts

↓

Fire Request

↓

Nearest Available Station

↓

Dispatch Fire Engine

↓

Drive To Building

↓

Extinguish Fire

↓

Return To Station
```

Failure to respond quickly may result in:

- Building destruction
- Fire spreading
- Citizen deaths

---

## Fire Spread

Fire propagates based on:

- Building distance
- Wind direction
- Building material
- Fire intensity

Higher-density districts experience faster fire spread.

---

# 8.5 Police Department

Police services reduce crime and maintain public safety.

Police stations dispatch:

- Patrol Cars
- Prison Vans
- Helicopters (future DLC)

Crime increases due to:

- Unemployment
- Low education
- Poor land value
- Lack of police coverage

---

## Crime Simulation

Every building maintains a crime probability.

High crime results in:

- Reduced happiness
- Lower commercial income
- Increased abandonment
- Reduced tourism

Police patrols reduce local crime over time.

---

# 8.6 Prison System

Arrested criminals are transported to prisons.

Process:

```text
Crime

↓

Police Arrest

↓

Prison Van

↓

Prison

↓

Sentence

↓

Release
```

Prisons reduce repeat offenses but incur operating costs.

---

# 8.7 Healthcare

Healthcare improves citizen health and life expectancy.

Facilities include:

- Medical Clinic
- Hospital
- Medical Center

Each facility maintains:

- Ambulances
- Patient Capacity
- Operating Budget

---

## Ambulance Dispatch

```text
Citizen Sick

↓

Medical Request

↓

Nearest Hospital

↓

Dispatch Ambulance

↓

Pickup Citizen

↓

Treatment

↓

Return Vehicle
```

Insufficient ambulances increase mortality.

---

# 8.8 Health Simulation

Health depends on:

- Pollution
- Water Quality
- Healthcare
- Noise
- Parks
- Age

Poor health reduces productivity.

---

# 8.9 Deathcare

When a citizen dies:

```text
Citizen Dies

↓

Body Remains

↓

Hearse Request

↓

Hearse Pickup

↓

Cemetery

or

Crematorium
```

Buildings containing uncollected deceased citizens lose happiness.

---

## Cemeteries

Provide:

- Burial Capacity
- Hearses

Capacity eventually fills.

Filled cemeteries stop accepting bodies until emptied.

---

## Crematoriums

Provide:

- Hearses
- Cremation

Bodies are permanently removed.

---

# 8.10 Education

Education determines workforce quality.

Education buildings:

- Elementary School
- High School
- University

Each tracks:

- Student Capacity
- Teachers
- Budget

---

## Education Progression

```text
Child

↓

Elementary

↓

Teen

↓

High School

↓

Young Adult

↓

University

↓

Highly Educated Worker
```

Higher education unlocks office employment.

---

# 8.11 School Enrollment

Students attend the nearest school with available capacity.

Enrollment considers:

- Distance
- Capacity
- Transportation
- Budget

Schools with insufficient capacity leave citizens uneducated.

---

# 8.12 Parks & Recreation

Parks improve:

- Happiness
- Land Value
- Tourism

Supported park types:

- Small Park
- Plaza
- Playground
- Botanical Garden
- Zoo
- Nature Reserve
- Amusement Park

Parks generate maintenance costs.

Some parks generate income.

---

# 8.13 Libraries

Libraries increase education efficiency.

Effects:

- Faster education
- Higher graduation rates
- Better workforce

Libraries do not replace schools.

---

# 8.14 Childcare

Childcare improves household efficiency.

Effects:

- Parents spend more time working.
- Education improves.
- Happiness increases.

---

# 8.15 Eldercare

Eldercare improves senior health.

Benefits:

- Longer lifespan
- Lower healthcare demand
- Reduced death waves

---

# 8.16 Mail Service

Mail buildings generate delivery vehicles.

Process:

```text
Post Office

↓

Sorting

↓

Mail Truck

↓

Buildings

↓

Collection

↓

Processing
```

Mail improves commercial productivity.

---

# 8.17 Disaster Response

Disaster response buildings provide:

- Helicopters
- Rescue Teams
- Emergency Shelters

Used during:

- Tornadoes
- Earthquakes
- Floods
- Fires
- Meteors
- Tsunamis

---

# 8.18 Service Vehicles

Every service vehicle stores:

```cpp
VehicleID

Current Request

Origin

Destination

Fuel

Maintenance

Status
```

Possible states:

```text
Idle

↓

Dispatch

↓

Travel

↓

Service

↓

Return
```

---

# 8.19 Service Budget

Each service has an independent budget.

Budget affects:

- Vehicle Count
- Operating Hours
- Capacity
- Coverage

Reducing the budget decreases effectiveness.

Increasing the budget improves response time.

---

# 8.20 District Policies

District policies influence services.

Examples:

- Smoke Detectors
- Neighborhood Watch
- Free Education
- Free Public Transport
- Recreational Use
- Recycling

Policies modify service demand and effectiveness.

---

# 8.21 Statistics

Each service tracks:

Fire

- Fires
- Average Response Time

Police

- Crimes
- Arrests

Healthcare

- Patients
- Ambulance Usage

Education

- Students
- Graduation Rate

Deathcare

- Bodies Collected
- Cemetery Usage

Garbage

- Tons Collected

---

# 8.22 Serialization

Saved data includes:

- Service Buildings
- Vehicle States
- Requests
- Coverage
- Budgets
- Statistics
- Capacities

Vehicles resume their current assignments after loading.

---

# 8.23 Performance Optimization

Optimizations include:

- Shared dispatch queues
- Vehicle pooling
- Cached coverage maps
- Incremental service updates
- Spatial indexing
- Priority request batching

Only active service requests consume simulation resources.

---

# 8.24 Design Goals

The Public Service System is designed to provide:

- Realistic emergency response
- Dynamic service coverage
- Capacity-based simulation
- Individual service vehicle dispatch
- Integrated education and healthcare
- Crime and fire management
- Deathcare logistics
- District policy interaction
- Efficient large-city performance
- Deterministic service behavior

The Public Service System acts as the operational backbone of the city, ensuring that citizens receive essential governmental services while directly influencing happiness, land value, population growth, economic development, and overall city stability.

# 9. Economy, Taxes & Financial Simulation

The Economy System governs every financial transaction within the city. Every citizen, household, business, service, utility, vehicle, tourist, and government facility contributes to a continuously simulated economic model.

Unlike simplified city builders, the economy is not merely an income-versus-expense ledger. It is a dynamic simulation where employment, production, education, transportation, land value, taxation, imports, exports, and citizen behavior collectively determine financial stability and city growth.

---

# 9.1 Economy Architecture

The Economy Manager coordinates every financial subsystem.

```text
EconomyManager

├── Treasury
├── Taxation
├── Budget
├── Loans
├── Building Economy
├── Citizen Economy
├── Industrial Economy
├── Imports
├── Exports
├── Tourism
├── Production Chains
└── Statistics
```

The economy updates incrementally throughout the simulation rather than only at fixed intervals.

---

# 9.2 Treasury

The treasury stores the city's available funds.

```cpp
struct Treasury
{
    int64 CurrentBalance;

    int64 WeeklyIncome;

    int64 WeeklyExpenses;

    int64 LifetimeIncome;

    int64 LifetimeExpenses;
};
```

Every financial transaction modifies the treasury immediately.

---

# 9.3 Income Sources

The city generates income from:

- Residential Taxes
- Commercial Taxes
- Industrial Taxes
- Office Taxes
- Tourism
- Public Transport Tickets
- Park Admission
- Unique Buildings
- Toll Booths
- Airport Income
- Harbor Income
- DLC Industries
- Campus DLC
- Hotels (future expansion)

Each income source is tracked independently.

---

# 9.4 Expenses

The city continuously pays:

- Public Service Maintenance
- Utility Maintenance
- Road Maintenance
- Vehicle Maintenance
- Public Transport
- Parks
- Unique Buildings
- Landscaping
- Loan Payments
- Disaster Recovery
- Policy Costs

Expenses are deducted regardless of tax income.

---

# 9.5 Tax System

Taxes are configured independently for each zone.

Supported categories:

- Residential
- Commercial
- Industrial
- Office

Each category supports adjustable tax percentages.

Example:

```text
Residential

11%

Commercial

10%

Industrial

12%

Office

9%
```

---

# 9.6 Tax Effects

Increasing taxes raises revenue but negatively affects demand.

High taxes may cause:

Residential

- Emigration
- Lower Happiness
- Reduced Growth

Commercial

- Business Failure
- Reduced Investment

Industrial

- Factory Closure
- Reduced Production

Office

- Reduced Expansion

Lower taxes stimulate development but reduce government income.

---

# 9.7 Building Economy

Every zoned building continuously evaluates profitability.

Commercial buildings calculate:

```text
Revenue

↓

Operating Cost

↓

Taxes

↓

Profit
```

Industrial buildings evaluate:

- Resource Cost
- Worker Availability
- Freight Efficiency
- Export Revenue

Residential buildings pay taxes instead of generating profit.

---

# 9.8 Citizen Economy

Every household stores wealth.

Household income depends upon:

- Employment
- Education
- Taxes
- Policies

Higher-income households spend more on:

- Shopping
- Leisure
- Tourism
- Transportation

---

# 9.9 Employment Economy

Workers generate economic output.

Productivity depends on:

- Education
- Health
- Commute Time
- Happiness
- Workplace Efficiency

Long commutes reduce productivity.

---

# 9.10 Goods Production

Industrial buildings produce goods.

Production chain:

```text
Raw Materials

↓

Industry

↓

Processed Goods

↓

Commercial

↓

Citizens
```

Goods shortages reduce commercial sales.

---

# 9.11 Imports

If local production cannot satisfy demand, goods are imported.

Imports arrive via:

- Highway
- Cargo Rail
- Harbor
- Cargo Airport

Imports increase traffic and transportation costs.

---

# 9.12 Exports

Surplus production is exported.

Exports generate income.

Export routes include:

- Cargo Trains
- Ships
- Trucks
- Aircraft

Congested export routes reduce profitability.

---

# 9.13 Tourism Economy

Tourists generate revenue through:

- Hotels
- Shopping
- Public Transport
- Parks
- Attractions
- Restaurants
- Entertainment

Tourist spending contributes directly to city income.

---

# 9.14 Land Value Economy

Land value influences:

- Building Levels
- Tax Revenue
- Demand
- Property Value

Higher land value produces greater long-term tax income.

---

# 9.15 Budget System

Each department has an independent budget.

Examples:

Fire

80%

Police

100%

Healthcare

120%

Education

90%

Budgets influence:

- Vehicle Count
- Operating Hours
- Service Capacity
- Maintenance

---

# 9.16 Loans

Players may borrow money.

Loan properties:

```cpp
Principal

Interest Rate

Weekly Payment

Remaining Balance

Duration
```

Loan repayments occur automatically.

Multiple concurrent loans may be supported depending on progression.

---

# 9.17 Cash Flow

Every simulation day updates:

```text
Income

↓

Expenses

↓

Treasury

↓

Forecast

↓

Reports
```

Cash flow reports help identify financial issues before bankruptcy.

---

# 9.18 Bankruptcy

If funds become insufficient:

- Construction halts.
- Service budgets suffer.
- New loans may be required.

Persistent insolvency significantly reduces city performance.

---

# 9.19 Economic Policies

Policies influence economic behavior.

Examples:

- High Tech Housing
- Small Business Enthusiast
- Industry 4.0
- Tax Relief
- Heavy Traffic Ban
- Free Public Transport

Policies may increase expenses while improving long-term growth.

---

# 9.20 Resource Economy

Natural resources directly affect industrial profitability.

Resources:

- Ore
- Oil
- Forestry
- Fertile Land

Ore and oil deposits gradually deplete.

Forests regenerate naturally.

Fertile land remains renewable.

---

# 9.21 Inflation & Price Stability

The simulation assumes relatively stable prices over gameplay.

Construction costs, maintenance costs, and tax rates remain deterministic unless modified by scenarios, policies, or scripted events.

This avoids unnecessary economic complexity while maintaining predictable gameplay.

---

# 9.22 Financial Reports

The economy manager generates detailed reports.

Categories include:

Income

- Taxes
- Tourism
- Transport
- Parks

Expenses

- Services
- Utilities
- Roads
- Loans

Assets

- Cash
- Buildings
- Infrastructure

Forecasts

- Weekly Balance
- Trend Analysis

---

# 9.23 Statistics

Tracked metrics include:

- Weekly Income
- Weekly Expenses
- Cash Balance
- Profit/Loss
- GDP (simulation statistic)
- Employment Rate
- Average Household Wealth
- Commercial Revenue
- Industrial Production
- Office Productivity
- Tourist Spending

These statistics feed the city information panels and advisor systems.

---

# 9.24 Serialization

Saved financial data includes:

- Treasury
- Tax Rates
- Budgets
- Active Loans
- Weekly Reports
- Production Statistics
- Import/Export Totals
- Economic Policies

Financial history is preserved across saves.

---

# 9.25 Performance Optimization

Economic calculations use:

- Cached tax summaries
- Incremental building evaluation
- District aggregation
- Batched production updates
- Deferred report generation

Most calculations operate on district-level aggregates rather than recalculating every building each frame.

---

# 9.26 Design Goals

The Economy System is designed to provide:

- Dynamic taxation
- Realistic municipal budgeting
- Production and consumption simulation
- Import/export logistics
- Tourism revenue
- Resource-based industry
- Household wealth simulation
- Financial forecasting
- Scalable city-wide economics
- Deterministic and efficient updates

The Economy System ties together every major gameplay mechanic—citizens, zoning, services, transportation, industries, and policies—creating a living financial ecosystem that rewards thoughtful planning and long-term city management.

# 10. Districts, Policies & Specialized Areas

Districts allow players to divide the city into independently managed administrative regions. Every district can have unique policies, taxation, specialization, service priorities, transportation rules, and statistics.

Districts do not modify the terrain itself; instead they act as administrative overlays that influence the simulation of buildings, citizens, transportation, services, and the economy.

---

# 10.1 District Architecture

The District Manager maintains all player-created districts.

```text
DistrictManager

├── District Registry
├── District Borders
├── Policies
├── Tax Overrides
├── Specializations
├── Service Priorities
├── Statistics
└── Overlays
```

Each district functions independently while remaining connected to the overall city simulation.

---

# 10.2 District Definition

Every district stores:

```cpp
struct District
{
    DistrictID id;

    string name;

    Color color;

    Polygon boundary;

    Policies policies;

    Specialization specialization;

    TaxSettings taxes;

    Statistics stats;
};
```

District boundaries are stored as polygons rather than fixed tiles, allowing flexible shapes.

---

# 10.3 District Painting

Players create districts using a brush tool.

Painting workflow:

```text
Select Brush

↓

Paint Area

↓

District Created

↓

Buildings Assigned

↓

Policies Applied
```

Buildings automatically inherit the district in which their lot center resides.

---

# 10.4 District Boundaries

District borders are visual overlays only.

They do not:

- Block traffic
- Prevent zoning
- Affect terrain
- Stop utilities

They define administrative regions for simulation purposes.

---

# 10.5 District Naming

Each district supports:

- Custom Name
- Custom Color
- Custom Icon (future expansion)

District names appear in:

- Info Views
- Statistics
- Building Panels
- Citizen Information
- Transport Lines

---

# 10.6 Policy System

Policies modify simulation behavior inside a district.

Policies are categorized by:

- City Planning
- Economy
- Education
- Transportation
- Environment
- Public Services
- Recreation

Each policy has:

- Weekly Cost
- Simulation Effects
- Unlock Requirement

---

# 10.7 Residential Policies

Examples include:

- High-Rise Ban
- Self-Sufficient Housing
- Combustion Engine Ban
- Recycling
- Smoke Detector Distribution

Effects may include:

- Lower pollution
- Increased happiness
- Reduced electricity consumption
- Slower building growth

---

# 10.8 Commercial Policies

Supported examples:

- Small Business Enthusiast
- Organic Produce
- Local Produce
- Night Tours
- Tourism Promotion

Commercial policies influence:

- Customer traffic
- Profitability
- Tourist attraction
- Freight demand

---

# 10.9 Industrial Policies

Examples:

- Industry 4.0
- Automation
- Worker Safety
- Recycling
- Resource Efficiency

Industrial policies affect:

- Production
- Pollution
- Worker education requirements
- Resource consumption

---

# 10.10 Office Policies

Office-focused policies include:

- IT Cluster
- High-Tech Development
- Remote Work (future expansion)

Benefits:

- Higher productivity
- Reduced industrial demand
- Increased education requirements

---

# 10.11 Transportation Policies

Examples:

- Heavy Traffic Ban
- Encourage Cycling
- Old Town
- Free Public Transport
- Parking Restrictions

These policies influence:

- Vehicle choice
- Traffic flow
- Public transport usage
- Pedestrian activity

---

# 10.12 Service Policies

Districts may receive enhanced services.

Examples:

- Free Wi-Fi
- Improved Healthcare
- Community Policing
- Increased Fire Safety

Policies improve service quality while increasing maintenance costs.

---

# 10.13 Tax Overrides

Districts may override city-wide tax rates.

Supported categories:

```text
Residential

Commercial

Industrial

Office
```

Overrides apply only within the selected district.

---

# 10.14 District Specializations

Districts may specialize in unique economic activities.

Supported specializations:

- Forestry
- Farming
- Ore
- Oil
- Leisure
- Tourism
- Organic Commercial
- IT Cluster
- Self-Sufficient Residential
- Wall-to-Wall Housing

Specialization replaces or modifies the default building behavior.

---

# 10.15 Forestry Specialization

Forestry districts:

- Consume forest resources
- Produce timber
- Generate forestry jobs
- Export wood products

Trees regenerate naturally over time.

---

# 10.16 Farming Specialization

Requirements:

- Fertile Land

Produces:

- Crops
- Animal Products
- Agricultural Freight

Fertile land is renewable.

---

# 10.17 Ore Specialization

Requirements:

- Ore Deposits

Produces:

- Minerals
- Industrial Materials

Ore deposits gradually deplete.

---

# 10.18 Oil Specialization

Requirements:

- Oil Deposits

Produces:

- Petroleum
- Fuel

Oil deposits deplete over time.

---

# 10.19 Tourism Districts

Tourism specialization encourages:

- Hotels
- Attractions
- Restaurants
- Shopping
- Entertainment

Benefits:

- Increased tourist visits
- Higher commercial income

Drawbacks:

- More traffic
- Higher service demand

---

# 10.20 Leisure Districts

Leisure specialization promotes:

- Bars
- Restaurants
- Nightlife
- Entertainment Venues

Benefits:

- Higher tax income
- Greater tourism

Drawbacks:

- Increased noise
- Night traffic

---

# 10.21 District Statistics

Each district tracks:

Population

Employment

Education

Land Value

Traffic

Crime

Health

Fire Safety

Pollution

Tourism

Income

Expenses

Demand

Service Coverage

Statistics update continuously.

---

# 10.22 District Service Evaluation

Every district periodically evaluates:

```text
Population

↓

Service Capacity

↓

Coverage

↓

Citizen Happiness

↓

Growth Rate
```

Poor service slows development.

---

# 10.23 Citizen Interaction

Citizens recognize district boundaries for:

- Home location
- Workplace
- Taxes
- Policies
- Happiness modifiers

Citizens may live in one district while working in another.

---

# 10.24 Building Interaction

Buildings inherit:

- Policies
- Tax Rates
- Specialization
- Service Priorities

Changing district settings immediately affects all assigned buildings.

---

# 10.25 Transportation Integration

Transport systems use district information for:

- Passenger statistics
- Line naming
- Demand estimation
- Policy effects

District policies may encourage or discourage private vehicle usage.

---

# 10.26 Economy Integration

Districts maintain independent financial summaries.

Tracked values include:

- Tax Revenue
- Maintenance Costs
- Service Expenses
- Production
- Tourism Revenue
- Import/Export Volume

These reports help players optimize district performance.

---

# 10.27 Serialization

Saved district data includes:

- Boundaries
- Names
- Colors
- Policies
- Tax Overrides
- Specializations
- Statistics
- Service Priorities

District overlays are reconstructed exactly when loading a save.

---

# 10.28 Performance Optimization

The District Manager uses:

- Spatial indexing
- Cached boundary lookups
- Incremental statistics updates
- Dirty district recalculation
- Shared policy evaluation

Only districts affected by simulation changes are recalculated.

---

# 10.29 Design Goals

The District System is designed to provide:

- Administrative city management
- Fine-grained policy control
- Regional tax customization
- Specialized industrial development
- Localized service management
- Detailed district statistics
- Flexible player creativity
- Scalable simulation performance

Districts provide players with powerful tools to shape neighborhoods into unique communities, industrial centers, tourist destinations, and specialized economic hubs while integrating seamlessly with every major gameplay system.

# 11. Industries & Supply Chain Simulation

The Industries System extends the standard industrial economy into a complete production chain simulation. Instead of generic factories simply producing goods, specialized industries extract raw resources, process intermediate materials, manufacture finished products, store cargo, and export surplus production.

Every stage of production depends on transportation, workforce availability, utilities, storage capacity, and supply chain efficiency.

---

# 11.1 Industry Architecture

The Industry Manager oversees all industrial production.

```text
IndustryManager

├── Resource Extraction
├── Processing Facilities
├── Manufacturing
├── Warehouses
├── Cargo Logistics
├── Supply Chains
├── Production Statistics
└── Export Management
```

Each industrial area operates independently while contributing to the city's economy.

---

# 11.2 Industry Categories

Supported industries include:

- Generic Industry
- Forestry
- Farming
- Ore
- Oil

(DLC Expansion)

- Unique Factories
- Warehouses
- Production Chains

Each specialization defines:

- Raw Resources
- Processed Materials
- Finished Goods
- Workforce Requirements
- Pollution
- Vehicle Traffic

---

# 11.3 Production Chain

Every product moves through multiple stages.

```text
Natural Resource

↓

Extractor

↓

Raw Material

↓

Processor

↓

Processed Material

↓

Factory

↓

Commercial Goods

↓

Commercial Building

↓

Citizen Purchase
```

Interruptions at any stage reduce production.

---

# 11.4 Resource Extraction

Extraction buildings harvest:

Forestry

- Logs

Farming

- Crops
- Livestock

Ore

- Minerals

Oil

- Crude Oil

Extraction speed depends on:

- Resource Density
- Worker Availability
- Electricity
- Water
- Building Efficiency

---

# 11.5 Processing Facilities

Processors convert raw materials into industrial inputs.

Examples:

Logs

↓

Planed Timber

Crude Oil

↓

Petroleum

Ore

↓

Metals

Crops

↓

Food Products

Processing buildings require incoming freight deliveries before production begins.

---

# 11.6 Manufacturing

Factories consume processed materials to create commercial goods.

Requirements:

- Processed Inputs
- Workers
- Utilities
- Freight Access

Factories continuously evaluate inventory before producing.

---

# 11.7 Warehouses

Warehouses buffer supply chains.

Storage modes include:

- Balanced
- Fill
- Empty

Stored resources include:

- Raw Materials
- Processed Goods
- Commercial Goods

Warehouses reduce transport inefficiencies.

---

# 11.8 Inventory Simulation

Every industrial building maintains inventories.

```cpp
RawInput

ProcessedInput

OutputStorage

MaximumCapacity
```

Production halts if:

- Inputs unavailable
- Output storage full

---

# 11.9 Freight Logistics

Freight is transported by:

- Trucks
- Cargo Trains
- Cargo Ships
- Cargo Aircraft

The transport system selects the most efficient available route.

---

# 11.10 Internal Deliveries

Factories request deliveries automatically.

Workflow:

```text
Inventory Low

↓

Delivery Request

↓

Warehouse

↓

Truck Assigned

↓

Delivery

↓

Inventory Updated
```

Traffic congestion delays production.

---

# 11.11 External Imports

When resources are unavailable locally:

```text
Outside Connection

↓

Import Vehicle

↓

Warehouse

↓

Factory
```

Imports satisfy shortages but increase transportation costs.

---

# 11.12 Exports

Surplus products are exported.

Workflow:

```text
Warehouse

↓

Cargo Terminal

↓

Outside Region

↓

Income Generated
```

Efficient exports improve profitability.

---

# 11.13 Unique Factories

Unique factories consume multiple processed resources.

Example:

```text
Steel

+

Plastics

+

Paper

↓

Luxury Products
```

Unique factories generate high tax income and export value.

---

# 11.14 Workforce

Industrial buildings require workers.

Worker requirements vary by building type.

Generic Industry

- Mostly Uneducated

Processing

- Mixed Education

Unique Factories

- Highly Educated

Lack of workers reduces efficiency.

---

# 11.15 Resource Depletion

Finite resources:

- Ore
- Oil

Renewable resources:

- Forestry
- Farming

Extraction reduces local resource density until depletion.

---

# 11.16 Industrial Pollution

Industrial facilities generate:

- Air Pollution
- Ground Pollution
- Noise

Pollution intensity depends on:

- Building Type
- Production Rate
- Policies
- Technology Level

---

# 11.17 Supply Chain Failures

Production may stop due to:

- Missing Inputs
- No Workers
- Utility Failure
- Traffic Congestion
- Full Warehouses
- Export Blockage

Buildings display warning indicators until resolved.

---

# 11.18 Production Efficiency

Efficiency depends upon:

- Worker Education
- Utility Availability
- Freight Speed
- Policies
- Nearby Warehouses
- Resource Quality

Higher efficiency increases production without increasing building size.

---

# 11.19 Industry Areas

Industrial districts maintain:

- Production Rate
- Employment
- Resource Output
- Export Volume
- Import Volume
- Storage Capacity

Area statistics update continuously.

---

# 11.20 Budget & Maintenance

Industrial buildings incur:

- Maintenance
- Utility Costs
- Workforce Costs
- Vehicle Costs

Higher budgets improve production efficiency but increase expenses.

---

# 11.21 Policies

Industry policies include:

- Automation
- Worker Safety
- Resource Efficiency
- Recycling
- Improved Logistics

Policies modify production speed, pollution, and operating costs.

---

# 11.22 Statistics

Tracked metrics include:

- Raw Resource Production
- Processed Goods
- Commercial Goods
- Imports
- Exports
- Warehouse Utilization
- Factory Efficiency
- Employment
- Freight Traffic

These statistics are available through industry information panels.

---

# 11.23 Serialization

Saved industry data includes:

- Building States
- Inventories
- Warehouse Contents
- Active Deliveries
- Production Rates
- Statistics
- Policies

Freight deliveries resume normally after loading.

---

# 11.24 Performance Optimization

The Industry Manager uses:

- Cached production graphs
- Batched inventory updates
- Warehouse indexing
- Shared freight routing
- Incremental production evaluation

Only buildings with changing inventories or active deliveries are recalculated each update cycle.

---

# 11.25 Design Goals

The Industries System is designed to provide:

- Complete production chains
- Resource extraction simulation
- Realistic freight logistics
- Warehouse management
- Dynamic imports and exports
- Worker-dependent production
- Pollution interaction
- Transportation integration
- Scalable industrial simulation
- Efficient performance

The Industries System transforms industrial zones from simple tax generators into complex economic ecosystems where transportation, workforce, utilities, and logistics determine productivity and profitability.

# 12. Building System & Asset Lifecycle

The Building System is responsible for every placeable structure within the city. Buildings are persistent simulation entities that consume land, connect to infrastructure, provide services, house citizens, employ workers, store resources, and generate economic activity.

Unlike decorative objects, buildings participate continuously in the simulation and transition through a complete lifecycle from construction to demolition.

---

# 12.1 Building Architecture

The Building Manager owns every building in the simulation.

```text
BuildingManager

├── Zoned Buildings
├── Service Buildings
├── Unique Buildings
├── Monument Buildings
├── DLC Buildings
├── Construction
├── Upgrades
├── Damage
└── Demolition
```

Every building is uniquely identified and managed through a centralized entity pool.

---

# 12.2 Building Categories

Buildings are divided into major categories.

### Zoned Buildings

- Low Density Residential
- High Density Residential
- Low Density Commercial
- High Density Commercial
- Generic Industry
- Office

---

### Service Buildings

- Fire Station
- Police Station
- Hospital
- Clinic
- School
- University
- Landfill
- Power Plant
- Water Pump
- Sewage Plant
- Public Transport Depot

---

### Special Buildings

- Parks
- Monuments
- Unique Buildings
- Stadiums
- Airports
- Harbors
- Government Buildings

---

# 12.3 Building Data

Each building stores:

```cpp
struct Building
{
    BuildingID id;

    BuildingType type;

    Vector3 position;

    Quaternion rotation;

    DistrictID district;

    ZoneType zone;

    Level level;

    Occupancy occupancy;

    Flags state;

    Health health;

    Construction construction;
};
```

Simulation data is separated from rendering data.

---

# 12.4 Building Lifecycle

Every building progresses through a deterministic lifecycle.

```text
Planned

↓

Construction

↓

Completed

↓

Occupied

↓

Operational

↓

Upgrade

↓

Damaged

↓

Abandoned

↓

Demolished
```

Not every building reaches every state.

---

# 12.5 Construction Phase

When placed or spawned:

```text
Building Selected

↓

Foundation

↓

Construction

↓

Completed

↓

Activated
```

During construction:

- Utilities unavailable
- No occupants
- No tax income
- No service provided

---

# 12.6 Activation

After construction completes, initialization occurs.

Validation checks:

- Road access
- Electricity
- Water
- Sewage
- Terrain
- Collision
- District assignment

Failure prevents activation.

---

# 12.7 Occupancy

Buildings maintain occupancy information.

Residential

- Households
- Citizens

Commercial

- Employees
- Customers

Industrial

- Workers
- Freight

Office

- Employees

Service Buildings

- Patients
- Students
- Prisoners
- Vehicles

---

# 12.8 Building States

Possible building states include:

```text
Constructing

Operational

No Power

No Water

No Workers

No Customers

No Goods

Sick Citizens

Crime

Fire

Flooded

Collapsed

Abandoned

Historical

Destroyed
```

Multiple states may exist simultaneously.

---

# 12.9 Service Buildings

Service buildings provide simulation capabilities rather than housing.

Examples:

Hospital

Provides:

- Healthcare
- Ambulances

School

Provides:

- Education

Fire Station

Provides:

- Fire Engines

Police Station

Provides:

- Patrol Vehicles

Each maintains:

- Capacity
- Vehicles
- Budget
- Coverage

---

# 12.10 Building Levels

Zoned buildings evolve over time.

Example:

```text
Level 1

↓

Level 2

↓

Level 3

↓

Level 4

↓

Level 5
```

Higher levels generally provide:

- More residents
- More workers
- Increased tax revenue
- Better appearance

---

# 12.11 Upgrade Evaluation

Every update cycle evaluates:

- Land Value
- Happiness
- Services
- Education
- Pollution
- Building Age
- Occupancy

Eligible buildings begin upgrading automatically.

---

# 12.12 Historical Buildings

Players may mark buildings as historical.

Historical status:

- Prevents automatic rebuilding
- Prevents visual replacement
- Preserves architecture

Simulation behavior remains unchanged.

---

# 12.13 Abandonment

Buildings become abandoned if operational requirements remain unmet.

Common causes:

Residential

- No utilities
- High pollution
- Crime
- Taxes

Commercial

- No goods
- No customers
- No workers

Industrial

- No workers
- No freight

Office

- No educated workers

Abandoned buildings cease economic activity.

---

# 12.14 Damage System

Buildings may suffer damage from:

- Fire
- Flood
- Earthquake
- Tornado
- Meteor
- Tsunami
- Building Collapse

Damage is represented as a percentage.

```text
100%

↓

75%

↓

50%

↓

25%

↓

Destroyed
```

---

# 12.15 Collapse

Buildings collapse when structural integrity reaches zero.

Effects:

- Citizens evacuated
- Roads blocked
- Debris created
- Emergency services dispatched

Collapsed buildings require demolition before rebuilding.

---

# 12.16 Relocation

Certain service buildings may be relocated.

Process:

```text
Move Requested

↓

Validate New Location

↓

Transfer Simulation Data

↓

Reconnect Services

↓

Remove Old Building
```

Service interruption is minimized during relocation.

---

# 12.17 Demolition

Players may demolish buildings manually.

Demolition process:

```text
Demolish

↓

Evacuate

↓

Remove Occupants

↓

Remove Building

↓

Free Lot
```

Utilities and zoning remain where applicable.

---

# 12.18 Building Connections

Buildings maintain references to:

- Road
- Utility Network
- District
- Zone
- Service Coverage
- Nearby Buildings

Connections are updated whenever infrastructure changes.

---

# 12.19 Service Consumption

Each building consumes:

- Electricity
- Water
- Sewage
- Heating
- Garbage Collection

Consumption scales with occupancy and building level.

---

# 12.20 Pollution

Buildings may generate:

- Air Pollution
- Ground Pollution
- Noise Pollution

Pollution depends on:

- Building Type
- Production
- Traffic
- Policies

Residential buildings generate minimal pollution.

Industrial facilities generate the most.

---

# 12.21 Building AI

Every simulation cycle, buildings evaluate:

```text
Utilities

↓

Occupancy

↓

Services

↓

Economy

↓

Growth

↓

Warnings

↓

State Update
```

Only relevant systems are recalculated if the building's state changes.

---

# 12.22 Asset System

Every building references an asset definition.

Assets define:

- Meshes
- Textures
- LOD Models
- Construction Cost
- Maintenance
- Service Type
- Capacity
- Upgrade Rules

Simulation instances reference immutable asset templates.

---

# 12.23 Asset Variations

Buildings may have multiple visual variants.

Selection depends on:

- Zone
- Level
- District Style
- Theme
- Random Seed

Simulation properties remain identical unless explicitly overridden.

---

# 12.24 Statistics

Each building tracks:

- Construction Date
- Occupancy
- Visitors
- Lifetime Taxes
- Maintenance
- Resource Consumption
- Service Requests
- Happiness Modifier

Statistics are available through the building information panel.

---

# 12.25 Serialization

Saved building data includes:

- Position
- Rotation
- State
- Occupancy
- Level
- Construction Progress
- Damage
- Policies
- Historical Status
- Statistics

Asset definitions are loaded separately from simulation data.

---

# 12.26 Performance Optimization

The Building Manager uses:

- Entity pooling
- Dirty building updates
- Spatial indexing
- Cached service coverage
- Incremental occupancy evaluation
- Shared asset templates

Only buildings affected by simulation changes are reprocessed.

---

# 12.27 Design Goals

The Building System is designed to provide:

- Deterministic building lifecycles
- Dynamic upgrades
- Realistic occupancy
- Service integration
- Disaster interaction
- Efficient asset reuse
- Flexible zoning support
- Scalable simulation
- Incremental processing
- High-performance rendering integration

The Building System serves as the physical backbone of the city, connecting zoning, citizens, services, transportation, utilities, economy, and disasters into a unified simulation while supporting cities containing tens of thousands of simultaneously active structures.

# 13. Vehicle Simulation & Traffic AI

The Vehicle Simulation System is responsible for every moving vehicle within the city. Every private car, taxi, bus, tram, train, metro, aircraft, ship, bicycle, emergency vehicle, cargo truck, garbage truck, hearse, and service vehicle exists as an independent simulation entity.

Unlike abstract transportation models, every vehicle follows roads or transport networks, obeys traffic laws, avoids collisions, responds to congestion, and dynamically recalculates routes when the network changes.

---

# 13.1 Vehicle Architecture

The Vehicle Manager owns all moving entities.

```text
VehicleManager

├── Private Cars
├── Service Vehicles
├── Emergency Vehicles
├── Cargo Vehicles
├── Public Transport
├── Trains
├── Ships
├── Aircraft
├── Bicycles
└── Pedestrians (Movement Layer)
```

Vehicles are allocated from memory pools and updated through deterministic simulation ticks.

---

# 13.2 Vehicle Entity

Each vehicle stores:

```cpp
struct Vehicle
{
    VehicleID id;

    VehicleType type;

    OwnerID owner;

    CurrentLane lane;

    CurrentSegment segment;

    float speed;

    float acceleration;

    float maxSpeed;

    Route currentRoute;

    VehicleState state;

    Cargo cargo;

    PassengerData passengers;
};
```

Vehicles never store rendering information.

---

# 13.3 Vehicle Categories

Supported vehicle classes include:

### Private

- Car
- Motorcycle (Future)
- Bicycle

---

### Public Transport

- Bus
- Tram
- Metro
- Train
- Ferry
- Monorail
- Cable Car
- Taxi

---

### Emergency

- Ambulance
- Fire Engine
- Police Car
- Prison Van
- Rescue Helicopter

---

### Utility

- Garbage Truck
- Hearse
- Maintenance Truck
- Snow Plow

---

### Cargo

- Cargo Truck
- Cargo Train
- Cargo Ship
- Cargo Aircraft

---

### External

- Tourist Car
- Outside Connection Truck
- Passenger Train
- Cruise Ship
- Passenger Aircraft

---

# 13.4 Vehicle Lifecycle

Vehicles transition through the following lifecycle:

```text
Spawn Request

↓

Allocated

↓

Path Generated

↓

Travel

↓

Destination

↓

Unload

↓

Return

↓

Despawn

↓

Pool
```

Vehicles are never permanently destroyed during normal gameplay.

---

# 13.5 Vehicle Spawning

Vehicles spawn only when required.

Examples:

Citizen leaves home

↓

Spawn Car

Fire request

↓

Spawn Fire Engine

Garbage collection

↓

Spawn Garbage Truck

Bus departure

↓

Spawn Bus

Spawn locations include:

- Parking
- Depots
- Stations
- Outside Connections

---

# 13.6 Vehicle Despawning

Vehicles despawn only when:

- Returning to depots
- Leaving outside connections
- Completing one-time trips
- Simulation optimization (where appropriate)

Private vehicles generally return to parking rather than disappearing.

---

# 13.7 Vehicle State Machine

Each vehicle operates using a finite state machine.

```text
Idle

↓

Spawn

↓

Find Lane

↓

Accelerate

↓

Cruise

↓

Brake

↓

Stop

↓

Turn

↓

Unload

↓

Return

↓

Despawn
```

---

# 13.8 Driving Model

Each simulation step evaluates:

```text
Desired Speed

↓

Lane Speed Limit

↓

Vehicle Ahead

↓

Traffic Signal

↓

Intersection

↓

Acceleration

↓

New Position
```

Vehicles obey acceleration and deceleration limits rather than changing speed instantly.

---

# 13.9 Speed Limits

Each road defines a maximum speed.

Vehicles calculate:

```text
Road Speed

↓

Vehicle Maximum

↓

Traffic

↓

Weather

↓

Actual Speed
```

Heavy vehicles accelerate more slowly.

Emergency vehicles may exceed standard limits under emergency response.

---

# 13.10 Lane Following

Vehicles remain attached to lane splines.

Movement is calculated using:

```text
Lane Curve

↓

Distance Along Curve

↓

Interpolated Position

↓

Heading

↓

Vehicle Transform
```

This produces smooth movement regardless of road curvature.

---

# 13.11 Lane Changing

Lane changes occur when:

- Preparing for turns
- Avoiding congestion
- Overtaking (where permitted)
- Following route requirements

Lane changes require:

- Safe gap
- Valid target lane
- Route compatibility

Unsafe lane changes are rejected.

---

# 13.12 Intersection Behavior

Vehicles approaching intersections evaluate:

- Traffic lights
- Stop signs
- Yield rules
- Priority roads
- Occupied intersections

Vehicles enter only when the path ahead is clear.

---

# 13.13 Collision Avoidance

Vehicles continuously monitor:

- Lead vehicle
- Crossing traffic
- Pedestrians
- Service vehicles

Following distance increases with speed.

Vehicles brake before collisions occur.

---

# 13.14 Parking System

Private vehicles attempt to park near destinations.

Parking locations include:

- Roadside parking
- Parking lots
- Parking garages

If parking is unavailable:

```text
Search Nearby

↓

Alternative Parking

↓

Walk Remaining Distance
```

Parking demand influences traffic congestion.

---

# 13.15 Emergency Vehicle Priority

Emergency vehicles receive routing priority.

Behavior includes:

- Ignore certain traffic rules
- Higher speed limits
- Priority at intersections
- Reduced congestion penalties

Other vehicles yield when possible.

---

# 13.16 Cargo Vehicle Simulation

Cargo vehicles transport:

- Raw resources
- Processed materials
- Commercial goods
- Mail (where enabled)

Cargo deliveries directly influence industrial productivity and commercial supply.

---

# 13.17 Public Transport Vehicles

Public transport vehicles additionally manage:

- Passenger boarding
- Passenger unloading
- Capacity
- Timetables (simulation-based)
- Stop queues

Vehicles pause at designated stops before continuing.

---

# 13.18 Vehicle Routing

Every route is generated using the transportation graph.

Routing considers:

- Distance
- Estimated travel time
- Congestion
- Road hierarchy
- Vehicle restrictions
- Transport mode

Routes are recalculated when the road network changes significantly.

---

# 13.19 Traffic Congestion

Congestion increases travel time.

Causes include:

- High vehicle density
- Poor intersections
- Insufficient road hierarchy
- Freight bottlenecks
- Poor public transport

Congestion feeds back into citizen route planning.

---

# 13.20 Traffic Flow

The simulation continuously measures:

- Average speed
- Road occupancy
- Queue lengths
- Intersection throughput
- Travel times

Traffic heatmaps visualize these metrics.

---

# 13.21 Vehicle Maintenance

Service vehicles return to depots for maintenance.

Poor maintenance budgets may:

- Reduce available vehicles
- Slow dispatch
- Increase downtime

Private vehicles do not require explicit maintenance simulation.

---

# 13.22 Weather Effects

Weather influences driving behavior.

Examples:

Rain

- Reduced speeds
- Increased stopping distance

Snow

- Slower acceleration
- Snow plow dependency

Fog

- Reduced visibility (visual effect)
- Slightly lower speeds

---

# 13.23 Vehicle Statistics

Each vehicle records:

- Lifetime Distance
- Fuel/Energy Usage (simulation statistic)
- Cargo Delivered
- Passengers Transported
- Average Speed
- Waiting Time

Aggregated statistics feed transportation reports.

---

# 13.24 Serialization

Saved vehicle data includes:

- Position
- Lane
- Route
- Speed
- Destination
- Cargo
- Passengers
- Current State

Vehicles resume exactly where they were after loading a save.

---

# 13.25 Performance Optimization

To support thousands of simultaneous vehicles, the Vehicle Manager uses:

- Vehicle pooling
- Shared lane graphs
- Incremental AI updates
- Hierarchical pathfinding
- Spatial partitioning
- Cached routes
- Distance-based simulation LOD

Only vehicles affected by network or traffic changes perform expensive recalculations.

---

# 13.26 Design Goals

The Vehicle Simulation System is designed to provide:

- Individually simulated vehicles
- Realistic traffic behavior
- Dynamic lane selection
- Collision avoidance
- Intelligent routing
- Public transport integration
- Freight logistics
- Emergency response priority
- Weather-aware driving
- Scalable performance
- Deterministic simulation

The Vehicle Simulation System forms the dynamic movement layer of the city, connecting citizens, industries, services, transportation, and the economy through a realistic and highly scalable traffic simulation.

# 14. Pathfinding & Navigation System

The Pathfinding System is the core navigation engine used by every moving entity in the simulation. Citizens, vehicles, public transport, service vehicles, cargo deliveries, tourists, and emergency responders all rely on a shared routing framework to navigate the city.

Unlike simple shortest-path systems, the navigation engine evaluates travel time, transportation modes, congestion, road hierarchy, service restrictions, and dynamic network conditions.

---

# 14.1 Navigation Architecture

The Navigation Manager coordinates all routing requests.

```text
NavigationManager

├── Road Graph
├── Rail Graph
├── Pedestrian Graph
├── Water Graph
├── Air Graph
├── Transport Transfers
├── Route Cache
├── Traffic Cost Map
└── Path Scheduler
```

Each transportation network maintains its own navigation graph while remaining connected through transfer points.

---

# 14.2 Navigation Graph

The world is represented as multiple interconnected graphs.

```text
Road Graph

↓

Pedestrian Graph

↓

Metro Graph

↓

Rail Graph

↓

Ferry Graph

↓

Air Graph
```

Transfer nodes connect these graphs into one unified transportation network.

---

# 14.3 Graph Nodes

Each navigation node stores:

```cpp
NodeID

Position

ConnectedEdges

NodeType

Flags
```

Examples:

- Road Intersection
- Bus Stop
- Metro Station
- Train Platform
- Ferry Dock
- Airport Gate
- Pedestrian Crossing

---

# 14.4 Graph Edges

Edges connect nodes.

Each edge stores:

```cpp
Length

Travel Time

Speed Limit

Transportation Mode

Capacity

Current Cost
```

Travel cost changes dynamically.

---

# 14.5 Supported Travel Modes

The navigation engine supports:

- Walking
- Bicycle
- Private Vehicle
- Taxi
- Bus
- Tram
- Metro
- Train
- Ferry
- Monorail
- Cable Car
- Airplane
- Ship

Each mode has unique costs and restrictions.

---

# 14.6 Route Requests

Entities generate routing requests.

Examples:

Citizen

```
Home

↓

Work
```

Garbage Truck

```
Depot

↓

Building

↓

Landfill
```

Cargo Truck

```
Factory

↓

Warehouse
```

Each request is processed asynchronously.

---

# 14.7 Route Cost

Pathfinding minimizes generalized travel cost rather than physical distance.

Cost factors include:

- Distance
- Speed
- Congestion
- Waiting Time
- Transfers
- Parking Search
- Traffic Lights
- Road Hierarchy
- Vehicle Restrictions

---

# 14.8 A\* Search

Primary route generation uses the A\* algorithm.

Evaluation:

```text
Current Cost (G)

+

Estimated Remaining Cost (H)

=

Total Cost (F)
```

The heuristic uses straight-line travel time rather than Euclidean distance.

---

# 14.9 Hierarchical Routing

Large cities require hierarchical navigation.

Routing stages:

```text
District

↓

Road Hierarchy

↓

Local Roads

↓

Destination
```

This dramatically reduces search complexity.

---

# 14.10 Route Caching

Frequently used routes are cached.

Examples:

- Home → Work
- Home → School
- Depot → Service Area
- Warehouse → Factory

Cache invalidation occurs only when network topology changes.

---

# 14.11 Dynamic Recalculation

Routes may be recalculated when:

- Roads are removed
- Traffic becomes severe
- Flooding blocks roads
- Disaster destroys infrastructure
- Policies restrict access
- New roads shorten travel time

Only affected routes are recalculated.

---

# 14.12 Lane Routing

Vehicle routing produces:

```text
Road Path

↓

Lane Assignment

↓

Turn Planning

↓

Driving Instructions
```

Lane selection is deferred until approaching intersections.

---

# 14.13 Pedestrian Routing

Pedestrians travel using:

- Sidewalks
- Crosswalks
- Pedestrian Streets
- Paths
- Parks
- Station Connections

Walking routes may combine with public transport.

---

# 14.14 Public Transport Routing

Citizens evaluate multimodal journeys.

Example:

```text
Walk

↓

Bus

↓

Metro

↓

Walk
```

Transfers incur additional travel cost.

---

# 14.15 Cargo Routing

Cargo routes prioritize:

- Fast Roads
- Highways
- Cargo Rail
- Harbors
- Airports

Freight avoids residential streets whenever practical.

---

# 14.16 Emergency Routing

Emergency vehicles receive modified routing rules.

Priority:

- Ignore congestion penalties
- Prefer arterial roads
- Minimize response time
- Override some traffic restrictions

---

# 14.17 Road Restrictions

Navigation respects:

- One-Way Roads
- Vehicle Restrictions
- Bus Lanes
- Tram Tracks
- Heavy Traffic Bans
- Old Town Policies

Invalid edges are excluded during search.

---

# 14.18 Congestion Costs

Traffic continuously modifies edge weights.

Example:

```text
Empty Road

Cost = 1.0

↓

Moderate Traffic

Cost = 1.5

↓

Heavy Congestion

Cost = 3.8
```

Citizens naturally seek faster alternatives.

---

# 14.19 Outside Connections

Routes may terminate outside the map.

Supported:

- Highway
- Rail
- Ship
- Aircraft

Outside nodes connect directly to regional simulation.

---

# 14.20 Transfer Hubs

Stations connect multiple transportation graphs.

Example:

```text
Bus Stop

↓

Metro Station

↓

Train Station

↓

Airport
```

Transfer penalties prevent unrealistic route choices.

---

# 14.21 Route Failures

If no valid path exists:

Possible causes:

- Road disconnected
- Missing utilities
- Destroyed bridge
- Blocked tunnel
- Closed station

Affected entities wait until a valid route becomes available.

---

# 14.22 Batch Scheduling

Route requests are processed using priority queues.

Priority order:

```text
Emergency

↓

Public Transport

↓

Cargo

↓

Citizens

↓

Tourists

↓

Statistics
```

Critical routes are always computed first.

---

# 14.23 Pathfinding Statistics

Tracked metrics include:

- Total Requests
- Successful Routes
- Failed Routes
- Average Search Time
- Cache Hit Rate
- Average Travel Time
- Recalculation Count

These metrics assist with performance analysis.

---

# 14.24 Serialization

Saved navigation data includes:

- Cached Routes
- Graph Versions
- Pending Requests
- Route Seeds

Navigation graphs themselves are regenerated from roads and transport networks during loading.

---

# 14.25 Performance Optimization

The navigation engine employs:

- Hierarchical A\*
- Route caching
- Shared graph structures
- Parallel pathfinding jobs
- Incremental graph rebuilding
- Spatial partitioning
- Batch scheduling

Only modified graph regions are rebuilt after infrastructure changes.

---

# 14.26 Design Goals

The Pathfinding System is designed to provide:

- Fast large-scale routing
- Multimodal transportation
- Dynamic traffic adaptation
- Intelligent lane selection
- Efficient emergency response
- Scalable navigation
- Deterministic route generation
- High-performance pathfinding
- Seamless integration with transportation and citizen AI

The Pathfinding System is the central navigation layer that connects every transportation mode, allowing millions of simulated journeys to occur efficiently while adapting dynamically to the evolving city infrastructure.

# 15. Public Transportation System

The Public Transportation System enables citizens and tourists to travel efficiently without relying on private vehicles. It consists of transport networks, depots, stations, lines, vehicles, schedules, passenger routing, ticket revenue, and operational management.

Unlike simple transport mechanics, every passenger independently selects routes, transfers between transportation modes, waits at stations, boards vehicles based on capacity, and dynamically adjusts travel plans according to congestion and network changes.

---

# 15.1 Transport Architecture

The Transport Manager coordinates every transportation network.

```text
TransportManager

├── Bus Network
├── Tram Network
├── Metro Network
├── Train Network
├── Ferry Network
├── Monorail Network
├── Cable Car Network
├── Taxi System
├── Airport System
├── Harbor System
├── Passenger Routing
└── Ticket Economy
```

Each transportation mode maintains its own infrastructure, vehicles, and operational rules while sharing a unified passenger routing system.

---

# 15.2 Transport Modes

Supported transport systems include:

Ground

- Bus
- Tram
- Taxi

Rail

- Metro
- Passenger Train
- Monorail

Water

- Ferry

Air

- Passenger Airport

Special

- Cable Car

Each mode has unique:

- Speed
- Capacity
- Infrastructure
- Maintenance
- Operating Cost

---

# 15.3 Transport Infrastructure

Infrastructure consists of:

- Depots
- Stops
- Stations
- Tracks
- Roads
- Platforms
- Terminals
- Maintenance Facilities

Vehicles originate from depots and operate along assigned transport lines.

---

# 15.4 Depots

Every transport system requires a depot.

Depot responsibilities:

- Spawn vehicles
- Store idle vehicles
- Perform maintenance
- Replace broken vehicles

Without a depot, no transport line can operate.

---

# 15.5 Stops & Stations

Passengers board and leave vehicles only at designated stops.

Stop types:

- Bus Stop
- Tram Stop
- Metro Station
- Train Station
- Ferry Pier
- Monorail Station
- Cable Car Station
- Airport Terminal

Stops store waiting passengers until a suitable vehicle arrives.

---

# 15.6 Transport Lines

Players create transport lines manually.

Workflow:

```text
Create Line

↓

Select Stops

↓

Complete Loop or Route

↓

Assign Color

↓

Assign Vehicles

↓

Activate
```

Lines may be:

- Circular
- Bidirectional
- Point-to-point

---

# 15.7 Line Data

Each transport line stores:

```cpp
LineID

TransportType

Stops[]

Vehicles[]

Color

Name

Budget

PassengerStatistics
```

Lines are persistent simulation entities.

---

# 15.8 Passenger Simulation

Every passenger independently evaluates:

```text
Origin

↓

Destination

↓

Walking Distance

↓

Available Lines

↓

Transfers

↓

Travel Time

↓

Chosen Route
```

Passengers are never teleported.

---

# 15.9 Waiting System

Passengers arriving at a stop enter a waiting queue.

Queue properties:

- Arrival Time
- Desired Line
- Destination
- Boarding Priority

Passengers continue waiting until an appropriate vehicle arrives or they abandon the trip.

---

# 15.10 Boarding

When a vehicle reaches a stop:

```text
Vehicle Arrives

↓

Passengers Exit

↓

Waiting Queue Evaluated

↓

Passengers Board

↓

Capacity Reached

↓

Departure
```

Passengers unable to board remain in the queue.

---

# 15.11 Vehicle Capacity

Each vehicle has a maximum capacity.

Examples:

Bus

- ~30 passengers

Tram

- ~90 passengers

Metro

- ~180 passengers

Train

- Multiple carriages

Capacity affects passenger waiting times and line efficiency.

---

# 15.12 Transfers

Passengers may transfer between transport modes.

Example:

```text
Walk

↓

Bus

↓

Metro

↓

Train

↓

Walk
```

Transfer penalties increase total travel cost but expand reachable destinations.

---

# 15.13 Transport AI

Vehicles continuously evaluate:

- Current Stop
- Passenger Count
- Schedule Progress
- Congestion
- Route Completion

Vehicles follow predefined lines unless rerouted due to network changes.

---

# 15.14 Timetables

Vehicles do not use strict real-world timetables.

Instead, service frequency emerges from:

- Number of Vehicles
- Route Length
- Traffic Conditions
- Stop Dwell Time

Adding vehicles naturally reduces passenger waiting times.

---

# 15.15 Stop Dwell Time

Vehicles remain stopped while:

- Passengers exit
- Passengers board

Dwell time scales with passenger volume.

Busy stations produce longer stop durations.

---

# 15.16 Ticket Revenue

Each completed passenger journey generates ticket income.

Revenue depends on:

- Transport Mode
- Passenger Count
- City Policies

Ticket income contributes directly to the city's treasury.

---

# 15.17 Operating Costs

Transport systems incur continuous expenses.

Examples:

- Vehicle Maintenance
- Depot Maintenance
- Fuel/Electricity
- Personnel
- Infrastructure Maintenance

Long transport lines require larger operating budgets.

---

# 15.18 Line Budget

Each transport type has an independent budget.

Budget influences:

- Maximum Vehicles
- Service Frequency
- Vehicle Maintenance
- Operating Hours

Higher budgets improve service quality but increase expenses.

---

# 15.19 Taxi System

Taxis operate differently from fixed-route transport.

Workflow:

```text
Passenger Requests Taxi

↓

Nearest Available Taxi

↓

Pickup

↓

Destination

↓

Drop-off

↓

Await Next Request
```

Taxi availability depends on the number of taxi depots and active vehicles.

---

# 15.20 Airport Simulation

Passenger airports handle:

- Tourist arrivals
- Tourist departures
- Citizen travel
- Outside connections

Airports consist of:

- Runways
- Taxiways
- Gates
- Passenger Terminals

Aircraft follow scheduled arrival and departure paths.

---

# 15.21 Harbor Simulation

Passenger harbors support ferry and cruise traffic.

Ships:

- Arrive
- Dock
- Exchange passengers
- Depart

Harbors connect to the city's road and public transport networks.

---

# 15.22 Passenger Satisfaction

Passengers evaluate:

- Walking Distance
- Waiting Time
- Crowding
- Transfers
- Travel Time

Poor public transport encourages private vehicle usage.

---

# 15.23 Transport Statistics

Tracked metrics include:

- Daily Ridership
- Average Waiting Time
- Passenger Capacity
- Ticket Revenue
- Vehicle Utilization
- Line Efficiency
- Transfers
- Operating Cost
- Profit/Loss

Statistics are available through transport information panels.

---

# 15.24 Transport Overlays

Available overlays include:

- Passenger Density
- Line Colors
- Vehicle Locations
- Stop Congestion
- Coverage
- Traffic Integration
- Ridership Heatmap

These overlays assist players in optimizing transport networks.

---

# 15.25 Serialization

Saved transport data includes:

- Lines
- Stops
- Vehicles
- Passenger Queues
- Budgets
- Ticket Revenue
- Statistics

Passengers resume their journeys seamlessly after loading.

---

# 15.26 Performance Optimization

The Transport Manager uses:

- Passenger pooling
- Incremental queue updates
- Shared route caches
- Station occupancy aggregation
- Vehicle LOD simulation
- Batch passenger routing

Only stations or lines experiencing activity are updated every simulation tick.

---

# 15.27 Design Goals

The Public Transportation System is designed to provide:

- Fully simulated passenger movement
- Dynamic route selection
- Multimodal transportation
- Capacity-based boarding
- Efficient transport management
- Revenue generation
- Congestion reduction
- Realistic station behavior
- Seamless integration with citizen AI and pathfinding
- Scalable performance for large metropolitan cities

The Public Transportation System serves as the backbone of urban mobility, providing efficient alternatives to private vehicles while directly influencing traffic flow, citizen happiness, economic productivity, tourism, and overall city development.

# 16. Traffic Management & Road AI

The Traffic Management System governs how vehicles interact with the road network. While the Vehicle Simulation System controls individual vehicle behavior, the Traffic Management System manages lane utilization, intersections, congestion, road hierarchy, traffic control devices, parking demand, and overall traffic flow.

The system continuously analyzes road usage and dynamically influences routing decisions across the city.

---

# 16.1 Traffic Architecture

The Traffic Manager coordinates every road-related simulation.

```text
TrafficManager

├── Road Network
├── Lane Manager
├── Intersection Manager
├── Traffic Signals
├── Parking Manager
├── Traffic Flow Analysis
├── Congestion Detection
├── Road Restrictions
└── Statistics
```

The Traffic Manager works alongside the Navigation Manager to optimize city-wide mobility.

---

# 16.2 Road Hierarchy

Efficient traffic relies on a hierarchical road network.

Road classes include:

- Highway
- Large Avenue
- Medium Avenue
- Collector Road
- Local Road
- Service Road
- Pedestrian Street

Vehicles prefer higher-capacity roads for long-distance travel and local roads for access to destinations.

---

# 16.3 Lane System

Each road consists of one or more lanes.

Lane types include:

- Forward Driving Lane
- Reverse Driving Lane
- Turning Lane
- Bus Lane
- Tram Lane
- Bicycle Lane
- Parking Lane
- Emergency Lane (future expansion)

Every lane stores:

```cpp
LaneID

RoadSegment

Direction

SpeedLimit

AllowedVehicles

Connections[]
```

Lane connectivity determines valid vehicle movement through intersections.

---

# 16.4 Lane Selection

Vehicles continuously evaluate lane choice.

Factors include:

- Upcoming turn
- Congestion
- Lane speed
- Vehicle restrictions
- Destination

Lane changes are planned well before reaching intersections to minimize abrupt movements.

---

# 16.5 Road Capacity

Every road segment has a theoretical maximum throughput.

Capacity depends on:

- Lane count
- Speed limit
- Intersection spacing
- Traffic lights
- Parking activity
- Public transport usage

Traffic volume approaching capacity results in slower average speeds and longer queues.

---

# 16.6 Congestion Simulation

Congestion emerges naturally from vehicle density.

The system tracks:

- Average speed
- Queue length
- Lane occupancy
- Intersection delay
- Throughput

Road segments are classified as:

```text
Free Flow

↓

Moderate

↓

Heavy

↓

Gridlocked
```

Congestion directly influences route planning.

---

# 16.7 Traffic Signals

Signalized intersections regulate vehicle movement.

Signal phases include:

```text
North/South Green

↓

Yellow

↓

All Red

↓

East/West Green

↓

Yellow

↓

Repeat
```

Signal timing adapts to road configuration but remains deterministic.

---

# 16.8 Stop Signs & Yield

Unsignalized intersections use priority rules.

Supported controls:

- Stop Sign
- Yield Sign
- Priority Road
- Roundabout Priority

Vehicles must wait until gaps are available before entering.

---

# 16.9 Roundabouts

Roundabouts prioritize circulating traffic.

Rules:

- Entering vehicles yield.
- Vehicles exit using assigned lanes.
- Lane changes within the roundabout are minimized.

Properly designed roundabouts improve flow compared to signalized intersections.

---

# 16.10 Turning Movements

Vehicles classify intended movement as:

- Left Turn
- Right Turn
- Straight
- U-turn (where permitted)

Dedicated turning lanes improve intersection throughput.

---

# 16.11 One-Way Roads

One-way roads simplify traffic flow.

Benefits include:

- Increased capacity
- Fewer conflict points
- Better traffic distribution

Navigation respects one-way restrictions at all times.

---

# 16.12 Highway Rules

Highways support:

- High speed limits
- No zoning
- Limited intersections
- Entrance ramps
- Exit ramps

Pedestrians and bicycles are prohibited unless explicitly allowed by asset configuration.

---

# 16.13 Parking Simulation

Parking demand is simulated for private vehicles.

Available parking:

- Roadside
- Parking Lot
- Parking Garage
- Building Parking

If nearby parking is unavailable, drivers search surrounding areas before walking to their destination.

Parking shortages increase local congestion.

---

# 16.14 Public Transport Priority

Public transport vehicles receive priority where infrastructure allows.

Examples:

- Dedicated bus lanes
- Tram tracks
- Signal priority (future expansion)

Priority reduces delays and improves service reliability.

---

# 16.15 Freight Traffic

Cargo vehicles prioritize:

- Highways
- Cargo terminals
- Industrial roads

Heavy freight avoids residential streets whenever viable.

Industrial districts naturally generate concentrated freight traffic.

---

# 16.16 Emergency Traffic

Emergency vehicles override normal routing.

Features:

- Priority pathfinding
- Reduced intersection delays
- Higher travel speeds
- Vehicles yield where possible

Response time is critical to service effectiveness.

---

# 16.17 Traffic Incidents

Traffic flow may be disrupted by:

- Vehicle breakdowns (scenario support)
- Building fires
- Flooding
- Collapsed roads
- Disaster debris
- Construction

Blocked roads are temporarily removed from routing until cleared.

---

# 16.18 Road Restrictions

Road segments may restrict access.

Supported restrictions include:

- Heavy Traffic Ban
- Bus Only
- Tram Only
- Service Vehicles Only
- Pedestrian Only
- Bicycle Priority

Restrictions influence navigation and traffic distribution.

---

# 16.19 Traffic Flow Metrics

The system continuously calculates:

- Vehicles per minute
- Average speed
- Delay
- Road occupancy
- Queue growth
- Travel time

These metrics are visualized using overlays.

---

# 16.20 Traffic Heatmaps

Players can inspect:

- Congestion
- Average speed
- Road usage
- Freight traffic
- Public transport usage
- Parking demand

Heatmaps update in real time with the simulation.

---

# 16.21 Road Maintenance

Road condition is abstracted into maintenance costs.

Well-funded maintenance ensures:

- Full speed limits
- Reliable service access
- Consistent traffic flow

Road maintenance expenses scale with total road length.

---

# 16.22 Traffic Statistics

Tracked metrics include:

- Overall Traffic Flow (%)
- Total Vehicles
- Average Commute Time
- Road Utilization
- Congested Segments
- Parking Occupancy
- Freight Volume
- Public Transport Share

Traffic flow percentage is a key city performance indicator.

---

# 16.23 Serialization

Saved traffic data includes:

- Active queues
- Signal states
- Lane occupancy
- Parking occupancy
- Traffic statistics
- Congestion history

Road geometry and lane graphs are regenerated from the road network during loading.

---

# 16.24 Performance Optimization

The Traffic Manager employs:

- Lane occupancy caching
- Incremental congestion updates
- Spatial partitioning
- Shared lane graphs
- Dirty intersection recalculation
- Hierarchical road evaluation

Only roads affected by significant traffic changes are recalculated.

---

# 16.25 Design Goals

The Traffic Management System is designed to provide:

- Realistic road hierarchy
- Intelligent lane usage
- Dynamic congestion
- Efficient intersections
- Parking simulation
- Freight optimization
- Emergency vehicle priority
- Public transport integration
- High-performance traffic analysis
- Deterministic city-wide traffic behavior

The Traffic Management System transforms the road network into a living transportation ecosystem where infrastructure design, zoning, public transport, and citizen behavior collectively determine traffic efficiency and the long-term success of the city.

# 17. Disaster Simulation & Emergency Management

The Disaster System introduces large-scale dynamic events that can significantly alter the city. Disasters destroy infrastructure, disrupt utilities, block transportation, injure citizens, damage buildings, and create long-term recovery challenges.

Unlike scripted events, disasters are fully integrated into the simulation. Emergency services, citizens, transportation, utilities, economy, and public services all respond dynamically based on the severity and location of the event.

---

# 17.1 Disaster Architecture

The Disaster Manager coordinates every disaster event.

```text
DisasterManager

├── Disaster Generator
├── Active Disasters
├── Damage Simulation
├── Emergency Response
├── Evacuation System
├── Shelter Management
├── Recovery System
├── Terrain Effects
└── Statistics
```

Disasters are deterministic once generated, ensuring save/load consistency.

---

# 17.2 Disaster Types

Supported disasters include:

Natural

- Earthquake
- Tornado
- Tsunami
- Flood
- Forest Fire
- Sinkhole
- Thunderstorm

Space

- Meteor Strike

Human

- Building Fire
- Industrial Explosion
- Infrastructure Failure (Scenario)

Each disaster has unique behaviors, damage patterns, and recovery requirements.

---

# 17.3 Disaster Lifecycle

Every disaster progresses through defined stages.

```text
Spawn

↓

Warning

↓

Growth

↓

Peak

↓

Decay

↓

Recovery

↓

Cleanup

↓

Completed
```

Each phase influences emergency response and citizen behavior.

---

# 17.4 Early Warning System

Some disasters provide advance notice.

Examples:

- Tornado Warning
- Tsunami Warning
- Severe Storm Alert
- Meteor Detection

Warnings allow players to activate emergency procedures before impact.

---

# 17.5 Earthquakes

Earthquakes generate:

- Ground shaking
- Road damage
- Building collapse
- Utility failures
- Aftershocks

Damage severity depends on:

- Magnitude
- Distance from epicenter
- Building resilience
- Terrain type

---

# 17.6 Tornadoes

Tornadoes move dynamically across the map.

Effects include:

- Building destruction
- Vehicle displacement
- Fallen trees
- Utility damage
- Citizen casualties

Wind speed determines destruction radius.

---

# 17.7 Tsunamis

Tsunamis propagate across water.

Simulation:

```text
Wave Generated

↓

Ocean Travel

↓

Landfall

↓

Flooding

↓

Retreat

↓

Cleanup
```

Coastal infrastructure is especially vulnerable.

---

# 17.8 Floods

Flooding results from:

- Tsunamis
- River overflow
- Heavy rainfall (Scenario)

Floodwater spreads according to terrain elevation.

Flooded areas experience:

- Building shutdown
- Road closures
- Utility disruption
- Citizen evacuation

---

# 17.9 Forest Fires

Forests ignite under appropriate conditions.

Fire spreads according to:

- Tree density
- Wind direction
- Dryness
- Distance

Firefighters attempt containment before urban areas are reached.

---

# 17.10 Meteor Strikes

Meteor events consist of:

```text
Detection

↓

Atmospheric Entry

↓

Impact

↓

Explosion

↓

Shockwave

↓

Fire

↓

Recovery
```

Large impacts may permanently alter terrain.

---

# 17.11 Sinkholes

Sinkholes destroy infrastructure below ground.

Possible effects:

- Road collapse
- Utility disruption
- Building destruction
- Traffic rerouting

---

# 17.12 Disaster Damage

Damage categories include:

- Structural Damage
- Utility Damage
- Road Damage
- Rail Damage
- Water Network Damage
- Citizen Injuries
- Vehicle Losses

Buildings store continuous damage values rather than simple destroyed/not destroyed states.

---

# 17.13 Citizen Response

Citizens react immediately.

Possible behaviors:

```text
Continue Activity

↓

Notice Warning

↓

Evacuate

↓

Seek Shelter

↓

Return Home

↓

Recovery
```

Behavior depends on disaster type and available shelters.

---

# 17.14 Evacuation System

Players may designate evacuation routes.

Process:

```text
Warning

↓

Shelter Assigned

↓

Citizen Travels

↓

Shelter Arrival

↓

Remain Safe

↓

Return Home
```

Road congestion directly affects evacuation success.

---

# 17.15 Emergency Shelters

Shelters provide:

- Temporary housing
- Food
- Medical support
- Safety

Each shelter stores:

- Capacity
- Occupancy
- Supplies

Overflowing shelters reduce effectiveness.

---

# 17.16 Emergency Services

Disasters generate large numbers of service requests.

Responding agencies include:

- Fire Department
- Police
- Ambulances
- Rescue Helicopters
- Disaster Response Units

Emergency dispatch prioritizes life-threatening situations.

---

# 17.17 Utility Failures

Disasters may disable:

- Electricity
- Water
- Sewage
- Heating
- Communications (future expansion)

Infrastructure repairs begin once emergency conditions stabilize.

---

# 17.18 Transportation Impact

Transportation systems experience:

- Road blockages
- Bridge collapse
- Tunnel flooding
- Rail disruption
- Airport shutdown
- Harbor closure

Navigation automatically reroutes where possible.

---

# 17.19 Economic Impact

Disasters reduce:

- Tax revenue
- Tourism
- Productivity
- Commercial activity

Recovery expenses include:

- Building reconstruction
- Road repair
- Utility restoration
- Emergency services

---

# 17.20 Building Recovery

Damaged buildings progress through:

```text
Damaged

↓

Inspection

↓

Repair

↓

Reoccupation

↓

Operational
```

Destroyed buildings require complete reconstruction.

---

# 17.21 Terrain Modification

Certain disasters permanently modify terrain.

Examples:

- Meteor craters
- Sinkholes
- Landslides (future expansion)

Terrain changes affect future construction and routing.

---

# 17.22 Disaster Statistics

Tracked metrics include:

- Buildings Destroyed
- Citizens Injured
- Citizens Deceased
- Roads Damaged
- Utilities Disabled
- Economic Loss
- Emergency Response Time
- Shelter Occupancy
- Recovery Cost

Statistics remain accessible after recovery.

---

# 17.23 Disaster Scenarios

Scenario support includes:

- Forced disasters
- Timed objectives
- Difficulty modifiers
- Budget constraints
- Special victory conditions

Scenario events may chain multiple disasters together.

---

# 17.24 Serialization

Saved disaster data includes:

- Active disasters
- Disaster progression
- Damaged buildings
- Floodwater state
- Emergency requests
- Shelter occupancy
- Recovery progress

Simulation resumes seamlessly after loading.

---

# 17.25 Performance Optimization

The Disaster Manager employs:

- Localized damage simulation
- Spatial partitioning
- Incremental flood propagation
- Event-based updates
- Parallel damage evaluation
- Deferred cleanup

Only affected regions receive intensive simulation updates during active disasters.

---

# 17.26 Design Goals

The Disaster System is designed to provide:

- Dynamic large-scale disasters
- Integrated emergency response
- Citizen evacuation
- Infrastructure destruction
- Terrain interaction
- Long-term recovery
- Economic consequences
- Transportation disruption
- Deterministic simulation
- High-performance event processing

The Disaster System transforms unexpected catastrophic events into fully simulated gameplay experiences, challenging players to design resilient cities, prepare effective emergency services, and recover efficiently while maintaining the long-term prosperity of the city.

# 18. Environment, Pollution & Natural Resources

The Environment System simulates the interaction between terrain, vegetation, water, weather, pollution, climate, and natural resources. Environmental conditions directly influence citizen health, land value, industry, tourism, agriculture, and long-term city development.

Unlike purely visual systems, environmental simulation continuously affects gameplay through measurable consequences.

---

# 18.1 Environment Architecture

The Environment Manager coordinates all environmental subsystems.

```text
EnvironmentManager

├── Air Pollution
├── Ground Pollution
├── Noise Pollution
├── Water Pollution
├── Wind Simulation
├── Weather
├── Temperature
├── Natural Resources
├── Forest System
├── Water Simulation
└── Statistics
```

Environmental systems update independently while sharing terrain and simulation data.

---

# 18.2 Environmental Layers

The world contains multiple environmental overlays.

Supported layers include:

- Air Pollution
- Ground Pollution
- Noise Pollution
- Water Pollution
- Wind
- Temperature
- Ore
- Oil
- Fertile Land
- Forest Density

Each layer is stored independently.

---

# 18.3 Air Pollution

Air pollution is generated by:

- Coal Power Plants
- Oil Power Plants
- Generic Industry
- Specialized Industry
- Heavy Traffic
- Airports

Air pollution diffuses outward over time.

Consequences include:

- Reduced citizen health
- Lower land value
- Building abandonment
- Reduced desirability

---

# 18.4 Ground Pollution

Ground pollution affects terrain directly.

Sources:

- Industry
- Landfills
- Oil Extraction
- Ore Extraction
- Sewage Leakage
- Disaster Debris

Ground pollution dissipates slowly.

Residential zoning performs poorly in polluted areas.

---

# 18.5 Noise Pollution

Noise originates from:

- Busy Roads
- Highways
- Railways
- Airports
- Industrial Areas
- Entertainment Districts
- Stadiums

Noise decreases with distance.

High noise reduces:

- Residential happiness
- Land value
- Residential demand

Commercial districts tolerate higher noise levels.

---

# 18.6 Water Pollution

Water pollution enters rivers, lakes, and oceans through:

- Sewage Outlets
- Industrial Waste
- Flood Debris
- Landfill Leakage

Pollution follows water flow.

Consequences:

- Contaminated drinking water
- Increased illness
- Fish mortality
- Lower tourism
- Reduced waterfront land value

---

# 18.7 Pollution Propagation

Pollution spreads according to its type.

```text
Air

↓

Wind

Ground

↓

Terrain

Water

↓

Water Flow

Noise

↓

Distance
```

Each propagation model is simulated independently.

---

# 18.8 Wind Simulation

Wind direction influences:

- Air pollution movement
- Forest fire spread
- Tornado movement
- Wind turbine efficiency

Wind is represented as a continuous vector field across the map.

---

# 18.9 Temperature

Temperature affects:

- Heating demand
- Snow accumulation
- Citizen comfort
- Disaster likelihood (scenario dependent)

Cold maps require active heating infrastructure.

---

# 18.10 Weather

Supported weather conditions include:

- Clear
- Cloudy
- Rain
- Snow
- Fog
- Storm

Weather influences:

- Traffic
- Electricity production
- Tourism
- Citizen behavior

---

# 18.11 Water Simulation

Water is fully simulated.

Water sources include:

- Rivers
- Lakes
- Ocean
- Water Outlets
- Rainfall (scenario)

Water interacts with:

- Terrain elevation
- Dams
- Flood barriers
- Sewage
- Pollution

---

# 18.12 Forest System

Trees exist as simulation objects.

Functions:

- Improve land value
- Absorb pollution
- Support forestry
- Reduce noise
- Beautify the environment

Forests regenerate naturally if left undisturbed.

---

# 18.13 Tree Lifecycle

```text
Sapling

↓

Young Tree

↓

Mature Tree

↓

Old Tree

↓

Removed

or

Natural Regrowth
```

Tree growth rates vary by map type and climate.

---

# 18.14 Natural Resources

Supported resources:

- Ore
- Oil
- Fertile Land
- Forest

Each map defines unique resource distributions.

---

# 18.15 Resource Depletion

Finite resources:

- Ore
- Oil

Extraction reduces local deposits over time.

Renewable resources:

- Forestry
- Agriculture

These regenerate naturally through environmental processes.

---

# 18.16 Land Value

Environmental quality strongly affects land value.

Positive influences:

- Parks
- Clean Air
- Waterfront
- Trees
- Good Services

Negative influences:

- Pollution
- Noise
- Traffic
- Crime

Higher land value promotes building upgrades.

---

# 18.17 Citizen Health

Environmental conditions modify health.

Negative factors:

- Dirty Water
- Air Pollution
- Noise
- Ground Pollution

Positive factors:

- Parks
- Healthcare
- Clean Environment

Poor health increases healthcare demand.

---

# 18.18 Agriculture

Farming specialization requires fertile land.

Crop productivity depends on:

- Fertile soil
- Water availability
- Worker efficiency
- Policies

Fertile land remains permanently renewable.

---

# 18.19 Renewable Energy

Weather affects renewable production.

Wind Turbines

Output depends on:

- Wind Speed

Solar Plants

Output depends on:

- Daylight
- Weather

Hydroelectric Dams

Output depends on:

- Water Flow

Renewable production fluctuates continuously.

---

# 18.20 Environmental Policies

Policies include:

- Recycling
- Energy Conservation
- Smoke Detectors
- Combustion Engine Ban
- Electric Vehicle Promotion
- Sustainable Fishing (future expansion)

Policies influence pollution generation and resource consumption.

---

# 18.21 Environment Overlays

Players may inspect:

- Air Pollution
- Ground Pollution
- Noise
- Water Pollution
- Wind
- Natural Resources
- Temperature
- Land Value

Each overlay updates continuously with simulation changes.

---

# 18.22 Environmental Statistics

Tracked metrics include:

- Average Air Quality
- Water Quality
- Noise Exposure
- Forest Coverage
- Resource Extraction
- Pollution Production
- Renewable Energy Output
- Average Temperature

Statistics feed advisor panels and city reports.

---

# 18.23 Serialization

Saved environmental data includes:

- Pollution Maps
- Water State
- Wind Field
- Temperature
- Tree Growth
- Resource Depletion
- Weather State

Terrain data is stored separately from dynamic environmental simulation.

---

# 18.24 Performance Optimization

The Environment Manager uses:

- Chunk-based pollution maps
- Incremental diffusion
- Sparse update grids
- Cached wind fields
- Tree instancing
- Deferred environmental updates

Only regions affected by environmental changes are recalculated.

---

# 18.25 Design Goals

The Environment System is designed to provide:

- Realistic pollution simulation
- Dynamic weather
- Water interaction
- Renewable resource management
- Natural resource depletion
- Environmental consequences
- Renewable energy integration
- Terrain interaction
- Scalable simulation
- Efficient environmental processing

The Environment System ensures that the natural world remains an active participant in city development, rewarding sustainable planning while imposing meaningful consequences for pollution, overexploitation, and poor environmental management.

# 18. Environment, Pollution & Natural Resources

The Environment System simulates the interaction between terrain, vegetation, water, weather, pollution, climate, and natural resources. Environmental conditions directly influence citizen health, land value, industry, tourism, agriculture, and long-term city development.

Unlike purely visual systems, environmental simulation continuously affects gameplay through measurable consequences.

---

# 18.1 Environment Architecture

The Environment Manager coordinates all environmental subsystems.

```text
EnvironmentManager

├── Air Pollution
├── Ground Pollution
├── Noise Pollution
├── Water Pollution
├── Wind Simulation
├── Weather
├── Temperature
├── Natural Resources
├── Forest System
├── Water Simulation
└── Statistics
```

Environmental systems update independently while sharing terrain and simulation data.

---

# 18.2 Environmental Layers

The world contains multiple environmental overlays.

Supported layers include:

- Air Pollution
- Ground Pollution
- Noise Pollution
- Water Pollution
- Wind
- Temperature
- Ore
- Oil
- Fertile Land
- Forest Density

Each layer is stored independently.

---

# 18.3 Air Pollution

Air pollution is generated by:

- Coal Power Plants
- Oil Power Plants
- Generic Industry
- Specialized Industry
- Heavy Traffic
- Airports

Air pollution diffuses outward over time.

Consequences include:

- Reduced citizen health
- Lower land value
- Building abandonment
- Reduced desirability

---

# 18.4 Ground Pollution

Ground pollution affects terrain directly.

Sources:

- Industry
- Landfills
- Oil Extraction
- Ore Extraction
- Sewage Leakage
- Disaster Debris

Ground pollution dissipates slowly.

Residential zoning performs poorly in polluted areas.

---

# 18.5 Noise Pollution

Noise originates from:

- Busy Roads
- Highways
- Railways
- Airports
- Industrial Areas
- Entertainment Districts
- Stadiums

Noise decreases with distance.

High noise reduces:

- Residential happiness
- Land value
- Residential demand

Commercial districts tolerate higher noise levels.

---

# 18.6 Water Pollution

Water pollution enters rivers, lakes, and oceans through:

- Sewage Outlets
- Industrial Waste
- Flood Debris
- Landfill Leakage

Pollution follows water flow.

Consequences:

- Contaminated drinking water
- Increased illness
- Fish mortality
- Lower tourism
- Reduced waterfront land value

---

# 18.7 Pollution Propagation

Pollution spreads according to its type.

```text
Air

↓

Wind

Ground

↓

Terrain

Water

↓

Water Flow

Noise

↓

Distance
```

Each propagation model is simulated independently.

---

# 18.8 Wind Simulation

Wind direction influences:

- Air pollution movement
- Forest fire spread
- Tornado movement
- Wind turbine efficiency

Wind is represented as a continuous vector field across the map.

---

# 18.9 Temperature

Temperature affects:

- Heating demand
- Snow accumulation
- Citizen comfort
- Disaster likelihood (scenario dependent)

Cold maps require active heating infrastructure.

---

# 18.10 Weather

Supported weather conditions include:

- Clear
- Cloudy
- Rain
- Snow
- Fog
- Storm

Weather influences:

- Traffic
- Electricity production
- Tourism
- Citizen behavior

---

# 18.11 Water Simulation

Water is fully simulated.

Water sources include:

- Rivers
- Lakes
- Ocean
- Water Outlets
- Rainfall (scenario)

Water interacts with:

- Terrain elevation
- Dams
- Flood barriers
- Sewage
- Pollution

---

# 18.12 Forest System

Trees exist as simulation objects.

Functions:

- Improve land value
- Absorb pollution
- Support forestry
- Reduce noise
- Beautify the environment

Forests regenerate naturally if left undisturbed.

---

# 18.13 Tree Lifecycle

```text
Sapling

↓

Young Tree

↓

Mature Tree

↓

Old Tree

↓

Removed

or

Natural Regrowth
```

Tree growth rates vary by map type and climate.

---

# 18.14 Natural Resources

Supported resources:

- Ore
- Oil
- Fertile Land
- Forest

Each map defines unique resource distributions.

---

# 18.15 Resource Depletion

Finite resources:

- Ore
- Oil

Extraction reduces local deposits over time.

Renewable resources:

- Forestry
- Agriculture

These regenerate naturally through environmental processes.

---

# 18.16 Land Value

Environmental quality strongly affects land value.

Positive influences:

- Parks
- Clean Air
- Waterfront
- Trees
- Good Services

Negative influences:

- Pollution
- Noise
- Traffic
- Crime

Higher land value promotes building upgrades.

---

# 18.17 Citizen Health

Environmental conditions modify health.

Negative factors:

- Dirty Water
- Air Pollution
- Noise
- Ground Pollution

Positive factors:

- Parks
- Healthcare
- Clean Environment

Poor health increases healthcare demand.

---

# 18.18 Agriculture

Farming specialization requires fertile land.

Crop productivity depends on:

- Fertile soil
- Water availability
- Worker efficiency
- Policies

Fertile land remains permanently renewable.

---

# 18.19 Renewable Energy

Weather affects renewable production.

Wind Turbines

Output depends on:

- Wind Speed

Solar Plants

Output depends on:

- Daylight
- Weather

Hydroelectric Dams

Output depends on:

- Water Flow

Renewable production fluctuates continuously.

---

# 18.20 Environmental Policies

Policies include:

- Recycling
- Energy Conservation
- Smoke Detectors
- Combustion Engine Ban
- Electric Vehicle Promotion
- Sustainable Fishing (future expansion)

Policies influence pollution generation and resource consumption.

---

# 18.21 Environment Overlays

Players may inspect:

- Air Pollution
- Ground Pollution
- Noise
- Water Pollution
- Wind
- Natural Resources
- Temperature
- Land Value

Each overlay updates continuously with simulation changes.

---

# 18.22 Environmental Statistics

Tracked metrics include:

- Average Air Quality
- Water Quality
- Noise Exposure
- Forest Coverage
- Resource Extraction
- Pollution Production
- Renewable Energy Output
- Average Temperature

Statistics feed advisor panels and city reports.

---

# 18.23 Serialization

Saved environmental data includes:

- Pollution Maps
- Water State
- Wind Field
- Temperature
- Tree Growth
- Resource Depletion
- Weather State

Terrain data is stored separately from dynamic environmental simulation.

---

# 18.24 Performance Optimization

The Environment Manager uses:

- Chunk-based pollution maps
- Incremental diffusion
- Sparse update grids
- Cached wind fields
- Tree instancing
- Deferred environmental updates

Only regions affected by environmental changes are recalculated.

---

# 18.25 Design Goals

The Environment System is designed to provide:

- Realistic pollution simulation
- Dynamic weather
- Water interaction
- Renewable resource management
- Natural resource depletion
- Environmental consequences
- Renewable energy integration
- Terrain interaction
- Scalable simulation
- Efficient environmental processing

The Environment System ensures that the natural world remains an active participant in city development, rewarding sustainable planning while imposing meaningful consequences for pollution, overexploitation, and poor environmental management.

# 20. Save System, Serialization & Game State Management

The Save System is responsible for capturing the complete simulation state of the city and restoring it exactly as it existed. Every simulation entity, network, citizen, vehicle, economy, district, policy, weather condition, disaster, and runtime setting is serialized into a deterministic save format.

The system is designed to support cities containing millions of simulation objects while maintaining compatibility across future game updates and downloadable content.

---

# 20.1 Save Architecture

The Save Manager coordinates all save and load operations.

```text
SaveManager

├── Serializer
├── Version Manager
├── Compression
├── Chunk Writer
├── Integrity Checker
├── Asset Resolver
├── Mod Compatibility
├── Autosave Manager
└── Recovery System
```

The Save Manager never accesses simulation systems directly. Each manager is responsible for serializing its own state.

---

# 20.2 Save Philosophy

The save system follows these principles:

- Deterministic
- Versioned
- Incremental
- Fault Tolerant
- Platform Independent
- DLC Compatible
- Mod Aware

Loading the same save always restores the identical simulation state.

---

# 20.3 Save File Structure

```text
Header

↓

Metadata

↓

World Information

↓

Terrain

↓

Road Network

↓

Utilities

↓

Buildings

↓

Citizens

↓

Vehicles

↓

Economy

↓

Districts

↓

Policies

↓

Transport

↓

Industries

↓

Environment

↓

Disasters

↓

Statistics

↓

Mods

↓

Footer
```

Each section is serialized independently.

---

# 20.4 Save Metadata

Metadata stores:

- Save Name
- City Name
- Mayor Name
- Save Date
- Play Time
- Population
- Money
- Game Version
- DLC List
- Enabled Mods
- Thumbnail

Metadata is loaded without loading the full simulation.

---

# 20.5 World Serialization

World information includes:

- Map
- Climate
- Theme
- Seed
- Terrain Modifications
- Water State

Terrain is stored separately from gameplay entities.

---

# 20.6 Road Network

Serialized data includes:

- Nodes
- Segments
- Lanes
- Speed Limits
- Elevation
- Bridges
- Tunnels
- Decorations

Road graphs are reconstructed during loading.

---

# 20.7 Building Serialization

Every building stores:

```cpp
BuildingID

AssetID

Position

Rotation

Level

Occupancy

Health

Flags

Construction State

District

Historical Status
```

Simulation resumes exactly where it stopped.

---

# 20.8 Citizen Serialization

Each citizen stores:

- Identity
- Household
- Workplace
- Education
- Health
- Happiness
- Current Activity
- Current Destination
- Lifetime Statistics

Citizen appearance is regenerated from stored seeds.

---

# 20.9 Vehicle Serialization

Vehicles save:

- Position
- Lane
- Speed
- Destination
- Cargo
- Passengers
- Route
- AI State

Vehicles continue moving immediately after loading.

---

# 20.10 Economy Serialization

Economic data includes:

- Treasury
- Loans
- Tax Rates
- Budgets
- Weekly Income
- Weekly Expenses
- Resource Production
- Imports
- Exports

No financial calculations are recomputed during loading.

---

# 20.11 District Serialization

District data includes:

- Borders
- Colors
- Policies
- Taxes
- Specializations
- Statistics

District assignments remain unchanged.

---

# 20.12 Utility Networks

Serialized systems include:

- Electricity Graph
- Water Graph
- Sewage Graph
- Heating Graph

Connection graphs are validated during loading before simulation resumes.

---

# 20.13 Public Transport

Transport saves include:

- Lines
- Stops
- Vehicles
- Passenger Queues
- Depot Assignments
- Ticket Revenue

Transport resumes without resetting routes.

---

# 20.14 Industries

Industry state includes:

- Inventories
- Warehouses
- Active Deliveries
- Production Chains
- Statistics

Factories continue production immediately after loading.

---

# 20.15 Environment

Environmental state includes:

- Pollution Maps
- Weather
- Temperature
- Wind
- Flood Water
- Tree Growth
- Resource Depletion

Dynamic environmental simulations resume from their previous state.

---

# 20.16 Disaster State

If disasters are active:

Save includes:

- Disaster Progress
- Damage
- Emergency Requests
- Shelters
- Recovery Progress

Ongoing disasters continue after loading.

---

# 20.17 Statistics

Persistent statistics include:

- Population History
- Financial Graphs
- Traffic History
- Pollution History
- Happiness Trends
- Achievement Progress

Historical graphs remain intact.

---

# 20.18 Serialization Strategy

Each manager implements a common interface.

```cpp
interface ISerializable
{
    Save();

    Load();

    Validate();

    UpgradeVersion();
}
```

The Save Manager orchestrates execution order.

---

# 20.19 Version Compatibility

Every save stores:

```text
Major Version

Minor Version

Patch Version

Build Number
```

Compatibility rules:

- Same version → Direct load
- Minor upgrade → Automatic migration
- Major upgrade → Conversion pipeline

---

# 20.20 DLC Compatibility

Each save records enabled DLC.

Missing DLC behavior:

- Unsupported buildings become inactive.
- Missing assets display placeholders.
- Simulation data remains intact.

Re-enabling DLC restores full functionality.

---

# 20.21 Mod Compatibility

Save data stores:

- Mod IDs
- Mod Versions
- Asset References

Missing mods trigger compatibility warnings before loading.

Simulation continues whenever possible.

---

# 20.22 Autosave System

Autosaves occur:

- Fixed intervals
- Manual triggers
- Before disasters (optional)
- Before quitting (optional)

Old autosaves rotate automatically according to user settings.

---

# 20.23 Compression

Large cities generate substantial save data.

Compression pipeline:

```text
Serialize

↓

Chunk

↓

Compress

↓

Write

↓

Verify
```

Compression is transparent to gameplay.

---

# 20.24 Integrity Validation

Every save performs validation.

Checks include:

- Checksums
- Section Sizes
- Version Compatibility
- Missing Assets
- Invalid References

Corrupted sections generate recovery attempts before failing.

---

# 20.25 Incremental Saving

To reduce save times:

Only modified systems may be rewritten internally before final save generation.

Examples:

- Economy
- Vehicles
- Citizens
- Buildings

Unchanged data may reuse cached serialization buffers.

---

# 20.26 Multithreaded Saving

Saving occurs across worker threads.

Example:

```text
Thread 1

Buildings

Thread 2

Citizens

Thread 3

Vehicles

Thread 4

Economy

↓

Merge

↓

Compression

↓

Disk
```

Gameplay may continue during autosaves depending on platform capabilities.

---

# 20.27 Recovery System

If a save operation fails:

```text
Temporary Save

↓

Validation

↓

Replace Original
```

Original saves are never overwritten until validation succeeds.

---

# 20.28 Cloud Saves

Optional cloud synchronization stores:

- Save Files
- Metadata
- Screenshots
- Version Information

Synchronization occurs after successful local validation.

---

# 20.29 Performance Optimization

The Save Manager employs:

- Binary serialization
- Chunked writing
- Parallel serialization
- Compression
- Streaming I/O
- Shared asset references
- Delta caching

Large cities remain saveable without excessive loading times.

---

# 20.30 Design Goals

The Save & Serialization System is designed to provide:

- Exact simulation restoration
- Deterministic save files
- Fast loading
- Efficient storage
- Version compatibility
- DLC compatibility
- Mod compatibility
- Reliable autosaves
- Fault tolerance
- Scalable performance

The Save System guarantees that every aspect of the city—from the position of a single citizen to the state of an ongoing disaster—can be restored precisely, ensuring consistent gameplay across sessions, updates, and hardware platforms.

# 21. Rendering Engine & Graphics Pipeline

The Rendering System is responsible for transforming the simulation into a visually rich, scalable, and performant city. It renders millions of objects, dynamic lighting, terrain, weather, water, vegetation, shadows, reflections, particles, and user interface while maintaining high frame rates across cities of every size.

Rendering is completely separated from simulation. The renderer only visualizes the current simulation state and never modifies gameplay.

---

# 21.1 Rendering Architecture

The Render Manager coordinates all rendering subsystems.

```text
RenderManager

├── Camera System
├── Terrain Renderer
├── Building Renderer
├── Road Renderer
├── Vegetation Renderer
├── Water Renderer
├── Vehicle Renderer
├── Citizen Renderer
├── Lighting System
├── Shadow System
├── Weather Renderer
├── Particle System
├── UI Renderer
├── Post Processing
└── GPU Resource Manager
```

Each renderer consumes immutable simulation snapshots generated during simulation ticks.

---

# 21.2 Rendering Pipeline

The graphics pipeline executes in multiple stages.

```text
Simulation Snapshot

↓

Visibility Culling

↓

LOD Selection

↓

Render Queue

↓

Shadow Pass

↓

Geometry Pass

↓

Lighting Pass

↓

Transparent Pass

↓

Particles

↓

Post Processing

↓

UI

↓

Present Frame
```

Simulation updates are decoupled from frame rendering.

---

# 21.3 Camera System

Supported camera controls include:

- Free Camera
- Orbit
- Zoom
- Pan
- Rotate
- Cinematic Camera
- First-Person (Developer)

Camera properties:

```cpp
Position

Rotation

FOV

Near Plane

Far Plane

Exposure
```

The camera never influences simulation.

---

# 21.4 Terrain Rendering

Terrain consists of:

- Heightmap
- Terrain Materials
- Terrain Blending
- Grass
- Cliffs
- Beaches
- Snow
- Roads

Terrain is rendered using chunked meshes.

---

# 21.5 Terrain Chunks

The map is divided into chunks.

```text
Terrain

↓

Chunk

↓

Mesh

↓

GPU Buffer
```

Only visible chunks are rendered.

---

# 21.6 Building Rendering

Buildings use:

- High Detail Mesh
- Medium LOD
- Low LOD
- Billboard (Very Long Distance)

LOD selection depends on camera distance.

---

# 21.7 GPU Instancing

Repeated assets share GPU resources.

Examples:

- Trees
- Street Lights
- Benches
- Utility Poles
- Cars
- Citizens

Thousands of identical assets are rendered using a single draw call where possible.

---

# 21.8 Road Rendering

Roads are generated procedurally.

Components include:

- Asphalt
- Lane Markings
- Sidewalks
- Curbs
- Medians
- Decorations

Road meshes regenerate only when edited.

---

# 21.9 Water Rendering

Water rendering supports:

- Reflections
- Refraction
- Waves
- Shoreline Foam
- Transparency
- Dynamic Water Level

Rendering remains synchronized with the water simulation.

---

# 21.10 Vegetation Rendering

Vegetation includes:

- Trees
- Bushes
- Grass
- Crops

Wind animation affects vegetation movement without changing simulation.

---

# 21.11 Vehicle Rendering

Vehicle rendering includes:

- LOD Models
- Wheel Animation
- Brake Lights
- Turn Signals
- Headlights
- Shadows

Rendering interpolates positions between simulation updates for smooth motion.

---

# 21.12 Citizen Rendering

Citizens use:

- Skeletal Animation
- Clothing Variations
- Hair Variations
- Accessory Variations
- Walking Animation
- Idle Animation

Visual diversity is generated procedurally.

---

# 21.13 Lighting

Lighting supports:

- Directional Sun
- Ambient Light
- Point Lights
- Spot Lights
- Vehicle Lights
- Building Lights

Lighting changes dynamically based on the day/night cycle.

---

# 21.14 Shadow System

Shadow sources include:

- Sun
- Buildings
- Trees
- Vehicles

Techniques include:

- Cascaded Shadow Maps
- Distance-Based Shadow LOD
- Contact Shadows

Far objects may render simplified shadows.

---

# 21.15 Day/Night Rendering

Time of day affects:

- Sky Color
- Sun Position
- Shadows
- Building Lights
- Vehicle Lights
- Street Lights
- Ambient Lighting

The lighting system transitions smoothly throughout the day.

---

# 21.16 Weather Rendering

Visual weather effects include:

- Rain
- Snow
- Fog
- Clouds
- Lightning
- Wet Roads

Weather rendering is synchronized with environmental simulation.

---

# 21.17 Particle System

Particles support:

- Smoke
- Fire
- Dust
- Steam
- Rain
- Snow
- Water Splashes
- Industrial Exhaust
- Explosions

Particles are GPU-accelerated where supported.

---

# 21.18 Reflection System

Reflections include:

- Water Reflections
- Glass Reflections
- Wet Surface Reflections

Reflection quality scales with graphics settings.

---

# 21.19 Post Processing

Supported effects include:

- HDR
- Bloom
- Ambient Occlusion
- Tone Mapping
- Motion Blur
- Depth of Field
- Color Grading
- Anti-Aliasing

Players may enable or disable individual effects.

---

# 21.20 Occlusion Culling

Objects hidden behind large geometry are skipped.

Workflow:

```text
Camera

↓

Visibility Test

↓

Occlusion Test

↓

Visible Objects

↓

Render
```

Invisible objects generate no draw calls.

---

# 21.21 Frustum Culling

Objects outside the camera frustum are ignored.

This significantly reduces rendering workload in large cities.

---

# 21.22 Level of Detail (LOD)

Every asset supports multiple LOD levels.

```text
LOD0

↓

LOD1

↓

LOD2

↓

LOD3

↓

Billboard
```

LOD transitions use fading to minimize popping.

---

# 21.23 Texture Streaming

Textures are streamed based on camera distance.

Benefits:

- Lower VRAM usage
- Faster loading
- Higher quality nearby assets

Unused textures are released when no longer needed.

---

# 21.24 Render Statistics

Tracked metrics include:

- FPS
- Frame Time
- Draw Calls
- Visible Objects
- GPU Memory
- CPU Time
- GPU Time
- Active Lights
- Shadow Casters

Developer overlays display these metrics in real time.

---

# 21.25 Graphics Settings

Supported options include:

- Resolution
- Render Scale
- Texture Quality
- Shadow Quality
- LOD Distance
- Water Quality
- Reflection Quality
- Anti-Aliasing
- Post Processing
- V-Sync
- Frame Rate Limit

Settings can be adjusted without restarting where supported.

---

# 21.26 Screenshot System

Players can capture:

- Standard Screenshot
- High Resolution Screenshot
- UI Hidden Screenshot
- Cinematic Screenshot

Screenshots do not interrupt simulation.

---

# 21.27 Serialization

Rendering data is generally not serialized.

Only persistent settings are saved:

- Graphics Options
- Camera Position (Optional)
- Photo Mode Settings
- UI Preferences

All render resources are regenerated when loading a save.

---

# 21.28 Performance Optimization

The Rendering Engine employs:

- GPU Instancing
- Occlusion Culling
- Frustum Culling
- Asynchronous Asset Streaming
- Texture Streaming
- LOD Rendering
- Deferred Rendering
- Dynamic Batching
- Multi-threaded Command Generation

Rendering performance scales efficiently from small towns to massive metropolitan regions.

---

# 21.29 Design Goals

The Rendering Engine is designed to provide:

- High visual fidelity
- Massive city rendering
- Smooth frame rates
- Efficient GPU utilization
- Dynamic weather visuals
- Day/night transitions
- Large-scale vegetation rendering
- Modern graphics features
- Scalable quality settings
- Complete separation from simulation

The Rendering Engine transforms the underlying simulation into a living, visually immersive city, accurately representing millions of simulation events while maintaining the performance necessary for real-time city management.

# 22. Game Progression, Milestones & Unlock System

The Progression System governs how the city evolves over time by rewarding population growth with new buildings, services, policies, transportation options, infrastructure, and gameplay mechanics. Rather than exposing every system immediately, the game gradually unlocks features through deterministic milestones, allowing the player to expand naturally while learning increasingly complex management systems.

The progression system is completely data-driven, enabling custom scenarios, sandbox modes, and future expansions to redefine progression without changing core game code.

---

# 22.1 Progression Architecture

The Progression Manager controls all progression-related systems.

```text
ProgressionManager

├── Population Tracker
├── Milestone Manager
├── Unlock Registry
├── Reward System
├── XP & Statistics
├── Advisor System
├── Achievement Tracker
├── Tutorial Progress
└── Scenario Progress
```

Progression never directly modifies simulation systems. It simply enables additional gameplay features.

---

# 22.2 Population Milestones

The primary progression metric is total population.

Example progression:

```text
Tiny Village

↓

Village

↓

Small Town

↓

Town

↓

Boom Town

↓

Small City

↓

Big Town

↓

Large City

↓

Capital City

↓

Metropolis

↓

Megalopolis
```

Each milestone unlocks new gameplay systems.

---

# 22.3 Milestone Definition

Each milestone contains:

```cpp
struct Milestone
{
    MilestoneID id;

    string name;

    uint64 populationRequirement;

    UnlockList unlocks;

    Rewards rewards;

    AdvisorMessages messages;
};
```

Milestones are evaluated whenever city population changes.

---

# 22.4 Unlock Categories

Unlockable content includes:

Infrastructure

- Roads
- Highways
- Bridges
- Tunnels

Services

- Police
- Fire
- Healthcare
- Education

Utilities

- Power
- Water
- Heating
- Waste Processing

Transportation

- Bus
- Metro
- Train
- Airport
- Harbor

Zoning

- High Density Residential
- High Density Commercial
- Offices

Economy

- Loans
- Tax Controls
- District Policies

---

# 22.5 Initial Gameplay

A new city begins with limited functionality.

Available:

- Basic Roads
- Low Density Residential
- Low Density Commercial
- Industry
- Electricity
- Water
- Sewage

Unavailable systems remain hidden until unlocked.

---

# 22.6 Unlock Process

Whenever population reaches a milestone:

```text
Population Increased

↓

Milestone Reached

↓

Rewards Granted

↓

Menus Updated

↓

Advisor Notification

↓

Player Continues
```

Unlocks occur immediately without requiring manual activation.

---

# 22.7 Rewards

Milestones may grant:

- Construction Budget
- Additional Tiles
- New Buildings
- New Policies
- New Roads
- Additional Loans
- Public Transport
- Landscaping Tools

Rewards are applied only once.

---

# 22.8 Construction Budget Rewards

Population milestones often grant direct cash.

Purpose:

- Support city expansion
- Offset infrastructure costs
- Encourage experimentation

Reward values are data-driven.

---

# 22.9 Additional Map Tiles

Milestones unlock purchasable land.

Workflow:

```text
Milestone

↓

Tile Purchase Enabled

↓

Player Selects Tile

↓

Terrain Added

↓

Simulation Expanded
```

Tile ownership persists permanently.

---

# 22.10 Feature Visibility

Locked systems remain hidden.

Examples:

Before unlocking Metro:

- Metro toolbar hidden
- Metro buildings unavailable
- Metro tutorial hidden

After unlocking:

- Full interface becomes available.

---

# 22.11 Advisors

Advisors guide progression.

Advisor notifications include:

- New Services
- New Roads
- Budget Suggestions
- Education Warnings
- Traffic Advice
- Pollution Advice

Advisors react dynamically to city conditions.

---

# 22.12 Unlock Dependencies

Some unlocks require:

Population

AND

Specific Service

Example:

Airport

Requires:

- Population milestone
- Outside connection
- Sufficient city finances

---

# 22.13 Progress Tracking

Players can inspect:

- Current Milestone
- Next Milestone
- Remaining Population
- Upcoming Unlocks
- Earned Rewards

Progress updates continuously.

---

# 22.14 Sandbox Mode

Sandbox mode disables progression restrictions.

All systems become immediately available:

- Every road
- Every service
- Every transport mode
- Every policy
- Every landscaping tool

Milestones remain available for statistics only.

---

# 22.15 Scenario Progression

Scenarios define custom progression.

Possible triggers:

- Population
- Time
- Objectives
- Budget
- Disaster Completion
- Resource Production

Scenarios may ignore normal milestone rules.

---

# 22.16 Research (Expansion Support)

Future expansions may introduce research trees.

Potential categories:

- Transportation
- Industry
- Sustainability
- Healthcare
- Education
- Technology

Research integrates with existing milestone progression.

---

# 22.17 Achievement Integration

Progression contributes toward achievements.

Examples:

- Reach population milestone
- Unlock every transport mode
- Unlock every policy
- Complete all milestones

Achievements remain independent of gameplay unlocks.

---

# 22.18 Tutorial Integration

Tutorials appear when systems unlock.

Examples:

Bus Unlocked

↓

Bus Tutorial

Metro Unlocked

↓

Metro Tutorial

Districts Unlocked

↓

District Painting Tutorial

Tutorials may be skipped.

---

# 22.19 Statistics

Tracked progression metrics include:

- Current Milestone
- Highest Population
- Total Unlocks
- Milestones Completed
- Rewards Earned
- Total Expansion Tiles
- Tutorial Completion

Statistics persist across saves.

---

# 22.20 Serialization

Saved progression data includes:

- Current Milestone
- Unlock Flags
- Rewards Claimed
- Advisor State
- Tutorial State
- Purchased Tiles

Progress resumes exactly after loading.

---

# 22.21 Performance Optimization

The Progression Manager uses:

- Event-driven milestone evaluation
- Cached unlock registry
- Incremental population tracking
- Data-driven reward tables
- Lazy UI updates

Milestone evaluation occurs only when relevant city metrics change.

---

# 22.22 Design Goals

The Progression System is designed to provide:

- Gradual gameplay expansion
- Clear player goals
- Structured learning curve
- Meaningful rewards
- Flexible sandbox support
- Scenario compatibility
- Data-driven unlocks
- Seamless tutorial integration
- Long-term progression
- Future expansion support

The Progression System ensures that city growth feels rewarding and manageable, introducing increasingly sophisticated mechanics at the appropriate time while supporting unrestricted sandbox gameplay for experienced players.

# 23. Economy, Budget & Financial Simulation

The Economy System governs every financial transaction within the city. Every tax payment, salary, maintenance cost, loan, import, export, construction expense, service budget, transport ticket, industrial production, and disaster recovery cost contributes to the city's financial state.

Unlike simple resource counters, the economy is driven by millions of simulation events generated by citizens, businesses, industries, services, transportation, and government operations.

---

# 23.1 Economy Architecture

The Economy Manager coordinates every financial subsystem.

```text
EconomyManager

├── Treasury
├── Tax System
├── Budget Manager
├── Loan Manager
├── Construction Costs
├── Maintenance Costs
├── Service Expenses
├── Industry Revenue
├── Transport Revenue
├── Import/Export Economy
├── Weekly Accounting
└── Financial Statistics
```

All city managers report financial events through the Economy Manager.

---

# 23.2 Treasury

The treasury stores the city's available money.

```cpp
struct Treasury
{
    int64 CurrentBalance;

    int64 WeeklyIncome;

    int64 WeeklyExpenses;

    int64 LifetimeIncome;

    int64 LifetimeExpenses;
};
```

All transactions update the treasury immediately.

---

# 23.3 Income Sources

City income comes from:

Residential Taxes

Commercial Taxes

Industrial Taxes

Office Taxes

Transport Tickets

Industry Production

Exports

Tourism

Park Revenue

Special Buildings

Scenario Rewards

Grant Payments

---

# 23.4 Expense Sources

City expenses include:

Building Maintenance

Road Maintenance

Vehicle Maintenance

Public Transport

Electricity Production

Water Production

Healthcare

Education

Police

Fire Department

Garbage Collection

Loans

Disaster Recovery

Land Purchases

Construction

---

# 23.5 Weekly Economy Cycle

The economy updates every simulation week.

```text
Collect Taxes

↓

Calculate Expenses

↓

Industry Income

↓

Transport Revenue

↓

Import Costs

↓

Loan Payments

↓

Update Treasury

↓

Generate Financial Report
```

Weekly reports provide the primary financial overview.

---

# 23.6 Tax System

Each zoning type supports independent taxation.

Residential

Commercial

Industrial

Office

Taxes are adjustable within allowed limits.

---

# 23.7 Tax Calculation

Revenue depends upon:

```text
Building Value

×

Tax Rate

×

Occupancy

×

Economic Modifier

=

Tax Revenue
```

Higher-value buildings contribute more tax income.

---

# 23.8 Tax Effects

Increasing taxes may result in:

- Reduced happiness
- Slower growth
- Building abandonment
- Reduced demand

Lower taxes encourage growth but reduce revenue.

---

# 23.9 Budget System

Each city service has an independent operating budget.

Supported budgets include:

Police

Fire

Healthcare

Education

Garbage

Public Transport

Road Maintenance

Parks

Utilities

---

# 23.10 Budget Effects

Budget changes influence:

- Service Capacity
- Vehicle Count
- Operating Hours
- Response Speed
- Maintenance Quality

Budget changes do not instantly alter simulation outcomes but take effect over time.

---

# 23.11 Construction Costs

Every construction project has:

Construction Cost

Construction Time

Maintenance Cost

Demolition Cost

Construction costs are deducted immediately upon placement.

---

# 23.12 Maintenance

Every operational asset incurs recurring maintenance.

Examples:

Road

↓

Weekly Maintenance

Power Plant

↓

Fuel + Maintenance

School

↓

Teacher Costs

Vehicle Depot

↓

Fleet Maintenance

Maintenance scales with city size.

---

# 23.13 Loans

Players may borrow money.

Each loan stores:

```cpp
LoanID

Principal

InterestRate

WeeklyPayment

RemainingBalance
```

Loans increase available funds while creating future obligations.

---

# 23.14 Loan Repayment

Repayment occurs automatically.

```text
Weekly Income

↓

Loan Payment

↓

Treasury Updated
```

Loans may also be repaid early.

---

# 23.15 Industry Revenue

Industries generate income through:

Resource Extraction

↓

Processing

↓

Manufacturing

↓

Exports

↓

Tax Revenue

Industry profitability depends upon logistics and workforce availability.

---

# 23.16 Import Costs

When local production cannot satisfy demand:

```text
Outside Region

↓

Import

↓

City Pays Import Cost
```

Heavy imports reduce overall profitability.

---

# 23.17 Export Revenue

Surplus production generates export income.

Exports include:

- Raw Resources
- Processed Goods
- Manufactured Products

Efficient transportation increases export profitability.

---

# 23.18 Public Transport Economy

Revenue sources:

Passenger Tickets

Tourism

Operating expenses include:

Vehicles

Fuel

Electricity

Maintenance

Staff

Transport systems may operate at a profit or loss depending on usage.

---

# 23.19 Tourism Economy

Tourists contribute through:

Hotels

Commercial Shopping

Parks

Unique Buildings

Transport

Entertainment

Tourism increases commercial tax revenue.

---

# 23.20 Disaster Costs

Disasters generate expenses including:

Emergency Response

Road Repair

Utility Restoration

Building Reconstruction

Healthcare

Shelter Operations

Disasters may temporarily create budget deficits.

---

# 23.21 Financial Policies

Policies modify economic behavior.

Examples:

- Free Public Transport
- Recycling
- Small Business Support
- Industry 4.0
- High-Tech Housing

Policies increase expenses while providing indirect benefits.

---

# 23.22 Deficit Handling

If expenses exceed income:

Effects include:

- Negative treasury
- Loan recommendations
- Reduced construction capability
- Financial advisor warnings

The simulation continues unless scenario rules specify bankruptcy.

---

# 23.23 Bankruptcy (Scenario Support)

Scenarios may define bankruptcy conditions.

Example:

Treasury < 0

AND

Unable to obtain loans

↓

Scenario Failure

Sandbox mode ignores bankruptcy.

---

# 23.24 Financial Statistics

Tracked metrics include:

Weekly Income

Weekly Expenses

Net Profit

Tax Revenue

Transport Revenue

Tourism Revenue

Industry Revenue

Import Costs

Export Revenue

Maintenance Costs

Loans

Treasury Balance

Lifetime Revenue

Lifetime Expenses

---

# 23.25 Financial Graphs

Historical graphs display:

- Income
- Expenses
- Profit
- Population
- Service Costs
- Transport Usage
- Tax Revenue

Players may inspect trends across multiple time scales.

---

# 23.26 Advisors

Financial advisors notify players about:

- Budget deficits
- High taxes
- Low tax income
- Rising maintenance
- Expensive services
- Loan opportunities

Advisor recommendations are generated dynamically.

---

# 23.27 Serialization

Saved economy state includes:

- Treasury
- Loans
- Budgets
- Tax Rates
- Weekly Reports
- Historical Graphs
- Industry Revenue
- Transport Revenue

Financial history is preserved between sessions.

---

# 23.28 Performance Optimization

The Economy Manager employs:

- Event-driven accounting
- Cached tax calculations
- Incremental budget updates
- Weekly aggregation
- Batched financial reports
- Shared revenue pipelines

Most financial calculations occur during scheduled accounting cycles rather than every simulation tick.

---

# 23.29 Design Goals

The Economy System is designed to provide:

- Realistic municipal budgeting
- Dynamic taxation
- Meaningful financial decisions
- Industry-driven income
- Service-based expenses
- Long-term economic planning
- Historical financial analysis
- Scenario compatibility
- Scalable accounting
- Deterministic financial simulation

The Economy System serves as the financial backbone of the city, connecting every gameplay system into a unified municipal budget where infrastructure, services, transportation, industry, and citizen activity collectively determine the long-term prosperity and sustainability of the city.

# 24. User Interface, Information Views & Player Interaction

The User Interface (UI) System provides every tool required for city construction, management, monitoring, and analysis. It presents simulation data without modifying simulation behavior and acts as the primary communication layer between the player and the city.

The UI is fully event-driven, modular, scalable, keyboard/controller compatible, and independent from rendering and simulation systems.

---

# 24.1 UI Architecture

The UI Manager coordinates every interface subsystem.

```text
UIManager

├── HUD
├── Main Toolbar
├── Build Menus
├── Information Views
├── Inspector Panels
├── Statistics
├── Notifications
├── Advisors
├── Tool System
├── Overlay Manager
├── Dialog Manager
└── Input Manager
```

UI components subscribe to simulation events rather than polling continuously.

---

# 24.2 HUD

The Heads-Up Display continuously presents critical city information.

Displayed information includes:

- Treasury
- Weekly Income
- Population
- Happiness
- Current Date
- Current Time
- Simulation Speed
- Active Milestone
- Notifications

HUD elements update automatically whenever relevant simulation values change.

---

# 24.3 Main Toolbar

The toolbar provides access to every gameplay system.

Categories include:

- Roads
- Zoning
- Districts
- Electricity
- Water & Sewage
- Garbage
- Healthcare
- Fire & Rescue
- Police
- Education
- Public Transport
- Landscaping
- Parks
- Economy
- Policies
- Statistics
- Options

Unavailable tools remain hidden until unlocked.

---

# 24.4 Build Menus

Each build category contains:

- Asset List
- Search
- Filters
- Favorites
- Recently Used
- DLC Indicators
- Mod Indicators
- Asset Preview

Assets display:

- Cost
- Maintenance
- Size
- Requirements
- Unlock Conditions

---

# 24.5 Tool System

Supported tools include:

- Place
- Bulldoze
- Upgrade
- Move
- Replace
- Paint
- Measure
- Inspect

Only one tool may remain active at a time.

---

# 24.6 Placement Preview

Before confirming construction:

Preview displays:

- Footprint
- Cost
- Terrain Conflicts
- Road Connections
- Utility Connections
- Collision Warnings
- Elevation

Invalid placements display immediate visual feedback.

---

# 24.7 Inspector Panels

Selecting an object opens its information panel.

Supported inspectors:

- Building
- Citizen
- Vehicle
- District
- Road
- Utility
- Park
- Industry
- Public Transport Line

Inspectors expose simulation state without modifying it.

---

# 24.8 Building Information Panel

Displays:

- Building Name
- Level
- Occupancy
- Workers
- Visitors
- Maintenance
- Electricity
- Water
- Service Coverage
- Warnings
- Statistics

Context-sensitive actions are also available.

---

# 24.9 Citizen Information Panel

Displays:

- Name
- Age
- Education
- Home
- Workplace
- Happiness
- Health
- Current Activity
- Travel Route
- Vehicle
- Lifetime Statistics

Players can follow individual citizens.

---

# 24.10 Vehicle Information Panel

Displays:

- Vehicle Type
- Speed
- Destination
- Route
- Passengers
- Cargo
- Owner
- Current State

Useful for diagnosing transportation issues.

---

# 24.11 Information Views

Information Views visualize hidden simulation data.

Supported overlays include:

- Electricity
- Water
- Sewage
- Heating
- Garbage
- Healthcare
- Fire Safety
- Crime
- Education
- Happiness
- Land Value
- Pollution
- Noise
- Traffic
- Public Transport
- Wind
- Natural Resources
- Districts

Only one information view is active unless explicitly combined.

---

# 24.12 Overlay System

Overlays are layered independently.

Examples:

```text
Terrain

↓

Roads

↓

Buildings

↓

Heatmap

↓

Icons

↓

Labels

↓

Selection
```

Heatmaps never alter gameplay.

---

# 24.13 Notifications

The Notification Manager reports simulation events.

Examples:

- No Electricity
- No Water
- Crime
- Fire
- Death Wave
- Traffic Congestion
- Building Abandoned
- Budget Deficit
- Milestone Reached

Notifications persist until resolved or dismissed.

---

# 24.14 Advisor System

City advisors provide contextual recommendations.

Advisor categories:

- Economy
- Transportation
- Utilities
- Healthcare
- Education
- Industry
- Environment

Advisors generate recommendations from live simulation data.

---

# 24.15 Statistics Window

Provides detailed reports for:

- Population
- Economy
- Traffic
- Industry
- Public Transport
- Pollution
- Education
- Health
- Crime
- Tourism
- Utilities

Historical graphs support multiple time scales.

---

# 24.16 Search System

Search supports:

- Buildings
- Streets
- Districts
- Public Transport Lines
- Citizens
- Services

Search results center the camera on the selected entity.

---

# 24.17 Selection System

Players may select:

- Buildings
- Roads
- Citizens
- Vehicles
- Districts
- Trees
- Props

Selection highlights the target and displays contextual actions.

---

# 24.18 Camera Controls

Supported controls include:

- Pan
- Rotate
- Zoom
- Focus
- Follow Citizen
- Follow Vehicle
- Reset Camera

Camera transitions are smoothly interpolated.

---

# 24.19 Time Controls

Simulation controls include:

- Pause
- Normal Speed
- Double Speed
- Triple Speed

Simulation speed affects update frequency but not deterministic outcomes.

---

# 24.20 Undo Support

Construction actions support limited undo.

Examples:

- Road Placement
- Zoning
- Demolition
- Landscaping

Undo history is cleared after loading a save unless explicitly preserved.

---

# 24.21 Keyboard Shortcuts

Supported shortcut categories:

- Build Tools
- Camera
- Time Controls
- Information Views
- Search
- Statistics
- Screenshots

Bindings are fully configurable.

---

# 24.22 Controller Support

Controllers support:

- Cursor Movement
- Radial Menus
- Camera Controls
- Building Placement
- Selection
- UI Navigation

UI automatically adapts to the active input device.

---

# 24.23 Accessibility

Accessibility options include:

- Adjustable UI Scale
- Color-Blind Modes
- High Contrast
- Subtitle Support
- Keyboard Navigation
- Remappable Controls
- Reduced Motion
- Screen Reader Metadata (platform support)

Accessibility settings are independent of gameplay.

---

# 24.24 Localization

All UI text is localized.

Supported content includes:

- Menus
- Tooltips
- Advisor Messages
- Notifications
- Tutorials
- Statistics
- Error Messages

Localization updates dynamically without restarting.

---

# 24.25 UI Serialization

Persistent UI state includes:

- Window Positions
- Panel Visibility
- Active Filters
- Favorite Assets
- Camera Preferences
- Information View Preferences
- UI Scale

Transient windows are not restored unless explicitly configured.

---

# 24.26 Performance Optimization

The UI Manager employs:

- Event-driven updates
- Virtualized lists
- Lazy window creation
- Cached graph rendering
- Batched notifications
- Incremental statistics refresh

Only visible interface components receive update events.

---

# 24.27 Design Goals

The User Interface System is designed to provide:

- Immediate access to all city management tools
- Rich inspection of simulation entities
- Real-time visualization of hidden systems
- Responsive controls
- Scalable layouts
- Accessibility
- Efficient rendering
- Complete keyboard and controller support
- Modular architecture
- Full separation from simulation

The User Interface serves as the player's command center, exposing every aspect of the city's simulation through intuitive tools, detailed analytics, and responsive controls while remaining performant even in the largest metropolitan cities.

# 25. Audio System & Ambient Simulation

The Audio System provides spatial, dynamic, and event-driven sound throughout the city. Rather than simply playing background music, it reflects the state of the simulation by reacting to traffic, citizens, weather, disasters, construction, industries, public transport, wildlife, and environmental conditions.

Audio is entirely presentation-layer functionality and never affects gameplay simulation.

---

# 25.1 Audio Architecture

The Audio Manager coordinates every sound source.

```text
AudioManager

├── Music System
├── Ambient System
├── Environmental Audio
├── Vehicle Audio
├── Citizen Audio
├── Building Audio
├── Disaster Audio
├── UI Audio
├── Spatial Audio
├── Voice System
└── Audio Mixer
```

Each subsystem independently contributes audio events to the mixer.

---

# 25.2 Audio Categories

Supported sound categories include:

Music

Environment

Weather

Buildings

Construction

Traffic

Public Transport

Citizens

Animals

Industry

Utilities

Emergency Services

Disasters

UI

Voice

---

# 25.3 Music System

Music plays independently of simulation.

Features:

- Dynamic Playlist
- Shuffle
- Repeat
- DLC Radio Stations
- Custom Music Support
- Volume Control

Music continues seamlessly during gameplay.

---

# 25.4 Ambient Audio

Ambient sound depends on location.

Examples:

Residential

- Birds
- Wind
- Children

Commercial

- Crowds
- Store Activity
- Traffic

Industrial

- Machinery
- Heavy Equipment
- Steam

Parks

- Nature
- Wildlife
- Water

Ambient transitions smoothly while moving the camera.

---

# 25.5 Environmental Audio

Environment generates:

- Wind
- Ocean Waves
- Rivers
- Rain
- Thunder
- Snow
- Trees Rustling

Audio intensity scales with environmental conditions.

---

# 25.6 Weather Audio

Weather affects the soundscape.

Rain

- Rainfall
- Roof Impact

Storm

- Thunder
- Strong Wind

Snow

- Dampened Environment

Fog

- Reduced Ambient Activity

Weather transitions blend continuously.

---

# 25.7 Building Audio

Operational buildings emit sounds.

Examples:

Power Plants

- Turbines
- Generators

Factories

- Machinery
- Compressors

Schools

- Bells
- Children

Hospitals

- HVAC
- Ambulances

Building volume depends on distance.

---

# 25.8 Construction Audio

Construction produces:

- Excavation
- Jackhammers
- Cranes
- Welding
- Demolition
- Heavy Machinery

Construction sounds stop immediately after completion.

---

# 25.9 Vehicle Audio

Vehicles emit:

Cars

- Engines
- Tires

Buses

- Engines
- Doors

Trains

- Rail Noise
- Horns

Aircraft

- Jet Engines

Ships

- Horns
- Water Movement

Emergency Vehicles

- Sirens

Vehicle audio follows the moving entity.

---

# 25.10 Citizen Audio

Citizens generate:

- Walking
- Conversations
- Crowds
- Cheering
- Complaints

Individual conversations are abstracted rather than fully voiced.

---

# 25.11 Public Transport Audio

Transport stations produce:

- Passenger Announcements
- Vehicle Arrivals
- Doors
- Boarding
- Departure Signals

Busy stations become noticeably louder.

---

# 25.12 Disaster Audio

Disasters generate dynamic audio.

Earthquake

- Rumbles
- Structural Collapse

Tornado

- Wind
- Debris

Fire

- Flames
- Explosions

Meteor

- Atmospheric Entry
- Impact

Flood

- Water Movement
- Structural Damage

Disaster sounds fade naturally as recovery progresses.

---

# 25.13 Emergency Audio

Emergency response includes:

- Fire Engine Sirens
- Ambulance Sirens
- Police Sirens
- Rescue Helicopters

Priority vehicles remain audible at longer distances.

---

# 25.14 Wildlife Audio

Natural areas generate:

- Birds
- Insects
- Frogs
- Ocean Life
- Forest Wildlife

Wildlife decreases in heavily urbanized regions.

---

# 25.15 Spatial Audio

All major sounds use 3D positioning.

Audio attenuation depends on:

- Distance
- Direction
- Camera Position
- Obstruction

Stereo and surround sound are fully supported.

---

# 25.16 Audio Mixing

The mixer combines:

Music

-

Ambient

-

Effects

-

UI

-

Voice

↓

Master Output

Each category has an independent volume control.

---

# 25.17 Dynamic Audio

Audio intensity changes according to simulation.

Examples:

Traffic

↓

More Vehicles

↓

Louder Traffic

Fire

↓

Larger Fire

↓

More Intense Flames

Festival

↓

Larger Crowd

↓

More Crowd Noise

---

# 25.18 UI Audio

Interface sounds include:

- Button Clicks
- Notifications
- Error Sounds
- Menu Open
- Tool Selection
- Construction Confirmed

UI sounds never interfere with important gameplay audio.

---

# 25.19 Voice Announcements

Optional announcements include:

- Disaster Warnings
- Milestone Reached
- Advisor Messages
- Transport Announcements

Voice language follows localization settings.

---

# 25.20 Audio Prioritization

To avoid excessive simultaneous playback:

Priority order:

```text
Emergency

↓

Disasters

↓

Nearby Vehicles

↓

Nearby Buildings

↓

Environment

↓

Ambient

↓

Far Objects
```

Low-priority sounds may be culled.

---

# 25.21 Audio Zones

Maps define audio zones.

Examples:

- Downtown
- Forest
- Waterfront
- Airport
- Industrial Area
- Residential Neighborhood

Zone blending occurs gradually.

---

# 25.22 Audio Settings

Players may configure:

- Master Volume
- Music
- Effects
- Ambient
- UI
- Voice
- Dynamic Range
- Output Device

Changes apply immediately.

---

# 25.23 Serialization

Persistent settings include:

- Volume Levels
- Playlist
- Radio Station
- Output Configuration

Runtime sound playback is not serialized.

---

# 25.24 Performance Optimization

The Audio Manager employs:

- Audio pooling
- Distance culling
- Sound virtualization
- Streaming music
- Shared samples
- Event-driven playback

Only audible sound sources are actively processed.

---

# 25.25 Design Goals

The Audio System is designed to provide:

- Dynamic city soundscapes
- Spatial audio
- Weather-aware ambience
- Reactive disaster audio
- Immersive transport sounds
- Efficient audio streaming
- Independent mixing
- High-performance playback
- Accessibility
- Complete separation from simulation

The Audio System transforms the city into a living acoustic environment, allowing players to hear the pulse of their metropolis through evolving ambient sound, transportation, construction, disasters, weather, and citizen activity while maintaining efficient runtime performance.

# 26. Map Editor, Terrain Tools & Scenario Editor

The Editor System allows players and content creators to design custom maps, scenarios, themes, and gameplay experiences. It provides professional-grade tools for terrain sculpting, resource placement, road layout, outside connections, environmental configuration, scenario scripting, and map validation.

The editor operates independently of normal gameplay and exports fully deterministic content that integrates seamlessly with the simulation.

---

# 26.1 Editor Architecture

The Editor Manager coordinates all creation tools.

```text
EditorManager

├── Terrain Editor
├── Water Editor
├── Resource Editor
├── Road Editor
├── Outside Connections
├── Environment Editor
├── Theme Editor
├── Map Validation
├── Scenario Editor
├── Asset Placement
└── Export Manager
```

Editor operations never run during standard gameplay.

---

# 26.2 Editor Modes

Supported editor modes include:

- Map Editor
- Scenario Editor
- Theme Editor
- Asset Editor (future expansion)
- Testing Mode

Each mode exposes only the relevant editing tools.

---

# 26.3 Terrain Sculpting

Terrain tools include:

- Raise
- Lower
- Flatten
- Smooth
- Soften
- Level
- Erosion Brush
- Noise Brush

Brush settings:

- Radius
- Strength
- Falloff
- Shape

Terrain modifications update in real time.

---

# 26.4 Heightmap Support

Maps may be generated from grayscale heightmaps.

Import pipeline:

```text
Heightmap

↓

Normalize Elevation

↓

Generate Terrain

↓

Apply Theme

↓

Validate Map
```

Supported resolutions depend on map size.

---

# 26.5 Water Editing

Water tools include:

- River Sources
- Lakes
- Ocean Level
- Water Outlets
- Flow Direction
- Water Volume

Water simulation is previewed directly in the editor.

---

# 26.6 Road Layout

Road editing supports:

- Highways
- Roads
- Railways
- Pedestrian Paths
- Bridges
- Tunnels

Roads use the same procedural generation system as gameplay.

---

# 26.7 Outside Connections

Maps define external connections.

Supported connections:

- Highway
- Railway
- Ship Route
- Flight Path

Outside connections determine:

- Imports
- Exports
- Tourism
- Migration

At least one outside connection is required for normal gameplay.

---

# 26.8 Natural Resources

Resource painting supports:

- Oil
- Ore
- Fertile Land
- Forest Density

Resources may overlap where appropriate.

---

# 26.9 Vegetation

Tree tools include:

- Single Placement
- Brush Placement
- Forest Generation
- Random Density
- Species Selection

Trees become simulation objects after the map loads.

---

# 26.10 Environment Settings

Environment configuration includes:

- Climate
- Temperature
- Wind
- Weather Frequency
- Water Color
- Fog Density
- Ambient Lighting

Environment settings define the map's visual identity.

---

# 26.11 Theme Selection

Themes determine:

- Terrain Materials
- Road Materials
- Skybox
- Vegetation
- Water Appearance
- Lighting

Themes are interchangeable without altering gameplay mechanics.

---

# 26.12 Starting Tile

The editor defines:

- Initial Build Area
- Camera Spawn
- Highway Access
- Initial Infrastructure

Players always begin within the designated starting tile.

---

# 26.13 Buildable Area

Maps define:

- Maximum Playable Area
- Purchasable Tiles
- Water Boundaries
- Terrain Limits

Tile layouts follow the standard expansion system.

---

# 26.14 Decorative Assets

Creators may place:

- Rocks
- Trees
- Ruins
- Landmarks
- Fences
- Utility Props

Decorative assets do not affect simulation unless configured otherwise.

---

# 26.15 Map Metadata

Each map stores:

```cpp
MapName

Author

Description

Theme

Climate

Size

PreviewImage

Version
```

Metadata appears in the map browser.

---

# 26.16 Scenario Editor

Scenarios define custom gameplay objectives.

Supported objectives include:

- Reach Population
- Maintain Budget
- Build Services
- Survive Disaster
- Export Resources
- Achieve Happiness
- Complete Within Time Limit

Multiple objectives may run simultaneously.

---

# 26.17 Scenario Conditions

Triggers include:

- Population
- Time
- Treasury
- Disaster Occurrence
- Building Count
- Industry Production
- Tourism
- Pollution

Conditions are evaluated continuously.

---

# 26.18 Scenario Events

Supported events include:

- Spawn Disaster
- Award Money
- Unlock Feature
- Restrict Building
- Modify Policies
- Change Weather
- Display Message

Events can be chained together.

---

# 26.19 Victory Conditions

Examples:

- Population Goal
- Financial Goal
- Transport Efficiency
- Pollution Reduction
- Education Level
- Tourism Target

Multiple victory paths are supported.

---

# 26.20 Failure Conditions

Examples:

- Bankruptcy
- Population Loss
- Time Expired
- Utility Failure
- Disaster Casualties

Failure immediately ends the scenario unless retries are enabled.

---

# 26.21 Validation

Before export, the editor verifies:

- Outside Connections
- Water Sources
- Terrain Integrity
- Asset References
- Missing Dependencies
- Invalid Roads
- Broken Transport Networks

Validation errors must be resolved before publishing.

---

# 26.22 Testing Mode

Testing mode launches the map directly.

Creators may:

- Simulate Population
- Trigger Disasters
- Unlock Everything
- Inspect Performance

Changes can be made without restarting the editor.

---

# 26.23 Export Pipeline

```text
Validate

↓

Generate Metadata

↓

Compress

↓

Generate Thumbnail

↓

Package

↓

Export
```

Exported content is ready for local use or Workshop distribution.

---

# 26.24 Workshop Publishing

Maps and scenarios support:

- Upload
- Update
- Versioning
- Screenshots
- Tags
- Dependencies

Workshop metadata is managed separately from the map package.

---

# 26.25 Serialization

Editor projects store:

- Terrain
- Water
- Resources
- Roads
- Decorations
- Metadata
- Scenario Logic
- Environment Settings

Exported maps exclude editor-only data.

---

# 26.26 Performance Optimization

The Editor Manager employs:

- Chunked terrain editing
- GPU brush previews
- Incremental terrain updates
- Asynchronous validation
- Lazy asset loading
- Cached thumbnails

Large maps remain responsive during editing.

---

# 26.27 Design Goals

The Editor System is designed to provide:

- Professional terrain editing
- Flexible scenario creation
- Complete map customization
- Resource placement
- Environmental control
- Deterministic exports
- Comprehensive validation
- Workshop integration
- High-performance editing
- Future extensibility

The Editor System empowers creators to build entirely new gameplay experiences—from realistic geographic recreations to custom challenge scenarios—while ensuring every exported map integrates seamlessly with the game's simulation, rendering, and progression systems.

# 27. Achievements, Statistics & City Timeline

The Achievement & Statistics System records every meaningful event that occurs throughout the lifetime of a city. It provides long-term progression tracking, historical analytics, milestone records, and player accomplishments without affecting gameplay balance.

Achievements are optional goals, while statistics and the city timeline serve as analytical tools for understanding the city's development over time.

---

# 27.1 Statistics Architecture

The Statistics Manager coordinates all historical and analytical data.

```text
StatisticsManager

├── Population Statistics
├── Economy Statistics
├── Traffic Statistics
├── Service Statistics
├── Environment Statistics
├── Timeline Manager
├── Achievement Manager
├── Records Manager
├── Graph Generator
└── Export System
```

Statistics are collected continuously using event-driven updates.

---

# 27.2 Statistic Categories

The system records data for:

City

Population

Economy

Traffic

Utilities

Education

Healthcare

Crime

Fire Safety

Transportation

Industry

Tourism

Environment

Disasters

Construction

Performance

---

# 27.3 Population Statistics

Tracked metrics include:

- Current Population
- Highest Population
- Birth Rate
- Death Rate
- Immigration
- Emigration
- Household Count
- Average Household Size
- Education Distribution
- Age Distribution

Historical trends remain available throughout the city's lifetime.

---

# 27.4 Economic Statistics

Recorded values include:

- Treasury
- Weekly Income
- Weekly Expenses
- Tax Revenue
- Tourism Revenue
- Industry Revenue
- Import Costs
- Export Revenue
- Loans
- Net Profit

Financial history supports multiple zoom levels.

---

# 27.5 Traffic Statistics

Traffic metrics include:

- Traffic Flow
- Average Commute Time
- Vehicle Count
- Public Transport Usage
- Parking Occupancy
- Road Congestion
- Freight Volume
- Average Vehicle Speed

Traffic history assists long-term infrastructure planning.

---

# 27.6 Utility Statistics

Tracked systems:

- Electricity Production
- Electricity Consumption
- Water Production
- Water Consumption
- Sewage Capacity
- Heating Production
- Garbage Processing

Utility shortages are recorded for historical analysis.

---

# 27.7 Healthcare Statistics

Metrics include:

- Average Health
- Hospital Capacity
- Ambulance Availability
- Patient Count
- Disease Outbreaks
- Death Rate

Healthcare trends are correlated with environmental conditions.

---

# 27.8 Education Statistics

Recorded metrics:

- Student Count
- School Capacity
- Graduation Rate
- Average Education Level
- Literacy Distribution

Education statistics influence workforce analysis.

---

# 27.9 Crime Statistics

Crime reports include:

- Crime Rate
- Police Coverage
- Arrests
- Prison Population
- Emergency Response Time

Historical spikes remain visible for investigation.

---

# 27.10 Environmental Statistics

Environmental history includes:

- Air Pollution
- Ground Pollution
- Water Pollution
- Noise Pollution
- Forest Coverage
- Renewable Energy Usage
- Temperature
- Weather Events

Environmental improvements are visible over long periods.

---

# 27.11 Transportation Statistics

Recorded values:

- Passenger Counts
- Line Usage
- Station Usage
- Vehicle Occupancy
- Ticket Revenue
- On-Time Performance
- Transport Efficiency

Every transport mode maintains independent statistics.

---

# 27.12 Industry Statistics

Tracked information:

- Raw Material Production
- Manufacturing Output
- Warehouse Utilization
- Exports
- Imports
- Resource Depletion

Industry efficiency is monitored continuously.

---

# 27.13 Disaster Statistics

Historical disaster data includes:

- Disaster Type
- Damage Cost
- Buildings Destroyed
- Casualties
- Response Time
- Recovery Time

Completed disasters remain permanently recorded.

---

# 27.14 Construction Statistics

Construction history records:

- Buildings Constructed
- Roads Built
- Bridges
- Tunnels
- Parks
- Utilities
- Land Purchased

Construction trends illustrate city expansion.

---

# 27.15 Timeline System

Major events appear in chronological order.

Examples:

```text
City Founded

↓

Village Reached

↓

University Built

↓

Metro Opened

↓

Major Fire

↓

Population 100,000

↓

Airport Opened

↓

Metropolis
```

Timeline entries are permanent.

---

# 27.16 Graph System

Every major statistic supports graphs.

Time ranges include:

- Hour
- Day
- Week
- Month
- Year
- Entire City History

Graphs automatically scale to available data.

---

# 27.17 Records System

The game tracks lifetime records.

Examples:

- Highest Population
- Largest Weekly Income
- Longest Traffic Jam
- Biggest Disaster
- Highest Happiness
- Most Tourists
- Largest Budget Surplus

Records update immediately when surpassed.

---

# 27.18 Achievement Architecture

Achievements are event-driven.

```text
Simulation Event

↓

Requirement Check

↓

Achievement Progress

↓

Unlock

↓

Notification
```

Achievement evaluation never impacts simulation performance.

---

# 27.19 Achievement Categories

Categories include:

- Population
- Economy
- Transportation
- Industry
- Environment
- Education
- Healthcare
- Construction
- Disaster Recovery
- Tourism
- Scenario Completion
- Hidden Achievements

---

# 27.20 Example Achievements

Examples include:

- Reach 10,000 Population
- Build Every Transport Type
- Maintain Positive Income for One Year
- Eliminate Pollution
- Survive a Major Disaster
- Achieve Maximum Education
- Complete Every Milestone
- Export One Million Units of Goods

Achievements are fully data-driven.

---

# 27.21 Progress Tracking

Achievements may store progress.

Example:

```text
Plant Trees

Current:

4,250

Goal:

10,000
```

Progress updates automatically.

---

# 27.22 Scenario Statistics

Scenario-specific metrics include:

- Completion Time
- Objectives Completed
- Optional Objectives
- Score
- Final Population
- Remaining Budget

Scenario records remain separate from sandbox cities.

---

# 27.23 Leaderboards (Platform Support)

Optional online leaderboards may rank:

- Population
- City Value
- Scenario Completion
- Efficiency Scores
- Challenge Events

Leaderboard support depends on platform capabilities.

---

# 27.24 Export System

Players may export:

- Financial Reports
- Traffic Statistics
- Population Graphs
- Timeline
- Achievement List

Exports are intended for sharing and analysis.

---

# 27.25 Serialization

Persistent data includes:

- Historical Graphs
- Timeline
- Records
- Achievement Progress
- Unlocked Achievements
- Statistics Cache

No statistical history is lost after saving.

---

# 27.26 Performance Optimization

The Statistics Manager employs:

- Event aggregation
- Historical compression
- Incremental graph generation
- Cached calculations
- Lazy graph rendering
- Batched timeline updates

Older historical data may be summarized while preserving long-term trends.

---

# 27.27 Design Goals

The Achievement & Statistics System is designed to provide:

- Comprehensive city analytics
- Historical trend visualization
- Long-term player progression
- Optional gameplay goals
- Efficient data collection
- Rich reporting
- Timeline preservation
- Scenario tracking
- Platform integration
- High-performance historical analysis

The Achievement & Statistics System transforms millions of simulation events into meaningful insights, allowing players to analyze their city's growth, celebrate accomplishments, learn from past decisions, and compare their achievements across multiple cities and scenarios.

# 28. Artificial Intelligence, Advisors & Game Assistance

The Advisor System provides contextual guidance, recommendations, warnings, and educational information based on the current simulation state. Advisors do not play the game for the player or make automatic city decisions. Instead, they analyze city conditions and present actionable insights.

The system is entirely data-driven, allowing future advisors, scenarios, tutorials, and expansions to introduce new guidance without modifying core simulation systems.

---

# 28.1 Advisor Architecture

The Advisor Manager coordinates all advisor functionality.

```text
AdvisorManager

├── Economy Advisor
├── Transportation Advisor
├── Utilities Advisor
├── Healthcare Advisor
├── Education Advisor
├── Fire & Rescue Advisor
├── Police Advisor
├── Industry Advisor
├── Environment Advisor
├── Disaster Advisor
├── Tutorial Advisor
└── Notification System
```

Each advisor independently subscribes to simulation events.

---

# 28.2 Advisor Responsibilities

Advisors monitor:

- City health
- Budget
- Population
- Utilities
- Traffic
- Pollution
- Services
- Education
- Happiness
- Emergencies

Advisors never directly modify gameplay.

---

# 28.3 Advisor Evaluation Cycle

```text
Simulation Event

↓

Relevant Advisor

↓

Condition Evaluation

↓

Recommendation Generated

↓

Notification

↓

Player Decision
```

Only relevant advisors evaluate triggered events.

---

# 28.4 Economy Advisor

Monitors:

- Budget deficit
- Low cash reserves
- Loan opportunities
- High maintenance
- Tax efficiency
- Service spending
- Economic growth

Example recommendations:

- Increase residential taxes.
- Reduce transport budget.
- Expand industry.
- Delay expensive construction.

---

# 28.5 Transportation Advisor

Analyzes:

- Congestion
- Public transport usage
- Parking demand
- Freight bottlenecks
- Highway capacity
- Route efficiency

Suggestions include:

- Add bus lines.
- Upgrade roads.
- Build metro.
- Improve intersections.

---

# 28.6 Utility Advisor

Monitors:

- Electricity
- Water
- Sewage
- Heating
- Garbage

Warnings include:

- Low power production.
- Water shortage.
- Landfill nearly full.
- Heating capacity exceeded.

---

# 28.7 Healthcare Advisor

Tracks:

- Citizen health
- Hospital occupancy
- Ambulance response
- Disease outbreaks
- Cemetery capacity

Suggestions focus on expanding medical services before critical shortages occur.

---

# 28.8 Fire & Rescue Advisor

Evaluates:

- Fire coverage
- Fire incidents
- Response time
- Forest fire risk

Recommendations include:

- Build new fire stations.
- Improve road access.
- Increase fire budget.

---

# 28.9 Police Advisor

Monitors:

- Crime rate
- Prison occupancy
- Police coverage
- Emergency response

Suggested actions:

- Build police stations.
- Increase police funding.
- Improve road connectivity.

---

# 28.10 Education Advisor

Analyzes:

- School capacity
- University demand
- Graduation rates
- Workforce education

Recommendations:

- Build elementary schools.
- Expand universities.
- Improve education budget.

---

# 28.11 Industry Advisor

Monitors:

- Worker shortages
- Freight congestion
- Resource depletion
- Warehouse utilization
- Production efficiency

Suggestions encourage balanced industrial expansion.

---

# 28.12 Environment Advisor

Tracks:

- Pollution
- Noise
- Renewable energy
- Forest coverage
- Water quality

Recommendations include:

- Plant trees.
- Reduce industrial pollution.
- Build cleaner power plants.
- Improve sewage treatment.

---

# 28.13 Disaster Advisor

Activated during emergency situations.

Provides:

- Evacuation reminders
- Shelter status
- Recovery progress
- Emergency vehicle shortages
- Infrastructure priorities

Disaster recommendations change dynamically throughout the event.

---

# 28.14 Tutorial Advisor

Introduces newly unlocked systems.

Examples:

Metro Unlocked

↓

Metro Tutorial

Districts Unlocked

↓

District Tutorial

Policies Unlocked

↓

Policy Tutorial

Tutorials never interrupt active emergencies.

---

# 28.15 Warning Levels

Recommendations are categorized by urgency.

```text
Information

↓

Suggestion

↓

Warning

↓

Critical

↓

Emergency
```

Critical alerts receive visual and audio emphasis.

---

# 28.16 Notification Prioritization

Priority order:

1. Disasters
2. Utility failures
3. Fire
4. Crime
5. Healthcare
6. Budget
7. Traffic
8. Education
9. General suggestions

Lower-priority notifications may be grouped together.

---

# 28.17 Advisor History

Players may review previous recommendations.

Stored information includes:

- Time
- Advisor
- Recommendation
- Severity
- Resolution Status

Resolved recommendations remain archived.

---

# 28.18 Recommendation Rules

Each recommendation defines:

```cpp
Condition

Priority

Cooldown

Message

Suggested Action

Dismissable
```

Cooldowns prevent repetitive notifications.

---

# 28.19 Advisor Personalities

Different advisors have unique communication styles while maintaining factual accuracy.

Examples:

Economy Advisor

- Financially conservative

Transport Advisor

- Efficiency focused

Environment Advisor

- Sustainability focused

Personality affects wording only—not gameplay logic.

---

# 28.20 Advisor Settings

Players may configure:

- Advisor frequency
- Notification sounds
- Pop-up behavior
- Tutorial visibility
- Alert categories

Individual advisors may be muted independently.

---

# 28.21 Contextual Recommendations

Recommendations consider multiple systems simultaneously.

Example:

Heavy traffic

-

High pollution

-

Growing population

↓

Recommend Metro instead of widening roads.

Cross-system reasoning produces more useful guidance.

---

# 28.22 Scenario Integration

Scenario creators may define:

- Custom advisors
- Scenario hints
- Objective reminders
- Dynamic story events

Scenario advice integrates with standard advisors.

---

# 28.23 Statistics

Tracked metrics include:

- Advisor Messages Generated
- Recommendations Accepted
- Recommendations Dismissed
- Tutorials Completed
- Average Response Time

These statistics assist player learning analysis.

---

# 28.24 Serialization

Saved advisor data includes:

- Active Recommendations
- Notification History
- Tutorial Progress
- Advisor Settings
- Cooldowns

Advisor state resumes seamlessly after loading.

---

# 28.25 Performance Optimization

The Advisor Manager employs:

- Event-driven evaluation
- Cached thresholds
- Cooldown timers
- Rule batching
- Deferred recommendation generation

Advisors only evaluate when relevant simulation events occur.

---

# 28.26 Design Goals

The Advisor System is designed to provide:

- Context-aware recommendations
- Non-intrusive guidance
- Educational tutorials
- Cross-system analysis
- Configurable notifications
- Efficient event processing
- Scenario compatibility
- Long-term player assistance
- Fully data-driven rules
- Complete separation from gameplay simulation

The Advisor System helps players understand the increasingly complex interactions within their city by transforming simulation data into meaningful recommendations, allowing both new and experienced players to make informed decisions without reducing the depth or challenge of city management.

# 29. Game Modes, Scenarios & Sandbox

The Game Mode System defines how players interact with the simulation by enabling different rulesets, objectives, progression systems, and gameplay constraints. Each game mode uses the same core simulation engine while altering progression, unlocks, victory conditions, budgets, disasters, and player limitations.

All game modes are data-driven and can be extended through DLC, mods, or custom scenarios.

---

# 29.1 Game Mode Architecture

The Game Mode Manager coordinates all gameplay modes.

```text
GameModeManager

├── Standard Mode
├── Sandbox Mode
├── Scenario Mode
├── Tutorial Mode
├── Challenge Mode
├── Custom Rules
├── Victory Manager
├── Failure Manager
└── Scenario Script Engine
```

The selected mode configures simulation rules before a city is initialized.

---

# 29.2 Standard Mode

Standard Mode represents the default gameplay experience.

Features include:

- Population-based progression
- Milestone unlocks
- Economy simulation
- Budget management
- Loans
- Achievements
- Natural disasters (optional)
- Full simulation

This mode closely mirrors the default Cities: Skylines gameplay loop.

---

# 29.3 Sandbox Mode

Sandbox Mode removes gameplay restrictions.

Features:

- Unlimited money (optional)
- Unlimited resources (optional)
- All milestones unlocked
- All buildings unlocked
- All transport systems unlocked
- Optional disasters
- Optional achievements (configurable)

Players may freely experiment without progression limits.

---

# 29.4 Scenario Mode

Scenarios define structured objectives.

Each scenario includes:

- Starting city
- Starting budget
- Starting population
- Objectives
- Failure conditions
- Rewards
- Custom events

Scenarios may modify nearly every gameplay system.

---

# 29.5 Tutorial Mode

Tutorial Mode introduces gameplay systems gradually.

Lessons include:

- Road construction
- Zoning
- Utilities
- Budget
- Services
- Public transport
- Districts
- Policies

Tutorials monitor player actions and advance upon completion.

---

# 29.6 Challenge Mode

Challenge Mode applies gameplay modifiers.

Examples:

- High disaster frequency
- Low starting budget
- No loans
- Resource scarcity
- Heavy traffic
- Pollution restrictions
- Time limits

Challenges encourage alternative city-building strategies.

---

# 29.7 Custom Rules

Players may customize:

- Starting money
- Population
- Unlock progression
- Natural resources
- Day/night cycle
- Traffic difficulty
- Weather frequency
- Disaster frequency
- Citizen aging
- Economy difficulty

Custom rules are stored with the save.

---

# 29.8 Starting Conditions

Every game defines:

```cpp
StartingMoney

StartingPopulation

UnlockedFeatures

AvailableTiles

Weather

MapTheme

ScenarioState
```

These values initialize the simulation.

---

# 29.9 Victory Conditions

Supported victory conditions include:

Population Target

City Value

Treasury

Education Level

Traffic Flow

Pollution Reduction

Tourism

Industry Production

Transport Usage

Disaster Survival

Multiple victory conditions may be active simultaneously.

---

# 29.10 Failure Conditions

Possible failure conditions:

- Bankruptcy
- Population Collapse
- Time Limit
- Disaster Casualties
- Objective Failure
- Infrastructure Failure
- Environmental Collapse

Sandbox mode ignores failure unless explicitly enabled.

---

# 29.11 Scenario Objectives

Objective types include:

Build Structure

Reach Population

Maintain Budget

Reduce Crime

Increase Education

Increase Happiness

Export Goods

Construct Transport

Recover After Disaster

Objectives may contain multiple stages.

---

# 29.12 Dynamic Scenario Events

Scenarios support scripted events.

Examples:

```text
Population 20,000

↓

Earthquake

↓

Emergency Objective

↓

Recovery Reward
```

Events may trigger other events recursively.

---

# 29.13 Timed Objectives

Objectives may include timers.

Example:

```text
Build Hospital

↓

Within

↓

90 Days
```

Failure may trigger penalties or scenario loss.

---

# 29.14 Difficulty Settings

Difficulty modifies:

- Construction costs
- Maintenance costs
- Tax efficiency
- Disaster severity
- Citizen tolerance
- Traffic behavior
- Economic growth
- Service demand

Simulation logic remains unchanged.

---

# 29.15 Scoring

Scenarios may calculate scores.

Possible scoring factors:

- Completion Time
- Budget Remaining
- Population
- Happiness
- Pollution
- Traffic
- Disaster Damage
- Education

Scores enable leaderboard integration.

---

# 29.16 Event Scripting

Scenario scripts support:

Conditions

↓

Events

↓

Actions

↓

Consequences

↓

Objectives

Scripts execute deterministically.

---

# 29.17 Reward System

Scenarios may reward:

- Achievements
- Unlocks
- Cosmetic Items
- Titles
- Statistics
- Workshop Ratings

Rewards do not affect unrelated saves.

---

# 29.18 Checkpoints

Scenarios may define checkpoints.

Checkpoint stores:

- Simulation State
- Objectives
- Progress
- Budget
- Population

Players may restart from the latest checkpoint.

---

# 29.19 Save Restrictions

Scenario creators may choose:

- Manual Saves Allowed
- Autosaves Only
- Single Save Slot
- Ironman Mode

Restrictions are enforced by the Save Manager.

---

# 29.20 Random Events

Game modes may enable:

- Weather events
- Economic events
- Tourism booms
- Industrial demand spikes
- Infrastructure failures
- Emergency situations

Random events follow deterministic seeded generation.

---

# 29.21 Mod Integration

Game modes support:

- Custom scenarios
- Script extensions
- Custom objectives
- New victory conditions
- Gameplay modifiers

Mods interact through the public scenario API.

---

# 29.22 Statistics

Tracked metrics include:

- Active Game Mode
- Scenario Progress
- Objectives Completed
- Scenario Score
- Total Play Time
- Challenge Modifiers
- Victory Status

Statistics remain available after completion.

---

# 29.23 Serialization

Saved game mode data includes:

- Active Mode
- Scenario State
- Objectives
- Difficulty
- Custom Rules
- Event Queue
- Victory Progress

Game mode resumes exactly after loading.

---

# 29.24 Performance Optimization

The Game Mode Manager employs:

- Event-driven objective evaluation
- Cached rule lookups
- Incremental progress tracking
- Deferred event execution
- Lightweight scripting engine

Inactive objectives consume no runtime resources.

---

# 29.25 Design Goals

The Game Mode System is designed to provide:

- Flexible gameplay experiences
- Structured progression
- Open-ended sandbox play
- Replayable scenarios
- Custom difficulty
- Dynamic objectives
- Scriptable events
- Deterministic execution
- Mod compatibility
- Future expansion support

The Game Mode System enables players to experience the same simulation engine through a wide range of play styles—from unrestricted creative building to highly structured challenge scenarios—while ensuring every mode remains deterministic, extensible, and fully integrated with the city's core simulation.

# 30. Engine Architecture, Simulation Loop & Core Systems

The Engine Architecture defines how every subsystem communicates, updates, renders, saves, and scales while maintaining deterministic simulation. It is the foundation upon which every gameplay feature—from a single citizen walking to a city containing millions of simulated entities—is built.

The engine separates simulation, rendering, audio, user interface, asset management, and persistence into independent modules that communicate through well-defined interfaces.

---

# 30.1 High-Level Architecture

```text
Application

├── Core Engine
├── Simulation Engine
├── Rendering Engine
├── Physics Engine
├── Audio Engine
├── UI System
├── Asset System
├── Save System
├── Mod System
├── Networking (Future)
└── Platform Layer
```

Each subsystem operates independently and communicates through event-driven interfaces.

---

# 30.2 Core Engine

The Core Engine is responsible for:

- Engine startup
- Initialization order
- Main loop
- Configuration
- Service registration
- Thread management
- Memory management
- Shutdown

It never contains gameplay logic.

---

# 30.3 Engine Initialization

Startup order:

```text
Platform

↓

Configuration

↓

Asset System

↓

Rendering

↓

Audio

↓

Simulation

↓

UI

↓

Mods

↓

Load Save / New Game

↓

Gameplay
```

Initialization order is deterministic.

---

# 30.4 Main Game Loop

The engine follows a fixed simulation timestep.

```text
Input

↓

Simulation Tick

↓

Physics

↓

AI

↓

Economy

↓

Rendering Snapshot

↓

Audio

↓

UI

↓

Present Frame
```

Simulation always executes before rendering.

---

# 30.5 Fixed Simulation Tick

Simulation runs using a fixed timestep.

Example:

```text
60 Simulation Updates / Second
```

Benefits:

- Determinism
- Stable AI
- Consistent economy
- Reliable pathfinding

Rendering may run faster or slower independently.

---

# 30.6 Render Interpolation

Rendering interpolates between simulation states.

```text
Simulation A

↓

Interpolation

↓

Simulation B
```

This produces smooth animation without affecting gameplay.

---

# 30.7 Entity Component System (ECS)

Simulation entities are composed from components.

Example:

```text
Entity

↓

Transform

↓

Citizen Component

↓

Pathfinding Component

↓

Needs Component

↓

Rendering Component
```

Behavior is provided by systems rather than inheritance.

---

# 30.8 Simulation Systems

Major systems include:

- Citizens
- Vehicles
- Economy
- Buildings
- Utilities
- Transport
- Industry
- Environment
- Disasters
- Policies

Each system updates in a deterministic order.

---

# 30.9 Update Order

Example update sequence:

```text
Time

↓

Environment

↓

Utilities

↓

Buildings

↓

Citizens

↓

Vehicles

↓

Economy

↓

Services

↓

Statistics
```

Stable ordering prevents inconsistent simulation outcomes.

---

# 30.10 Event Bus

Subsystems communicate through events.

Example:

```text
Building Burned

↓

Fire Event

↓

Fire Department

↓

Citizen Notification

↓

Statistics

↓

UI Update
```

Direct subsystem dependencies are minimized.

---

# 30.11 Job System

Expensive workloads execute in parallel.

Examples:

- Pathfinding
- Pollution diffusion
- Citizen updates
- Traffic analysis
- Visibility calculation

Jobs synchronize before dependent systems continue.

---

# 30.12 Thread Architecture

Example worker layout:

```text
Main Thread

↓

Simulation

Worker Threads

↓

Citizens

↓

Vehicles

↓

Environment

↓

Traffic

↓

Rendering Preparation
```

Rendering submission may occur on dedicated graphics threads where supported.

---

# 30.13 Memory Management

The engine uses:

- Object pools
- Arena allocators
- Resource caches
- Shared asset references
- Chunk allocators

Heap allocations during simulation are minimized.

---

# 30.14 Resource Management

Resources include:

- Meshes
- Textures
- Audio
- Materials
- Animations
- Localization
- Scripts

Unused resources are released automatically.

---

# 30.15 Configuration System

Configuration supports:

- Graphics
- Audio
- Gameplay
- Controls
- Accessibility
- Mods
- Debug Options

Configuration changes are validated before applying.

---

# 30.16 Platform Layer

Platform-specific functionality includes:

- Window Management
- Input
- File System
- GPU Interface
- Audio Device
- Networking APIs
- Cloud Saves

Gameplay systems remain platform independent.

---

# 30.17 Error Handling

Runtime errors are categorized as:

- Warning
- Recoverable Error
- Fatal Error

Recoverable errors allow gameplay to continue whenever possible.

---

# 30.18 Logging

Log categories include:

- Engine
- Rendering
- Simulation
- Audio
- Assets
- Mods
- Save System
- Performance

Logging levels:

- Debug
- Info
- Warning
- Error

---

# 30.19 Debug Tools

Developer tools include:

- Entity Inspector
- Simulation Visualizer
- Performance Overlay
- Memory Viewer
- Job Timeline
- Event Viewer
- Console

Debug tools are unavailable in normal gameplay.

---

# 30.20 Performance Profiling

Profiler tracks:

- CPU Time
- GPU Time
- Memory
- Draw Calls
- Simulation Cost
- AI Cost
- Pathfinding Cost
- Traffic Cost

Profiling supports frame-by-frame analysis.

---

# 30.21 Determinism

The simulation guarantees:

Same Seed

-

Same Inputs

↓

Same Simulation

↓

Same Save

↓

Same Results

Random number generation is fully seeded.

---

# 30.22 Scalability

The engine is designed to scale across:

- Small villages
- Medium towns
- Large cities
- Megacities

Performance scales through:

- Spatial partitioning
- Parallel simulation
- LOD systems
- Streaming
- Incremental updates

---

# 30.23 Extensibility

Future systems integrate through public interfaces.

Supported extension points:

- Mods
- DLC
- Assets
- Scenarios
- UI Extensions
- Advisors
- Policies

Core engine code remains unchanged.

---

# 30.24 Shutdown Sequence

Shutdown order:

```text
Pause Simulation

↓

Save (Optional)

↓

Unload Mods

↓

Unload Assets

↓

Destroy Rendering

↓

Destroy Audio

↓

Release Memory

↓

Exit
```

Resources are released in dependency order.

---

# 30.25 Serialization Interfaces

Every major subsystem implements:

```cpp
interface ISystem
{
    Initialize();

    Update();

    Serialize();

    Deserialize();

    Shutdown();
}
```

This standardizes communication across the engine.

---

# 30.26 Performance Optimization

The engine employs:

- Multi-threaded simulation
- Fixed timestep updates
- GPU instancing
- Resource streaming
- Object pooling
- Event-driven communication
- Chunked world simulation
- Incremental updates
- Cache-friendly data layouts
- Deterministic scheduling

---

# 30.27 Design Goals

The Engine Architecture is designed to provide:

- Deterministic simulation
- Massive scalability
- High performance
- Modular systems
- Clean subsystem separation
- Efficient multithreading
- Stable save/load behavior
- Cross-platform compatibility
- Long-term extensibility
- Reliable future expansion

The Engine Architecture serves as the foundation of the entire game, ensuring that every subsystem—from rendering and audio to economy and citizen AI—operates cohesively, efficiently, and deterministically while supporting cities of virtually any scale.

# 31. Asset System, Asset Editor & Content Pipeline

The Asset System is responsible for loading, validating, managing, and rendering every piece of content used by the game. Every building, road, vehicle, citizen, prop, tree, texture, material, animation, audio clip, and UI element exists as an asset.

The Asset System is fully data-driven and supports official content, DLC, user-generated assets, and mods without requiring changes to the simulation engine.

---

# 31.1 Asset Architecture

The Asset Manager coordinates all content.

```text
AssetManager

├── Asset Database
├── Asset Loader
├── Dependency Manager
├── Asset Validator
├── Asset Streaming
├── Thumbnail Generator
├── Asset Cache
├── Asset Editor
├── Workshop Assets
└── Package Manager
```

Assets are identified by globally unique identifiers (GUIDs).

---

# 31.2 Asset Categories

Supported asset types include:

Buildings

Roads

Bridges

Railways

Vehicles

Citizens

Props

Trees

Terrain Materials

Water Materials

Animations

Textures

Meshes

Materials

Audio

UI

Icons

Fonts

Scripts (Editor Only)

---

# 31.3 Asset Metadata

Each asset stores:

```cpp
AssetID

Name

Author

Version

Category

Thumbnail

Tags

Dependencies

License

Package
```

Metadata is searchable within the Asset Browser.

---

# 31.4 Asset Loading

Loading pipeline:

```text
Read Package

↓

Validate

↓

Resolve Dependencies

↓

Load Resources

↓

Create Runtime Asset

↓

Cache

↓

Available
```

Assets become available immediately after successful validation.

---

# 31.5 Asset Packages

Assets are distributed inside packages.

Example:

```text
Package

├── Metadata
├── Meshes
├── Textures
├── Materials
├── Animations
├── Audio
└── Preview
```

Packages may contain one or multiple assets.

---

# 31.6 Dependency Management

Assets may depend on:

- Textures
- Materials
- Meshes
- Animations
- DLC
- Other Assets

Missing dependencies prevent asset loading unless fallback resources exist.

---

# 31.7 Building Assets

Building assets define:

- Mesh
- LOD Meshes
- Footprint
- Service Type
- Construction Cost
- Maintenance Cost
- Electricity Usage
- Water Usage
- Workers
- Capacity

Simulation data is stored separately from visual resources.

---

# 31.8 Vehicle Assets

Vehicle assets contain:

- Mesh
- LOD Models
- Animations
- Speed Limits
- Capacity
- Turning Radius
- Sounds
- Lights

Vehicle behavior is controlled by simulation systems.

---

# 31.9 Road Assets

Road definitions include:

- Lane Count
- Lane Directions
- Speed Limit
- Sidewalk Width
- Median
- Decoration
- Tram Tracks
- Bus Lanes
- Bicycle Lanes

Road geometry is generated procedurally.

---

# 31.10 Citizen Assets

Citizen assets define:

- Skeleton
- Clothing Variants
- Hair Styles
- Accessories
- Animations
- LOD Models

Simulation identity is independent of appearance.

---

# 31.11 Prop Assets

Examples:

- Benches
- Street Lights
- Mailboxes
- Traffic Lights
- Fences
- Signs

Props generally do not participate in gameplay simulation.

---

# 31.12 Tree Assets

Tree definitions include:

- Species
- Seasonal Appearance
- Wind Animation
- Growth Scale
- Collision
- LOD

Large forests are rendered using GPU instancing.

---

# 31.13 Material System

Materials define:

- Albedo
- Normal Map
- Roughness
- Metallic
- Ambient Occlusion
- Emissive

Materials are shared across multiple assets whenever possible.

---

# 31.14 Texture System

Supported texture types:

- Diffuse
- Normal
- Height
- Roughness
- Metallic
- Emission
- Mask

Textures stream based on camera distance.

---

# 31.15 Mesh System

Meshes support:

- Multiple LODs
- Collision Mesh
- Shadow Mesh
- Navigation Mesh
- Vertex Colors

Meshes are optimized during import.

---

# 31.16 Animation Assets

Supported animations:

- Walking
- Running
- Sitting
- Construction
- Vehicle Doors
- Wheel Rotation

Animations are shared using common skeletons.

---

# 31.17 Audio Assets

Audio resources include:

- Ambient
- Music
- UI
- Vehicles
- Buildings
- Weather
- Disasters

Audio assets are streamed when appropriate.

---

# 31.18 Asset Browser

The browser supports:

- Search
- Categories
- Favorites
- Tags
- DLC Filter
- Workshop Filter
- Recently Used

Thumbnails are generated automatically.

---

# 31.19 Asset Editor

The editor allows creators to modify:

- Mesh
- Materials
- Textures
- Metadata
- Categories
- Thumbnail
- LODs

Simulation properties are validated before export.

---

# 31.20 Validation

Asset validation checks:

- Missing Meshes
- Missing Materials
- Missing Textures
- Invalid Metadata
- Broken Dependencies
- Invalid Collision
- Invalid LODs

Invalid assets cannot be published.

---

# 31.21 Thumbnail Generation

Preview images are generated automatically.

Preview includes:

- Neutral Lighting
- Transparent Background
- Standard Camera Angle
- Consistent Scale

Creators may replace thumbnails manually.

---

# 31.22 Streaming

Assets stream dynamically.

Workflow:

```text
Camera Moves

↓

Determine Required Assets

↓

Load Nearby Assets

↓

Unload Distant Assets

↓

Update Cache
```

Streaming minimizes memory usage.

---

# 31.23 Asset Caching

Frequently used assets remain cached.

Examples:

- Common Trees
- Roads
- Residential Buildings
- Vehicles

Least recently used assets may be evicted.

---

# 31.24 Workshop Integration

Workshop assets support:

- Upload
- Download
- Ratings
- Comments
- Dependencies
- Version Updates

Workshop metadata remains separate from runtime assets.

---

# 31.25 DLC Support

Official DLC packages register assets through the same pipeline.

DLC assets may introduce:

- Buildings
- Roads
- Maps
- Vehicles
- Themes
- Music

The Asset Manager treats official and user assets consistently.

---

# 31.26 Serialization

Save files reference asset identifiers rather than embedding resources.

Example:

```text
Building Instance

↓

Asset GUID

↓

Transform

↓

Simulation Data
```

This minimizes save file size.

---

# 31.27 Performance Optimization

The Asset System employs:

- Asset streaming
- Shared resources
- Texture atlases
- GPU instancing
- Incremental loading
- Compression
- Runtime caching
- Lazy initialization

Only required assets occupy memory.

---

# 31.28 Design Goals

The Asset System is designed to provide:

- Fully data-driven content
- Efficient runtime loading
- Modular asset management
- User-generated content support
- Workshop compatibility
- Reliable dependency handling
- Memory-efficient streaming
- High-performance rendering
- Future DLC compatibility
- Complete separation from simulation

The Asset System enables both official developers and community creators to expand the game with new buildings, roads, vehicles, props, maps, and visual content while maintaining deterministic simulation, efficient resource usage, and seamless integration with every other engine subsystem.

# 32. Modding API, Plugin Framework & Extensibility

The Modding System enables developers and players to extend nearly every aspect of the game without modifying the core engine. Mods can introduce new gameplay mechanics, buildings, roads, vehicles, UI panels, scenarios, policies, AI behaviors, rendering effects, editor tools, and custom assets.

The core engine remains deterministic and sandboxed while exposing carefully designed APIs that maintain save compatibility, performance, and stability.

---

# 32.1 Modding Architecture

The Mod Manager coordinates every installed modification.

```text
ModManager

├── Plugin Loader
├── Assembly Loader
├── Dependency Resolver
├── Version Manager
├── API Registry
├── Event Dispatcher
├── Asset Integration
├── UI Extensions
├── Save Compatibility
└── Sandbox Security
```

Mods communicate only through the public API.

---

# 32.2 Mod Types

Supported mod categories include:

Gameplay Mods

UI Mods

Asset Packs

Map Packs

Scenario Packs

Code Plugins

Visual Mods

Audio Mods

Localization Packs

Editor Extensions

Multiple mod types may coexist.

---

# 32.3 Plugin Lifecycle

Every plugin follows the same lifecycle.

```text
Load

↓

Initialize

↓

Register

↓

Running

↓

Disable

↓

Unload
```

The engine guarantees proper cleanup during unloading.

---

# 32.4 Plugin Interface

Example lifecycle interface:

```cpp
interface IPlugin
{
    OnLoad();

    OnInitialize();

    OnEnable();

    OnDisable();

    OnUnload();
}
```

Plugins are initialized after engine startup and before gameplay begins.

---

# 32.5 API Categories

Public APIs include:

Simulation

Economy

Citizens

Vehicles

Buildings

Roads

Terrain

Environment

Rendering

Audio

UI

Editor

Scenario

Statistics

Save System

Mods cannot access internal engine implementations directly.

---

# 32.6 Event System

Plugins subscribe to engine events.

Examples:

```text
BuildingCreated

CitizenMoved

RoadBuilt

DisasterStarted

BudgetUpdated

VehicleSpawned

BuildingDestroyed
```

Events are dispatched asynchronously where safe.

---

# 32.7 Asset Registration

Mods may register:

- Buildings
- Roads
- Trees
- Vehicles
- Props
- Themes
- Materials
- Audio
- Icons

Assets become available through the normal asset browser.

---

# 32.8 Custom Buildings

Building mods define:

- Appearance
- Footprint
- Cost
- Maintenance
- Capacity
- Electricity
- Water
- Workers

Simulation integration follows existing building rules.

---

# 32.9 Custom Roads

Road plugins may define:

- Lane Layout
- Speed Limits
- Sidewalks
- Decorations
- Medians
- Bus Lanes
- Tram Tracks
- Bicycle Lanes

Procedural generation automatically uses these definitions.

---

# 32.10 UI Extensions

Mods may create:

- Tool Windows
- Inspector Panels
- Statistics Views
- Toolbar Buttons
- Information Views
- Notifications

UI extensions follow the native interface style.

---

# 32.11 Custom Tools

Supported editor/gameplay tools include:

- Terrain Brushes
- Placement Tools
- Analysis Tools
- Selection Tools
- Measurement Tools

Custom tools integrate with the standard input system.

---

# 32.12 Scenario Extensions

Scenario plugins may define:

- Objectives
- Events
- Rewards
- Failure Conditions
- Story Scripts
- Dialog

Scenario APIs are completely data-driven.

---

# 32.13 Policy Extensions

Mods may introduce new city policies.

Examples:

- Green Energy Initiative
- Remote Work
- Congestion Pricing
- Recycling Incentives

Policies interact through the existing simulation framework.

---

# 32.14 Advisor Extensions

Plugins may register:

- New Advisors
- Custom Recommendations
- Tutorial Messages
- Scenario Advisors

Advisor priorities follow the standard notification system.

---

# 32.15 Localization

Mods support localized text.

Resources include:

- UI
- Descriptions
- Tutorials
- Notifications
- Dialog

Localization files are loaded automatically.

---

# 32.16 Dependency Resolution

Mods may depend upon:

- Other Mods
- DLC
- Asset Packs
- API Versions

Dependency graph:

```text
Plugin

↓

Dependencies

↓

Load Order

↓

Initialization
```

Missing dependencies prevent loading.

---

# 32.17 Version Compatibility

Every plugin defines:

```cpp
Plugin Version

Minimum API Version

Maximum API Version

Compatible DLC

Dependencies
```

The engine warns players about incompatible versions.

---

# 32.18 Save Compatibility

Save files store:

- Enabled Mods
- Plugin Versions
- Asset References
- Custom Data

Missing mods generate compatibility warnings rather than immediate corruption whenever possible.

---

# 32.19 Mod Settings

Each plugin may expose:

- Checkboxes
- Sliders
- Dropdowns
- Keybindings
- Text Fields

Settings are stored independently of save files unless specified.

---

# 32.20 Script Security

Plugins execute within a restricted environment.

Restrictions include:

- Controlled filesystem access
- Controlled networking
- API-only simulation modification
- Permission validation

Unsafe operations require explicit user approval where supported.

---

# 32.21 Performance Budget

The engine monitors:

- CPU Time
- Memory Usage
- Update Time
- Event Cost

Plugins exceeding configurable limits generate warnings in developer mode.

---

# 32.22 Debugging Support

Plugin developers may access:

- Console
- Logging
- Event Viewer
- Performance Profiler
- Exception Viewer

Debug information is isolated per plugin.

---

# 32.23 Workshop Integration

Workshop supports:

- Automatic Updates
- Version History
- Ratings
- Screenshots
- Tags
- Dependency Lists
- Changelogs

Installed mods are synchronized through the Workshop client where available.

---

# 32.24 Serialization

Plugins may serialize custom data.

```text
Plugin

↓

Serialize()

↓

Save File

↓

Deserialize()

↓

Restore State
```

Plugin data remains isolated from core engine serialization.

---

# 32.25 Performance Optimization

The Mod Manager employs:

- Lazy plugin initialization
- Event filtering
- Dependency caching
- Incremental loading
- Plugin profiling
- Shared asset references

Inactive plugins consume minimal runtime resources.

---

# 32.26 Design Goals

The Modding System is designed to provide:

- Extensive engine extensibility
- Stable public APIs
- Deterministic simulation integration
- Safe plugin execution
- Efficient asset loading
- Reliable dependency management
- Save compatibility
- Workshop integration
- High-performance mod execution
- Long-term API stability

The Modding System allows the game to evolve far beyond its base feature set by enabling developers and players to create sophisticated extensions that integrate seamlessly with the engine while preserving performance, stability, and deterministic simulation.

# 33. Save System, Serialization & Persistence

The Save System is responsible for preserving the complete deterministic state of a city. Every simulation object—including citizens, vehicles, buildings, roads, economy, utilities, transport networks, weather, disasters, policies, and progression—is serialized so that loading a save reproduces the exact same city state.

The save system is versioned, extensible, deterministic, and designed to support future DLC, mods, and engine upgrades without invalidating existing saves whenever possible.

---

# 33.1 Save System Architecture

The Save Manager coordinates all persistence operations.

```text
SaveManager

├── Serialization Engine
├── Deserialization Engine
├── Version Manager
├── Compression System
├── Autosave Manager
├── Backup Manager
├── Save Validator
├── Cloud Save Interface
├── Thumbnail Generator
└── Migration Manager
```

Every simulation subsystem implements a common serialization interface.

---

# 33.2 Save Philosophy

The save system stores **simulation state**, not rendered visuals.

Saved:

- Citizens
- Buildings
- Roads
- Economy
- Policies
- Utilities
- Weather
- Vehicles
- Progression
- Statistics

Not Saved:

- GPU Buffers
- Render Queues
- Cached Shadows
- Temporary Particle Effects
- Temporary UI Windows

Visual resources are rebuilt after loading.

---

# 33.3 Save File Structure

```text
Save File

├── Header
├── Metadata
├── World
├── Terrain
├── Roads
├── Buildings
├── Citizens
├── Vehicles
├── Economy
├── Utilities
├── Transportation
├── Environment
├── Statistics
├── Progression
├── Mods
├── Custom Data
└── Footer
```

Every section is independently versioned.

---

# 33.4 Save Metadata

Metadata includes:

```cpp
SaveName

CityName

Author

GameVersion

SaveVersion

Timestamp

Population

Treasury

PlayTime

MapName

Scenario

Mods

Thumbnail
```

Metadata is readable without loading the entire save.

---

# 33.5 Serialization Interface

Every subsystem implements:

```cpp
interface ISerializable
{
    Serialize(Stream);

    Deserialize(Stream);

    GetVersion();
}
```

This standardizes save behavior across the engine.

---

# 33.6 Save Order

Serialization follows a deterministic order.

```text
World

↓

Terrain

↓

Road Network

↓

Buildings

↓

Utilities

↓

Citizens

↓

Vehicles

↓

Economy

↓

Environment

↓

Statistics

↓

UI State
```

Ordering ensures object references remain valid.

---

# 33.7 Object References

Entities reference one another through stable IDs.

Example:

```text
Citizen

↓

Home ID

↓

Building

↓

Road Connection

↓

District
```

Pointers are never written directly to disk.

---

# 33.8 Stable Identifiers

Every runtime object receives:

```cpp
EntityID

AssetGUID

ParentID

OwnerID
```

Stable identifiers simplify migration and mod compatibility.

---

# 33.9 Incremental Serialization

Large collections serialize incrementally.

Examples:

- Citizens
- Vehicles
- Buildings
- Trees

This reduces memory spikes during saving.

---

# 33.10 Compression

After serialization:

```text
Raw Data

↓

Compress

↓

Checksum

↓

Write File
```

Compression reduces storage requirements without altering simulation data.

---

# 33.11 Save Validation

Validation checks:

- Missing Objects
- Broken References
- Invalid IDs
- Duplicate IDs
- Corrupted Sections
- Version Compatibility

Invalid saves are rejected before loading.

---

# 33.12 Loading Pipeline

```text
Open File

↓

Validate Header

↓

Check Version

↓

Decompress

↓

Deserialize

↓

Resolve References

↓

Initialize Systems

↓

Resume Simulation
```

The simulation begins only after all systems finish loading.

---

# 33.13 Versioning

Every save contains:

```cpp
MajorVersion

MinorVersion

BuildVersion

SchemaVersion
```

Migration rules convert older saves when possible.

---

# 33.14 Save Migration

When versions differ:

```text
Old Save

↓

Migration Rules

↓

Updated Save Structure

↓

Load
```

Migration preserves gameplay whenever technically feasible.

---

# 33.15 Autosave

Autosaves occur periodically.

Configurable options:

- Every 5 Minutes
- Every 10 Minutes
- Every 15 Minutes
- Every 30 Minutes
- Disabled

Autosaves use rotating save slots.

---

# 33.16 Backup Saves

Before overwriting:

```text
Current Save

↓

Backup

↓

Write New Save
```

Backups protect against interrupted saves or corruption.

---

# 33.17 Manual Saves

Players may:

- Create
- Rename
- Overwrite
- Delete
- Duplicate

Manual saves are unlimited except where platform restrictions apply.

---

# 33.18 Cloud Saves

Platform support may provide:

- Automatic Upload
- Automatic Download
- Conflict Resolution
- Cross-device Synchronization

Cloud functionality is abstracted behind the platform layer.

---

# 33.19 Save Browser

The browser displays:

- Thumbnail
- City Name
- Population
- Treasury
- Save Date
- Play Time
- Game Version
- Map
- Active Mods

Sorting options include:

- Name
- Date
- Population
- Play Time

---

# 33.20 Corruption Recovery

Recovery attempts include:

- Checksum Validation
- Backup Restore
- Partial Recovery
- Reference Repair

Players receive detailed error information if recovery fails.

---

# 33.21 Mod Data

Each plugin serializes independently.

```text
Core Save

↓

Plugin Section

↓

Plugin Serializer

↓

Plugin Restore
```

Missing plugins do not invalidate the core save unless essential data cannot be reconstructed.

---

# 33.22 Scenario Saves

Scenario saves additionally store:

- Objective Progress
- Active Events
- Timers
- Scores
- Checkpoints

Scenario state resumes exactly after loading.

---

# 33.23 Multiplayer Compatibility (Future)

The save format is designed to support future multiplayer synchronization by maintaining deterministic world state and stable object identifiers.

---

# 33.24 Performance Optimization

The Save Manager employs:

- Background serialization
- Incremental writing
- Compression
- Chunked saves
- Object pooling
- Reference caching
- Parallel serialization where safe

Saving minimizes gameplay interruption.

---

# 33.25 Design Goals

The Save System is designed to provide:

- Deterministic persistence
- Version compatibility
- Efficient serialization
- Reliable corruption recovery
- Mod compatibility
- Fast loading
- Background saving
- Stable object references
- Cloud integration
- Future extensibility

The Save System ensures that every aspect of a city—from individual citizens to complex transportation networks—is preserved accurately, allowing players to resume their cities exactly as they left them while maintaining compatibility across updates, DLC, and modded content.

# 34. Platform Services, Distribution & Online Features

The Platform Services System integrates the game with operating system and distribution platform functionality while remaining completely separate from the core simulation. Platform features include achievements, cloud saves, Workshop integration, leaderboards, DLC management, telemetry (optional), licensing, input devices, and platform-specific services.

All platform services are abstracted behind a unified interface, allowing the game to run consistently across multiple operating systems and storefronts.

---

# 34.1 Platform Architecture

The Platform Manager coordinates all external platform integrations.

```text
PlatformManager

├── Platform API
├── Steam Integration
├── Epic Integration
├── GOG Integration
├── Xbox Integration
├── PlayStation Integration
├── Cloud Save Manager
├── Achievement Service
├── Leaderboard Service
├── DLC Manager
├── Workshop Interface
└── Licensing Manager
```

The simulation engine never communicates directly with platform-specific APIs.

---

# 34.2 Supported Platforms

Platform abstraction supports:

- Windows
- Linux
- macOS
- Xbox
- PlayStation
- Future Platforms

Each platform implements the same public interface.

---

# 34.3 Platform Initialization

Initialization sequence:

```text
Launch Game

↓

Detect Platform

↓

Initialize Platform API

↓

Authenticate User

↓

Load Cloud Data

↓

Initialize DLC

↓

Start Game
```

Platform failures never corrupt gameplay data.

---

# 34.4 User Profiles

Platform profiles provide:

- Username
- Avatar
- Unique Platform ID
- Achievement Progress
- Cloud Save Access
- Friends List (Platform Support)

The game maintains its own local player configuration independently.

---

# 34.5 Achievement Integration

Platform achievements synchronize with:

- Population Milestones
- Scenario Completion
- Construction Goals
- Transport Goals
- Environmental Goals
- Hidden Achievements

Synchronization occurs asynchronously.

---

# 34.6 Leaderboards

Leaderboards may track:

- Scenario Completion Time
- Population
- City Value
- Efficiency Score
- Tourism
- Industry Production

Leaderboard participation is optional.

---

# 34.7 Cloud Saves

Cloud synchronization includes:

```text
Local Save

↓

Upload

↓

Cloud Storage

↓

Download

↓

Restore
```

Conflict resolution supports:

- Newest Save
- Local Version
- Cloud Version
- Manual Selection

---

# 34.8 DLC Management

The DLC Manager detects installed content.

Examples:

- Official Expansions
- Content Creator Packs
- Radio Stations
- Cosmetic Packs

DLC content is registered through the Asset Manager.

---

# 34.9 Dynamic DLC Loading

Installed DLC becomes available during startup.

```text
Scan Installed DLC

↓

Validate

↓

Register Assets

↓

Register Systems

↓

Enable Features
```

Missing DLC gracefully disables related content.

---

# 34.10 Workshop Integration

Workshop functionality includes:

- Browse
- Subscribe
- Unsubscribe
- Automatic Updates
- Dependency Resolution
- Version Detection

Downloaded content is validated before loading.

---

# 34.11 User Content

Players may publish:

- Maps
- Assets
- Mods
- Scenarios
- Themes

Metadata includes:

- Screenshots
- Tags
- Description
- Version
- Dependencies

---

# 34.12 Friend Features (Platform Support)

Optional platform functionality:

- Friend List
- Recently Played
- Shared Workshop Items
- Shared Scenarios
- Challenge Scores

No gameplay depends upon social features.

---

# 34.13 Licensing

The Licensing Manager verifies:

- Base Game
- DLC Ownership
- Expansion Access
- Trial Restrictions

Licensing checks never affect save integrity.

---

# 34.14 Offline Mode

Offline mode supports:

- Full Simulation
- Local Saves
- Local Assets
- Local Mods
- Scenario Play

Unavailable features:

- Cloud Saves
- Workshop
- Online Leaderboards

Offline functionality is fully supported.

---

# 34.15 Telemetry (Optional)

Optional analytics may collect:

- Crash Reports
- Performance Metrics
- Hardware Information
- Anonymous Gameplay Statistics

Telemetry never includes personal city save data unless explicitly approved.

---

# 34.16 Crash Reporting

Crash reports may include:

- Engine Version
- Stack Trace
- Loaded Mods
- Hardware
- Active DLC

Reports require user consent where applicable.

---

# 34.17 Localization Services

Platform integration provides:

- System Language Detection
- Region Detection
- Currency Detection
- Time Format

Players may override automatic settings.

---

# 34.18 Input Services

Supported input devices:

- Keyboard
- Mouse
- Controller
- Touch (Platform Support)

Input abstraction ensures identical gameplay logic across devices.

---

# 34.19 Update System

Game updates support:

```text
Detect Update

↓

Download

↓

Verify

↓

Install

↓

Migration

↓

Launch
```

Save migration occurs automatically when necessary.

---

# 34.20 Mod Compatibility

Platform updates verify:

- API Version
- Plugin Compatibility
- Asset Versions
- Save Compatibility

Incompatible mods are disabled rather than loaded unsafely.

---

# 34.21 Security

Platform services enforce:

- Package Validation
- Signed DLC
- Secure Save Locations
- Safe Plugin Loading
- Permission Management

Security operates independently of gameplay systems.

---

# 34.22 Error Recovery

Platform failures are categorized as:

- Network Failure
- Authentication Failure
- Cloud Conflict
- Workshop Error
- DLC Validation Error

Most failures fall back to local gameplay without interruption.

---

# 34.23 Statistics

Platform metrics include:

- Total Play Time
- Sessions
- Achievements Earned
- Workshop Downloads
- DLC Usage
- Cloud Sync Status

Platform statistics remain separate from gameplay statistics.

---

# 34.24 Performance Optimization

The Platform Manager employs:

- Asynchronous networking
- Background downloads
- Cached authentication
- Deferred synchronization
- Incremental Workshop updates

Online services never block the simulation thread.

---

# 34.25 Design Goals

The Platform Services System is designed to provide:

- Cross-platform compatibility
- Reliable cloud saves
- Seamless DLC integration
- Safe Workshop support
- Optional online features
- Offline gameplay
- Stable licensing
- Efficient background networking
- Future platform extensibility
- Complete separation from simulation

The Platform Services System connects the game to modern platform ecosystems while ensuring that the core city simulation remains deterministic, fully playable offline, and independent of external services.

# 35. Final Integration, Boot Sequence & Complete Runtime Flow

This chapter describes how every subsystem comes together to produce a fully playable city simulation. It defines the complete runtime lifecycle from launching the game to shutting it down.

At this point, every major gameplay system has been implemented:

- Terrain
- Roads
- Zoning
- Buildings
- Citizens
- Vehicles
- Economy
- Utilities
- Public Services
- Public Transport
- Districts
- Policies
- Weather
- Water
- Disasters
- Rendering
- UI
- Audio
- Progression
- Statistics
- Saving
- Modding
- Asset Management
- Scenario System
- Platform Services

This final chapter explains how these systems cooperate.

---

# 35.1 Engine Boot Sequence

When the game starts, initialization follows a deterministic order.

```text
Executable

↓

Platform Layer

↓

Configuration

↓

Asset Database

↓

Graphics Engine

↓

Audio Engine

↓

Input System

↓

Physics Engine

↓

Simulation Engine

↓

UI

↓

Mods

↓

Workshop Assets

↓

Load Save / New Game

↓

Gameplay Starts
```

Every subsystem must complete initialization before gameplay begins.

---

# 35.2 Creating a New City

When starting a new city:

```text
Choose Map

↓

Choose Theme

↓

Select Starting Options

↓

Generate Terrain

↓

Generate Resources

↓

Initialize Outside Connections

↓

Initialize Simulation

↓

Spawn Initial Infrastructure

↓

Begin Simulation
```

The simulation always begins in a valid state.

---

# 35.3 Loading an Existing Save

Loading performs:

```text
Load Save

↓

Version Check

↓

Migration

↓

Deserialize Systems

↓

Resolve References

↓

Initialize Managers

↓

Generate Render State

↓

Resume Simulation
```

No simulation updates occur until loading completes successfully.

---

# 35.4 Main Runtime Loop

During gameplay:

```text
Read Input

↓

Simulation Tick

↓

AI

↓

Traffic

↓

Economy

↓

Utilities

↓

Services

↓

Statistics

↓

Generate Snapshot

↓

Rendering

↓

Audio

↓

UI

↓

Present Frame
```

Rendering never modifies simulation.

---

# 35.5 Typical Gameplay Example

Player builds a road.

```text
Road Tool

↓

Road Validation

↓

Terrain Modified

↓

Road Network Updated

↓

Pathfinding Updated

↓

Utility Network Updated

↓

Traffic Graph Updated

↓

Zones Updated

↓

Rendering Updated

↓

Save Dirty Flag
```

Only affected systems update.

---

# 35.6 Example Building Placement

Player places a hospital.

```text
Hospital Asset

↓

Construction Cost

↓

Terrain Validation

↓

Road Connection

↓

Power Connection

↓

Water Connection

↓

Healthcare System Updated

↓

Coverage Updated

↓

Rendering Updated
```

Citizens immediately begin considering the new service.

---

# 35.7 Example Citizen Simulation

Citizen wakes up.

```text
Home

↓

Need Generated

↓

Choose Destination

↓

Find Route

↓

Walk

↓

Bus

↓

Walk

↓

Work

↓

Return Home
```

Every decision is produced through simulation rules.

---

# 35.8 Example Traffic Update

A road becomes congested.

```text
Traffic Density

↓

Congestion Detected

↓

Vehicles Slow

↓

Travel Time Increases

↓

Citizens Recalculate Routes

↓

Transport Demand Changes

↓

Statistics Updated
```

Congestion propagates naturally through the transportation system.

---

# 35.9 Example Economy Update

Weekly budget.

```text
Collect Taxes

↓

Calculate Expenses

↓

Industry Income

↓

Transport Revenue

↓

Loan Payments

↓

Treasury Updated

↓

Statistics Updated
```

Every financial system contributes.

---

# 35.10 Example Disaster

A tornado begins.

```text
Spawn Tornado

↓

Weather Updated

↓

Buildings Damaged

↓

Citizens Evacuate

↓

Emergency Services Respond

↓

Recovery Begins

↓

Statistics Recorded
```

Disasters integrate with every simulation subsystem.

---

# 35.11 Autosave

Every autosave:

```text
Pause Save Thread

↓

Serialize

↓

Compress

↓

Write

↓

Resume
```

Simulation continues whenever background saving is available.

---

# 35.12 Shutdown

When exiting:

```text
Pause

↓

Autosave (Optional)

↓

Unload Mods

↓

Unload Assets

↓

Destroy Rendering

↓

Destroy Audio

↓

Release Memory

↓

Exit
```

Resources are released in reverse dependency order.

---

# 35.13 Complete Dependency Graph

```text
Platform
    │
    ▼
Core Engine
    │
    ├──────────────┐
    ▼              ▼
Simulation      Rendering
    │              │
    ▼              ▼
Economy       GPU Renderer
Traffic       Lighting
Citizens      Shadows
Buildings     Terrain
Utilities     Water
Weather       Particles
Disasters     UI
    │              │
    └──────┬───────┘
           ▼
        Save System
           │
           ▼
        Mod System
```

Subsystems communicate through events and public interfaces rather than direct dependencies.

---

# 35.14 Performance Goals

Target performance:

- Fixed simulation timestep
- Stable frame pacing
- Efficient multithreading
- GPU instancing
- Chunk-based world updates
- Background streaming
- Background saving
- Deterministic simulation
- Minimal runtime allocations

The engine is designed to scale from villages with hundreds of entities to metropolitan regions containing millions of simulated objects.

---

# 35.15 Future Expansion Support

The engine is intentionally designed to support future additions without breaking existing systems.

Planned extension points include:

- Multiplayer
- Multiplayer synchronization
- Seasons
- Dynamic economy
- Construction phases
- Procedural events
- Additional transport modes
- Expanded industries
- Agriculture overhaul
- Mixed-use zoning
- Government simulation
- Elections
- Regional simulation
- Multiple connected cities
- Asset bundles
- New AI behaviors
- New disaster types
- New progression systems

These features can be implemented through existing extension interfaces.

---

# 35.16 Final Design Principles

The engine follows several core principles.

### Determinism

Given identical input:

- Same seed
- Same simulation
- Same outcome

---

### Data-Oriented Design

Simulation data is separated from rendering.

Logic is separated from presentation.

---

### Event-Driven Communication

Systems communicate through events instead of direct coupling.

This improves maintainability and scalability.

---

### Modularity

Every subsystem can evolve independently.

Examples:

- Replace rendering engine.
- Replace UI framework.
- Expand transport system.

No gameplay rewrite is required.

---

### Extensibility

Every major subsystem exposes controlled public interfaces.

Official DLC and community mods use the same architecture.

---

### Performance First

Large cities remain performant through:

- ECS architecture
- Spatial partitioning
- Parallel jobs
- Incremental updates
- Streaming
- LOD
- Culling
- Object pooling
- Cache-friendly memory layouts

---

### Save Stability

Every save stores only simulation state.

Visual resources regenerate automatically.

This minimizes save size while maximizing compatibility.

---

# 35.17 Project Completion

At the completion of this specification, the engine defines:

✓ Complete simulation architecture

✓ Terrain generation

✓ Road network system

✓ Procedural road geometry

✓ Citizen AI

✓ Vehicle AI

✓ Advanced pathfinding

✓ Zone simulation

✓ Building simulation

✓ Economy

✓ Budgeting

✓ Utilities

✓ Electricity

✓ Water & sewage

✓ Garbage

✓ Healthcare

✓ Education

✓ Police

✓ Fire services

✓ Public transportation

✓ Districts

✓ Policies

✓ Industries

✓ Tourism

✓ Environment

✓ Pollution

✓ Weather

✓ Dynamic water

✓ Natural disasters

✓ Progression

✓ Milestones

✓ Rendering engine

✓ UI framework

✓ Audio engine

✓ Asset pipeline

✓ Map editor

✓ Scenario editor

✓ Statistics

✓ Achievements

✓ Save/load

✓ Modding API

✓ Workshop support

✓ Platform services

✓ Performance architecture

✓ Extensibility

✓ Deterministic simulation

---

# 35.18 Specification Status

This concludes the functional game specification.

The document now defines a complete city-building simulation architecture comparable in scope to a modern commercial city simulation engine. It specifies the interactions, responsibilities, lifecycle, and data flow of all major gameplay and engine systems required to build a fully playable city-building game with deterministic simulation, extensible architecture, and long-term maintainability.

From a systems-design perspective, the specification is sufficiently complete to begin implementation of the engine and gameplay systems.

---

# Appendix A — Remaining Implementation Artifacts (Not Covered by This Specification)

The following items are intentionally excluded because they are implementation details rather than game design specifications:

- Database/file formats
- Source code architecture (packages/namespaces)
- Exact ECS component layouts
- Network protocol (future multiplayer)
- GPU shader implementations
- Physics solver implementation
- Pathfinding algorithm optimizations
- Build system (CMake, Meson, etc.)
- Continuous Integration/Deployment
- Testing framework
- Localization file formats
- Asset compiler implementation
- Binary serialization format
- Memory allocator implementation
- Rendering backend (DirectX/Vulkan/Metal/OpenGL)
- Exact editor UI implementation
- Engine coding standards
- Unit/integration testing strategy
- Performance benchmark suite

These are engineering decisions to be made during implementation rather than gameplay requirements.
