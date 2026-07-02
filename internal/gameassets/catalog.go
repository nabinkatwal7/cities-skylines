package gameassets

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Catalog holds Kenney CC0 models and surface textures loaded at startup.
type Catalog struct {
	BuildingSets map[string][]rl.Model
	Trees        [5]rl.Model // oak, pine, birch, palm, dead
	Vehicles     VehicleModels
	Grass        rl.Texture2D
	Road         rl.Texture2D
	Water        rl.Texture2D
}

type VehicleModels struct {
	Car       rl.Model
	Bus       rl.Model
	Truck     rl.Model
	Emergency rl.Model
	Bike      rl.Model
	HasAny    bool
}

var buildingSetDirs = []string{
	"residential_low", "residential_high",
	"commercial_low", "commercial_high",
	"industrial", "office",
}

func Load() (*Catalog, error) {
	c := &Catalog{BuildingSets: make(map[string][]rl.Model)}
	for _, name := range buildingSetDirs {
		dir := filepath.Join("assets", "buildings", name)
		models, err := loadModelDir(dir)
		if err != nil {
			return nil, fmt.Errorf("buildings %s: %w", dir, err)
		}
		c.BuildingSets[name] = models
	}
	treeNames := []string{"oak", "pine", "birch", "palm", "dead"}
	for i, name := range treeNames {
		path := filepath.Join("assets", "trees", name+".obj")
		if _, err := os.Stat(path); err == nil {
			c.Trees[i] = rl.LoadModel(path)
		}
	}
	vehMap := map[string]*rl.Model{
		"car": &c.Vehicles.Car, "bus": &c.Vehicles.Bus, "truck": &c.Vehicles.Truck,
		"emergency": &c.Vehicles.Emergency, "bike": &c.Vehicles.Bike,
	}
	for name, slot := range vehMap {
		path := filepath.Join("assets", "vehicles", name+".obj")
		if _, err := os.Stat(path); err == nil {
			*slot = rl.LoadModel(path)
			c.Vehicles.HasAny = true
		}
	}
	c.Grass = loadTex("assets/textures/grass.png")
	c.Road = loadTex("assets/textures/road.png")
	c.Water = loadTex("assets/textures/water.png")
	return c, nil
}

func (c *Catalog) Unload() {
	if c == nil {
		return
	}
	for _, list := range c.BuildingSets {
		for i := range list {
			if list[i].MeshCount > 0 {
				rl.UnloadModel(list[i])
			}
		}
	}
	for i := range c.Trees {
		if c.Trees[i].MeshCount > 0 {
			rl.UnloadModel(c.Trees[i])
		}
	}
	for _, m := range []*rl.Model{&c.Vehicles.Car, &c.Vehicles.Bus, &c.Vehicles.Truck, &c.Vehicles.Emergency, &c.Vehicles.Bike} {
		if m.MeshCount > 0 {
			rl.UnloadModel(*m)
		}
	}
	if c.Grass.ID != 0 {
		rl.UnloadTexture(c.Grass)
		c.Grass = rl.Texture2D{}
	}
	if c.Road.ID != 0 {
		rl.UnloadTexture(c.Road)
		c.Road = rl.Texture2D{}
	}
	if c.Water.ID != 0 {
		rl.UnloadTexture(c.Water)
		c.Water = rl.Texture2D{}
	}
}

func (c *Catalog) BuildingModel(set string, id uint32) (rl.Model, bool) {
	if c == nil || set == "" {
		return rl.Model{}, false
	}
	list := c.BuildingSets[set]
	if len(list) == 0 {
		return rl.Model{}, false
	}
	return list[int(id)%len(list)], true
}

func (c *Catalog) TreeModel(species int, dead bool) rl.Model {
	if c == nil {
		return rl.Model{}
	}
	if dead && c.Trees[4].MeshCount > 0 {
		return c.Trees[4]
	}
	if species < 0 || species >= 4 {
		species = 0
	}
	if c.Trees[species].MeshCount > 0 {
		return c.Trees[species]
	}
	for i := range c.Trees {
		if c.Trees[i].MeshCount > 0 {
			return c.Trees[i]
		}
	}
	return rl.Model{}
}

func (c *Catalog) VehicleModel(vt int) (rl.Model, bool) {
	if c == nil || !c.Vehicles.HasAny {
		return rl.Model{}, false
	}
	switch vt {
	case 0:
		return c.Vehicles.Car, c.Vehicles.Car.MeshCount > 0
	case 1:
		return c.Vehicles.Bus, c.Vehicles.Bus.MeshCount > 0
	case 2:
		return c.Vehicles.Truck, c.Vehicles.Truck.MeshCount > 0
	case 3:
		return c.Vehicles.Emergency, c.Vehicles.Emergency.MeshCount > 0
	case 4:
		return c.Vehicles.Bike, c.Vehicles.Bike.MeshCount > 0
	default:
		return c.Vehicles.Car, c.Vehicles.Car.MeshCount > 0
	}
}

// FitScale returns uniform scale so model footprint fits target width/depth.
func FitScale(model rl.Model, targetW, targetD float32) float32 {
	if model.MeshCount == 0 {
		return 1
	}
	bb := rl.GetModelBoundingBox(model)
	w := bb.Max.X - bb.Min.X
	d := bb.Max.Z - bb.Min.Z
	maxDim := w
	if d > maxDim {
		maxDim = d
	}
	if maxDim < 0.01 {
		return 1
	}
	target := targetW
	if targetD > target {
		target = targetD
	}
	return target / maxDim
}

// GroundOffset returns Y translation so model base sits on terrain.
func GroundOffset(model rl.Model, scale float32) float32 {
	if model.MeshCount == 0 {
		return 0
	}
	bb := rl.GetModelBoundingBox(model)
	return -bb.Min.Y * scale
}

func loadModelDir(dir string) ([]rl.Model, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".obj") {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no obj models in %s", dir)
	}
	// Stable order: 0.obj, 1.obj, ...
	out := make([]rl.Model, 0, len(paths))
	for i := 0; ; i++ {
		p := filepath.Join(dir, fmt.Sprintf("%d.obj", i))
		if _, err := os.Stat(p); err != nil {
			break
		}
		m := rl.LoadModel(p)
		if !rl.IsModelValid(m) {
			continue
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("failed to load models from %s", dir)
	}
	return out, nil
}

func loadTex(path string) rl.Texture2D {
	if _, err := os.Stat(path); err != nil {
		return rl.Texture2D{}
	}
	return rl.LoadTexture(path)
}

// VehicleYawDeg converts simulation heading to model rotation (degrees).
func VehicleYawDeg(yawRad float32) float32 {
	return yawRad * 180 / float32(math.Pi)
}
