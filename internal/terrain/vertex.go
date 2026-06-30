package terrain

type Vertex struct {
	Height      float32
	WaterLevel  float32
	Pollution   float32
	Fertility   float32
	Ore         float32
	Oil         float32
	Forestry    float32
	Buildability float32
}

type Biome int

const (
	BiomeOcean      Biome = 0
	BiomeBeach      Biome = 1
	BiomeGrassland  Biome = 2
	BiomeForest     Biome = 3
	BiomeDesert     Biome = 4
	BiomeTundra     Biome = 5
	BiomeMountain   Biome = 6
)
