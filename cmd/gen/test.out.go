package main

import "pulp"

var _ pulp.LiveComponent = &Simple3{}

type Simple3 struct {
	Username string
	Age      int
	Nested   struct {
		X, Y int
	}
}

func (t *Simple3) Mount(socket pulp.Socket) {
	t.Username = "Donald Duck"
	t.Age = 14

	socket.Changes(t).Do()
}

func (t *Simple3) HandleEvent(event pulp.Event, socket pulp.Socket) {

	switch event.Name {
	case "inc":
		t.Age++
	case "dec":
		t.Age--
	case "username":
		t.Username = event.Data["value"]
	}

	socket.Changes(t).Do()
}

func (t Simple3) Render() pulp.HTML {
	return func() pulp.StaticDynamic {
	x1 := t.Username
x2 := t.Username
x3 := pulp.If{
		Condition:  t.Age > 15 ,
		True: pulp.StaticDynamic{
			Static:  []string{"\n\t\t\t<h4>name: ", " </h4>\n\t\t"},
			Dynamic: pulp.Dynamics{t.Username, },
		},
		False: pulp.StaticDynamic{
			Static:  []string{"\n\t\t\thello world\n\t\t"},
			Dynamic: pulp.Dynamics{},
		},
	}
	x4 := pulp.If{
		Condition:  t.Age > 10 ,
		True: pulp.StaticDynamic{
			Static:  []string{" \n\t\tage > 10, wait for it.. \n\n\t\t", "\n\n\t"},
			Dynamic: pulp.Dynamics{x3, },
		},
		False: pulp.StaticDynamic{
			Static:  []string{"\n\t\t<p> ", " </p>\n\t"},
			Dynamic: pulp.Dynamics{t.Age, },
		},
	}
	x5 := pulp.NewStaticDynamic("`\n\t<input type=\"text\" value=\"{}\" amigo-input=\"username\">{}</input>\n\t<p>{}</p>\n\t<button amigo-click=\"inc\"> increment </button>\n\t<button amigo-click=\"dec\"> decrement </button>\n\t\n\n\t" , x1, x2, x4)
	return x5
}()
}

func (Simple3) Name() string { return "Simple3" }
