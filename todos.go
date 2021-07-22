package amigo

// var _ LiveComponent = &TodosComponent{}

// type TodosComponent struct {
// 	Todos map[string]Todo

// 	NewTodoInputValue string
// 	Loading           bool

// 	ShowFlashError bool

// 	templateString string
// }

// type Todo struct {
// 	Title string
// }

// func (t *TodosComponent) Mount(socket Socket, events chan<- Event) error {
// 	t.Todos = map[string]Todo{}
// 	t.Todos[shortid.MustGenerate()] = Todo{"todo 1"}
// 	t.Todos[shortid.MustGenerate()] = Todo{"todo 2"}
// 	t.Todos[shortid.MustGenerate()] = Todo{"todo 1"}

// 	file, err := os.Open("todos.temp.html")
// 	if err != nil {
// 		return err
// 	}

// 	bs, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		return err
// 	}
// 	t.templateString = string(bs)

// 	return nil
// }

// func (t *TodosComponent) HandleEvent(event Event, changes chan<- LiveComponent) error {

// 	switch event.Name {
// 	case "input":
// 		t.NewTodoInputValue = event.Data["value"]
// 		t.ShowFlashError = len(strings.Trim(t.NewTodoInputValue, " \n\t")) > 10

// 	case "submit":
// 		if t.NewTodoInputValue == "" {
// 			break
// 		}

// 		t.Loading = true

// 		go func() {
// 			time.Sleep(time.Second / 8)
// 			t.Loading = false
// 			t.Todos[shortid.MustGenerate()] = Todo{Title: strings.Trim(t.NewTodoInputValue, " \n\t")}
// 			t.NewTodoInputValue = ""
// 			changes <- t
// 		}()

// 		return nil

// 	case "delete":
// 		id := event.Data["value"]
// 		if id == "" {
// 			return fmt.Errorf("empty id")
// 		}
// 		delete(t.Todos, id)
// 	}

// 	return nil
// }

// func (t TodosComponent) Render() (Assets, string) {
// 	assets := Assets{
// 		"messageLength": len(t.NewTodoInputValue),
// 	}

// 	return assets, `
// 	<input amigo-input="input" type="text" value="{{.C.NewTodoInputValue}}"> <span> {{.A.messageLength}} </span>

// 	<button amigo-click="submit" {{if .C.ShowFlashError}} disabled {{end}}> create </button>

// 	</br>

// 	</br> {{if .C.Loading}}

// 	<div class="spinner-border text-danger" role="status">
// 			<span class="visually-hidden">Loading...</span>
// 	</div>

// 	{{end}}

// 	</br>

// 	<ul>
// 			{{ range $index, $todo := .C.Todos}}

// 			<div class="card" style="width: 18rem;">
// 					<ul class="list-group list-group-flush">
// 							<li class="list-group-item">
// 									<button amigo-click="delete" amigo-value="{{$index}}" type="button" class="btn-close" aria-label="Close"></button> {{$todo.Title}}
// 							</li>
// 					</ul>
// 			</div>

// 			<!-- <li> <button amigo-click="delete" amigo-values="{{$index}}"> {{$index}} - {{$todo.Title}} </button> </li> -->
// 			{{end}}
// 	</ul>

// 	{{if .C.ShowFlashError}}

// 	<div class="card text-white bg-danger mb-3  animate__animated animate__fadeOutUp animate__delay-4s" style="max-width: 18rem;">
// 			<div class="card-body">
// 					<p class="card-text">Message is too long</p>
// 			</div>
// 	</div>

// 	{{end}}`
// }

// func (TodosComponent) Name() string {
// 	return "TodosComponent"
// }
