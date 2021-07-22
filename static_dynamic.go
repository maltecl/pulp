package amigo

import (
	"fmt"
	"strings"
)

type StaticDynamic struct {
	Static  []string      `json:"s"`
	Dynamic []interface{} `json:"d"`
}

func NewStaticDynamic(format string, values ...interface{}) StaticDynamic {
	static := strings.Split(format, "{}")

	return StaticDynamic{
		Static:  static,
		Dynamic: values,
	}
}

func (s StaticDynamic) String() string {
	res := strings.Builder{}

	for i := range s.Static {
		res.WriteString(s.Static[i])

		if ok := i < len(s.Dynamic); ok {
			res.WriteString(fmt.Sprint(s.Dynamic[i]))
		}
	}

	return res.String()
}

func Comparable(sd1, sd2 StaticDynamic) bool {
	return len(sd1.Dynamic) == len(sd2.Dynamic) && len(sd1.Static) == len(sd2.Static)
}

type Patches map[int]interface{}

func Diff(sd1, sd2 StaticDynamic) *Patches {
	needsPatch, err := diff(sd1, sd2)
	if err != nil {
		return nil
	}

	patches := Patches{}
	for _, patchIndex := range needsPatch {
		patches[patchIndex] = sd2.Dynamic[patchIndex]
	}

	return &patches
}

func diff(sd1, sd2 StaticDynamic) ([]int, error) {
	if !Comparable(sd1, sd2) {
		return []int{}, fmt.Errorf(("err 1"))
	}

	ret := make([]int, 0, len(sd1.Dynamic))

	for i := 0; i < len(sd1.Dynamic); i++ {
		if sd1.Dynamic[i] != sd2.Dynamic[i] {
			ret = append(ret, i)
		}
	}

	return ret, nil
}

// this is here for the typechecker,
// every call to this function will be replaced by generated code
func L() StaticDynamic {
	return StaticDynamic{}
}

// notes:

func _() string {

	var arg0 interface{}
	length := 10

	if length > 10 {
		arg0 = 10
	} else {
		arg0 = length
	}

	_ = StaticDynamic{
		Static:  []string{"hello", "world"},
		Dynamic: []interface{}{arg0},
	}

	// to be eq

	return `
		hello 
		{{if length > 10 then 10 else length}}

		world
	`
}

func _() string {

	post := &struct {
		title, body string
	}{
		title: "new post",
		body:  "post body",
	}

	var arg0 interface{}

	if post != nil {
		arg0 = fmt.Sprint(post.title) + " - " + fmt.Sprint(post.body) // variant without deep compare
	} else {
		arg0 = "10"
	}

	_ = StaticDynamic{
		Static:  []string{"hello", "world"},
		Dynamic: []interface{}{arg0},
	}

	// to be eq

	return `
		hello 
		{{if post != nil}}
			{{post.title}} - {{post.body}}
		{{else}}
			10
		{{end}}

		world
	`
}

func _() string {

	post := &struct {
		title, body string
	}{
		title: "new post",
		body:  "post body",
	}

	arg0 := IfTemplate{}
	arg0.Condition = post != nil

	if arg0.Condition {
		arg0.Static = PartialStatic("{{post.title}} - {{post.body}}")
		arg0.Dynamic = []interface{}{post.title, post.body}
	} else {
		arg0.Static = PartialStatic("")
		arg0.Dynamic = []interface{}{}
	}

	_ = StaticDynamic{
		Static:  []string{"hello", "world"},
		Dynamic: []interface{}{arg0},
	}

	// to be eq

	return `
		hello 
		{{if post != nil}}
			{{post.title}} - {{post.body}}
		{{else}}
			10
		{{end}}

		world
	`
}

type IfTemplate struct {
	Condition bool
	StaticDynamic
}

func _() string {

	posts := []struct {
		title, body string
	}{
		{
			title: "new post",
			body:  "post body",
		},
	}

	var arg0 interface{}

	str0 := ""
	for _, post := range posts {
		str0 += fmt.Sprint(post.title) + " - " + fmt.Sprint(post.body) // variant without deep compare
	}
	arg0 = str0

	_ = StaticDynamic{
		Static:  []string{"hello", "world"},
		Dynamic: []interface{}{arg0},
	}

	// to be eq

	return `
		hello 

		{{for _, post := range posts}}
			{{post.title}} - {{post.body}}
		{{end}}

		world
	`
}

func PartialStatic(string) []string {
	return []string{}
}

// deep compare version of for
func _() string {

	posts := []struct {
		title, body string
	}{
		{
			title: "new post",
			body:  "post body",
		},
	}

	arg0 := ForTemplate{}
	arg0.template = PartialStatic("{{post.title}} - {{post.body}}")
	arg0.dynamic = make([][]interface{}, 0)

	for _, post := range posts {
		arg0.dynamic = append(arg0.dynamic, []interface{}{post.title, post.body})
	}

	_ = StaticDynamic{
		Static:  []string{"hello", "world"},
		Dynamic: []interface{}{arg0},
	}

	// to be eq

	return `
		hello 

		{{for _, post := range posts}}
			{{post.title}} - {{post.body}}
		{{end}}

		world
	`
}

// for deep compare?
type ForTemplate struct {
	template []string
	dynamic  [][]interface{}
}
