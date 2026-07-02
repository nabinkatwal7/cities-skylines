package ui

// StatisticsPanel displays city-wide metrics (24.15).
type StatisticsPanel struct {
	open bool
}

func NewStatisticsPanel() *StatisticsPanel { return &StatisticsPanel{} }

func (s *StatisticsPanel) Draw() {}

func (s *StatisticsPanel) Open() bool { return s.open }
