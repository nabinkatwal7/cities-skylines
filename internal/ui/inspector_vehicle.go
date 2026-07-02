package ui

import (
	"fmt"

	"github.com/katwate/js-skylines/internal/core"
	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/sim"
)

func buildVehicleSelection(sm *sim.SimulationManager, slot int32, v *road.Vehicle) Selection {
	dest := "—"
	routeLen := len(v.Path)
	if routeLen > 0 && v.PathIdx < routeLen {
		if v.PathIdx+1 < routeLen {
			dest = fmt.Sprintf("Node %d", v.Path[v.PathIdx+1])
		} else {
			dest = fmt.Sprintf("Node %d", v.Path[routeLen-1])
		}
	}
	passengers := 0
	cargo := "—"
	if v.Type == road.VehicleBus {
		passengers = 8 // ponytail: until passenger sim hooks in
	}
	if v.Type == road.VehicleTruck {
		cargo = "Freight"
	}
	state := vehicleState(v)
	return Selection{
		Kind:  InspVehicle,
		Title: road.VehicleTypeName(v.Type),
		Lines: []string{
			fmt.Sprintf("Speed: %.1f (target %.1f)", v.Speed, v.TargetSpeed),
			fmt.Sprintf("Destination: %s", dest),
			fmt.Sprintf("Route: %d nodes", routeLen),
			fmt.Sprintf("Passengers: %d", passengers),
			fmt.Sprintf("Cargo: %s", cargo),
			fmt.Sprintf("Owner: %s", ownerName(v.Owner)),
			fmt.Sprintf("State: %s", state),
			fmt.Sprintf("Segment: %d", v.RoadSeg),
		},
		Actions:     []InspectorAction{{ID: "follow", Label: "Follow"}},
		vehicleSlot: slot,
		followX:     v.Position.X,
		followZ:     v.Position.Z,
	}
}

func vehicleState(v *road.Vehicle) string {
	if v.ParkTimer > 0 || v.ParkSpotIdx >= 0 {
		return "Parked"
	}
	if v.Waiting > 0 {
		return "Waiting"
	}
	if v.Speed < 0.5 {
		return "Stopped"
	}
	return "Driving"
}

func ownerName(o core.OwnerType) string {
	switch o {
	case core.OwnerVehicle:
		return "Vehicle"
	case core.OwnerBuilding:
		return "Building"
	case core.OwnerCitizen:
		return "Citizen"
	default:
		return "City"
	}
}
