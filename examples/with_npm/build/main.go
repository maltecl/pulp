package main

import (
	"log"
	"net/http"
	"time"

	"github.com/amblified/pulp"
)

type index struct {
	msg     string
	seconds int
	counter int
}

func (c *index) Mount(socket pulp.Socket) {
	c.counter = 10

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
	return func() pulp.StaticDynamic {
		x1 := pulp.StaticDynamic{
			Static:  []string{"<h2> ", " </h2><span> ", " </span> seconds passed </br><button :click=\"increment\"> increment </button> <span> you have pressed the button ", " times </span> "},
			Dynamic: pulp.Dynamics{c.msg, c.seconds, c.counter},
		}

		return x1
	}(), nil
}

func (c index) Unmount() {
	log.Println("heading out.")
}

func main() {

	http.HandleFunc("/bundle.js", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/bundle.js")
	})

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "web/index.html")
	})

	http.HandleFunc("/socket", pulp.LiveSocket(func() pulp.LiveComponent {
		return &index{msg: "hello world"}
	}))
	http.ListenAndServe(":4000", nil)
}
