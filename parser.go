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

	Error error
}

func (p *parser) assertf(cond bool, format string, args ...interface{}) {
	if p.Error != nil || cond {
		return
	}
	p.Error = fmt.Errorf(format, args...)
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

func (g *Generator) WriteNamedWithID(source func(id) string) id {
	ident := g.nextID()
	g.sourceWriter.WriteString(string(ident) + " := " + source(ident))
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

func (p *parser) Parse() *staticDynamicExpr {
	sd := staticDynamicExpr{}

	go p.runLexer()

	ret, _ := parseAllUntil(p, []string{})
	sd.dynamic = ret.dynamic
	sd.static = ret.static

	return &sd
}

func parseAllUntil(p *parser, delimiters []string) (ret staticDynamicExpr, endedWith string) {
	for {
		next := p.next()

		if next == nil {
			break
		}

		for _, delimiter := range delimiters {
			if p.lastTrimmed.value == delimiter {
				endedWith = delimiter
				return
			}
		}

		if next.typ == tokGoSource {
			keyWord := strings.Split(strings.TrimSpace(next.value), " ")[0]
			parser, foundMatchingKeyword := parserMap[keyWord]

			if !foundMatchingKeyword {
				parser = parseRawString
			}

			ret.dynamic = append(ret.dynamic, parser(p))
		} else if next.typ == tokOtherSource {
			ret.static = append(ret.static, p.last.value)
		} else {
			notreached()
		}
	}

	return
}

type id string

type expr interface {
	Gen(*Generator) id
}

type parserFunc func(p *parser) expr

var parserMap map[string]parserFunc

func init() {
	parserMap = map[string]parserFunc{
		"for": parseFor,
		"if":  parseIf,
	}
}

type rawStringExpr string

func parseRawString(p *parser) expr {
	return rawStringExpr(p.lastTrimmed.value)
}

type staticDynamicExpr struct {
	static  []string
	dynamic []expr
}

type ifExpr struct {
	condStr string
	True    staticDynamicExpr
	False   staticDynamicExpr
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

func parseIf(p *parser) expr {
	ret := ifExpr{}
	ret.condStr = p.last.value[len("if "):]

	var endedWith string
	ret.True, endedWith = parseAllUntil(p, []string{"else"})

	gotElseBranch := endedWith == "else"
	p.assertf(gotElseBranch, fmt.Sprintf("!gotElseBranch: %q", endedWith))

	if gotElseBranch {
		ret.False, endedWith = parseAllUntil(p, []string{"end"})
		p.assertf(endedWith == "end", fmt.Sprintf("expected \"end\", got: %q", endedWith))
	}

	return &ret
}

type forExpr struct {
	rangeStr string
	staticDynamicExpr
}

func parseFor(p *parser) expr {
	ret := forExpr{}
	ret.rangeStr = p.last.value[len("for "):]

	var endedWith string
	ret.staticDynamicExpr, endedWith = parseAllUntil(p, []string{"end"})

	p.assertf(endedWith == "end", fmt.Sprintf(`expected "end", got: `, endedWith))

	return ret
}
