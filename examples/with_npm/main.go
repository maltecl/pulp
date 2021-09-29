package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/maltecl/pulp"
)

type index struct {
	msg     string
	seconds int
	counter int
}

func (c *index) Mount(socket pulp.Socket) {
	c.counter = 10

	// If you keep reference to the socket in another go-routine, make sure to shutdown that go-routine once the socket is done.
	// Otherwise this will leak the socket and end up crashing your app
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.seconds++
				socket.Update()
			case <-socket.Done():
				return
			}
		}
	}()
}

func (c *index) HandleEvent(event pulp.Event, socket pulp.Socket) {

	if _, ok := event.(pulp.RouteChangedEvent); ok {
		return
	}

	e := event.(pulp.UserEvent)

	switch e.Name {
	case "increment":
		c.counter++
		socket.Update()
	}

}

func (c index) Render(pulp.Socket) (pulp.HTML, pulp.Assets) {
	return pulp.L(`
		<h2> {{ c.msg }} </h2>
		<span> {{ c.seconds }} </span> seconds passed </br>
		<button :click="increment"> increment </button> <span> you have pressed the button {{ c.counter }} times </span> 
	`), nil
}

func (c index) Unmount() {
	log.Println("heading out.")
}

func main() {

	http.HandleFunc("/socket", pulp.LiveSocket(func() pulp.LiveComponent {
		return &index{msg: "hello world"}
	}))
	// serve your html however you like
	http.HandleFunc("/bundle.js", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/bundle.js")
	})

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/index.html")
	})

	fmt.Println("listening on localhost:4000")
	http.ListenAndServe(":4000", nil)
}
