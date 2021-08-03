package amigo

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type StaticDynamic struct {
	Static  []string      `json:"s"`
	Dynamic []interface{} `json:"d"`
}

func NewStaticDynamic(format string, dynamics ...interface{}) StaticDynamic {
	static := strings.Split(format, "{}")
	return StaticDynamic{static, dynamics}
}

func (s StaticDynamic) String() string {
	res := strings.Builder{}

	for i := range s.Static {
		res.WriteString(s.Static[i])

		if ok := i < len(s.Dynamic); ok {

			switch r := s.Dynamic[i].(type) {
			case IfTemplate:
				ifStr := ""

				if *r.Condition {
					ifStr = StaticDynamic{r.StaticTrue, r.Dynamic}.String()
				} else {
					ifStr = StaticDynamic{r.StaticFalse, r.Dynamic}.String()
				}
				res.WriteString(ifStr)

			case ForTemplate:
				forStr := strings.Builder{}

				for _, dynamic := range r.Dynamics {
					fmt.Fprint(&forStr, StaticDynamic{Static: r.Static, Dynamic: dynamic}.String())
				}

				res.WriteString(forStr.String())
			default:
				res.WriteString(fmt.Sprint(s.Dynamic[i]))
			}
		}
	}

	return res.String()
}

func Comparable(sd1, sd2 StaticDynamic) bool {
	return len(sd1.Dynamic) == len(sd2.Dynamic) && len(sd1.Static) == len(sd2.Static)
}

type Patches map[int]interface{}

func Diff(sd1, sd2 StaticDynamic) (*Patches, bool) {
	needsPatch, err := diff(sd1.Dynamic, sd2.Dynamic)
	if err != nil {
		return nil, false
	}

	nonEmptyPatch := len(needsPatch) != 0

	patches := Patches{}
	for _, patchIndex := range needsPatch {
		patch := sd2.Dynamic[patchIndex]

		switch new_ := patch.(type) {
		case IfTemplate:
			old := sd1.Dynamic[patchIndex].(IfTemplate)

			diff := IfTemplate{}
			if *old.Condition != *new_.Condition {
				fmt.Println("was here!")

				diff.Condition = new_.Condition
			}

			if !cmp.Equal(old.Dynamic, new_.Dynamic) {
				// TODO: do diffing here too. maybe make diff() more general (it does not use the static part) and reuse it here
				diff.Dynamic = new_.Dynamic
			}

			patches[patchIndex] = diff

		default:
			_ = new_
			patches[patchIndex] = patch
		}

	}

	return &patches, nonEmptyPatch
}

// TODO applies to many patches?
func diff(d1, d2 []interface{}) ([]int, error) {
	if len(d1) != len(d2) {
		return []int{}, fmt.Errorf(("err 1"))
	}

	ret := make([]int, 0, len(d1))

	for i := 0; i < len(d1); i++ {
		if !cmp.Equal(d1[i], d2[i]) {
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

	cond0 := post != nil
	arg0 := IfTemplate{
		Condition:   &cond0,
		StaticTrue:  PartialStatic("{{post.title}} - {{post.body}}"),
		StaticFalse: PartialStatic(""),
	}

	if *arg0.Condition {
		arg0.Dynamic = []interface{}{post.title, post.body}
	} else {
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
	Condition   *bool         `json:"c,omitempty"`
	StaticTrue  []string      `json:"t,omitempty"`
	StaticFalse []string      `json:"f,omitempty"`
	Dynamic     []interface{} `json:"d,omitempty"`
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
	arg0.Static = PartialStatic("{{post.title}} - {{post.body}}")
	arg0.Dynamics = make([][]interface{}, 0)

	for _, post := range posts {
		arg0.Dynamics = append(arg0.Dynamics, []interface{}{post.title, post.body})
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
	Static   []string        `json:"s"`
	Dynamics [][]interface{} `json:"ds"`
}
