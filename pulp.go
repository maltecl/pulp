package pulp

import (
	"context"
	"fmt"
	"log"
)

type Assigns map[string]interface{}

type Socket struct {
	updates   chan Socket
	lastState LiveComponent
	Err       error
	context.Context
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

type LiveComponent interface {
	Mount(Socket)
	Render() StaticDynamic
	HandleEvent(Event, Socket)
	Name() string
}

type Event struct {
	Name string
	Data map[string]string
}

func New(ctx context.Context, component LiveComponent, events chan Event, errors chan<- error, onMount chan<- StaticDynamic) <-chan Patches {

	socket := Socket{Context: ctx, updates: make(chan Socket)}
	patchesStream := make(chan Patches)

	component.Mount(socket)
	socket = <-socket.updates
	if socket.Err != nil {
		errors <- socket.Err
		return nil
	}

	lastRender := component.Render()
	go func() { onMount <- lastRender }()
	// onMount is closed

	go func() {
		defer func() {
			close(socket.updates)
			close(patchesStream)
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				break outer
			case event := <-events:
				component.HandleEvent(event, socket)
				continue outer
			case socket = <-socket.updates:
				if socket.Err != nil {
					errors <- socket.Err
				}
				component = socket.lastState
			}

			newRender := component.Render()
			patches := lastRender.Dynamic.Diff(newRender.Dynamic)
			if patches == nil {
				log.Println("empty patches")
				continue
			}

			lastRender = newRender
			select {
			case <-ctx.Done():
				break outer
			case patchesStream <- *patches:
			}
		}
	}()

	return patchesStream
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type Assets map[string]interface{}

type WithAssets struct {
	A Assets
	C LiveComponent
}

type HTML interface{ HTML() }

type L struct{ Out string }

func (L) HTML() {}

func (StaticDynamic) HTML() {}
