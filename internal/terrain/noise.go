package terrain

import (
	"math"
	"math/rand"
)

type Noise struct {
	perm [512]int
}

func NewNoise(seed int64) *Noise {
	n := &Noise{}
	rng := rand.New(rand.NewSource(seed))
	p := make([]int, 256)
	for i := 0; i < 256; i++ {
		p[i] = i
	}
	for i := 255; i > 0; i-- {
		j := rng.Intn(i + 1)
		p[i], p[j] = p[j], p[i]
	}
	for i := 0; i < 256; i++ {
		n.perm[i] = p[i]
		n.perm[i+256] = p[i]
	}
	return n
}

func (n *Noise) Noise2D(x, y float64) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)

	u := fade(xf)
	v := fade(yf)

	aa := n.perm[n.perm[xi]+yi]
	ab := n.perm[n.perm[xi]+yi+1]
	ba := n.perm[n.perm[xi+1]+yi]
	bb := n.perm[n.perm[xi+1]+yi+1]

	return lerpDouble(
		lerpDouble(grad(aa, xf, yf), grad(ba, xf-1, yf), u),
		lerpDouble(grad(ab, xf, yf-1), grad(bb, xf-1, yf-1), u),
		v,
	)
}

func (n *Noise) Fbm(x, y float64, octaves int, lacunarity, gain float64) float64 {
	var value float64
	var amplitude float64 = 1
	var frequency float64 = 1
	var maxVal float64

	for i := 0; i < octaves; i++ {
		value += amplitude * n.Noise2D(x*frequency, y*frequency)
		maxVal += amplitude
		amplitude *= gain
		frequency *= lacunarity
	}

	return value / maxVal
}

func (n *Noise) Ridge(x, y float64, octaves int, lacunarity, gain float64) float64 {
	var value float64
	var amplitude float64 = 1
	var frequency float64 = 1

	for i := 0; i < octaves; i++ {
		noiseVal := 1.0 - math.Abs(n.Noise2D(x*frequency, y*frequency))
		noiseVal = noiseVal * noiseVal
		value += amplitude * noiseVal
		amplitude *= gain
		frequency *= lacunarity
	}

	return value
}

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerpDouble(a, b, t float64) float64 {
	return a + t*(b-a)
}

func grad(hash int, x, y float64) float64 {
	h := hash & 3
	switch h {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	default:
		return -x - y
	}
}
