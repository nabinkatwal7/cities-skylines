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
	if uiFont.Texture.ID == 0 {
		rl.DrawText(text, x, y, size, col)
		return
	}
	rl.DrawTextEx(uiFont, text, rl.NewVector2(float32(x), float32(y)), float32(size), 1, col)
}

func MeasureUIText(text string, size int32) int32 {
	if uiFont.Texture.ID == 0 {
		return rl.MeasureText(text, size)
	}
	return int32(rl.MeasureTextEx(uiFont, text, float32(size), 1).X)
}
