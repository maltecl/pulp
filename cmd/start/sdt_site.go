package main

import "amigo"

var _ amigo.LiveComponent = &TestSite{}

type TestSite struct {
	Username string
	Age      int
}

func (t *TestSite) Mount(socket amigo.Socket, events chan<- amigo.Event) error {
	t.Username = "Donald Duck"
	t.Age = 14
	return nil
}

func (t TestSite) Render() amigo.StaticDynamic {
	return amigo.NewStaticDynamic(
		`<button amigo-click="increment">increment</button> <h2>{}</h2> - <p>{}</p>`,
		t.Username,
		t.Age,
	)
}

func (t *TestSite) HandleEvent(event amigo.Event, changes chan<- amigo.LiveComponent) error {

	// if t.Age%2 == 0 {
	// 	t.Username += ", Donald"
	// }

	switch event.Name {
	case "increment":
		t.Age++
	}

	return nil
}

func (TestSite) Name() string { return "TestSite" }
