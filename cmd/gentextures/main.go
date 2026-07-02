package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	os.MkdirAll("assets/textures", 0o755)
	writeGrass("assets/textures/grass.png")
	writeWater("assets/textures/water.png")
}

func writeGrass(path string) {
	size := 1024
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			ux := float64(x) / float64(size)
			uy := float64(y) / float64(size)
			n := seamlessNoise(ux, uy, 4)
			detail := seamlessNoise(ux*8, uy*8, 2)
			r := 0.18 + n*0.08 + detail*0.03
			g := 0.42 + n*0.18 + detail*0.06
			b := 0.12 + n*0.05
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

func seamlessNoise(ux, uy float64, octaves int) float64 {
	angle := ux * 2 * math.Pi
	theta := uy * math.Pi
	sx := math.Sin(theta) * math.Cos(angle)
	sy := math.Sin(theta) * math.Sin(angle)
	sz := math.Cos(theta)
	n, amp, freq, maxAmp := 0.0, 1.0, 1.0, 0.0
	for i := 0; i < octaves; i++ {
		wave := math.Sin(sx*freq*3.1+sy*freq*2.7+sz*freq*1.9)*0.5 + 0.5
		n += wave * amp
		maxAmp += amp
		amp *= 0.5
		freq *= 2.0
	}
	return n / maxAmp
}

func clamp01(v float64) float64 {
	return math.Min(1, math.Max(0, v))
}
