package terrain

import "math"

type Generator struct {
	seed      int64
	noise     *Noise
	heightmap *Heightmap
}

func NewGenerator(seed int64) *Generator {
	return &Generator{
		seed:      seed,
		noise:     NewNoise(seed),
		heightmap: NewHeightmap(),
	}
}

func (g *Generator) Generate() *Heightmap {
	g.applyBaseNoise()
	g.applyMountainPass()
	g.carveRiver()
	g.applyErosion()
	g.applyResources()
	return g.heightmap
}

func (g *Generator) Heightmap() *Heightmap {
	return g.heightmap
}

func (g *Generator) applyBaseNoise() {
	scale := 0.015
	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			nx := float64(x) * scale
			nz := float64(z) * scale

			continent := g.noise.Fbm(nx, nz, 4, 2.0, 0.5)
			mountain := g.noise.Ridge(nx*0.5+100, nz*0.5+100, 4, 2.2, 0.6)
			hill := g.noise.Fbm(nx*2+200, nz*2+200, 3, 2.0, 0.5)
			detail := g.noise.Fbm(nx*4+300, nz*4+300, 2, 2.0, 0.5)

			height := continent*0.6 + mountain*0.25 + hill*0.1 + detail*0.05
			height = math.Max(0, math.Min(1, height))

			edge := g.edgeFalloff(x, z)
			height *= edge

			g.heightmap.Set(x, z, float32(height))
		}
	}
}

func (g *Generator) applyMountainPass() {
	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			h := g.heightmap.Get(x, z)
			if h > 0.6 {
				ridge := (h - 0.6) / 0.4
				h = h*(1-ridge*0.3) + 0.9*ridge*0.3
				g.heightmap.Set(x, z, float32(math.Min(1, float64(h))))
			}
		}
	}
}

func (g *Generator) carveRiver() {
	type point struct{ x, z int }
	center := HeightmapSize / 2

	startX := center - 40 + int(g.noise.Noise2D(0.5, 0.7)*60)
	startZ := center - 30 + int(g.noise.Noise2D(0.3, 0.9)*40)

	path := []point{{startX, startZ}}
	current := point{startX, startZ}
	steps := 0

	for current.z < HeightmapSize-10 && steps < 300 {
		best := current
		bestScore := float32(math.MaxFloat32)

		nx := g.noise.Noise2D(float64(steps)*0.1+5, 0.5)
		nz := g.noise.Noise2D(0.5, float64(steps)*0.1+5)

		dz := 1
		offsets := []int{-1, 0, 1}
		for _, ox := range offsets {
			cx := current.x + ox
			cz := current.z + dz
			if cx < 1 || cx >= HeightmapSize-1 || cz >= HeightmapSize-1 {
				continue
			}
			h := g.heightmap.Get(cx, cz)
			score := h + float32(math.Abs(float64(cx-center))*0.05) + float32(math.Abs(nx+float64(ox)*0.5)*0.1) + float32(math.Abs(nz)*0.1)
			if score < bestScore {
				bestScore = score
				best = point{cx, cz}
			}
		}
		current = best
		path = append(path, current)
		steps++
	}

	for _, p := range path {
		width := 2 + int(g.noise.Noise2D(float64(p.x)*0.05, float64(p.z)*0.05)*2)
		for dx := -width; dx <= width; dx++ {
			for dz := -1; dz <= 1; dz++ {
				cx := p.x + dx
				cz := p.z + dz
				if cx >= 0 && cx < HeightmapSize && cz >= 0 && cz < HeightmapSize {
					dist := float32(math.Abs(float64(dx))) / float32(width+1)
					carve := float32(0.15) * (1 - dist)
					h := g.heightmap.Get(cx, cz)
					g.heightmap.Set(cx, cz, h-carve)
				}
			}
		}
	}
}

func (g *Generator) applyErosion() {
	const iterations = 3
	for i := 0; i < iterations; i++ {
		for z := 1; z < HeightmapSize-1; z++ {
			for x := 1; x < HeightmapSize-1; x++ {
				h := g.heightmap.Get(x, z)
				avg := (g.heightmap.Get(x-1, z) + g.heightmap.Get(x+1, z) +
					g.heightmap.Get(x, z-1) + g.heightmap.Get(x, z+1)) / 4
				if h > avg+0.02 {
					g.heightmap.Set(x, z, h-0.005)
				}
			}
		}
	}
}

func (g *Generator) applyResources() {
	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			h := g.heightmap.Get(x, z)
			g.heightmap.Data[z][x] = h
		}
	}
}

func (g *Generator) edgeFalloff(x, z int) float64 {
	dist := math.Sqrt(
		math.Pow(float64(x-HeightmapSize/2), 2) +
			math.Pow(float64(z-HeightmapSize/2), 2),
	)
	maxDist := float64(HeightmapSize) * 0.42
	if dist > maxDist {
		return math.Max(0, 1-(dist-maxDist)/(float64(HeightmapSize)*0.08))
	}
	return 1.0
}
