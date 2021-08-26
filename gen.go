package pulp

import (
	"fmt"
	"strings"

	"github.com/kr/pretty"
)

func (r staticDynamicExpr) Gen(g *Generator) id {
	staticsString := strings.Join(r.static, "{}")

	dynamicString := &strings.Builder{}

	for _, d := range r.dynamic {
		dynamicString.WriteString(", " + string(d.Gen(g)))
	}

	return g.WriteNamed(fmt.Sprintf("pulp.NewStaticDynamic(%q %s)", staticsString, dynamicString.String()))
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

func (e forExpr) Gen(g *Generator) id {
	return g.WriteNamedWithID(func(currentID id) string {
		return fmt.Sprintf(`pulp.For{
		Statics: %s,
		ManyDynamics: make([]pulp.Dynamics, 0),
		DiffStrategy: pulp.Append,
	}

	for %s {
		%s.ManyDynamics = append(%s.ManyDynamics, pulp.Dynamics%s)
	}
	`, pretty.Sprint(e.static), e.rangeStr, string(currentID), string(currentID), sprintDynamic(e.dynamic))

	})
}

func sprintDynamic(dynamics []expr) string {
	ret := fmt.Sprint(dynamics)
	ret = strings.ReplaceAll(ret, " ", ", ")
	ret = ret[1 : len(ret)-1]
	return "{" + ret + "}"
}
