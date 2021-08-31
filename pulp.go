package pulp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/kr/pretty"
)

func init() {
	fmt.Println("MARKER1")
}

type LiveComponent interface {
	Mount(Socket)
	Render() HTML // guranteed to be StaticDynamic after code generation
	HandleEvent(Event, Socket)
	Name() string
}

type UnMountable interface {
	UnMount()
}

type Event struct {
	Name string
	Data map[string]interface{}
}

var socketID = uint32(0)

func New(ctx context.Context, component LiveComponent, events chan Event, errors chan<- error, onMount chan<- StaticDynamic) <-chan Patches {

	socket := Socket{Context: ctx, updates: make(chan Socket), events: events, ID: socketID}
	fmt.Printf("new socket: %d\n", socketID)

	atomic.AddUint32(&socketID, 1)
	patchesStream := make(chan Patches)

	component.Mount(socket)
	socket = <-socket.updates
	if socket.Err != nil {
		errors <- socket.Err
		return nil
	}

	lastRender := component.Render().(StaticDynamic)
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
				fmt.Printf("socket %d event %v\n", socket.ID, pretty.Sprint(event))
				component.HandleEvent(event, socket)
				continue outer
			case socket = <-socket.updates:
				if socket.Err != nil {
					errors <- socket.Err
					return
				}

				// fmt.Printf("socket %v got updates: %v\n", socket.ID, pretty.Sprint(socket.lastState))

				component = socket.lastState
			}

			fmt.Printf("socket %d render\n", socket.ID)
			newRender := component.Render().(StaticDynamic)
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
		errors := make(chan error, 2)

		upgrader := websocket.Upgrader{}

		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		events := make(chan Event, 1024)
		onMount := make(chan StaticDynamic)

		ctx, canc := context.WithCancel(context.Background())

		component := newComponent()
		patchesStream := New(ctx, component, events, errors, onMount)

		// send mount message

		{
			payload, err := json.Marshal(<-onMount)
			if err != nil {
				errors <- err
			}

			err = conn.WriteMessage(websocket.BinaryMessage, payload)
			if err != nil {
				errors <- err
			}
			close(onMount)
		}

		wg := &sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			for patches := range patchesStream {

				payload, err := json.Marshal(patches)
				if err != nil {
					select {
					case errors <- err:
					case <-ctx.Done():
						return
					}
				}

				err = conn.WriteMessage(websocket.BinaryMessage, payload)
				if err != nil {
					select {
					case errors <- err:
					case <-ctx.Done():
						return
					}
				}

			}
			select {
			case errors <- nil:
			case <-ctx.Done():
				return
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				var msg = map[string]interface{}{}

				err := conn.ReadJSON(&msg)
				if err != nil {
					select {
					case errors <- err:
						return
					case <-ctx.Done():
						return
					}
				}

				// fmt.Println(msg)

				t := msg["type"].(string)
				delete(msg, "type")

				select {
				case <-ctx.Done():
					return
				case events <- Event{Name: t, Data: msg}:
				}
			}

		}()

		fmt.Printf("connection error: %v", <-errors)
		canc()

		if unmountable, ok := component.(UnMountable); ok {
			unmountable.UnMount()
		}
		conn.Close()

		wg.Wait()
		close(events)
		close(errors)
	}
}
