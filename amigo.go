package amigo

import (
	"context"
	"fmt"
)

type Assigns map[string]interface{}

type Socket struct {
}

type LiveComponent interface {
	Mount(Socket, chan<- Event) error
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

	if err := component.Mount(Socket{}, events); err != nil {
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
				// case component = <-changes:
			}

			newRender := component.Render()
			patches := Diff(lastRender, newRender)
			if patches == nil {
				errors <- fmt.Errorf("nil patches")
				return
			}

			// do this on the client side
			// for k, patch := range map[int]interface{}(*patches) {
			// 	lastRender.Dynamic[k] = patch
			// }

			// fmt.Println(lastRender.String())

			patchesStream <- *patches
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
