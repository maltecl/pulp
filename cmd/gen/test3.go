package main

import "pulp"

var _ pulp.LiveComponent = &Simple3{}

type Simple3 struct {
	Username string
	Age      int
	Nested   struct {
		X, Y int
	}
}

func (t *Simple3) Mount(socket pulp.Socket) {
	t.Username = "Donald Duck"
	t.Age = 14

	// go func() {
	// 	for range time.NewTicker(time.Second).C {
	// 		t.Age++
	// 		socket.Changes(t).Do()
	// 	}
	// }()

	socket.Changes(t).Do()
}

func (t *Simple3) HandleEvent(event pulp.Event, socket pulp.Socket) {

	switch event.Name {
	case "inc":
		t.Age++
	case "dec":
		t.Age--
	case "username":
		t.Username = event.Data["value"]
	}

	socket.Changes(t).Do()
}

func (t Simple3) Render() pulp.HTML {
	return pulp.L(`
	<input type="text" value="{{ t.Username }}" amigo-input="username">{}</input>
	<p>{{ t.Username }}</p>
	<button amigo-click="inc"> increment </button>
	<button amigo-click="dec"> decrement </button>
	

	{{ if t.Age > 10 }} 
		age > 10, wait for it.. 

		{{ if t.Age > 15 }}
			<h4>name: {{ t.Username }} </h4>
		{{ else }}
			hello world
		{{ end }}

	{{ else }}
		<p> {{ t.Age }} </p>
	{{ end }}
	`)
}

func (Simple3) Name() string { return "Simple3" }
