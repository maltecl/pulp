package main

import (
	"amigo"
	"fmt"
)

func init() {
	amigo.MustHaveValidTemplate(&amigo.TodosComponent{})
}

func main() {

	sd := amigo.StaticDynamic{}
	fmt.Println(sd)

	amigo.AmigoMain()
}
