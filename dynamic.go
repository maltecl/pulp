package pulp

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type rootNode struct {
	DynHTML    StaticDynamic `json:"html"`
	UserAssets Assets        `json:"assets"`
}

func (r rootNode) Diff(new_ interface{}) *Patches {
	new := new_.(rootNode)

	assetsPatches := r.UserAssets.Diff(new.UserAssets)
	htmlPatches := r.DynHTML.Diff(new.DynHTML)
	if assetsPatches == nil && htmlPatches == nil {
		return nil
	}

	patches := Patches{}
	if assetsPatches != nil {
		patches["assets"] = assetsPatches
	}

	if htmlPatches != nil {
		patches["html"] = htmlPatches
	}

	return &patches
}

func (old Assets) Diff(new_ interface{}) *Patches {
	new := new_.(Assets)

	patches := Patches{}

	for key, val := range new {
		if oldValue, isOld := old[key]; isOld {
			if oldValue != val {
				patches[key] = val
			}
		} else {
			patches[key] = val // new value, push the whole state
		}
	}

	for key := range old {
		if _, ok := new[key]; !ok {
			patches[key] = nil // deleted value, push nil
			fmt.Println("DIFF: NIL")
		}
	}

	if patches.IsEmpty() {
		return nil
	}

	return &patches
}

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

func notreached() {
	panic("should not be reached")
}

// Patches can point to actual value itself or another layer of Patches
type Patches map[string]interface{}

func (p Patches) IsEmpty() bool {
	return len(map[string]interface{}(p)) == 0
}

type Diffable interface {
	Diff(new_ interface{}) *Patches
}

// TODO: not quite working yet
func (sd StaticDynamic) Diff(new_ interface{}) *Patches {
	new := new_.(StaticDynamic)

	return sd.Dynamic.Diff(new.Dynamic)
}

// Dynamics can be filled by actual values or itself by other Diffables
type Dynamics []interface{}

func (d Dynamics) Diff(new_ interface{}) *Patches {
	new := new_.(Dynamics)

	if len(d) != len(new) {
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
			if diff := d1Diffable.Diff(new[i]); diff != nil {
				ret[key] = diff
			}
		} else {
			if !cmp.Equal(d[i], new[i]) {
				ret[key] = new[i]
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

func (old If) Diff(new_ interface{}) *Patches {
	new := new_.(If)

	patches := Patches{}

	if old.Condition != new.Condition {
		patches["c"] = new.Condition
	}

	// if new.Condition {
	if trueDiff := old.True.Dynamic.Diff(new.True.Dynamic); trueDiff != nil {
		patches["t"] = trueDiff
	}
	// } else {
	if falseDiff := old.False.Dynamic.Diff(new.False.Dynamic); falseDiff != nil {
		patches["f"] = falseDiff
	}
	// }

	if patches.IsEmpty() {
		return nil
	}

	return &patches
}

type For struct {
	// keyOrder []string

	Statics      []string            `json:"s"`
	ManyDynamics map[string]Dynamics `json:"ds"`
	// DiffStrategy `json:"strategy"`
}

// for some reason the elements are already rendered in the right order... so this is not needed, it seems
// DiffStrategy is the strategy used for when a new node is pushed (i.e. the key of that node was unknown to the client)
// The problem this (for now) solves is, that when a new node is pushed, the client does not know to which position in the
// array this node belongs.
// DiffStrategy allows for assumptions about that.
// type DiffStrategy uint8

// // the order here is reflected in types.js
// const (
// 	// When a new node is pushed, display it after all other nodes (as the last element)
// 	Append DiffStrategy = iota

// 	// ... display it before all other nodes (as the first element)
// 	Prepend

// 	// ... also diff&patch the keyOrder, making sure everything is displayed in the proper order (as it was pushed into ManyDynamics)
// 	Intuitiv
// )

func (old For) Diff(new_ interface{}) *Patches {
	new := new_.(For)

	patches := Patches{}

	for key, val := range new.ManyDynamics {
		if oldVal, ok := old.ManyDynamics[key]; ok {
			if diff := oldVal.Diff(val); diff != nil {
				patches[key] = diff // old value, push the diff
			}
		} else {
			patches[key] = val // new value, push the whole state
		}
	}

	for key := range old.ManyDynamics {
		if _, ok := new.ManyDynamics[key]; !ok {
			patches[key] = nil // deleted value, push nil
			fmt.Println("DIFF: NIL")
		}
	}

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
