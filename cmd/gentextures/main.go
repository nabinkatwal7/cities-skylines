package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
)

func main() {
	os.MkdirAll("assets/textures", 0o755)
	writeGrass("assets/textures/grass.png")
	writeWater("assets/textures/water.png")
	writeRoad("assets/textures/road.png")
}

func writeRoad(path string) {
	const size = 256
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			n := hashNoise(x, y, 41)*0.06 + hashNoise(x*4, y*4, 7)*0.03
			v := 0.20 + n
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(clamp01(v*0.92) * 255),
				G: uint8(clamp01(v) * 255),
				B: uint8(clamp01(v*0.88) * 255),
				A: 255,
			})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func writeGrass(path string) {
	const size = 1024
	rng := rand.New(rand.NewSource(42))
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			ux := float64(x) / float64(size)
			uy := float64(y) / float64(size)
			n := hashNoise(x, y, 91) * 0.55
			n += hashNoise(x*3, y*3, 17) * 0.25
			n += hashNoise(x*7, y*7, 53) * 0.12
			n += hashNoise(x*13, y*13, 29) * 0.08
			patch := hashNoise(x/24, y/24, 7)
			dry := hashNoise(x/64+int(rng.Intn(3)), y/64, 3)
			r := 0.14 + n*0.10 + patch*0.04 + dry*0.03
			g := 0.34 + n*0.22 + patch*0.10 + dry*0.02
			b := 0.10 + n*0.06 + patch*0.02
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(clamp01(r) * 255),
				G: uint8(clamp01(g) * 255),
				B: uint8(clamp01(b) * 255),
				A: 255,
			})
			_ = ux
			_ = uy
		}
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func writeWater(path string) {
	size := 2048
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			ux := float64(x) / float64(size)
			uy := float64(y) / float64(size)
			angle := ux * 2 * math.Pi
			theta := uy * math.Pi
			sx := math.Sin(theta) * math.Cos(angle)
			sy := math.Sin(theta) * math.Sin(angle)
			sz := math.Cos(theta)

			n := 0.0
			amp := 1.0
			freq := 1.0
			maxAmp := 0.0
			for oct := 0; oct < 6; oct++ {
				nx := sx * freq
				ny := sy * freq
				nz := sz * freq
				wave := math.Sin(nx*3.7+ny*5.2+nz*2.1)*math.Cos(ny*4.3+nz*1.8+nx*6.5)*0.5 + 0.5
				n += wave * amp
				maxAmp += amp
				amp *= 0.45
				freq *= 2.1
			}
			n /= maxAmp
			swell := math.Sin(sx*1.2+sz*0.8)*0.5 + 0.5
			swell2 := math.Sin(sx*0.7+sy*0.5+sz*1.1)*0.5 + 0.5
			h := n*0.6 + swell*0.25 + swell2*0.15

			r := 0.05 + h*0.30
			g := 0.12 + h*0.40
			b := 0.28 + h*0.25
			foam := math.Max(0, (h-0.55)*4.0)
			if foam > 0 {
				r += foam * 0.6
				g += foam * 0.6
				b += foam * 0.6
			}
			streak := math.Abs(math.Sin(sx*15.0+sz*12.0)) * 0.08
			r += streak
			g += streak
			b += streak

			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(clamp01(r) * 255),
				G: uint8(clamp01(g) * 255),
				B: uint8(clamp01(b) * 255),
				A: 255,
			})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func hashNoise(x, y, seed int) float64 {
	h := uint32(x*374761393 + y*668265263 + seed*1274126177)
	h = (h ^ (h >> 13)) * 1274126177
	return float64(h&0xffff) / 65535.0
}

func clamp01(v float64) float64 {
	return math.Min(1, math.Max(0, v))
}
