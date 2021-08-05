package main

import (
	"amigo"
)

var _ amigo.LiveComponent = &Simple1{}

type Simple1 struct {
	Username string
	Age      int
	Nested   struct {
		X, Y int
	}
}

func (t *Simple1) Mount(socket amigo.Socket) {
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

func (t *Simple1) HandleEvent(event amigo.Event, socket amigo.Socket) {

	switch event.Name {
	case "inc":
		t.Age++
	}

	socket.Changes(t).Do()
}

func (t Simple1) Render() amigo.StaticDynamic {
	return amigo.NewStaticDynamic(
		`<h4>text: {}</h4>
		<p>{}</p>
		<button amigo-click="inc"> increment </button>`,
		t.Username,
		t.Age,
	)
}

func (Simple1) Name() string { return "Simple1" }
