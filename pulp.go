package pulp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kr/pretty"
)

type LiveComponent interface {
	Mount(Socket)
	Render() HTML // guranteed to be StaticDynamic after code generation
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
				component.HandleEvent(event, socket)
				continue outer
			case socket = <-socket.updates:
				if socket.Err != nil {
					errors <- socket.Err
				}
				component = socket.lastState
			}

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

func L(source string) LL { return LL{source} }

type LL struct{ Out string }

func (LL) HTML() {}

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
func LiveHandler(route string, component LiveComponent) {
	fmt.Println("was here")

	http.HandleFunc(route+"/bundle/bundle.js", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("hit1")
		http.ServeFile(rw, r, "web/bundle/bundle.js")
	})

	http.HandleFunc(route+"/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("hit2")
		http.ServeFile(rw, r, "web/index.html")
	})

	http.HandleFunc(route+"/ws", handler(component))
}

func handler(component LiveComponent) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("hit3")
		errors := make(chan error, 2)

		upgrader := websocket.Upgrader{}

		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		events := make(chan Event)
		onMount := make(chan StaticDynamic)

		ctx, canc := context.WithCancel(context.Background())

		patchesStream := New(ctx, component, events, errors, onMount)

		// send mount message

		{

			mountedWith := <-onMount
			fmt.Println("mounted: ", mountedWith.String())

			payload, err := json.Marshal(mountedWith)

			fmt.Println("payload: ", string(payload))

			if err != nil {
				errors <- err
			}

			err = conn.WriteMessage(websocket.BinaryMessage, payload)
			if err != nil {
				errors <- err
			}
			close(onMount)
		}

		go func() {
			for patches := range patchesStream {
				pretty.Println(patches)

				payload, err := json.Marshal(patches)
				if err != nil {
					errors <- err
				}

				err = conn.WriteMessage(websocket.BinaryMessage, payload)
				if err != nil {
					errors <- err
				}

			}
			errors <- nil
		}()

		go func() {
			for {
				var msg = map[string]string{}

				err := conn.ReadJSON(&msg)
				if err != nil {
					select {
					case errors <- err:
					case <-ctx.Done():
						return
					}
				}

				fmt.Println(msg)

				t := msg["type"]
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
		conn.Close()
		close(events)

		close(errors)
	}
}
