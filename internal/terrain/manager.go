package terrain

import (
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Manager struct {
	Generator  *Generator
	Heightmap  *Heightmap
	Water      *WaterSystem
	Terraform  *TerraformSystem
	Trees      *TreeSystem
	Chunks     []*Chunk
	Models     []rl.Model
	Seed       int64
}

func NewManager(seed int64) *Manager {
	m := &Manager{Seed: seed}
	return m
}

func (m *Manager) Generate() {
	m.Generator = NewGenerator(m.Seed)
	m.Heightmap = m.Generator.Generate()
	m.Water = NewWaterSystem()
	m.Water.Init(m.Heightmap)
	m.Terraform = NewTerraformSystem(m)
	m.Trees = NewTreeSystem(m.Seed)
	m.Trees.Generate(m.Heightmap, m.Water)
	m.Chunks = BuildChunks(m.Heightmap)
	m.buildModels()
}

func (m *Manager) buildModels() {
	m.Models = make([]rl.Model, len(m.Chunks))
	for i, c := range m.Chunks {
		BuildChunkMesh(c)
		m.Models[i] = chunkToModel(c)
	}
}

func (m *Manager) RebuildChunk(chunkIdx int) {
	if chunkIdx < 0 || chunkIdx >= len(m.Chunks) {
		return
	}
	c := m.Chunks[chunkIdx]
	if !c.Dirty {
		return
	}
	rl.UnloadModel(m.Models[chunkIdx])
	BuildChunkMesh(c)
	m.Models[chunkIdx] = chunkToModel(c)
}

func (m *Manager) Update(dt float64) {
	if m.Water != nil {
		m.Water.Update(dt)
	}
	if m.Trees != nil {
		m.Trees.Update(dt)
	}
}

func (m *Manager) Draw() {
	for _, model := range m.Models {
		rl.DrawModel(model, rl.NewVector3(0, 0, 0), 1, rl.White)
	}
	if m.Trees != nil {
		m.Trees.Draw(m.Heightmap)
	}
	if m.Water != nil {
		m.Water.Draw()
	}
}

func (m *Manager) Unload() {
	for _, model := range m.Models {
		rl.UnloadModel(model)
	}
	if m.Water != nil {
		m.Water.Unload()
	}
	m.Models = nil
}

func chunkToModel(c *Chunk) rl.Model {
	md := c.MeshData
	mesh := rl.Mesh{
		VertexCount:   md.VertexCount,
		TriangleCount: md.IndexCount / 3,
		Vertices:      &md.Vertices[0],
		Normals:       &md.Normals[0],
		Texcoords:     &md.TexCoords[0],
		Indices:       (*uint16)(unsafe.Pointer(&md.Indices[0])),
	}
	rl.UploadMesh(&mesh, false)
	return rl.LoadModelFromMesh(mesh)
}
