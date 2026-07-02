package ui

// InspectorKind identifies which simulation entity is selected (24.7).
type InspectorKind int

const (
	InspNone InspectorKind = iota
	InspBuilding
	InspCitizen
	InspVehicle
	InspDistrict
	InspRoad
	InspUtility
	InspPark
	InspIndustry
	InspTransportLine
	InspZone
)

// InspectorAction is a context-sensitive UI action (presentation only until wired).
type InspectorAction struct {
	ID    string
	Label string
}

// Selection is a read-only handle to an inspected entity.
type Selection struct {
	Kind    InspectorKind
	Title   string
	Lines   []string
	Actions []InspectorAction

	buildingIdx int
	vehicleSlot int32
	followX     float32
	followZ     float32
}

func (s Selection) FollowTarget() (x, z float32, ok bool) {
	if s.Kind == InspNone {
		return 0, 0, false
	}
	return s.followX, s.followZ, true
}
