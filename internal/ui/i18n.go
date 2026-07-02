package ui

// i18n provides localized UI strings with dynamic switching (24.24).
var currentLocale = "en"

var catalogs = map[string]map[string]string{
	"en": enCatalog,
	"es": esCatalog,
}

var enCatalog = map[string]string{
	"hud.speed":           "Speed",
	"hud.paused":          "Paused",
	"options.title":       "Options",
	"options.accessibility": "Accessibility",
	"options.bindings":    "Shortcuts",
	"options.language":    "Language",
	"a11y.ui_scale":       "UI Scale",
	"a11y.high_contrast":  "High Contrast",
	"a11y.subtitles":      "Subtitles",
	"a11y.reduced_motion": "Reduced Motion",
	"a11y.color_blind":    "Color Blind Mode",
	"bind.pause":          "Pause",
	"bind.speed1":         "Normal Speed",
	"bind.speed2":         "Double Speed",
	"bind.speed3":         "Triple Speed",
	"bind.undo":           "Undo",
	"bind.screenshot":     "Screenshot",
	"bind.cam_reset":      "Reset Camera",
	"bind.info_cycle":     "Cycle Info View",
	"bind.info_clear":     "Clear Info View",
	"bind.advisors":       "Advisors",
	"bind.search":         "Search",
	"bind.statistics":     "Statistics",
	"search.title":        "Search (Enter go, Esc close)",
	"search.placeholder":  "Type to search...",
	"advisors.title":      "City Advisors",
	"stats.title":         "City Statistics",
	"input.keyboard":      "Keyboard",
	"input.gamepad":       "Gamepad",
	"subtitle.prefix":     "Alert",
}

var esCatalog = map[string]string{
	"hud.speed":           "Velocidad",
	"hud.paused":          "Pausa",
	"options.title":       "Opciones",
	"options.accessibility": "Accesibilidad",
	"options.bindings":    "Atajos",
	"options.language":    "Idioma",
	"a11y.ui_scale":       "Escala UI",
	"a11y.high_contrast":  "Alto Contraste",
	"a11y.subtitles":      "Subtítulos",
	"a11y.reduced_motion": "Movimiento Reducido",
	"a11y.color_blind":    "Modo Daltónico",
	"bind.pause":          "Pausa",
	"bind.speed1":         "Velocidad Normal",
	"bind.speed2":         "Doble Velocidad",
	"bind.speed3":         "Triple Velocidad",
	"bind.undo":           "Deshacer",
	"bind.screenshot":     "Captura",
	"bind.cam_reset":      "Restablecer Cámara",
	"bind.info_cycle":     "Ciclar Vista",
	"bind.info_clear":     "Limpiar Vista",
	"bind.advisors":       "Asesores",
	"bind.search":         "Buscar",
	"bind.statistics":     "Estadísticas",
	"search.title":        "Buscar (Enter ir, Esc cerrar)",
	"search.placeholder":  "Escribe para buscar...",
	"advisors.title":      "Asesores de la Ciudad",
	"stats.title":         "Estadísticas",
	"input.keyboard":      "Teclado",
	"input.gamepad":       "Mando",
	"subtitle.prefix":     "Aviso",
}

func T(key string) string {
	if cat, ok := catalogs[currentLocale]; ok {
		if s, ok := cat[key]; ok {
			return s
		}
	}
	if s, ok := catalogs["en"][key]; ok {
		return s
	}
	return key
}

func SetLocale(lang string) {
	if _, ok := catalogs[lang]; ok {
		currentLocale = lang
	}
}

func CurrentLocale() string { return currentLocale }

func AvailableLocales() []string { return []string{"en", "es"} }
