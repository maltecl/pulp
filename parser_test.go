package pulp

import (
	"testing"
)

func TestParser(t *testing.T) {
	p := NewParser(testSource2)

	g := &Generator{}
	expr := p.Parse()
	t.Logf("parse: %v, %v", expr.Gen(g), g.sourceWriter.String())

	t.Log("\n\n\n")

	replace()
}
