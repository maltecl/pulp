// +build dont

package main

import "time"

func (c MyComponent) HandleEvent(e Event, s Socket, newS chan<- Socket) {
	if e.Name == "submitted" {
		data := e.Params

		go func() {
			time.Sleep(time.Second)

			newS <- s.FlashError("message is to long").Changes(c)
		}()

		return

	}

	newS <- s.Error("weird event")
}

// markdown editor
func (c MyComponent) Render() string {

	return `
	
		{{ render counter.Component{
				Count: len(c.todos),
				Color: counter.Black,
			} 
		}}


		<span> {{if len(c.todos) > 10 then "message too long" else "okay"}} </span>



		{{// the document goes here}}
		{{ for index, line := range strings.Split(string(c.raw), "\n") }}
			<span amigo-id="{{ index }}"" class="document-line">
				{{ index }} :- {{ line }}
			</span>
		{{ end }}
	`

}
