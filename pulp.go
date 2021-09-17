package pulp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

type LiveComponent interface {
	Mount(Socket)
	Render(Socket) (HTML, Assets) // HTML guranteed to be StaticDynamic after code generation
	HandleEvent(Event, Socket)
	Name() string
}

type UnMountable interface {
	UnMount()
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
		lastState: component,
	}

	fmt.Printf("new socket: %d\n", socketID)

	atomic.AddUint32(&socketID, 1)

	errors := make(chan error)
	patchesStream := make(chan Patches)

	socket.lastState.Mount(socket)

	initalTemplate, initialUserAssets := socket.lastState.Render(socket)
	lastTemplate := initalTemplate.(StaticDynamic)

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
					fmt.Println("event: ", userEvent.Name)
					socket.lastState.HandleEvent(userEvent, socket)
					continue outer
				}

				if routeEvent, ok := event.(RouteChangedEvent); ok {
					socket.lastState.HandleEvent(routeEvent, socket)
					socket.Prepare().Redirect(routeEvent.To).Do()
				}
			case update, ok := <-socket.updates:
				if !ok {
					return
				}
				if socket.Err != nil {
					errors <- socket.Err
					return
				}

				update.apply(&socket)
			}

			fmt.Printf("socket %d render\n", socket.ID)
			newTemplate, newAssets := socket.lastState.Render(socket)
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

type HTML interface{ HTML() }

type L string

func (L) HTML() {}

func (StaticDynamic) HTML() {}

func ServeWebFiles() {
	http.HandleFunc("/bundle/bundle.js", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/bundle/bundle.js")
	})

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/index.html")
	})

	http.HandleFunc("/index.css", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/index.css")
	})

}

// TODO: the api needs to be improved ALOT
func LiveHandler(route string, newComponent func() LiveComponent) {

	http.HandleFunc(filepath.Join(route, "/bundle/bundle.js"), func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/bundle/bundle.js")
	})

	http.HandleFunc(filepath.Join(route, "/"), func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/index.html")
	})

	http.HandleFunc(filepath.Join(route, "/ws"), handler(newComponent))
}

func handler(newComponent func() LiveComponent) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		upgrader := websocket.Upgrader{}

		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		events := make(chan Event, 1024)

		errGroup, ctx := errgroup.WithContext(context.Background())

		component := newComponent()
		route := r.URL.RawFragment
		initialRender, patchesStream, componentErrors := newPatchesStream(ctx, component, events, route)

		// send mount message

		conn.SetCloseHandler(func(code int, text string) error {
			fmt.Println("CLOSED")
			return nil
		})

		payload, err := json.Marshal(initialRender)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		errGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-componentErrors:
				return err
			}
		})

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
						log.Println("expected name in event: ", msg)
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

		if err := errGroup.Wait(); err != nil {
			log.Println("errGroup.Error: ", err)
		}
		log.Println("done with: ", err)

		if unmountable, ok := component.(UnMountable); ok {
			unmountable.UnMount()
		}
		close(events)
		conn.Close()

	}
}
