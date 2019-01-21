package pusher

// DummyPusher is a do-nothing pusher which can be used in place of an actual pusher
type DummyPusher struct {
}

// PushState does nothing for a DummyPusher
func (p DummyPusher) PushState() error {
	return nil
}
