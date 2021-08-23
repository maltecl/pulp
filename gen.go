package pulp

import (
	"fmt"
	"strings"

	"github.com/kr/pretty"
)

func (r rootExpr) Gen(g *Generator) id {
	staticsString := strings.Join(r.static, "{}")

	dynamicString := &strings.Builder{}

	for _, d := range r.dynamic {
		dynamicString.WriteString(", " + string(d.Gen(g)))
	}

	return g.WriteNamed(fmt.Sprintf("pulp.NewStaticDynamic(%q %s)", staticsString, dynamicString.String()))
}

func (staticDynamicExpr) Gen(g *Generator) id {
	return id("")
}

func (i *ifExpr) Gen(g *Generator) id {
	return g.WriteNamed(
		fmt.Sprintf(
			`pulp.If{
		Condition: %s,
		True: pulp.StaticDynamic{
			Static:  %s,
			Dynamic: pulp.Dynamics%s,
		},
		False: pulp.StaticDynamic{
			Static:  %s,
			Dynamic: pulp.Dynamics%s,
		},
	}
	`,
			i.condStr,
			pretty.Sprint(i.True.static),
			sprintDynamic(i.True.dynamic),
			pretty.Sprint(i.False.static),
			sprintDynamic(i.False.dynamic),
		),
	)
}

func (e rawStringExpr) Gen(g *Generator) id {
	return g.WriteNamed(string(e) + "\n")
}
