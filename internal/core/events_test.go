package core

import "testing"

func TestEventNameConstants(t *testing.T) {
	if EventRoadPlaced != "road:placed" {
		t.Fatalf("EventRoadPlaced = %q", EventRoadPlaced)
	}
	if EventFloodStarted != "flood:started" {
		t.Fatalf("EventFloodStarted = %q", EventFloodStarted)
	}

	bus := NewEventBus()
	var got any
	bus.On(string(EventRoadPlaced), func(data any) { got = data })
	bus.Emit(string(EventRoadPlaced), 42)
	bus.ProcessQueue()
	if got != 42 {
		t.Fatalf("event payload: got %v want 42", got)
	}
}
