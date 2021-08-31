package pulp

import (
	"context"
	"fmt"
	"sync"
)

type Socket struct {
	ID uint32

	updates   chan Socket
	lastState LiveComponent
	Err       error
	context.Context
	events chan<- Event

	once sync.Once
}

type M map[string]interface{}

// don't use this yet. this is not working perfectly
func (s *Socket) Dispatch(event string, data M) {
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
