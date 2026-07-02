package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/sim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const statsHistoryLen = 48

// StatsReport holds detailed city metrics (24.15).
type StatsReport struct {
	Population   int
	Money        float32
	WeeklyIncome float32
	Happiness    float32
	Traffic      float32
	Industrial   float32
	TransitPax   int32
	Pollution    float32
	Education    float32
	Health       float32
	Crime        float32
	Tourism      float32
	Utilities    float32
}

// StatisticsPanel displays city reports with history (24.15).
type StatisticsPanel struct {
	open     bool
	view     ViewState
	report   StatsReport
	history  [statsHistoryLen]int
	histIdx  int
	histFull bool
	tab      int
}

func NewStatisticsPanel() *StatisticsPanel { return &StatisticsPanel{} }

func (s *StatisticsPanel) Sync(view ViewState) { s.view = view }

func (s *StatisticsPanel) SyncSim(sm *sim.SimulationManager, view ViewState) {
	s.view = view
	s.report = StatsReport{
		Population:   view.Population,
		Money:        view.Money,
		WeeklyIncome: view.WeeklyIncome,
		Happiness:    view.Happiness,
	}
	if sm == nil {
		return
	}
	if sm.Demand != nil {
		f := sm.Demand.Factors
		s.report.Industrial = sm.Demand.Industrial
		s.report.Pollution = f.Pollution
		s.report.Education = sm.Demand.Education
		s.report.Health = f.ServiceScore
		s.report.Crime = f.Crime
		s.report.Tourism = f.Tourism
		s.report.Traffic = f.FreightCongestion
		s.report.Utilities = f.ServiceScore
	}
	if sm.Transport != nil {
		for _, net := range sm.Transport.Networks {
			s.report.TransitPax += net.WeeklyPassengers
		}
	}
}

func (s *StatisticsPanel) RecordDay(pop int) {
	s.history[s.histIdx] = pop
	s.histIdx = (s.histIdx + 1) % statsHistoryLen
	if s.histIdx == 0 {
		s.histFull = true
	}
}

func (s *StatisticsPanel) Open() bool { return s.open }

func (s *StatisticsPanel) Draw() {
	if !s.open {
		return
	}
	w, h := int32(420), int32(300)
	x := int32((ScreenW - w) / 2)
	y := int32(TopBarH + 8)
	rl.DrawRectangle(x, y, w, h, rl.NewColor(0, 0, 0, 220))
	rl.DrawRectangleLines(x, y, w, h, rl.Gray)
	DrawUIText(T("stats.title"), x+10, y+8, 16, rl.White)

	tabs := []string{"Pop", "Economy", "Traffic", "Industry", "Transit", "Env"}
	for i, label := range tabs {
		bx := x + 10 + int32(i*62)
		sel := i == s.tab
		col := rl.NewColor(40, 40, 45, 200)
		if sel {
			col = rl.NewColor(60, 70, 90, 220)
		}
		uiBtn(bx, y+28, 58, 20, label, col, rl.White, sel)
	}

	ly := y + 56
	switch s.tab {
	case 0:
		s.drawPop(x, ly, w)
	case 1:
		s.drawEconomy(x, ly)
	case 2:
		s.drawLine(x, ly, w, "Traffic load", s.report.Traffic)
	case 3:
		s.drawLine(x, ly, w, "Industrial demand", s.report.Industrial)
	case 4:
		DrawUIText(fmt.Sprintf("Transit passengers/wk: %d", s.report.TransitPax), x+10, ly, 14, rl.LightGray)
	case 5:
		DrawUIText(fmt.Sprintf("Pollution: %.0f%%  Crime: %.0f%%", s.report.Pollution*100, s.report.Crime*100), x+10, ly, 14, rl.LightGray)
		DrawUIText(fmt.Sprintf("Education: %.0f%%  Health: %.0f%%", s.report.Education*100, s.report.Health*100), x+10, ly+18, 14, rl.LightGray)
		DrawUIText(fmt.Sprintf("Tourism: %.0f%%  Utilities: %.0f%%", s.report.Tourism*100, s.report.Utilities*100), x+10, ly+36, 14, rl.LightGray)
	}
	s.drawHistoryGraph(x+w-160, y+h-70, 150, 60)
}

func (s *StatisticsPanel) drawPop(x, ly, w int32) {
	DrawUIText(fmt.Sprintf("Population: %d", s.report.Population), x+10, ly, 14, rl.LightGray)
	DrawUIText(fmt.Sprintf("Happiness: %.0f%%", s.report.Happiness*100), x+10, ly+18, 14, rl.LightGray)
	DrawUIText("Milestone: "+s.view.Milestone, x+10, ly+36, 14, rl.LightGray)
}

func (s *StatisticsPanel) drawEconomy(x, ly int32) {
	DrawUIText(fmt.Sprintf("Treasury: $%.0f", s.report.Money), x+10, ly, 14, rl.LightGray)
	DrawUIText(fmt.Sprintf("Weekly income: %+0.f", s.report.WeeklyIncome), x+10, ly+18, 14, rl.LightGray)
}

func (s *StatisticsPanel) drawLine(x, ly, w int32, label string, val float32) {
	DrawUIText(fmt.Sprintf("%s: %.0f%%", label, val*100), x+10, ly, 14, rl.LightGray)
	barW := w - 30
	fill := int32(float32(barW) * clamp01(val))
	rl.DrawRectangle(x+10, ly+22, barW, 10, rl.NewColor(30, 30, 30, 200))
	rl.DrawRectangle(x+10, ly+22, fill, 10, rl.NewColor(100, 180, 220, 200))
}

func (s *StatisticsPanel) drawHistoryGraph(x, y, w, h int32) {
	rl.DrawRectangle(x, y, w, h, rl.NewColor(20, 20, 25, 200))
	DrawUIText("Population", x+4, y+2, 10, rl.Gray)
	n := statsHistoryLen
	if !s.histFull {
		n = s.histIdx
	}
	if n < 2 {
		return
	}
	maxP := 1
	vals := make([]int, n)
	for i := 0; i < n; i++ {
		idx := (s.histIdx - n + i + statsHistoryLen) % statsHistoryLen
		vals[i] = s.history[idx]
		if vals[i] > maxP {
			maxP = vals[i]
		}
	}
	for i := 1; i < n; i++ {
		x1 := x + int32(i-1)*w/int32(n-1)
		x2 := x + int32(i)*w/int32(n-1)
		y1 := y + h - int32(float32(h)*float32(vals[i-1])/float32(maxP))
		y2 := y + h - int32(float32(h)*float32(vals[i])/float32(maxP))
		rl.DrawLine(x1, y1, x2, y2, rl.SkyBlue)
	}
}

func (s *StatisticsPanel) HandleClick(mx, my int32) bool {
	if !s.open {
		return false
	}
	w, h := int32(420), int32(300)
	x := int32((ScreenW - w) / 2)
	y := int32(TopBarH + 8)
	if mx < x || mx >= x+w || my < y || my >= y+h {
		return false
	}
	if my >= y+28 && my < y+48 {
		tab := int((mx - x - 10) / 62)
		if tab >= 0 && tab < 6 {
			s.tab = tab
		}
		return true
	}
	return true
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
