package core

const (
	dayLength   = 120.0
	nightLength = 60.0
	fullDay     = dayLength + nightLength
)

type Season int

const (
	Spring Season = 0
	Summer Season = 1
	Autumn Season = 2
	Winter Season = 3
)

type TimeSystem struct {
	TotalTime float64
	DayTime   float64
	DayCount  int
	Hour      int
	Minute    int
	Season    Season
}

func NewTimeSystem() *TimeSystem {
	return &TimeSystem{}
}

func (ts *TimeSystem) Tick(delta float64) {
	ts.TotalTime += delta
	ts.DayTime += delta

	if ts.DayTime >= fullDay {
		ts.DayTime -= fullDay
		ts.DayCount++
		ts.updateSeason()
	}

	dayProgress := ts.DayTime / fullDay
	ts.Hour = int(dayProgress * 24)
	ts.Minute = int((dayProgress*24 - float64(ts.Hour)) * 60)
}

func (ts *TimeSystem) updateSeason() {
	daysPerSeason := 90
	seasonIdx := (ts.DayCount / daysPerSeason) % 4
	ts.Season = Season(seasonIdx)
}

func (ts *TimeSystem) IsDaytime() bool {
	return ts.DayTime < dayLength
}

func (ts *TimeSystem) SunAngle() float64 {
	if ts.IsDaytime() {
		return (ts.DayTime / dayLength) * 180.0
	}
	return 0
}
