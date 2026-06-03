package messaging

// Register wires a handler to an event name.
func Register(bus *Bus, eventName string, h Handler) {
	bus.Subscribe(eventName, h)
}
