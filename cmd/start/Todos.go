package main

import "pulp"

var _ pulp.LiveComponent = &TodoPage{}

type TodoPage struct {
	InputValue string

	todos []todo
}

type todo struct{ title string }

func (t *TodoPage) Mount(socket pulp.Socket) {
	socket.Changes(t).Do()
}

func (t *TodoPage) HandleEvent(event pulp.Event, socket pulp.Socket) {

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

func (t TodoPage) Render() pulp.StaticDynamic {

	arg0 := pulp.For{
		Statics:      []string{"<li>", "</li>"},
		ManyDynamics: make([]pulp.Dynamics, len(t.todos)),
	}

	for i, todo := range t.todos {
		arg0.ManyDynamics[i] = pulp.Dynamics{todo.title}
	}

	return pulp.NewStaticDynamic(
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
