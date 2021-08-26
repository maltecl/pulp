package main

import (
	"fmt"
	"pulp"
)

type user struct {
	Username string
	Age      int
}

func _() pulp.HTML {
	t := user{Username: "Donald Duck", Age: 34}

	_ = t

	return func() pulp.StaticDynamic {
		x1 := t.Username
		x2 := t.Username
		x3 := t.Age
		x4 := pulp.If{
			Condition: t.Age > 10,
			True: pulp.StaticDynamic{
				Static:  []string{"\n\t\t<h4>name: ", " </h4>\n\t\t\n\t\t", "\n\t"},
				Dynamic: pulp.Dynamics{t.Username, 0xc0000da850},
			},
			False: pulp.StaticDynamic{
				Static:  []string{"\n\t\thello world\n\t"},
				Dynamic: pulp.Dynamics{},
			},
		}
		x5 := pulp.For{
			Statics:      []string{"\n\t\t\t<span> ", " - ", " </span>\n\t"},
			ManyDynamics: make([]pulp.Dynamics, 0),
			DiffStrategy: pulp.Append,
		}

		for i, x := range []int{1, 2, 3} {
			x5.ManyDynamics = append(x5.ManyDynamics, pulp.Dynamics{i, fmt.Sprint(x)})
		}
		x6 := pulp.NewStaticDynamic("`\n\t<input type=\"text\" value=\"{}\" amigo-input=\"username\">{}</input>\n\t<p>{}</p>\n\t<button amigo-click=\"inc\"> increment </button>\n\t<button amigo-click=\"dec\"> decrement </button>\n\t\n\t{}\n\n\n\n\t\n\t", x1, x2, x3, x4, x5)
		return x6
	}()
}
