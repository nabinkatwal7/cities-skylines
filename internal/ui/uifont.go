package ui

import rl "github.com/gen2brain/raylib-go/raylib"

var uiFont rl.Font

func LoadUIFont() {
	candidates := []string{
		"assets/fonts/JetBrainsMono-Regular.ttf",
		"C:/Windows/Fonts/segoeui.ttf",
		"C:/Windows/Fonts/consola.ttf",
	}
	for _, path := range candidates {
		f := rl.LoadFont(path)
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
	fs := int32(float32(size) * scale)
	if uiFont.Texture.ID == 0 {
		rl.DrawText(text, x, y, fs, col)
		return
	}
	rl.DrawTextEx(uiFont, text, rl.NewVector2(float32(x), float32(y)), float32(fs), 1, col)
}

func MeasureUIText(text string, size int32) int32 {
	if uiFont.Texture.ID == 0 {
		return rl.MeasureText(text, size)
	}
	return int32(rl.MeasureTextEx(uiFont, text, float32(size), 1).X)
}
