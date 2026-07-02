package road

// RoadTypeOptionNames matches toolbar / build-menu indices to RoadType values.
var RoadTypeOptionNames = []string{
	"2-Lane", "1-Way", "4-Lane", "Gravel", "Highway", "Roundabout",
	"6-Lane", "Avenue", "Bus Rd", "Tram Rd", "Bike Rd", "Tree Rd",
	"Asym Rd", "Pedestrian", "Quay",
}

// RoadTypeOptions maps option index → RoadType (must stay aligned with iota order).
var RoadTypeOptions = []RoadType{
	RoadTwoLane, RoadOneWay, RoadFourLane, RoadGravel, RoadHighway, RoadRoundabout,
	RoadSixLane, RoadAvenue, RoadBus, RoadTram, RoadBicycle, RoadTreeLined,
	RoadAsymmetric, RoadPedestrian, RoadQuay,
}

func RoadTypeFromOptionIndex(i int) RoadType {
	if i < 0 || i >= len(RoadTypeOptions) {
		return RoadTwoLane
	}
	return RoadTypeOptions[i]
}

func OptionIndexForRoadType(rt RoadType) int {
	for i, o := range RoadTypeOptions {
		if o == rt {
			return i
		}
	}
	return 0
}
