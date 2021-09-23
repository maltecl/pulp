package pulp

import (
	"context"
)

type Socket struct {
	ID uint32

	updates   chan socketUpdate
	component LiveComponent
	Err       error
	context.Context
	events chan<- Event

	Route string
}

type socketUpdate struct {
	err   *error
	route *string
}

type Assets map[string]interface{}

func (a Assets) mergeAndOverwrite(other Assets) Assets {
	if a == nil {
		return other
	}

	for key, val := range other {
		a[key] = val
	}
	return a
}

// TODO: make sure this works fine
// type M map[string]interface{}

// func (s *Socket) Dispatch(event string, data M) {
// 	select {
// 	case <-s.Done():
// 	case s.events <- UserEvent{Name: event, Data: data}:
// 	}
// }

// func (s *Socket) Errorf(format string, values ...interface{}) {
// 	err := fmt.Errorf(format, values...)
// 	s.sendUpdate(socketUpdate{err: &err})
// }

// TODO: add flash messages that will be sent as assets
// func (s *Socket) FlashError(route string) {
// }

// func (s *Socket) FlashInfo(route string) {
// }

// func (s *Socket) FlashWarning(route string) {
// }

func (s Socket) assets() Assets {
	return Assets{
		"route": s.Route,
	}
}

func (s *Socket) sendUpdate(update socketUpdate) {
	select {
	case <-s.Done():
	case s.updates <- update:
	}
}

func (s *Socket) Update() {
	s.sendUpdate(socketUpdate{})
}

func (s *Socket) Redirect(route string) {
	s.sendUpdate(socketUpdate{route: &route})
}

func (u socketUpdate) apply(socket *Socket) {
	if u.route != nil {
		socket.Route = *u.route
	}

	if u.err != nil {
		socket.Err = *u.err
	}
}
