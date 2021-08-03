package amigo

import (
	"context"
	"fmt"
)

type Assigns map[string]interface{}

type Socket struct {
}

type LiveComponent interface {
	Mount(Socket, chan<- Event, chan<- LiveComponent) error
	Render() StaticDynamic
	HandleEvent(Event, chan<- LiveComponent) error
	Name() string
}

type Event struct {
	Name string
	Data map[string]string
}

func New(ctx context.Context, component LiveComponent, events chan Event, errors chan<- error, onMount chan<- StaticDynamic) <-chan Patches {
	patchesStream := make(chan Patches)
	changes := make(chan LiveComponent)

	if err := component.Mount(Socket{}, events, changes); err != nil {
		errors <- err
		return nil
	}

	lastRender := component.Render()

	go func() {
		onMount <- lastRender
	}()

	// onMount is closed

	go func() {

	outer:
		for {
			select {
			case <-ctx.Done():
				break outer
			case event := <-events:
				err := component.HandleEvent(event, changes)
				if err != nil {
					break outer
				}
			case component = <-changes:
			}

			newRender := component.Render()
			patches, patchesNotEmpty := Diff(lastRender, newRender)
			if patches == nil {
				errors <- fmt.Errorf("nil patches")
				return
			}

			if patchesNotEmpty {
				lastRender = newRender
				select {
				case <-ctx.Done():
					break outer
				case patchesStream <- *patches:
				}
			}
		}

		close(changes)
		close(patchesStream)
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
