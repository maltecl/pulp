package main

import (
	"amigo"
	"fmt"
	"html/template"
	"os"

	"github.com/kr/pretty"
)

func init() {

	pretty.Println(amigo.Diff(amigo.StaticDynamic{Dynamic: []interface{}{""}}, amigo.StaticDynamic{Dynamic: []interface{}{"hello"}}))
	os.Exit(1)

	tt, err := template.New("test").Parse(`
		<p> {{.Assets.FromMap}} </p>

		<p> {{.In.FromStruct}} </p>
	`)
	fmt.Println(err)

	type in interface{}

	type comp struct {
		FromStruct string
	}

	data := struct {
		Assets map[string]interface{}
		In     in
	}{
		Assets: map[string]interface{}{
			"FromMap": "hello from map",
		},
		In: comp{
			FromStruct: "hello from struct",
		},
	}

	err = tt.Execute(os.Stdout, data)
	fmt.Println(err)

	os.Exit(1)
}

func main() {
	amigo.AmigoMain()
}
