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
		// if p.last.typ == tokEof {
		// 	return nil
		// }

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
	shouldBreak := false
	for !shouldBreak {
		next := p.next()

		shouldBreak = next.typ == tokEof

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
		} else if next.typ == tokOtherSource || next.typ == tokEof {
			ret.static = append(ret.static, next.value)
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
		"key": parseKeyedSection,
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

func parseIf(p *parser) expr {
	ret := ifExpr{}
	ret.condStr = p.last.value[len("if "):]

	var endedWith string
	ret.True, endedWith = parseAllUntil(p, []string{"else", "end"})

	gotElseBranch := endedWith == "else"

	if gotElseBranch {
		ret.False, endedWith = parseAllUntil(p, []string{"end"})
	} else {
		ret.False = staticDynamicExpr{static: []string{}, dynamic: []expr{}}
	}
	p.assertf(endedWith == "end", "expected \"end\", got: %q", endedWith)

	return &ret
}

type forExpr struct {
	rangeStr string
	sd       staticDynamicExpr
}

func parseFor(p *parser) expr {
	ret := forExpr{}
	ret.rangeStr = p.last.value[len("for "):]

	var endedWith string
	ret.sd, endedWith = parseAllUntil(p, []string{"end"})
	p.assertf(endedWith == "end", `expected "end", got: %q`, endedWith)

	return ret
}

type keyedSectionExpr struct {
	keyString string
	sd        staticDynamicExpr
}

func parseKeyedSection(p *parser) expr {
	ret := keyedSectionExpr{keyString: p.last.value[len("key "):]}

	var endedWith string
	ret.sd, endedWith = parseAllUntil(p, []string{"end"})
	p.assertf(endedWith == "end", `expected "end", got: %q`, endedWith)

	return ret
}
