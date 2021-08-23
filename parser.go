package pulp

import (
	"fmt"
	"strings"
)

type parser struct {
	tokens      <-chan *token
	runLexer    func()
	done        <-chan struct{}
	last        *token
	lastTrimmed *token
}

type Generator struct {
	idCounter    int
	sourceWriter strings.Builder
}

func (g *Generator) WriteNamed(source string) id {
	ident := g.nextID()
	g.sourceWriter.WriteString(string(ident) + " := " + source)
	return ident
}

func (g Generator) Out() string {
	return fmt.Sprintf(`func() pulp.StaticDynamic {
	%s
	return %s
}()`, g.sourceWriter.String(), string(g.lastID()))
}

func (g *Generator) nextID() id {
	g.idCounter++
	return id("x" + fmt.Sprint(g.idCounter))
}

func (g *Generator) lastID() id {
	return id("x" + fmt.Sprint(g.idCounter))
}

func NewParser(input string) *parser {
	tokens := make(chan *token)

	l := &lexer{tokens: tokens, input: input, state: lexUntilLBrace}

	return &parser{
		tokens:   tokens,
		runLexer: l.run,
	}
}

func (p *parser) next() *token {
	select {
	case <-p.done:
		return nil
	case p.last = <-p.tokens:
		if p.last.typ == tokEof {
			return nil
		}

		p.lastTrimmed = &token{typ: p.last.typ, value: strings.TrimSpace(p.last.value)}
		return p.last
	}
}

func (p *parser) Parse() *rootExpr {
	sd := rootExpr{}

	go p.runLexer()

	for {
		next := p.next()

		if next == nil {
			break
		}

		if next.typ == tokGoSource {
			keyWord := strings.Split(strings.TrimSpace(next.value), " ")[0]
			parser, foundMatchingKeyword := parserMap[keyWord]

			if !foundMatchingKeyword {
				parser = parseRawString
			}

			sd.dynamic = append(sd.dynamic, parser(p))
		} else if next.typ == tokOtherSource {

			sd.static = append(sd.static, next.value)
		} else {
			notreached()
		}
	}

	return &sd
}

type id string

type expr interface {
	Gen(*Generator) id
}

type parserFunc func(p *parser) expr

var parserMap = map[string]parserFunc{
	"if": parseIf,
}

type rawStringExpr string

func parseRawString(p *parser) expr {
	return rawStringExpr(p.last.value)
}

type rootExpr struct {
	static  []string
	dynamic []expr
}

type staticDynamicExpr struct {
	static, dynamic []string
}

type ifExpr struct {
	condStr string
	True    staticDynamicExpr
	False   staticDynamicExpr
}

func sprintDynamic(dynamics []string) string {
	ret := fmt.Sprint(dynamics)
	ret = strings.ReplaceAll(ret, " ", ", ")
	ret = ret[1 : len(ret)-1]
	return "{" + ret + "}"
}

func parseIf(p *parser) expr {
	ret := ifExpr{}
	ret.condStr = p.last.value[len("if "):]

	gotElseBranch := false

	for {
		next := p.next()
		if p.lastTrimmed.value == "else" {
			gotElseBranch = true
			break
		}
		if p.lastTrimmed.value == "end" {
			break
		}

		if next.typ == tokGoSource {
			ret.True.dynamic = append(ret.True.dynamic, next.value)
		} else if next.typ == tokOtherSource {
			ret.True.static = append(ret.True.static, next.value)
		} else {
			notreached()
		}
	}

	if gotElseBranch {
		for {
			next := p.next()
			if strings.TrimSpace(next.value) == "end" {
				break
			}

			if next.typ == tokGoSource {
				ret.False.dynamic = append(ret.False.dynamic, next.value)
			} else if next.typ == tokOtherSource {
				ret.False.static = append(ret.False.static, next.value)
			} else {
				notreached()
			}
		}
	}

	return &ret
}
