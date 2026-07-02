package ui

import (
	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Overlay layers render above world geometry (24.12). Heatmaps never alter gameplay.
//
//	Terrain → Roads → Buildings → Heatmap → Icons → Labels → Selection
type OverlayLayer uint8

const (
	LayerHeatmap OverlayLayer = 1 << iota
	LayerIcons
	LayerLabels
	LayerSelection
)

// OverlayManager coordinates independent presentation layers.
type OverlayManager struct {
	enabled OverlayLayer
	labels  []overlayLabel
}

type overlayLabel struct {
	x, z    float32
	text    string
	iconCol rl.Color
}

func NewOverlayManager() *OverlayManager {
	return &OverlayManager{
		enabled: LayerHeatmap | LayerIcons | LayerLabels | LayerSelection,
	}
}

func (o *OverlayManager) SetLayer(layer OverlayLayer, on bool) {
	if on {
		o.enabled |= layer
	} else {
		o.enabled &^= layer
	}
}

func (o *OverlayManager) LayerOn(layer OverlayLayer) bool {
	return o.enabled&layer != 0
}

func (o *OverlayManager) Visible() bool { return o.enabled != 0 }

// DrawWorld renders world-space overlay layers after buildings.
func (o *OverlayManager) DrawWorld(sm *sim.SimulationManager, info *InfoViews, sel *SelectionSystem) {
	if sm == nil {
		return
	}
	if o.LayerOn(LayerHeatmap) && info != nil {
		info.DrawWorld(sm)
	}
	if o.LayerOn(LayerIcons) {
		o.drawIcons(sm)
	}
	if o.LayerOn(LayerLabels) {
		o.drawLabels(sm)
	}
	if o.LayerOn(LayerSelection) && sel != nil {
		sel.DrawHighlight(sm)
	}
}

func (o *OverlayManager) drawIcons(sm *sim.SimulationManager) {
	if sm.Transport == nil {
		return
	}
	for _, stop := range sm.Transport.Stops {
		hy := sm.Heightmap.WorldHeight(stop.X, stop.Z) + 1.2
		rl.DrawSphere(rl.NewVector3(stop.X, hy, stop.Z), 0.35, rl.NewColor(255, 220, 80, 120))
	}
}

func (o *OverlayManager) drawLabels(sm *sim.SimulationManager) {
	o.labels = o.labels[:0]
	if sm.Buildings != nil {
		for i := range sm.Buildings.Buildings {
			b := &sm.Buildings.Buildings[i]
			if b.State != 0 && b.Level >= 3 {
				o.labels = append(o.labels, overlayLabel{
					x: b.WorldX, z: b.WorldZ,
					text: zoneTypeName(b.Type),
				})
			}
		}
	}
	for _, lb := range o.labels {
		hy := sm.Heightmap.WorldHeight(lb.x, lb.z) + 3
		rl.DrawSphere(rl.NewVector3(lb.x, hy, lb.z), 0.15, rl.NewColor(200, 200, 255, 100))
	}
}

func (o *OverlayManager) Draw() {
	// Screen-space overlay chrome (layer legend)
	if o.enabled == 0 {
		return
	}
	y := int32(TopBarH + 48)
	if o.LayerOn(LayerHeatmap) {
		DrawUIText("[Heat]", 8, y, 11, rl.Gray)
		y += 14
	}
	if o.LayerOn(LayerSelection) {
		DrawUIText("[Select]", 8, y, 11, rl.Gray)
	}
}
