package ui

import (
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// MainToolbar is the primary build-tool selector and option bar.
type MainToolbar struct{}

func NewMainToolbar() *MainToolbar { return &MainToolbar{} }

func (tb *MainToolbar) Draw(ts *ToolSystem) {
	tb.drawBar(ts)
	tb.drawOptions(ts)
}

func (tb *MainToolbar) HandleClick(ts *ToolSystem) GameTool {
	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)
	totalW := len(ToolbarItems)*ToolbarBtnW + (len(ToolbarItems)-1)*ToolbarPad
	startX := (ScreenW - totalW) / 2
	for i, item := range ToolbarItems {
		bx := int32(startX + i*(ToolbarBtnW+ToolbarPad))
		by := int32(ToolbarY)
		if mx >= bx && mx < bx+ToolbarBtnW && my >= by && my < by+ToolbarBtnH {
			ts.Selected = item.Tool
			return item.Tool
		}
	}
	return ts.Selected
}

func (tb *MainToolbar) drawBar(ts *ToolSystem) {
	rl.DrawRectangle(0, ToolbarY, ScreenW, ToolbarH, rl.NewColor(0, 0, 0, 200))
	totalW := len(ToolbarItems)*ToolbarBtnW + (len(ToolbarItems)-1)*ToolbarPad
	startX := (ScreenW - totalW) / 2
	for i, item := range ToolbarItems {
		bx := int32(startX + i*(ToolbarBtnW+ToolbarPad))
		by := int32(ToolbarY + (ToolbarH-ToolbarBtnH)/2)
		col := item.Color
		sel := ts.Selected == item.Tool
		textCol := rl.White
		if sel {
			textCol = rl.NewColor(255, 255, 200, 255)
		}
		uiBtn(bx, by, ToolbarBtnW, ToolbarBtnH, item.Label, col, textCol, sel)
	}
}

func (tb *MainToolbar) drawOptions(ts *ToolSystem) {
	if ts.Selected == ToolPointer || ts.Selected == ToolRemove || ts.Selected == ToolUpgrade {
		return
	}
	rl.DrawRectangle(0, ToolbarY-OptionsBarH, ScreenW, OptionsBarH, rl.NewColor(0, 0, 0, 160))

	mPos := rl.GetMousePosition()
	mx := int32(mPos.X)
	my := int32(mPos.Y)

	switch ts.Selected {
	case ToolRoad:
		item := &ToolbarItems[1]
		optW := int32(80)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (ScreenW - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			uiBtn(bx, by, optW, OptionsBarH-8, opt, rl.NewColor(40, 40, 40, 200), rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				ts.RoadType = road.RoadType(oi)
			}
		}
	case ToolParking:
		item := &ToolbarItems[2]
		optW := int32(95)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (ScreenW - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			uiBtn(bx, by, optW, OptionsBarH-8, opt, rl.NewColor(40, 40, 40, 200), rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				ts.setParkingMode(oi)
			}
		}
	case ToolTransport:
		item := &ToolbarItems[3]
		optW := int32(75)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (ScreenW - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			col := rl.NewColor(200, 150, 50, 200)
			if oi < int(transport.TransportModeCount) {
				col = transport.TransportStopColor(transport.TransportType(oi))
				col.A = 200
			}
			uiBtn(bx, by, optW, OptionsBarH-8, opt, col, rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				if oi >= int(transport.TransportModeCount) {
					ts.CargoMode = true
				} else {
					ts.CargoMode = false
					ts.TransportType = transport.TransportType(oi)
				}
			}
		}
	case ToolZone:
		item := &ToolbarItems[4]
		optW := int32(80)
		total := len(item.Options) * int(optW+ToolbarPad)
		sx := (ScreenW - total) / 2
		for oi, opt := range item.Options {
			bx := int32(sx + oi*(int(optW)+int(ToolbarPad)))
			by := int32(ToolbarY - OptionsBarH + 4)
			sel := oi == item.OptIndex
			col := rl.NewColor(40, 40, 40, 200)
			if oi < 6 {
				col = zoning.ZoneColor(zoning.ZoneType(oi + 1))
				col.A = 200
			}
			uiBtn(bx, by, optW, OptionsBarH-8, opt, col, rl.White, sel)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && mx >= bx && mx < bx+optW && my >= by && my < by+OptionsBarH-8 {
				item.OptIndex = oi
				ts.ZoneType = oi
			}
		}
	}
}
