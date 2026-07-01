package terrain

import "math"

const (
	ChunkSize     = 33
	ChunksPerSide = (HeightmapSize - 1) / (ChunkSize - 1)
	NumChunks     = ChunksPerSide * ChunksPerSide
	NumLODs       = 4
)

var lodResolutions = [NumLODs]int{33, 17, 9, 5}
var lodStrides = [NumLODs]int{1, 2, 4, 8}

var lodDistances = [NumLODs]float32{500, 1000, 2000, 3000}

type Chunk struct {
	IndexX, IndexZ int
	Dirty          bool
	Heights        [ChunkSize][ChunkSize]float32
	MeshData       *MeshData
	LODMeshes      [NumLODs]*MeshData
}

type MeshData struct {
	Vertices    []float32
	Normals     []float32
	TexCoords   []float32
	Indices     []uint16
	VertexCount int32
	IndexCount  int32
}

func BuildChunks(h *Heightmap) []*Chunk {
	chunks := make([]*Chunk, NumChunks)
	chunkIdx := 0
	for cz := 0; cz < ChunksPerSide; cz++ {
		for cx := 0; cx < ChunksPerSide; cx++ {
			c := &Chunk{
				IndexX: cx,
				IndexZ: cz,
				Dirty:  true,
			}
			baseX := cx * (ChunkSize - 1)
			baseZ := cz * (ChunkSize - 1)
			for lz := 0; lz < ChunkSize; lz++ {
				for lx := 0; lx < ChunkSize; lx++ {
					hz := baseZ + lz
					hx := baseX + lx
					if hz >= HeightmapSize {
						hz = HeightmapSize - 1
					}
					if hx >= HeightmapSize {
						hx = HeightmapSize - 1
					}
					c.Heights[lz][lx] = h.Get(hx, hz)
				}
			}
			chunks[chunkIdx] = c
			chunkIdx++
		}
	}
	return chunks
}

func BuildChunkMesh(c *Chunk) *MeshData {
	return buildChunkMeshLOD(c, 0)
}

func buildChunkMeshLOD(c *Chunk, lod int) *MeshData {
	res := lodResolutions[lod]
	stride := lodStrides[lod]
	scale := WorldSize / float32(HeightmapSize-1)

	numVerts := res * res
	numQuads := (res - 1) * (res - 1)

	md := &MeshData{
		Vertices:    make([]float32, numVerts*3),
		Normals:     make([]float32, numVerts*3),
		TexCoords:   make([]float32, numVerts*2),
		Indices:     make([]uint16, numQuads*6),
		VertexCount: int32(numVerts),
		IndexCount:  int32(numQuads * 6),
	}

	idx := 0
	for lz := 0; lz < res; lz++ {
		for lx := 0; lx < res; lx++ {
			srcZ := lz * stride
			srcX := lx * stride
			if srcZ >= ChunkSize {
				srcZ = ChunkSize - 1
			}
			if srcX >= ChunkSize {
				srcX = ChunkSize - 1
			}

			worldX := float32(c.IndexX*(ChunkSize-1)+srcX) * scale
			worldZ := float32(c.IndexZ*(ChunkSize-1)+srcZ) * scale
			h := c.Heights[srcZ][srcX] * MaxHeight

			md.Vertices[idx*3] = worldX - WorldSize/2
			md.Vertices[idx*3+1] = h
			md.Vertices[idx*3+2] = worldZ - WorldSize/2

			gx := (c.IndexX*(ChunkSize-1) + srcX)
			gz := (c.IndexZ*(ChunkSize-1) + srcZ)
			md.TexCoords[idx*2] = float32(gx) / float32(HeightmapSize-1) * 64
			md.TexCoords[idx*2+1] = float32(gz) / float32(HeightmapSize-1) * 64

			nl := max(0, srcX-stride)
			nr := min(ChunkSize-1, srcX+stride)
			nd := max(0, srcZ-stride)
			nu := min(ChunkSize-1, srcZ+stride)

			hl := c.Heights[srcZ][nl] * MaxHeight
			hr := c.Heights[srcZ][nr] * MaxHeight
			hd := c.Heights[nd][srcX] * MaxHeight
			hu := c.Heights[nu][srcX] * MaxHeight

			nx := hl - hr
			nz := hd - hu
			ny := 2.0 * scale * float32(stride)
			invLen := 1.0 / float32(math.Sqrt(float64(nx*nx+ny*ny+nz*nz)))
			md.Normals[idx*3] = nx * invLen
			md.Normals[idx*3+1] = ny * invLen
			md.Normals[idx*3+2] = nz * invLen

			idx++
		}
	}

	quadIdx := 0
	for lz := 0; lz < res-1; lz++ {
		for lx := 0; lx < res-1; lx++ {
			a := lz*res + lx
			b := lz*res + lx + 1
			c := (lz+1)*res + lx
			d := (lz+1)*res + lx + 1

			md.Indices[quadIdx*6] = uint16(a)
			md.Indices[quadIdx*6+1] = uint16(c)
			md.Indices[quadIdx*6+2] = uint16(b)
			md.Indices[quadIdx*6+3] = uint16(b)
			md.Indices[quadIdx*6+4] = uint16(c)
			md.Indices[quadIdx*6+5] = uint16(d)
			quadIdx++
		}
	}

	return md
}

func BuildAllLODMeshes(c *Chunk) {
	c.MeshData = buildChunkMeshLOD(c, 0)
	for lod := 0; lod < NumLODs; lod++ {
		c.LODMeshes[lod] = buildChunkMeshLOD(c, lod)
	}
	c.Dirty = false
}

func SelectLOD(dist float32) int {
	for lod := 0; lod < NumLODs; lod++ {
		if dist < lodDistances[lod] {
			return lod
		}
	}
	return NumLODs - 1
}
