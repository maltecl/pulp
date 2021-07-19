package amigo

import (
	"fmt"
	"time"
)

var _ LiveComponent = &Home{}

type Home struct {
	Counter          int
	InputValue       string
	ShowFlashMessage bool
}

func (h *Home) Mount(socket Socket, events chan<- Event) error {
	h.Counter = 4
	h.InputValue = "hellow orld"
	h.ShowFlashMessage = len(h.InputValue) > 10

	go func() {
		for range time.NewTicker(time.Second).C {
			events <- Event{
				Name: "ticker",
			}
		}
	}()

	return nil
}

func (h *Home) HandleEvent(event Event, changes chan<- LiveComponent) error {
	switch event.Name {
	case "increment":
		h.Counter += 10
	case "decrement":
		h.Counter--
		if h.Counter < 0 {
			h.Counter = 0
		}
	case "input":
		h.InputValue = event.Data["value"]
		h.ShowFlashMessage = len(h.InputValue) > 10
	case "ticker":
		h.Counter++

		if h.Counter > 16 {
			h.Counter = 0
		}
	}

	return nil
}

func (h *Home) Render2() StaticDynamic {

	return NewStaticDynamic(
		`<button amigo-click="increment"> increment </button>
			{}
		<button amigo-click="decrement"> decrement </button>`,
		h.Counter,
	)

	// return StaticDynamic{
	// 	Static: []string{
	// 		`<button amigo-click="increment"> increment </button>`,
	// 		`<button amigo-click="decrement"> decrement </button>`,
	// 	},
	// 	Dynamic: []interface{}{
	// 		h.Counter,
	// 	},
	// }
}

func (h *Home) Render() string {
	length := len(h.InputValue)

	return fmt.Sprintf(`
		<button amigo-click="increment"> increment </button>
		{{.Counter}}
		<button amigo-click="decrement"> decrement </button>


		<input
			amigo-input="input"
			value="{{.InputValue}}"
			type="text" >
	
		<span>length: %d </span>
	
		{{if .ShowFlashMessage}}
			</br>
			<div id="flash" style="background-color: red"> message is too long </div>
		{{end}}
	`, length)

}

func (h Home) Name() string {
	return "home"
}
