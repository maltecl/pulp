package main

import "pulp"

type user struct {
	Username string
	Age      int
}

func _() pulp.HTML {
	t := user{Username: "Donald Duck", Age: 34}

	_ = t

	return pulp.L(`
	<input type="text" value="{{t.Username}}" amigo-input="username">{{t.Username}}</input>
	<p>{{t.Age}}</p>
	<button amigo-click="inc"> increment </button>
	<button amigo-click="dec"> decrement </button>
	
	{{if t.Age > 10}}
		<h4>name: {{t.Username}} </h4>
		
		{{if t.Age > 10}}
			<h4>name: {{t.Username}} </h4>
		{{else}}
			<p> {{t.Age}} </p>
		{{end}}
	{{else}}
		hello world
	{{end}}



	
	{{ for i, x := range []int{1,2,3} }}
			<span> {{ i }} - {{ fmt.Sprint(x) }} </span>
	{{ end }}
	
	`)
}