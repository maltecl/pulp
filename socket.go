package pulp

import (
	"context"
	"fmt"
)

type Socket struct {
	updates   chan Socket
	lastState LiveComponent
	Err       error
	context.Context
	events chan<- Event
}

func (s *Socket) Dispatch(event string, data map[string]interface{}) {
	select {
	case <-s.Done():
	case s.events <- Event{Name: event, Data: data}:
	}
}

func (s *Socket) Errorf(format string, values ...interface{}) *Socket {
	s.Err = fmt.Errorf(format, values...)
	return s
}

func (s *Socket) Changes(state LiveComponent) *Socket {
	s.lastState = state
	return s
}

func (s Socket) Do() {
	go func() {
		select {
		case <-s.Context.Done():
		case s.updates <- s:
		}
	}()
}
