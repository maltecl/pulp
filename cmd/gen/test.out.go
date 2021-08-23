package main

import "pulp"

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
			Static:  []string{"\n\t\t<h4>name: ", " </h4>\n\t"},
			Dynamic: pulp.Dynamics{t.Username},
		},
		False: pulp.StaticDynamic{
			Static:  []string{"\n\t\thello world\n\t"},
			Dynamic: pulp.Dynamics{},
		},
	}
	x5 := pulp.If{
		Condition: t.Age > 10,
		True: pulp.StaticDynamic{
			Static:  []string{"\n\t\t<h4>name: ", " </h4>\n\t"},
			Dynamic: pulp.Dynamics{t.Username},
		},
		False: pulp.StaticDynamic{
			Static:  []string{"\n\t\t<p> ", " </p>\n\t"},
			Dynamic: pulp.Dynamics{t.Age},
		},
	}
	x6 := pulp.NewStaticDynamic("`\n\t<input type=\"text\" value=\"{}\" amigo-input=\"username\">{}</input>\n\t<p>{}</p>\n\t<button amigo-click=\"inc\"> increment </button>\n\t<button amigo-click=\"dec\"> decrement </button>\n\t\n\t{}\n\n\n\n\t" , x1, x2, x3, x4, x5)
	return x6
}()
}
