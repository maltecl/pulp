package pulp

import "testing"

var testSource = `
hellodasdasdasd  asdasd
asdsadasdas
{{if post != nil}}
	{{post.title}} - {{post.body}}
{{else}}
	20
{{end}}

adsa


world`

var testSource2 = `
	<input type="text" value="{{t.Username}}" amigo-input="username">{{t.Username}}</input>
	<p>{{t.Age}}</p>
	<button amigo-click="inc"> increment </button>
	<button amigo-click="dec"> decrement </button>
	
	{{if t.Age > 10}}
		<h4>name: {{t.Username}} </h4>
	{{else}}
		hello world
	{{end}}



	{{if t.Age > 10}}
		<h4>name: {{t.Username}} </h4>
	{{else}}
		<p> {{t.Age}} </p>
	{{end}}


	`

func TestLexer(t *testing.T) {

	l := lexer{input: testSource2, state: lexUntilLBrace, tokens: make(chan *token, 2)}

	// go func() {
	// 	for tok := range l.tokens {
	// 		t.Logf("token: %v", tok)
	// 	}
	// }()

	go l.run()

	for tok := range l.tokens {
		t.Logf("next: type(%v)  value: %v", tok.typ, tok.value)
	}

}
