package pulp

import (
	"context"
	"fmt"
	"sync"
)

type Socket struct {
	ID uint32

	updates   chan LiveComponent
	lastState LiveComponent
	Err       error
	context.Context
	events chan<- event

	once sync.Once

	assets struct {
		currentRoute string
		flash        struct {
			err, warning, info *string
		}
	}

	userAssets Assets
}

type Assets map[string]interface{}

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

func (s *Socket) FlashError(route string) {
}

func (s *Socket) FlashInfo(route string) {
}

func (s *Socket) FlashWarning(route string) {
}

func (s *Socket) Redirect(route string) {

}

func (s Socket) Do() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("socket panic: %d\n", s.ID)
			}
		}()

		select {
		case <-s.Context.Done():
			fmt.Println("socket done: ", s.ID)
		case s.updates <- s.lastState:
		}
	}()
}
