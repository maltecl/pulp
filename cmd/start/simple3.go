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

func (t Simple3) Render() pulp.StaticDynamic {

	var arg0 interface{} = pulp.If{
		Condition: t.Age > 15,
		True: pulp.StaticDynamic{
			Static:  []string{"<h4>name: ", "</h4>"},
			Dynamic: pulp.Dynamics{t.Username},
		},
		False: pulp.StaticDynamic{
			Static: []string{"hello world"},
		},
	}

	var arg1 interface{} = pulp.If{
		Condition: t.Age > 10,
		True: pulp.StaticDynamic{
			Static:  []string{"age > 10, wait for it..", ""},
			Dynamic: pulp.Dynamics{arg0},
		},
		False: pulp.StaticDynamic{
			Static:  []string{"<p>", "</p>"},
			Dynamic: pulp.Dynamics{t.Age},
		},
	}

	return pulp.NewStaticDynamic(
		`
		<input type="text" value="{}" amigo-input="username">{}</input>
		<p>{}</p>
		<button amigo-click="inc"> increment </button>
		<button amigo-click="dec"> decrement </button>
		{}`,
		t.Username,
		t.Username,
		t.Age,
		arg1,
	)
}

func (Simple3) Name() string { return "Simple3" }
