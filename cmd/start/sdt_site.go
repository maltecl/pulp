package main

import (
	"amigo"
	"time"
)

var _ amigo.LiveComponent = &TestSite{}

type TestSite struct {
	Username string
	Age      int
	Nested   struct {
		X, Y int
	}
}

func (t *TestSite) Mount(socket amigo.Socket, events chan<- amigo.Event, changes chan<- amigo.LiveComponent) error {
	t.Username = "Donald Duck"
	t.Age = 14

	go func() {
		for range time.NewTicker(time.Second).C {
			t.Nested.X++
			changes <- t
		}
	}()

	go func() {
		time.Sleep(time.Second / 2)
		for range time.NewTicker(time.Second).C {
			t.Nested.Y--
			changes <- t
		}
	}()

	return nil
}

func (t *TestSite) HandleEvent(event amigo.Event, changes chan<- amigo.LiveComponent) error {

	// if t.Age%2 == 0 {
	// 	t.Username += ", Donald"
	// }

	switch event.Name {
	case "increment":
		t.Age++
	case "name_changed":
		t.Username = event.Data["value"]
	case "reset":
		t.Username = ""
	}

	return nil
}

func (t TestSite) Render() amigo.StaticDynamic {

	cond0 := len(t.Username) > 5
	arg0 := amigo.IfTemplate{
		Condition:   &cond0,
		StaticTrue:  []string{"hello world", ""},
		StaticFalse: []string{"<span>count:", "</span> // <span>", "</span>"},
	}

	if *arg0.Condition { // this is cool, as it prevents silly rerenders, when the condition stays the same, but the dynamic value for the other case changes.

		// TODO: use two seperate staticdynamic pairs, so that this 10 is not sent across the wire, everytime the condition flips to true
		arg0.Dynamic = []interface{}{10}
	} else {
		arg0.Dynamic = []interface{}{t.Nested.X, t.Nested.Y}
	}

	arg1 := amigo.ForTemplate{
		Static: []string{"<h3>title: ", "</h3> <h5>body: ", "</h5>"},
	}

	arg1.Dynamics = make([][]interface{}, 0)
	arg1.Dynamics = append(arg1.Dynamics, []interface{}{": )", "good music"})
	arg1.Dynamics = append(arg1.Dynamics, []interface{}{"duster rocks", "i love duster"})

	return amigo.NewStaticDynamic(
		`<button amigo-click="increment">increment</button>
		<button amigo-click="reset">reset</button>

		<input amigo-input="name_changed" value="{}">

		<p>{}</p>

		{}
		{}`,

		t.Username,
		t.Age,
		arg0,
		arg1,
	)
}

func (TestSite) Name() string { return "TestSite" }
