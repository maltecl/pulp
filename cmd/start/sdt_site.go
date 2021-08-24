package main

import (
	"pulp"
	"time"
)

var _ pulp.LiveComponent = &TestSite{}

type TestSite struct {
	Username string
	Age      int
	Nested   struct {
		X, Y int
	}
}

func (t *TestSite) Mount(socket pulp.Socket) {
	t.Username = "Donald Duck"
	t.Age = 14

	go func() {
		for range time.NewTicker(time.Second).C {
			t.Nested.X++
			socket.Changes(t).Do()
		}
	}()

	go func() {
		time.Sleep(time.Second / 2)
		for range time.NewTicker(time.Second).C {
			t.Nested.Y--
			socket.Changes(t).Do()
		}
	}()

	socket.Changes(t).Do()
}

func (t *TestSite) HandleEvent(event pulp.Event, socket pulp.Socket) {

	// if t.Age%2 == 0 {
	// 	t.Username += ", Donald"
	// }

	switch event.Name {
	case "increment":
		t.Age++
	case "name_changed":
		t.Username = event.Data["value"]
	case "reset":
		// t.Username = ""

		socket.Errorf("not good :/ ")
	}

	socket.Changes(t).Do()
}

func (t TestSite) Render() pulp.HTML {

	cond0 := len(t.Username) > 5
	arg0 := pulp.IfTemplate{
		Condition: &cond0,
		True: pulp.StaticDynamic{
			Static:  []string{"hello world: ", ""},
			Dynamic: []interface{}{10},
		},
		False: pulp.StaticDynamic{
			Static:  []string{"<span>count:", "</span> // <span>", "</span>"},
			Dynamic: []interface{}{t.Nested.X, t.Nested.Y},
		},
	}

	// arg1 := pulp.ForTemplate{
	// 	Static: []string{"<h3>title: ", "</h3> <h5>body: ", "</h5>"},
	// }

	// arg1.Dynamics = make([][]interface{}, 0)
	// arg1.Dynamics = append(arg1.Dynamics, []interface{}{": )", "good music"})
	// arg1.Dynamics = append(arg1.Dynamics, []interface{}{"duster rocks", "i love duster"})

	return pulp.NewStaticDynamic(
		`text: <h4>{}</h4>
		<button amigo-click="increment">increment</button>
		<button amigo-click="reset">reset</button>

		<input amigo-input="name_changed" value="{}">

		<p>{}</p>

		{}
		{}`,

		t.Username,
		t.Username,
		t.Age,
		arg0,
		// arg1,
	)
}

func (TestSite) Name() string { return "TestSite" }
