package amigo

import (
	"bytes"
	"context"
	"html/template"
)

type Assigns map[string]interface{}

type Socket struct {
}

type LiveComponent interface {
	Mount(Socket, chan<- Event) error
	Render() (Assets, string)
	HandleEvent(Event, chan<- LiveComponent) error
	Name() string
}

type Event struct {
	Name string
	Data map[string]string
}

func New(ctx context.Context, component LiveComponent, events chan Event, errors chan<- error) <-chan string {
	renders := make(chan string)
	changes := make(chan LiveComponent)

	if err := component.Mount(Socket{}, events); err != nil {
		errors <- err
		return nil
	}

	go func() {
		render(component, errors, renders)

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

			render(component, errors, renders)
		}

		close(changes)
		close(renders)
	}()

	return renders
}

func render(component LiveComponent, errors chan<- error, renders chan string) {
	assets, templateString := component.Render()

	tt, err := template.New(component.Name()).Parse(templateString)
	if err != nil {
		errors <- err
	}

	renderBuff := &bytes.Buffer{}
	err = tt.Execute(renderBuff, WithAssets{A: assets, C: component})
	if err != nil {
		errors <- err
	}

	renders <- renderBuff.String()
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
