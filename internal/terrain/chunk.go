package terrain

import "math"

const (
	ChunkSize = 33
	ChunksPerSide = (HeightmapSize - 1) / (ChunkSize - 1)
	NumChunks = ChunksPerSide * ChunksPerSide
)

type Chunk struct {
	IndexX, IndexZ int
	Dirty          bool
	Heights        [ChunkSize][ChunkSize]float32
	MeshData       *MeshData
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
	vertsPerSide := ChunkSize
	numVerts := vertsPerSide * vertsPerSide
	numQuads := (vertsPerSide - 1) * (vertsPerSide - 1)

	md := &MeshData{
		Vertices:   make([]float32, numVerts*3),
		Normals:    make([]float32, numVerts*3),
		TexCoords:  make([]float32, numVerts*2),
		Indices:    make([]uint16, numQuads*6),
		VertexCount: int32(numVerts),
		IndexCount:  int32(numQuads * 6),
	}

	scale := WorldSize / float32(HeightmapSize-1)

	idx := 0
	for lz := 0; lz < vertsPerSide; lz++ {
		for lx := 0; lx < vertsPerSide; lx++ {
			worldX := float32(c.IndexX*(ChunkSize-1)+lx) * scale
			worldZ := float32(c.IndexZ*(ChunkSize-1)+lz) * scale
			h := c.Heights[lz][lx] * MaxHeight

			md.Vertices[idx*3] = worldX - WorldSize/2
			md.Vertices[idx*3+1] = h
			md.Vertices[idx*3+2] = worldZ - WorldSize/2

			gx := (c.IndexX*(ChunkSize-1) + lx)
			gz := (c.IndexZ*(ChunkSize-1) + lz)
			md.TexCoords[idx*2] = float32(gx) / float32(HeightmapSize-1) * 64
			md.TexCoords[idx*2+1] = float32(gz) / float32(HeightmapSize-1) * 64

			hl := c.Heights[lz][max(0, lx-1)] * MaxHeight
			hr := c.Heights[lz][min(vertsPerSide-1, lx+1)] * MaxHeight
			hd := c.Heights[max(0, lz-1)][lx] * MaxHeight
			hu := c.Heights[min(vertsPerSide-1, lz+1)][lx] * MaxHeight

			nx := hl - hr
			nz := hd - hu
			ny := 2.0 * scale
			invLen := 1.0 / float32(math.Sqrt(float64(nx*nx+ny*ny+nz*nz)))
			md.Normals[idx*3] = nx * invLen
			md.Normals[idx*3+1] = ny * invLen
			md.Normals[idx*3+2] = nz * invLen

			idx++
		}
	}

	quadIdx := 0
	for lz := 0; lz < vertsPerSide-1; lz++ {
		for lx := 0; lx < vertsPerSide-1; lx++ {
			a := lz*vertsPerSide + lx
			b := lz*vertsPerSide + lx + 1
			c := (lz+1)*vertsPerSide + lx
			d := (lz+1)*vertsPerSide + lx + 1

			md.Indices[quadIdx*6] = uint16(a)
			md.Indices[quadIdx*6+1] = uint16(c)
			md.Indices[quadIdx*6+2] = uint16(b)
			md.Indices[quadIdx*6+3] = uint16(b)
			md.Indices[quadIdx*6+4] = uint16(c)
			md.Indices[quadIdx*6+5] = uint16(d)
			quadIdx++
		}
	}

	c.MeshData = md
	c.Dirty = false
	return md
}
