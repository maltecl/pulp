package main

import (
	"amigo"
)

var _ amigo.LiveComponent = &Simple2{}

type Simple2 struct {
	Username string
	Age      int
	Nested   struct {
		X, Y int
	}
}

func (t *Simple2) Mount(socket amigo.Socket) {
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

func (t *Simple2) HandleEvent(event amigo.Event, socket amigo.Socket) {

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

func (t Simple2) Render() amigo.StaticDynamic {

	var arg0 interface{} = amigo.If{
		Condition: t.Age > 10,
		True: amigo.StaticDynamic{
			Static:  []string{"<h4>name: ", "</h4>"},
			Dynamic: amigo.Dynamics{t.Username},
		},
		False: amigo.StaticDynamic{
			Static: []string{"hello world"},
		},
	}

	var arg1 interface{} = amigo.If{
		Condition: t.Age > 10,
		True: amigo.StaticDynamic{
			Static:  []string{"<h4>name: ", "</h4>"},
			Dynamic: amigo.Dynamics{t.Username},
		},
		False: amigo.StaticDynamic{
			Static:  []string{"<p>", "</p>"},
			Dynamic: amigo.Dynamics{t.Age},
		},
	}

	return amigo.NewStaticDynamic(
		`
		<input type="text" value="{}" amigo-input="username">{}</input>
		<p>{}</p>
		<button amigo-click="inc"> increment </button>
		<button amigo-click="dec"> decrement </button>
		{}
		{}`,
		t.Username,
		t.Username,
		t.Age,
		arg0,
		arg1,
	)
}

func (Simple2) Name() string { return "Simple2" }
