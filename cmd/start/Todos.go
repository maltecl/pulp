package main

import (
	"amigo"
)

var _ amigo.LiveComponent = &TodoPage{}

type TodoPage struct {
	InputValue string

	todos []todo
}

type todo struct{ title string }

func (t *TodoPage) Mount(socket amigo.Socket) {
	socket.Changes(t).Do()
}

func (t *TodoPage) HandleEvent(event amigo.Event, socket amigo.Socket) {

	switch event.Name {

	case "changed":
		t.InputValue = event.Data["value"]
		socket.Changes(t).Do()

	case "submit":
		t.todos = append(t.todos, todo{t.InputValue})
		t.InputValue = ""
		socket.Changes(t).Do()
	}

}

func (t TodoPage) Render() amigo.StaticDynamic {

	arg0 := amigo.For{
		Statics:      []string{"<li>", "</li>"},
		ManyDynamics: make([]amigo.Dynamics, len(t.todos)),
	}

	for i, todo := range t.todos {
		arg0.ManyDynamics[i] = amigo.Dynamics{todo.title}
	}

	return amigo.NewStaticDynamic(
		`<input amigo-input="changed" type="text" value="{}"> <button amigo-click="submit"> go </button>
		
		</br>
		

		<ul>
		{}

		</ul>
		`,
		t.InputValue,
		arg0,
	)
}

func (TodoPage) Name() string { return "TodoPage" }
