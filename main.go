package amigo

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// var _ io.WriteCloser = &ChannelWriter{}

// func NewChannelWriter() *ChannelWriter {
// 	return &ChannelWriter{
// 		C: make(chan []byte, 1),
// 	}
// }

// type ChannelWriter struct {
// 	C chan []byte
// }

// func (c *ChannelWriter) Close() error {
// 	close(c.C)
// 	return nil
// }

// func (c *ChannelWriter) Write(bytes []byte) (int, error) {
// 	select {
// 	case c.C <- bytes:
// 	default:
// 		return 0, io.ErrClosedPipe
// 	}
// 	return len(bytes), nil
// }

func AmigoMain() {

	http.HandleFunc("/bundle/bundle.js", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/bundle/bundle.js")
	})

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/index.html")
	})

	http.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {

		upgrader := websocket.Upgrader{}

		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		events := make(chan Event)

		ctx, canc := context.WithCancel(context.Background())

		errors := make(chan error)
		renders := New(ctx, &TodosComponent{}, events, errors)

		go func() {
			for render := range renders {
				err := conn.WriteMessage(websocket.BinaryMessage, []byte(render))
				if err != nil {
					errors <- err
				}
			}
		}()

		go func() {
			for {
				var msg = map[string]string{}

				err := conn.ReadJSON(&msg)
				if err != nil {
					errors <- err
					continue
				}

				fmt.Println(msg)

				t := msg["type"]
				delete(msg, "type")

				events <- Event{
					Name: t,
					Data: msg,
				}
			}

		}()

		// h.Mount(Socket{})

		// time.Sleep(time.Second)
		// events <- Event("increment")
		// time.Sleep(time.Second)
		// events <- Event("increment")
		// time.Sleep(time.Second)
		// events <- Event("decrement")

		fmt.Printf("connection error: %v", <-errors)
		canc()
		conn.Close()
		close(events)

	})

	fmt.Printf("terminated with: %v", http.ListenAndServe(":8080", nil))

}

// var addr = flag.String("addr", ":8080", "http service address")

// func serveHome(w http.ResponseWriter, r *http.Request) {
// 	log.Println(r.URL)
// 	if r.URL.Path != "/" {
// 		http.Error(w, "Not found", http.StatusNotFound)
// 		return
// 	}
// 	if r.Method != "GET" {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	http.ServeFile(w, r, "home.html")
// }

// func main() {
// 	flag.Parse()
// 	hub := newHub()
// 	go hub.run()

// 	http.HandleFunc("/", serveHome)
// 	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
// 		serveWs(hub, w, r)
// 	})
// 	err := http.ListenAndServe(*addr, nil)
// 	if err != nil {
// 		log.Fatal("ListenAndServe: ", err)
// 	}
// }
