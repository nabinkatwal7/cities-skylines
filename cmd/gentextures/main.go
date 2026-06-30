package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	size := 2048
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Multi-octave wave noise for water surface
	// 4 octaves for detail
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			ux := float64(x) / float64(size)
			uy := float64(y) / float64(size)

			// Seamless noise using 4D spherical mapping
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
				// Use sin-based wave for seamless noise
				wave := math.Sin(nx*3.7+ny*5.2+nz*2.1)*math.Cos(ny*4.3+nz*1.8+nx*6.5)*0.5 + 0.5
				n += wave * amp
				maxAmp += amp
				amp *= 0.45
				freq *= 2.1
			}
			n /= maxAmp

			// Additional larger swell waves
			swell := math.Sin(sx*1.2+sz*0.8)*0.5 + 0.5
			swell2 := math.Sin(sx*0.7+sy*0.5+sz*1.1)*0.5 + 0.5

			// Combine
			h := n*0.6 + swell*0.25 + swell2*0.15

			// Deep water: dark blue (low areas)
			// Shallow/wave: lighter teal (high areas)
			// Foam caps: white (very high areas)

			// Base deep water color
			r := 0.05
			g := 0.12
			b := 0.28

			// Transition to teal/cyan
			r += h * 0.30
			g += h * 0.40
			b += h * 0.25

			// Foam crests at peaks
			foam := math.Max(0, (h-0.55)*4.0)
			if foam > 0 {
				r += foam * 0.6
				g += foam * 0.6
				b += foam * 0.6
			}

			// Subtle reflection streaks
			streak := math.Abs(math.Sin(sx*15.0+sz*12.0)) * 0.08
			r += streak
			g += streak
			b += streak

			// Clamp
			r = math.Min(1, math.Max(0, r))
			g = math.Min(1, math.Max(0, g))
			b = math.Min(1, math.Max(0, b))

			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: 255,
			})
		}
	}

	f, err := os.Create("assets/water.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}
