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

type Assets interface{}

type LiveComponent interface {
	Mount(Socket)
	Render() HTML // HTML guranteed to be StaticDynamic after code generation
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

func New(ctx context.Context, component LiveComponent, events chan Event) (*StaticDynamic, <-chan Patches, <-chan error) {

	socket := Socket{Context: ctx, updates: make(chan LiveComponent, 10), events: events, ID: socketID}
	fmt.Printf("new socket: %d\n", socketID)

	atomic.AddUint32(&socketID, 1)

	errors := make(chan error)
	patchesStream := make(chan Patches)

	component.Mount(socket)

	initialRender := component.Render().(StaticDynamic)
	lastRender := initialRender
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
				component.HandleEvent(event, socket)
				continue outer
			case newState, ok := <-socket.updates:
				if !ok {
					return
				}
				if socket.Err != nil {
					errors <- socket.Err
					return
				}

				component = newState
			}

			fmt.Printf("socket %d render\n", socket.ID)
			newRender := component.Render().(StaticDynamic)
			patches := lastRender.Dynamic.Diff(newRender.Dynamic)
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

	return &initialRender, patchesStream, errors
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

var i = 0

func handler(newComponent func() LiveComponent) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := i
		i++

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
		initialRender, patchesStream, componentErrors := New(ctx, component, events)

		// send mount message

		conn.SetCloseHandler(func(code int, text string) error {
			fmt.Println("CLOSED")
			return nil
		})

		payload, err := json.Marshal(*initialRender)
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

				t := msg["name"].(string)
				delete(msg, "name")

				select {
				case <-ctx.Done():
					return ctx.Err()
				case events <- Event{Name: t, Data: msg}:
				}
			}
		})

		go func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("outer socket panic: %d\n", id)
				}
			}()

		}()

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
