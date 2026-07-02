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

type graphLine struct {
	x1, y1, x2, y2 int32
}

type statsGraphCache struct {
	valid bool
	lines []graphLine
}

// StatisticsPanel displays city reports with history (24.15).
type StatisticsPanel struct {
	open     bool
	inited   bool
	dirty    bool
	view     ViewState
	report   StatsReport
	history  [statsHistoryLen]int
	histIdx  int
	histFull bool
	tab      int
	graph    statsGraphCache
}

func NewStatisticsPanel() *StatisticsPanel { return &StatisticsPanel{} }

func (s *StatisticsPanel) ensureInit() {
	if s.inited {
		return
	}
	s.inited = true
}

func (s *StatisticsPanel) MarkDirty() { s.dirty = true; s.graph.valid = false }

func (s *StatisticsPanel) Sync(view ViewState) {
	s.view = view
}

func (s *StatisticsPanel) SyncSim(sm *sim.SimulationManager, view ViewState) {
	if !s.open {
		return
	}
	s.ensureInit()
	s.view = view
	s.report = StatsReport{
		Population:   view.Population,
		Money:        view.Money,
		WeeklyIncome: view.WeeklyIncome,
		Happiness:    view.Happiness,
	}
	if sm == nil {
		s.dirty = false
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
		s.report.TransitPax = 0
		for _, net := range sm.Transport.Networks {
			s.report.TransitPax += net.WeeklyPassengers
		}
	}
	s.dirty = false
}

func (s *StatisticsPanel) RecordDay(pop int) {
	s.history[s.histIdx] = pop
	s.histIdx = (s.histIdx + 1) % statsHistoryLen
	if s.histIdx == 0 {
		s.histFull = true
	}
	s.graph.valid = false
}

func (s *StatisticsPanel) Open() bool { return s.open }

func (s *StatisticsPanel) Draw() {
	if !s.open {
		return
	}
	s.ensureInit()
	w, h := int32(440), int32(340)
	x := int32((ScreenW - w) / 2)
	y := int32(TopBarH + 12)
	drawPanel(x, y, w, h)
	drawLabel(T("stats.title"), x+14, y+12, FontLg, csText)

	tabs := []string{"Pop", "Economy", "Traffic", "Industry", "Transit", "Env"}
	for i, label := range tabs {
		bx := x + 14 + int32(i)*68
		sel := i == s.tab
		csOptionBtn(bx, y+36, 64, 26, label, csBtnIdle, sel)
	}

	ly := y + 72
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
		drawLabel(fmt.Sprintf("Transit passengers/wk: %d", s.report.TransitPax), x+14, ly, FontMd, csText)
	case 5:
		drawLabel(fmt.Sprintf("Pollution: %.0f%%  Crime: %.0f%%", s.report.Pollution*100, s.report.Crime*100), x+14, ly, FontMd, csTextDim)
		drawLabel(fmt.Sprintf("Education: %.0f%%  Health: %.0f%%", s.report.Education*100, s.report.Health*100), x+14, ly+22, FontMd, csTextDim)
		drawLabel(fmt.Sprintf("Tourism: %.0f%%  Utilities: %.0f%%", s.report.Tourism*100, s.report.Utilities*100), x+14, ly+44, FontMd, csTextDim)
	}
	s.drawHistoryGraph(x+w-170, y+h-80, 160, 68)
}

func (s *StatisticsPanel) drawPop(x, ly, w int32) {
	drawLabel(fmt.Sprintf("Population: %d", s.report.Population), x+14, ly, FontMd, csText)
	drawLabel(fmt.Sprintf("Happiness: %.0f%%", s.report.Happiness*100), x+14, ly+22, FontMd, csTextDim)
	drawLabel("Milestone: "+s.view.Milestone, x+14, ly+44, FontMd, csTextDim)
}

func (s *StatisticsPanel) drawEconomy(x, ly int32) {
	drawLabel(fmt.Sprintf("Treasury: $%.0f", s.report.Money), x+14, ly, FontMd, csMoney)
	drawLabel(fmt.Sprintf("Weekly income: %+0.f", s.report.WeeklyIncome), x+14, ly+22, FontMd, csTextDim)
}

func (s *StatisticsPanel) drawLine(x, ly, w int32, label string, val float32) {
	drawLabel(fmt.Sprintf("%s: %.0f%%", label, val*100), x+14, ly, FontMd, csText)
	barW := w - 36
	fill := int32(float32(barW) * clamp01(val))
	rl.DrawRectangle(x+14, ly+26, barW, 12, csInputBg)
	rl.DrawRectangle(x+14, ly+26, fill, 12, csBarLine)
}

func (s *StatisticsPanel) rebuildGraphCache(x, y, w, h int32) {
	n := statsHistoryLen
	if !s.histFull {
		n = s.histIdx
	}
	if n < 2 {
		s.graph.valid = true
		s.graph.lines = nil
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
	lines := make([]graphLine, 0, n-1)
	for i := 1; i < n; i++ {
		lines = append(lines, graphLine{
			x1: x + int32(i-1)*w/int32(n-1),
			y1: y + h - int32(float32(h)*float32(vals[i-1])/float32(maxP)),
			x2: x + int32(i)*w/int32(n-1),
			y2: y + h - int32(float32(h)*float32(vals[i])/float32(maxP)),
		})
	}
	s.graph.lines = lines
	s.graph.valid = true
}

func (s *StatisticsPanel) drawHistoryGraph(x, y, w, h int32) {
	rl.DrawRectangle(x, y, w, h, csInputBg)
	drawLabel("Population", x+6, y+4, FontXs, csTextDim)
	if !s.graph.valid {
		s.rebuildGraphCache(x, y, w, h)
	}
	for _, ln := range s.graph.lines {
		rl.DrawLine(ln.x1, ln.y1, ln.x2, ln.y2, csBarLine)
	}
}

func (s *StatisticsPanel) HandleClick(mx, my int32) bool {
	if !s.open {
		return false
	}
	s.ensureInit()
	w, h := int32(440), int32(340)
	x := int32((ScreenW - w) / 2)
	y := int32(TopBarH + 12)
	if mx < x || mx >= x+w || my < y || my >= y+h {
		return false
	}
	if my >= y+36 && my < y+64 {
		tab := int((mx - x - 14) / 68)
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
