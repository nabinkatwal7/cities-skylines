package ui

// ColorBlindMode adjusts presentation palettes (24.23).
type ColorBlindMode int

const (
	ColorBlindNone ColorBlindMode = iota
	ColorBlindProtanopia
	ColorBlindDeuteranopia
	ColorBlindTritanopia
)

// Accessibility settings are presentation-only (24.23).
type Accessibility struct {
	UIScale         float32
	ColorBlindMode  ColorBlindMode
	HighContrast    bool
	Subtitles       bool
	KeyboardNav     bool
	ReducedMotion   bool
	ScreenReaderMD  bool
}

func DefaultAccessibility() Accessibility {
	return Accessibility{
		UIScale:        1,
		ColorBlindMode: ColorBlindNone,
		KeyboardNav:    true,
		ScreenReaderMD: true,
	}
}

// PlayerSettings groups configurable player preferences (24.21–24.24).
type PlayerSettings struct {
	Bindings *KeyBindings
	A11y     Accessibility
	Locale   string
}

func NewPlayerSettings() *PlayerSettings {
	return &PlayerSettings{
		Bindings: DefaultBindings(),
		A11y:     DefaultAccessibility(),
		Locale:   "en",
	}
}

func (s *PlayerSettings) UIScale() float32 {
	if s == nil || s.A11y.UIScale < 0.75 {
		return 1
	}
	if s.A11y.UIScale > 1.5 {
		return 1.5
	}
	return s.A11y.UIScale
}

func ScaleY(base int32, scale float32) int32 {
	return int32(float32(base) * scale)
}

func ScaleSize(base int32, scale float32) int32 {
	if scale <= 0 {
		return base
	}
	return int32(float32(base) * scale)
}
