package amigo

import (
	"fmt"
	"strings"
)

type StaticDynamic struct {
	Static  []string
	Dynamic []interface{}
}

func Comparable(sd1, sd2 StaticDynamic) bool {
	return len(sd1.Dynamic) == len(sd2.Dynamic) && len(sd1.Static) == len(sd2.Static)
}

func Diff(sd1, sd2 StaticDynamic) *StaticDynamic {
	needsPatch, err := diff(sd1, sd2)
	if err != nil {
		return nil
	}

	sd := StaticDynamic{Dynamic: make([]interface{}, len(needsPatch))}
	for i, patchIndex := range needsPatch {
		sd.Dynamic[i] = sd2.Dynamic[patchIndex]
	}

	return &sd
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

func NewStaticDynamic(format string, values ...interface{}) StaticDynamic {
	static := strings.Split(format, "{}")

	return StaticDynamic{
		Static:  static,
		Dynamic: values,
	}
}
