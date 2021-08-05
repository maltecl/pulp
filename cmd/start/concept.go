// +build dont

package main

import "time"

type MyComponent struct{}

func (c MyComponent) Mount(s Socket, newS chan<- Socket) {

}

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


		{{ prepend/append for message := range c.newMessages}}
			 <span> {{ message.From.Name }} </span> <bold> {{ string(message.Body) }} </bold>
		{{ end }}
	`

}




// Patches can point to actual value itself or another layer of Patches
type Patches map[string]interface{}

func (p Patches) IsEmpty() bool {
	return len(map[string]interface{}(p)) == 0
}

type Diffable interface{
	Diff(new interace{}) *Patches
}


// Dynamics can be filled by actual values or itself by other Diffables
type Dynamics []interface{}

var _ = Dynamics{0, 1}

func (d *Dynamics) Diff(new interface{}) *Patches {
	new_ := new.(Dynamics) 


	if len(d1) != len(d2) {
		log.Fatalf("expected equal length in Dynamics")
		return nil
	}

	ret := Patches{}

	for i := 0; i < len(d1); i++ {
		if d1Diffable, isDiffable := d1[i].(Diffable), isDiffable {
			if diff := d1Diffable.Diff(d2[i]); diff != nil {
				ret[fmt.Sprint(i)] = diff
			}
		} else {
			if !cmp.Equal(d1[i], d2[i]) {
				ret[fmt.Sprint(i)] = d2[i]
			}
		}
	}

	if ret.IsEmpty() { // does this yield the length of keys in the map?
		return nil
	}

	return ret
}





