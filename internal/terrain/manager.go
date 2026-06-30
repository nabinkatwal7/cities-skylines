package terrain

import (
	"fmt"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ProgressInfo struct {
	Done  int
	Total int
}

type Manager struct {
	Generator   *Generator
	Heightmap   *Heightmap
	Water       *WaterSystem
	Terraform   *TerraformSystem
	Trees       *TreeSystem
	Resources   *ResourceSystem
	Connections *ConnectionSystem
	Chunks      []*Chunk
	Models      []rl.Model
	Seed        int64
	terrainTex  rl.Texture2D
	uploadIdx   int
}

func NewManager(seed int64) *Manager {
	return &Manager{Seed: seed}
}

func (m *Manager) GenerateData() {
	m.Generator = NewGenerator(m.Seed)
	m.Heightmap = m.Generator.Generate()
	m.Water = NewWaterSystem()
	m.Water.Init(m.Heightmap)
	m.Terraform = NewTerraformSystem(m)
	m.Trees = NewTreeSystem(m.Seed)
	m.Trees.Generate(m.Heightmap, m.Water)
	m.Resources = NewResourceSystem(m.Seed, m.Heightmap)
	m.Connections = NewConnectionSystem()
	m.Chunks = BuildChunks(m.Heightmap)
	for _, c := range m.Chunks {
		BuildChunkMesh(c)
	}
}

func (m *Manager) LoadAssets() error {
	model := rl.LoadModel("assets/tree/leaftree.obj")
	if rl.IsModelValid(model) {
		m.Trees.Model = model
	} else {
		rl.UnloadModel(model)
		return fmt.Errorf("failed to load tree model")
	}

	grassTex := rl.LoadTexture("assets/grass.png")
	if grassTex.ID != 0 {
		m.terrainTex = grassTex
	}
	return nil
}

func (m *Manager) PrepareUpload() {
	if m.terrainTex.ID == 0 {
		grassImg := rl.GenImageCellular(512, 512, 32)
		rl.ImageColorTint(grassImg, rl.NewColor(50, 130, 30, 255))
		rl.ImageColorBrightness(grassImg, -15)
		m.terrainTex = rl.LoadTextureFromImage(grassImg)
		rl.UnloadImage(grassImg)
	}

	m.Models = make([]rl.Model, len(m.Chunks))
	m.uploadIdx = 0
}

func (m *Manager) UploadNextBatch(count int) bool {
	end := m.uploadIdx + count
	if end > len(m.Chunks) {
		end = len(m.Chunks)
	}
	for i := m.uploadIdx; i < end; i++ {
		model := chunkToModel(m.Chunks[i])
		mats := model.GetMaterials()
		if len(mats) > 0 {
			rl.SetMaterialTexture(&mats[0], rl.MapAlbedo, m.terrainTex)
		}
		m.Models[i] = model
	}
	m.uploadIdx = end

	if m.uploadIdx >= len(m.Chunks) {
		return true
	}
	return false
}

func (m *Manager) UploadProgress() ProgressInfo {
	return ProgressInfo{Done: m.uploadIdx, Total: len(m.Chunks)}
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
	model := chunkToModel(c)
	mats := model.GetMaterials()
	if len(mats) > 0 {
		rl.SetMaterialTexture(&mats[0], rl.MapAlbedo, m.terrainTex)
	}
	m.Models[chunkIdx] = model
}

func (m *Manager) Update(dt float64) {
	if m.Water != nil {
		m.Water.Update(dt)
	}
	if m.Trees != nil {
		m.Trees.Update(dt)
	}
}

func (m *Manager) Draw(camX, camZ float32) {
	maxChunkDist := float32(3000)
	for i, model := range m.Models {
		cx, cz := chunkCenter(m.Chunks[i])
		dx := cx - camX
		dz := cz - camZ
		if dx*dx+dz*dz > maxChunkDist*maxChunkDist {
			continue
		}
		rl.DrawModel(model, rl.NewVector3(0, 0, 0), 1, rl.White)
	}
	if m.Trees != nil {
		m.Trees.Draw(m.Heightmap, camX, camZ)
	}
	if m.Water != nil {
		m.Water.Draw()
	}
}

func (m *Manager) Unload() {
	for _, model := range m.Models {
		rl.UnloadModel(model)
	}
	if m.Trees.Model.MeshCount > 0 {
		rl.UnloadModel(m.Trees.Model)
	}
	rl.UnloadTexture(m.terrainTex)
	m.Models = nil
}

func chunkCenter(c *Chunk) (float32, float32) {
	scale := WorldSize / float32(HeightmapSize-1)
	cx := float32(c.IndexX*(ChunkSize-1))*scale - WorldSize/2 + float32(ChunkSize-1)*scale/2
	cz := float32(c.IndexZ*(ChunkSize-1))*scale - WorldSize/2 + float32(ChunkSize-1)*scale/2
	return cx, cz
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

func clearModelMeshData(model *rl.Model) {
	meshes := model.GetMeshes()
	for i := range meshes {
		meshes[i].Vertices = nil
		meshes[i].Normals = nil
		meshes[i].Texcoords = nil
		meshes[i].Texcoords2 = nil
		meshes[i].Colors = nil
		meshes[i].Indices = nil
		meshes[i].Tangents = nil
		meshes[i].AnimVertices = nil
		meshes[i].AnimNormals = nil
		meshes[i].BoneIndices = nil
		meshes[i].BoneWeights = nil
	}
}
