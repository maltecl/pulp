package amigo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestDiff(t *testing.T) {

	cases := []struct {
		x, y          StaticDynamic
		diff          []int
		notComparable bool
	}{
		{
			x:    StaticDynamic{Dynamic: []interface{}{"hello"}},
			y:    StaticDynamic{Dynamic: []interface{}{"hello"}},
			diff: []int{},
		},
		{
			x:    StaticDynamic{Dynamic: []interface{}{""}},
			y:    StaticDynamic{Dynamic: []interface{}{"hello"}},
			diff: []int{0},
		},
		{
			x:             StaticDynamic{Dynamic: []interface{}{""}},
			y:             StaticDynamic{Dynamic: []interface{}{"hello", ""}},
			notComparable: true,
		},
	}

	for i, tc := range cases {
		want := tc.diff
		got, err := diff(tc.x, tc.y)

		expectedComparable := (err != nil) == tc.notComparable
		eq := cmp.Equal(got, want)

		if expectedComparable && (tc.notComparable || eq) {
			continue
		}

		errStr := &strings.Builder{}
		fmt.Fprintf(errStr, "test %v: ", i)

		if !expectedComparable {
			fmt.Fprintf(errStr, "tc.notComparable: %v (expected) != %v", tc.notComparable, err != nil)
		}

		if expectedComparable && !eq {
			fmt.Fprintf(errStr, "\n%v (expected) != %v,\ndiff: %s", want, got, cmp.Diff(got, want))
		}

		t.Error(errStr.String())
	}

}
