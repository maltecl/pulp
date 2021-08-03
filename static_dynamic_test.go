package amigo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
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
		got, err := diff(tc.x.Dynamic, tc.y.Dynamic)

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

func TestNewStaticDynamic(t *testing.T) {

	cases := []struct {
		static  string
		dynamic []interface{}

		expectedSD     StaticDynamic
		expectedString string
	}{
		{
			static:         "hello {}",
			dynamic:        []interface{}{0},
			expectedSD:     StaticDynamic{Static: []string{"hello ", ""}, Dynamic: []interface{}{0}},
			expectedString: "hello 0",
		},
	}

	for i, tc := range cases {
		got := NewStaticDynamic(tc.static, tc.dynamic...)
		eq := Comparable(got, tc.expectedSD) && cmp.Equal(got, tc.expectedSD)
		eqString := cmp.Equal(tc.expectedString, got.String())

		if eq && eqString {
			continue
		}

		errStr := &strings.Builder{}

		fmt.Fprintf(errStr, "test %v: ", i)

		if !eq {
			fmt.Fprintf(errStr, "!eq: ")
			fmt.Fprintf(errStr, "%v (expected) != %v", pretty.Sprint(tc.expectedSD), pretty.Sprint(got))
			fmt.Fprintf(errStr, "\ndiff:%v", cmp.Diff(tc.expectedSD, got))
		}

		if !eqString {
			fmt.Fprintf(errStr, "!eqString: ")
			fmt.Fprintf(errStr, "%q (expected) != %q", tc.expectedString, got.String())
			fmt.Fprintf(errStr, "\ndiff:%v", cmp.Diff(tc.expectedString, got.String()))
		}

		t.Error(errStr.String())

	}

}
