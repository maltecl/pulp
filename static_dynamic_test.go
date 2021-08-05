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
		from, to   Diffable
		patches    *Patches
		exptectErr bool
	}{
		{
			from: Dynamics{"hello"},
			to:   Dynamics{"hello"},
		},
		{
			from:    Dynamics{""},
			to:      Dynamics{"hello"},
			patches: &Patches{"0": "hello"},
		},
		{
			from:       Dynamics{""},
			to:         Dynamics{"hello", ""},
			exptectErr: true,
		},
		// if
		{
			from: If{True: StaticDynamic{Dynamic: Dynamics{"hello"}}},
			to:   If{True: StaticDynamic{Dynamic: Dynamics{"hello"}}},
		},
		{
			from:    If{True: StaticDynamic{Dynamic: Dynamics{"hello"}}},
			to:      If{Condition: true, True: StaticDynamic{Dynamic: Dynamics{"hello"}}},
			patches: &Patches{"c": true},
		},
		{
			from:    If{True: StaticDynamic{Dynamic: Dynamics{"hello"}}},
			to:      If{Condition: true, True: StaticDynamic{Dynamic: Dynamics{""}}},
			patches: &Patches{"c": true, "t": &Patches{"0": ""}},
		},
	}

	for i, tc := range cases {

		var (
			want = tc.patches
			got  *Patches
			err  error
		)

		errCatcher := func() {
			if mErr, isErr := recover().(error); isErr && mErr != nil {
				err = mErr
			}
		}

		func() {
			defer errCatcher()
			got = tc.from.Diff(tc.to)
		}()

		errMatches := tc.exptectErr == (err != nil)
		eq := cmp.Equal(got, want)

		if eq && errMatches {
			continue
		}

		errStr := &strings.Builder{}
		fmt.Fprintf(errStr, "test %v: ", i)

		if !eq {
			fmt.Fprintf(errStr, "\n%v (expected) != %v,\ndiff: %s", want, got, cmp.Diff(got, want))
		}

		if err != nil {
			fmt.Fprintf(errStr, "\ngot unexpected error: %v", err)
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
