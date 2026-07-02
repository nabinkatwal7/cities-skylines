package core

type EventName string

const (
	EventRoadPlaced          EventName = "road:placed"
	EventRoadRemoved         EventName = "road:removed"
	EventRoadUpgraded        EventName = "road:upgraded"
	EventDayNightCycle       EventName = "time:daynight"
	EventTaxCollected        EventName = "economy:tax"
	EventTimeMinute          EventName = "time:minute"
	EventTimeHour            EventName = "time:hour"
	EventTimeDay             EventName = "time:day"
	EventFloodStarted        EventName = "flood:started"
	EventFloodReceded        EventName = "flood:receded"
	EventParkingLotPlaced    EventName = "parkinglot:placed"
	EventParkingGaragePlaced EventName = "parkinggarage:placed"
	EventParkingLotRemoved   EventName = "parkinglot:removed"
	EventZonePlaced          EventName = "zone:placed"
	EventZoneRemoved         EventName = "zone:removed"
)
