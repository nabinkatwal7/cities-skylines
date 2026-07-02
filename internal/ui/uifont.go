package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const fontLoadSize int32 = 32

var uiFont rl.Font

// Standard readable sizes (loaded font is rasterized at fontLoadSize).
const (
	FontXs = 14
	FontSm = 16
	FontMd = 18
	FontLg = 22
	FontXl = 28
)

func LoadUIFont() {
	// ponytail: never use monospace for UI — JetBrains Mono was unreadable at UI sizes.
	candidates := []string{
		"C:/Windows/Fonts/segoeui.ttf",
		"C:/Windows/Fonts/SegoeUI.ttf",
		"C:/Windows/Fonts/arial.ttf",
		"C:/Windows/Fonts/calibri.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
	}
	for _, path := range candidates {
		f := rl.LoadFontEx(path, fontLoadSize, nil)
		if f.Texture.ID != 0 {
			uiFont = f
			return
		}
	}
	uiFont = rl.GetFontDefault()
}

func UnloadUIFont() {
	def := rl.GetFontDefault()
	if uiFont.Texture.ID != 0 && uiFont.Texture.ID != def.Texture.ID {
		rl.UnloadFont(uiFont)
	}
	uiFont = rl.Font{}
}

func DrawUIText(text string, x, y, size int32, col rl.Color) {
	drawUITextScaled(text, x, y, size, col, 1)
}

func DrawUITextScaled(text string, x, y, size int32, col rl.Color, scale float32) {
	drawUITextScaled(text, x, y, size, col, scale)
}

func drawUITextScaled(text string, x, y, size int32, col rl.Color, scale float32) {
	if scale <= 0 {
		scale = 1
	}
	fs := float32(size) * scale
	if uiFont.Texture.ID == 0 {
		rl.DrawText(text, x, y, int32(fs), col)
		return
	}
	rl.DrawTextEx(uiFont, text, rl.NewVector2(float32(x), float32(y)), fs, 1, col)
}

func MeasureUIText(text string, size int32) int32 {
	if uiFont.Texture.ID == 0 {
		return rl.MeasureText(text, size)
	}
	return int32(rl.MeasureTextEx(uiFont, text, float32(size), 1).X)
}
