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

func init() {
	fmt.Println("MARKER1")
}

type LiveComponent interface {
	Mount(Socket)
	Render(Socket) (HTML, Assets) // HTML guranteed to be StaticDynamic after code generation
	HandleEvent(Event, Socket)
	Name() string
}

type UnMountable interface {
	UnMount()
}

type event interface {
	event()
}
type Event struct {
	Name string
	Data map[string]interface{}
}

type routeChangedEvent struct {
	from, to string
}

func (Event) event()             {}
func (routeChangedEvent) event() {}

var socketID = uint32(0)

func New(ctx context.Context, component LiveComponent, events chan event, route string) (rootNode, <-chan Patches, <-chan error) {

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
		<-socket.Done()
		fmt.Printf("socket done %d\n", socket.ID)
	}()

	go func() {

	outer:
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-events:
				if userEvent, ok := event.(Event); ok {
					fmt.Println("event: ", userEvent.Name)
					socket.lastState.HandleEvent(userEvent, socket)
				}

				if routeEvent, ok := event.(routeChangedEvent); ok {
					fmt.Println("new route ", routeEvent) // TODO @router
				}
				continue outer
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

	go func() {
		<-socket.Done()
		close(errors)
		close(patchesStream)
		close(socket.updates)
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

		events := make(chan event, 1024)

		errGroup, ctx := errgroup.WithContext(context.Background())

		component := newComponent()
		route := r.URL.RawFragment
		initialRender, patchesStream, componentErrors := New(ctx, component, events, route)

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

				// event := routeChangedEvent{
				// 	from: msg["from"],
				// 	to:   msg["to"],
				// }

				// TODO: distinguish normal event message from route-changed message

				t := msg["name"].(string)
				delete(msg, "name")

				select {
				case <-ctx.Done():
					return ctx.Err()
				case events <- Event{Name: t, Data: msg}:
				}
			}
		})

		if err := errGroup.Wait(); err != nil {
			log.Println("errGroup.Error: ", err)
		}
		// canc()
		log.Println("done with: ", err)

		if unmountable, ok := component.(UnMountable); ok {
			unmountable.UnMount()
		}
		close(events)
		conn.Close()

	}
}
