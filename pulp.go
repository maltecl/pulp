package pulp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

type LiveComponent interface {
	Mount(Socket)
	Render(Socket) (HTML, Assets) // HTML guranteed to be StaticDynamic after code generation
	HandleEvent(Event, Socket)
}

type Unmountable interface {
	Unmount()
}

type Event interface {
	event()
}
type UserEvent struct {
	Name string
	Data map[string]interface{}
}

type RouteChangedEvent struct {
	From, To string
}

func (UserEvent) event()         {}
func (RouteChangedEvent) event() {}

var socketID = uint32(0)

func newPatchesStream(ctx context.Context, component LiveComponent, events chan Event, route string) (rootNode, <-chan Patches, <-chan error) {

	// TODO: @router get route from initial HTTP request
	socket := Socket{
		Context:   ctx,
		updates:   make(chan socketUpdate, 10),
		events:    events,
		ID:        socketID,
		Route:     route,
		component: component,
	}

	atomic.AddUint32(&socketID, 1)

	errors := make(chan error)
	patchesStream := make(chan Patches)

	socket.component.Mount(socket)

	initalTemplate, initialUserAssets := socket.component.Render(socket)
	lastTemplate, ok := initalTemplate.(StaticDynamic)
	if !ok {
		fmt.Println("the first return value of the call to the Render() method is not of type StaticDynamic, this means that you probably did not generate your code first")
		os.Exit(1)
	}

	lastRender := rootNode{DynHTML: lastTemplate, UserAssets: initialUserAssets.mergeAndOverwrite(socket.assets())}
	// onMount is closed

	go func() {
		defer func() {
			close(errors)
			close(patchesStream)
			close(socket.updates)
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-events:
				if userEvent, ok := event.(UserEvent); ok {
					socket.component.HandleEvent(userEvent, socket)
					continue outer
				}

				if routeEvent, ok := event.(RouteChangedEvent); ok {
					socket.component.HandleEvent(routeEvent, socket)
					socket.Redirect(routeEvent.To)
				}
			case update, ok := <-socket.updates:
				if !ok {
					return
				}
				update.apply(&socket)
				if socket.Err != nil {
					errors <- socket.Err
					return
				}
			}

			newTemplate, newAssets := socket.component.Render(socket)
			newRender := rootNode{DynHTML: newTemplate.(StaticDynamic), UserAssets: newAssets.mergeAndOverwrite(socket.assets())}
			patches := lastRender.Diff(newRender)
			if patches == nil {
				continue
			}

			lastRender = newRender

			select {
			case <-ctx.Done():
				return
			case patchesStream <- *patches:
			}
		}
	}()

	return lastRender, patchesStream, errors
}

type HTML interface{ html() }

type L string

func (L) html() {}

func (StaticDynamic) html() {}

func LiveSocket(newComponent func() LiveComponent) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		upgrader := websocket.Upgrader{}

		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		events := make(chan Event, 1024)

		ctx, canc := context.WithCancel(r.Context())
		errGroup, ctx := errgroup.WithContext(ctx)

		component := newComponent()
		route := r.URL.RawFragment
		initialRender, patchesStream, _ := newPatchesStream(ctx, component, events, route)

		// send mount message
		{
			payload, err := json.Marshal(initialRender)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				canc()
				return
			}

			if err = conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				canc()
				return
			}
		}

		// errGroup.Go(func() error {
		// 	select {
		// 	case <-ctx.Done():
		// 		return ctx.Err()
		// 	case err := <-componentErrors:
		// 		canc()
		// 		log.Println(err)
		// 		return err
		// 	}
		// })

		errGroup.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case patches := <-patchesStream:

					payload, err := json.Marshal(patches)
					if err != nil {
						return err
					}

					err = conn.WriteMessage(websocket.BinaryMessage, payload)
					if err != nil {
						return err
					}
				}
			}
		})

		errGroup.Go(func() error {
			for {
				var msg = map[string]interface{}{}

				err := conn.ReadJSON(&msg)
				if err != nil {
					return err
				}

				var e Event

				if _, ok := msg["to"]; ok { // got redirect event
					e = RouteChangedEvent{
						From: msg["from"].(string),
						To:   msg["to"].(string),
					}
				} else {
					t, ok := msg["name"].(string)
					if !ok {
						continue
					}
					delete(msg, "name")
					e = UserEvent{Name: t, Data: msg}
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case events <- e:
				}
			}
		})

		if err := errGroup.Wait(); err != nil && !websocket.IsUnexpectedCloseError(err) {
			log.Println("errGroup.Error: ", err)
		}
		canc()

		if unmountable, ok := component.(Unmountable); ok {
			unmountable.Unmount()
		}
		close(events)
		conn.Close()

	}
}
