package pulp

import (
	"fmt"
	"strings"

	"github.com/kr/pretty"
)

type Generator struct {
	idCounter int
	scopes    *scopeStack
}

func NewGenerator() *Generator {
	g := Generator{}
	g.pushScope()
	return &g
}

type scopeStack struct {
	prev *scopeStack
	scope
}

type scope struct {
	strings.Builder
}

func (g *Generator) pushScope() {
	newScopeEntry := scopeStack{prev: g.scopes}
	g.scopes = &newScopeEntry
}

func (g *Generator) popScope() string {
	if g.scopes == nil {
		return ""
	}
	ret := g.scopes.String()
	g.scopes = g.scopes.prev
	return ret
}

func (g *Generator) WriteNamed(source string) id {
	ident := g.nextID()
	g.scopes.WriteString(string(ident) + " := " + source)
	return ident
}

func (g *Generator) WriteNamedWithID(source func(id) string) id {
	ident := g.nextID()
	g.scopes.WriteString(string(ident) + " := " + source(ident))
	return ident
}

func (g Generator) Out() string {
	return fmt.Sprintf(`func() pulp.StaticDynamic {
	%s
	return %s
}()`, g.popScope(), string(g.lastID()))
}

func (g *Generator) nextID() id {
	g.idCounter++
	return id("x" + fmt.Sprint(g.idCounter))
}

func (g *Generator) lastID() id {
	return id("x" + fmt.Sprint(g.idCounter))
}

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
			sprintDynamic(i.True.dynamic, g),
			pretty.Sprint(i.False.static),
			sprintDynamic(i.False.dynamic, g),
		),
	)
}

func (e rawStringExpr) Gen(g *Generator) id {
	return g.WriteNamed(string(e) + "\n")
}

func (e forExpr) Gen(g *Generator) id {
	return g.WriteNamedWithID(func(currentID id) string {
		ret := fmt.Sprintf(`pulp.For{
		Statics: %s,
		ManyDynamics: make([]pulp.Dynamics, 0),
		DiffStrategy: pulp.Append,
	}

	for %s {
	`, pretty.Sprint(e.sd.static), e.rangeStr)

		g.pushScope()
		idStr := string(currentID)
		ids := sprintDynamic(e.sd.dynamic, g)
		scopeStr := g.popScope()

		ret += fmt.Sprintf(`%s
		%s.ManyDynamics = append(%s.ManyDynamics, pulp.Dynamics%s)
		}
	`, scopeStr, idStr, idStr, ids)

		return ret
	})
}

func (e keyedSectionExpr) Gen(g *Generator) id {
	return g.WriteNamed("MARKER\n")
}

func sprintDynamic(dynamics []expr, g *Generator) string {

	ret := &strings.Builder{}

	for _, e := range dynamics {
		if ee, ok := e.(rawStringExpr); ok {
			ret.WriteString(string(ee))
		} else {
			ret.WriteString(string(e.Gen(g)))
		}
		ret.WriteString(", ")
	}

	// ret := fmt.Sprint(dynamics)
	// ret = strings.ReplaceAll(ret, " ", ", ")
	// ret = ret[1 : len(ret)-1]

	retStr := ret.String()

	if len(dynamics) > 1 {
		retStr = retStr[:len(retStr)-1]
	}
	return "{" + retStr + "}"
}
