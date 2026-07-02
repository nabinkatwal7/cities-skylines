package core

type GameTime struct {
	TotalTime float64
	DayTime   float64
	DayCount  int32
	Hour      int32
	Minute    int32
	Speed     float64
	IsPaused  bool

	prevMinute int32
	prevHour   int32
	prevDay    int32

	Season Season
}

const (
	gameDayLength   = 120.0
	gameNightLength = 60.0
	gameFullDay     = gameDayLength + gameNightLength
	gameHoursPerDay = 24
	gameMinsPerHour = 60
)

func NewGameTime() *GameTime {
	return &GameTime{
		Speed: 1.0,
	}
}

func (gt *GameTime) Tick(dt float64) {
	if gt.IsPaused || dt == 0 {
		return
	}
	gt.TotalTime += dt
	gt.DayTime += dt

	if gt.DayTime >= gameFullDay {
		gt.DayTime -= gameFullDay
		gt.DayCount++
		gt.updateSeason()
	}

	dayProgress := gt.DayTime / gameFullDay
	gt.Hour = int32(dayProgress * gameHoursPerDay)
	gt.Minute = int32((dayProgress*gameHoursPerDay - float64(gt.Hour)) * gameMinsPerHour)
}

func (gt *GameTime) updateSeason() {
	daysPerSeason := int32(90)
	seasonIdx := (gt.DayCount / daysPerSeason) % 4
	gt.Season = Season(seasonIdx)
}

func (gt *GameTime) IsDaytime() bool {
	return gt.DayTime < gameDayLength
}

func (gt *GameTime) SunAngle() float64 {
	if gt.IsDaytime() {
		return (gt.DayTime / gameDayLength) * 180.0
	}
	return 0
}

func (gt *GameTime) MinuteChanged() bool {
	return gt.Minute != gt.prevMinute
}

func (gt *GameTime) HourChanged() bool {
	return gt.Hour != gt.prevHour
}

func (gt *GameTime) DayChanged() bool {
	return gt.DayCount != gt.prevDay
}

func (gt *GameTime) Snapshot() {
	gt.prevMinute = gt.Minute
	gt.prevHour = gt.Hour
	gt.prevDay = gt.DayCount
}

func (gt *GameTime) TimeString() string {
	h := gt.Hour
	ampm := "AM"
	if h >= 12 {
		ampm = "PM"
		if h > 12 {
			h -= 12
		}
	}
	if h == 0 {
		h = 12
	}
	return formatTime(h, gt.Minute, ampm)
}

func formatTime(h, m int32, ampm string) string {
	buf := []byte{
		byte('0' + h/10),
		byte('0' + h%10),
		':',
		byte('0' + m/10),
		byte('0' + m%10),
		' ',
	}
	buf = append(buf, []byte(ampm)...)
	return string(buf)
}
