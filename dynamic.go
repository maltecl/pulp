package pulp

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type StaticDynamic struct {
	Static  []string `json:"s"`
	Dynamic Dynamics `json:"d"`
}

func NewStaticDynamic(format string, dynamics ...interface{}) StaticDynamic {
	static := strings.Split(format, "{}")

	if dynamics == nil {
		dynamics = []interface{}{}
	}

	return StaticDynamic{static, dynamics}
}

func Comparable(sd1, sd2 StaticDynamic) bool {
	return len(sd1.Dynamic) == len(sd2.Dynamic) && len(sd1.Static) == len(sd2.Static)
}

// TODO: use this for the initial render over HTTP
func (s StaticDynamic) Render() string {
	res := strings.Builder{}

	for i := range s.Static {
		res.WriteString(s.Static[i])

		if ok := i < len(s.Dynamic); ok {

			switch r := s.Dynamic[i].(type) { // TODO: remove this switch and instead check if the type implements RenderDiff and use that Render method if this is the case

			case For:
				res.WriteString(r.Render())

			// case IfTemplate:
			// 	ifStr := ""

			// 	if *r.Condition {
			// 		ifStr = r.True.Render()
			// 	} else {
			// 		ifStr = r.False.Render()
			// 	}
			// 	res.WriteString(ifStr)

			// case ForTemplate:
			// 	notreached()

			// 	forStr := strings.Builder{}

			// 	for _, dynamic := range r.Dynamics {
			// 		fmt.Fprint(&forStr, StaticDynamic{Static: r.Static, Dynamic: dynamic}.Render())
			// 	}

			// 	res.WriteString(forStr.String())
			default:
				res.WriteString(fmt.Sprint(s.Dynamic[i]))
			}
		}
	}

	return res.String()
}

func notreached() {
	panic("should not be reached")
}

// Patches can point to actual value itself or another layer of Patches
type Patches map[string]interface{}

func (p Patches) IsEmpty() bool {
	return len(map[string]interface{}(p)) == 0
}

type Diffable interface {
	Diff(new interface{}) *Patches
}

type Renderable interface {
	Render() string
}

type RenderDiff interface {
	Diffable
	Renderable
}

// TODO: not quite working yet
func (sd StaticDynamic) Diff(new interface{}) *Patches {
	new_ := new.(StaticDynamic)

	return sd.Dynamic.Diff(new_.Dynamic)
}

// Dynamics can be filled by actual values or itself by other Diffables
type Dynamics []interface{}

func (d Dynamics) Diff(new interface{}) *Patches {
	new_ := new.(Dynamics)

	if len(d) != len(new_) {
		panic(fmt.Errorf("expected equal length in Dynamics"))
	}

	ret := Patches{}

	for i := 0; i < len(d); i++ {

		var key string
		if keyed, ok := d[i].(KeyedSection); ok {
			key = fmt.Sprint(keyed.Key)
		} else {
			key = fmt.Sprint(i)
		}

		if d1Diffable, isDiffable := d[i].(Diffable); isDiffable {
			if diff := d1Diffable.Diff(new_[i]); diff != nil {
				ret[key] = diff
			}
		} else {
			if !cmp.Equal(d[i], new_[i]) {
				ret[key] = new_[i]
			}
		}
	}

	if ret.IsEmpty() { // does this yield the length of keys in the map?
		return nil
	}

	return &ret
}

var _ Diffable = If{}

type If struct {
	Condition bool          `json:"c"`
	True      StaticDynamic `json:"t"`
	False     StaticDynamic `json:"f"`
}

func (old If) Diff(new interface{}) *Patches {
	new_ := new.(If)

	patches := Patches{}

	if old.Condition != new_.Condition {
		patches["c"] = new_.Condition
	}

	// if new_.Condition {
	if trueDiff := old.True.Dynamic.Diff(new_.True.Dynamic); trueDiff != nil {
		patches["t"] = trueDiff
	}
	// } else {
	if falseDiff := old.False.Dynamic.Diff(new_.False.Dynamic); falseDiff != nil {
		patches["f"] = falseDiff
	}
	// }

	if patches.IsEmpty() {
		return nil
	}

	return &patches
}

var _ RenderDiff = For{}

type For struct {
	Statics      []string            `json:"s"`
	ManyDynamics map[string]Dynamics `json:"ds"`
	DiffStrategy `json:"strategy"`
}

type DiffStrategy uint8

const (
	Append DiffStrategy = iota
	Prepend
)

func (f For) Render() string {
	str := strings.Builder{}

	for _, dynamics := range f.ManyDynamics {
		str.WriteString(StaticDynamic{f.Statics, dynamics}.Render())
	}

	return str.String()
}

func (old For) Diff(new interface{}) *Patches {
	new_ := new.(For)

	patches := Patches{}

	for key, newVal := range new_.ManyDynamics {
		if oldVal, notNew := old.ManyDynamics[key]; notNew {
			if diff := oldVal.Diff(newVal); diff != nil {
				patches[key] = diff
			}
		} else {
			patches[key] = newVal
		}
	}

	// TODO
	// if hasNewElements := len(old.ManyDynamics) < len(new_.ManyDynamics); hasNewElements {
	// 	base := len(old.ManyDynamics)
	// 	newElements := new_.ManyDynamics[base:]
	// 	for i, dynamics := range newElements {
	// 		patches[fmt.Sprint(i+base)] = dynamics
	// 	}
	// }

	if patches.IsEmpty() {
		return nil
	}

	return &Patches{
		"ds": patches,
	}
}

type KeyedSection struct {
	Key interface{} `json:"key"`
	StaticDynamic
}
