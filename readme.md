# JS Skylines

A modern city-building and simulation game inspired by Cities: Skylines, built entirely in **Go**.

The project focuses on a deterministic simulation engine, data-oriented architecture, high-performance manager-based systems, procedural world generation, and large-scale city simulation.

---

## Features

- 🚗 Deterministic simulation engine
- 🏙️ Dynamic zoning and city growth
- 🛣️ Procedural road generation
- 🚦 Advanced traffic simulation
- 🚍 Public transportation system
- 👥 Individual citizen simulation
- ⚡ Utility network simulation
- 🚓 Public service dispatch system
- 💰 Economy and taxation
- 🌦️ Weather simulation
- 🌋 Disaster simulation
- 🧵 Multithreaded job system
- 💾 Save / Load support
- 🎮 Real-time rendering

---

# Architecture

The engine is designed around a **Simulation Manager** that owns every gameplay subsystem.

<p align="center">
    <img src="./architecture.png" width="100%">
</p>

---

# Simulation Flow

Every simulation tick follows a deterministic update order.

<p align="center">
    <img src="./systemflow.png" width="100%">
</p>

---

# System Interaction

Subsystems communicate through the Event Bus and Simulation Manager rather than directly.

<p align="center">
    <img src="./interaction.png" width="100%">
</p>

---

# Gameplay

Current in-game progress.

<p align="center">
    <img src="./gameplay.png" width="100%">
</p>

---

# Project Structure

```text
.
├── assets/
├── cmd/
│   └── js-skylines/
├── internal/
│   ├── terrain/
│   ├── roads/
│   ├── simulation/
│   ├── rendering/
│   ├── citizens/
│   ├── transport/
│   ├── economy/
│   └── ...
├── architecture.png
├── gameplay.png
├── interaction.png
├── systemflow.png
├── features.md
├── implemented-features.md
├── README.md
├── go.mod
└── go.sum
```

---

# Build

```bash
CGO_ENABLED=0 go build -o js-skylines.exe ./cmd/js-skylines/
```

---

# Run

```bash
./js-skylines.exe
```

Windows

```powershell
js-skylines.exe
```

---

# Documentation

- `features.md` — Complete game specification
- `implemented-features.md` — Current implementation progress

---

# Engine Goals

- Deterministic simulation
- Data-oriented design
- Manager-based architecture
- Event-driven communication
- Multithreaded job scheduling
- High-performance pathfinding
- Large-scale cities
- Efficient serialization
- Modular gameplay systems
- Extensible architecture

---

# Development Status

This project is actively under development.

Current focus includes:

- Terrain generation
- Simulation scheduler
- Road system
- Traffic AI
- Building placement
- Citizen simulation
- Rendering improvements
