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
	if FlatTestMap {
		return g.generateFlatTest()
	}
	g.applyBaseNoise()
	g.applyMountainPass()
	g.applyErosion()
	g.carveLakeBasin()
	g.carveRivers()
	g.applyResources()
	return g.heightmap
}

func (g *Generator) Heightmap() *Heightmap {
	return g.heightmap
}

// generateFlatTest builds a flat playfield with low river channels for zoning/road tests.
func (g *Generator) generateFlatTest() *Heightmap {
	land := TestLandNorm()
	river := TestRiverNorm()
	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			g.heightmap.Set(x, z, land)
		}
	}
	carveChannel := func(x0, z0, x1, z1, halfW int) {
		dx := x1 - x0
		dz := z1 - z0
		steps := int(math.Hypot(float64(dx), float64(dz)))
		if steps < 1 {
			steps = 1
		}
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			cx := int(float64(x0) + float64(dx)*t)
			cz := int(float64(z0) + float64(dz)*t)
			for dz2 := -halfW; dz2 <= halfW; dz2++ {
				for dx2 := -halfW; dx2 <= halfW; dx2++ {
					px := cx + dx2
					pz := cz + dz2
					if px >= 0 && px < HeightmapSize && pz >= 0 && pz < HeightmapSize {
						g.heightmap.Set(px, pz, river)
					}
				}
			}
		}
	}
	mid := HeightmapSize / 2
	carveChannel(mid, 80, mid, HeightmapSize-80, 12)
	carveChannel(120, mid, HeightmapSize-120, mid, 10)
	carveChannel(200, 200, HeightmapSize-200, HeightmapSize-250, 8)
	return g.heightmap
}

func (g *Generator) applyBaseNoise() {
	scale := 0.018
	edgeMul := g.computeEdgeFalloff()
	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			nx := float64(x) * scale
			nz := float64(z) * scale

			continent := g.noise.Fbm(nx, nz, 4, 2.0, 0.5)
			mountain := g.noise.Ridge(nx*0.5+100, nz*0.5+100, 4, 2.2, 0.6)
			hill := g.noise.Fbm(nx*2+200, nz*2+200, 3, 2.0, 0.5)
			detail := g.noise.Fbm(nx*4+300, nz*4+300, 3, 2.0, 0.5)

			height := continent*0.55 + mountain*0.2 + hill*0.15 + detail*0.1
			height = math.Max(0, math.Min(1, height))

			height *= edgeMul[z][x]

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

func (g *Generator) carveRivers() {
	type point struct{ x, z int }
	center := HeightmapSize / 2

	dirs := []struct {
		startOffX, startOffZ int
		dx, dz               int
		seedX, seedZ         float64
	}{
		{-40, -30, 0, 1, 0.5, 0.7},
		{40, -20, 0, 1, 0.9, 0.3},
		{-30, 40, 1, 0, 0.2, 0.8},
		{20, 30, -1, 0, 0.7, 0.2},
	}

	for _, d := range dirs {
		startX := center + d.startOffX + int(g.noise.Noise2D(d.seedX, d.seedZ)*60)
		startZ := center + d.startOffZ + int(g.noise.Noise2D(d.seedZ, d.seedX)*40)
		startX = clamp(startX, 2, HeightmapSize-2)
		startZ = clamp(startZ, 2, HeightmapSize-2)

		path := []point{{startX, startZ}}
		current := point{startX, startZ}
		steps := 0
		maxSteps := 400

		for steps < maxSteps {
			atEdge := false
			if d.dx != 0 {
				if (d.dx > 0 && current.x >= HeightmapSize-10) || (d.dx < 0 && current.x <= 10) {
					atEdge = true
				}
			}
			if d.dz != 0 {
				if (d.dz > 0 && current.z >= HeightmapSize-10) || (d.dz < 0 && current.z <= 10) {
					atEdge = true
				}
			}
			if atEdge {
				break
			}

			best := current
			bestScore := float32(math.MaxFloat32)
			nx := g.noise.Noise2D(float64(steps)*0.1+5, d.seedX)
			nz := g.noise.Noise2D(d.seedZ, float64(steps)*0.1+5)

			offsets := []int{-1, 0, 1}
			for _, o := range offsets {
				cx := current.x
				cz := current.z
				if d.dx != 0 {
					cx += d.dx
					cz += o
				} else {
					cx += o
					cz += d.dz
				}
				if cx < 1 || cx >= HeightmapSize-1 || cz < 1 || cz >= HeightmapSize-1 {
					continue
				}
				h := g.heightmap.Get(cx, cz)
				score := h + float32(math.Abs(float64(cx-center))*0.02) + float32(math.Abs(nx+float64(o)*0.3)*0.1) + float32(math.Abs(nz)*0.1)
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
			width := 3 + int(g.noise.Noise2D(float64(p.x)*0.05, float64(p.z)*0.05)*2)
			for dx := -width; dx <= width; dx++ {
				for dz := -2; dz <= 2; dz++ {
					cx := p.x + dx
					cz := p.z + dz
					if cx >= 0 && cx < HeightmapSize && cz >= 0 && cz < HeightmapSize {
						dist := float32(math.Sqrt(float64(dx*dx+dz*dz))) / float32(width+1)
						carve := float32(0.3) * (1 - dist*0.9)
						if carve > 0 {
							h := g.heightmap.Get(cx, cz)
							g.heightmap.Set(cx, cz, h-carve)
						}
					}
				}
			}
		}
	}
}

func (g *Generator) carveLakeBasin() {
	center := HeightmapSize / 2
	radius := 50.0
	maxDepth := 0.3
	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			dist := math.Sqrt(float64((x-center)*(x-center) + (z-center)*(z-center)))
			if dist < radius {
				t := dist / radius
				carve := float32(maxDepth * (1 - t*t*0.8))
				h := g.heightmap.Get(x, z)
				g.heightmap.Set(x, z, float32(math.Max(0, float64(h)-float64(carve))))
			}
		}
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
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

func (g *Generator) computeEdgeFalloff() [HeightmapSize][HeightmapSize]float64 {
	var falloff [HeightmapSize][HeightmapSize]float64
	half := float64(HeightmapSize / 2)
	maxDist := half * 0.95
	startFalloff := half * 0.6
	falloffRange := maxDist - startFalloff
	for z := 0; z < HeightmapSize; z++ {
		for x := 0; x < HeightmapSize; x++ {
			dx := float64(x) - half
			dz := float64(z) - half
			dist := math.Sqrt(dx*dx + dz*dz)
			if dist >= maxDist {
				falloff[z][x] = 0
			} else if dist <= startFalloff {
				falloff[z][x] = 1.0
			} else {
				t := (dist - startFalloff) / falloffRange
				falloff[z][x] = math.Cos(t * math.Pi / 2)
			}
		}
	}
	return falloff
}
