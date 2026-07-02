package ui

// UnlockRegistry hides toolbar categories until population milestones are reached.
type UnlockRegistry struct {
	Population int
}

func NewUnlockRegistry() *UnlockRegistry { return &UnlockRegistry{} }

func (u *UnlockRegistry) SyncPopulation(pop int) { u.Population = pop }

func (u *UnlockRegistry) Unlocked(cat ToolbarCategory) bool {
	req := categoryUnlockPop[cat]
	return req <= u.Population
}

// ponytail: population gates only; full milestone system comes later.
var categoryUnlockPop = map[ToolbarCategory]int{
	CatRoads:           0,
	CatZoning:          0,
	CatPublicTransport: 0,
	CatOptions:         0,
	CatStatistics:      0,
	CatDistricts:         50,
	CatElectricity:       100,
	CatWater:             150,
	CatGarbage:           200,
	CatHealthcare:        300,
	CatFireRescue:        400,
	CatPolice:            500,
	CatEducation:         600,
	CatLandscaping:       250,
	CatParks:             350,
	CatEconomy:           800,
	CatPolicies:          1000,
}
