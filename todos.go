package amigo

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/teris-io/shortid"
)

var _ LiveComponent = &TodosComponent{}

type TodosComponent struct {
	Todos map[string]Todo

	NewTodoInputValue string
	Loading           bool

	ShowFlashError bool

	templateString string
}

type Todo struct {
	Title string
}

func (t *TodosComponent) Mount(socket Socket, events chan<- Event) error {
	t.Todos = map[string]Todo{}
	t.Todos[shortid.MustGenerate()] = Todo{"todo 1"}
	t.Todos[shortid.MustGenerate()] = Todo{"todo 2"}
	t.Todos[shortid.MustGenerate()] = Todo{"todo 1"}

	file, err := os.Open("todos.temp.html")
	if err != nil {
		return err
	}

	bs, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	t.templateString = string(bs)

	return nil
}

func (t *TodosComponent) HandleEvent(event Event, changes chan<- LiveComponent) error {

	switch event.Name {
	case "input":
		t.NewTodoInputValue = event.Data["value"]
		t.ShowFlashError = len(strings.Trim(t.NewTodoInputValue, " \n\t")) > 10

	case "submit":
		if t.NewTodoInputValue == "" {
			break
		}

		t.Loading = true

		go func() {
			time.Sleep(time.Second / 8)
			t.Loading = false
			t.Todos[shortid.MustGenerate()] = Todo{Title: strings.Trim(t.NewTodoInputValue, " \n\t")}
			t.NewTodoInputValue = ""
			changes <- t
		}()

		return nil

	case "delete":
		id := event.Data["value"]
		if id == "" {
			return fmt.Errorf("empty id")
		}
		delete(t.Todos, id)
	}

	return nil
}

func (t *TodosComponent) Render() string {
	return t.templateString
}

func (TodosComponent) Name() string {
	return "TodosComponent"
}
