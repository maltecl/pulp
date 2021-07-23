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
	case "append":
		t.Username += ", donald"
	case "reset":
		t.Username = ""
	}

	return nil
}

func (t TestSite) Render() amigo.StaticDynamic {

	cond0 := len(t.Username) > 20
	arg0 := amigo.IfTemplate{
		Condition:   &cond0,
		StaticTrue:  []string{"hello world"},
		StaticFalse: []string{"<h1>count:", "</h1> - <h1>", "</h1>"},
	}

	if *arg0.Condition { // this is cool, as it prevents silly rerenders, when the condition stays the same, but the dynamic value for the other case changes. DONT use two StaticDynamics as the IfTemplate, as this property would be lost
		arg0.Dynamic = []interface{}{}
	} else {
		arg0.Dynamic = []interface{}{t.Nested.X, t.Nested.Y}
	}

	return amigo.NewStaticDynamic(
		`<button amigo-click="increment">increment</button>
		<button amigo-click="append">append</button>
		<button amigo-click="reset">reset</button>

		<h2>{}</h2> - <p>{}</p>

		{}`,

		t.Username,
		t.Age,
		arg0,
	)
}

func (TestSite) Name() string { return "TestSite" }
