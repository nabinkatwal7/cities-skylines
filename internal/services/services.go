package services

// CityServices tracks city-wide utility networks.
// ponytail: city-wide flags until chapter 7 adds radius propagation from plants.
type CityServices struct {
	Electricity bool
	Water       bool
	Sewage      bool
}

func NewStarter() *CityServices {
	return &CityServices{Electricity: true, Water: true, Sewage: true}
}

func (c *CityServices) HasElectricity(_, _ float32) bool {
	return c != nil && c.Electricity
}

func (c *CityServices) HasWater(_, _ float32) bool {
	return c != nil && c.Water
}

func (c *CityServices) HasSewage(_, _ float32) bool {
	return c != nil && c.Sewage
}
