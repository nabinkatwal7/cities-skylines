package sim

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/katwate/js-skylines/internal/gameassets"
	"github.com/katwate/js-skylines/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ProgressInfo struct {
	Done  int
	Total int
}

type Manager struct {
	Sim        *SimulationManager
	Chunks     []*terrain.Chunk
	Models     [terrain.NumLODs][]rl.Model
	terrainTex rl.Texture2D
	uploadIdx  int
	Assets     *gameassets.Catalog
}

func NewManager(sim *SimulationManager) *Manager {
	return &Manager{Sim: sim}
}

func (m *Manager) InitChunks() {
	m.Chunks = terrain.BuildChunks(m.Sim.Heightmap)
	for _, c := range m.Chunks {
		terrain.BuildAllLODMeshes(c)
	}
}

func (m *Manager) LoadAssets() error {
	cat, err := gameassets.Load()
	if err != nil {
		return err
	}
	m.Assets = cat
	if m.Sim.Trees != nil {
		m.Sim.Trees.Models = cat.Trees
	}
	if m.Sim.Buildings != nil {
		m.Sim.Buildings.SetAssets(cat)
	}
	if m.Sim.Vehicles != nil {
		m.Sim.Vehicles.SetAssets(cat)
	}
	if cat.Road.ID != 0 && m.Sim.Roads != nil {
		m.Sim.Roads.SetRoadTexture(cat.Road)
	}
	if cat.Grass.ID != 0 {
		m.terrainTex = cat.Grass
	}
	if !m.Sim.Trees.HasModels() {
		return fmt.Errorf("no tree models loaded")
	}
	return nil
}

func (m *Manager) LoadBuildingAssets() {}

func (m *Manager) PrepareUpload() {
	if m.terrainTex.ID == 0 {
		grassImg := rl.GenImageCellular(512, 512, 32)
		rl.ImageColorTint(grassImg, rl.NewColor(50, 130, 30, 255))
		rl.ImageColorBrightness(grassImg, -15)
		m.terrainTex = rl.LoadTextureFromImage(grassImg)
		rl.UnloadImage(grassImg)
	}

	for lod := 0; lod < terrain.NumLODs; lod++ {
		m.Models[lod] = make([]rl.Model, len(m.Chunks))
	}
	m.uploadIdx = 0
}

func (m *Manager) UploadNextBatch(count int) bool {
	end := m.uploadIdx + count
	if end > len(m.Chunks) {
		end = len(m.Chunks)
	}
	for i := m.uploadIdx; i < end; i++ {
		c := m.Chunks[i]
		for lod := 0; lod < terrain.NumLODs; lod++ {
			model := chunkToModelLOD(c, lod)
			mats := model.GetMaterials()
			if len(mats) > 0 {
				rl.SetMaterialTexture(&mats[0], rl.MapAlbedo, m.terrainTex)
			}
			m.Models[lod][i] = model
		}
	}
	m.uploadIdx = end

	return m.uploadIdx >= len(m.Chunks)
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
	for lod := 0; lod < terrain.NumLODs; lod++ {
		rl.UnloadModel(m.Models[lod][chunkIdx])
	}
	terrain.BuildAllLODMeshes(c)
	for lod := 0; lod < terrain.NumLODs; lod++ {
		model := chunkToModelLOD(c, lod)
		mats := model.GetMaterials()
		if len(mats) > 0 {
			rl.SetMaterialTexture(&mats[0], rl.MapAlbedo, m.terrainTex)
		}
		m.Models[lod][chunkIdx] = model
	}
}

func (m *Manager) Draw(camX, camZ float32) {
	sim := m.Sim
	maxChunkDist := float32(3000)
	for i, c := range m.Chunks {
		cx, cz := chunkCenter(c)
		dx := cx - camX
		dz := cz - camZ
		distSq := dx*dx + dz*dz
		if distSq > maxChunkDist*maxChunkDist {
			continue
		}
		dist := float32(math.Sqrt(float64(distSq)))
		lod := terrain.SelectLOD(dist)
		model := m.Models[lod][i]
		rl.DrawModel(model, rl.NewVector3(0, 0, 0), 1, rl.White)
	}
	if sim.Trees != nil {
		sim.Trees.Draw(sim.Heightmap, camX, camZ)
	}
	if sim.Water != nil {
		sim.Water.Draw()
	}
	if sim.Roads != nil {
		sim.Roads.Draw(sim.Heightmap)
	}
	if sim.Vehicles != nil {
		sim.Vehicles.Draw(sim.Heightmap)
	}
	if sim.Parking != nil {
		sim.Parking.Draw(sim.Heightmap)
	}
	if sim.Transport != nil {
		sim.Transport.Draw(sim.Heightmap)
	}
	if sim.Zones != nil {
		sim.Zones.Draw(sim.Heightmap)
	}
	if sim.Buildings != nil {
		sim.Buildings.Draw(sim.Heightmap)
	}
}

func (m *Manager) Unload() {
	for lod := 0; lod < terrain.NumLODs; lod++ {
		for _, model := range m.Models[lod] {
			rl.UnloadModel(model)
		}
		m.Models[lod] = nil
	}
	if m.Sim.Trees != nil {
		m.Sim.Trees.Models = [5]rl.Model{}
	}
	if m.Assets != nil {
		m.Assets.Unload()
		m.Assets = nil
		m.terrainTex = rl.Texture2D{}
	} else if m.terrainTex.ID != 0 {
		rl.UnloadTexture(m.terrainTex)
	}
	if m.Sim.Vehicles != nil {
		m.Sim.Vehicles.Unload()
	}
	if m.Sim.Transport != nil {
		m.Sim.Transport.Unload()
	}
	if m.Sim.Roads != nil {
		m.Sim.Roads.Unload()
		m.Sim.Roads.ClearRoadTexture()
	}
}

func chunkCenter(c *terrain.Chunk) (float32, float32) {
	scale := terrain.WorldSize / float32(terrain.HeightmapSize-1)
	cx := float32(c.IndexX*(terrain.ChunkSize-1))*scale - terrain.WorldSize/2 + float32(terrain.ChunkSize-1)*scale/2
	cz := float32(c.IndexZ*(terrain.ChunkSize-1))*scale - terrain.WorldSize/2 + float32(terrain.ChunkSize-1)*scale/2
	return cx, cz
}

func chunkToModelLOD(c *terrain.Chunk, lod int) rl.Model {
	md := c.LODMeshes[lod]
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
