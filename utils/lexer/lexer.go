package lexer

import (
	"fmt"
	"text/scanner"

	"github.com/ioxe/graphql-go/errors"
)

type syntaxError string

type Lexer struct {
	sc          *scanner.Scanner
	next        rune
	descComment string
}

type Literal struct {
	Type rune
	Text string
}

type Ident struct {
	Name string
	Loc  errors.Location
}

type Variable string

func New(sc *scanner.Scanner) *Lexer {
	l := &Lexer{sc: sc}
	l.Consume()
	return l
}

func (l *Lexer) CatchSyntaxError(f func()) (errRes *errors.QueryError) {
	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(syntaxError); ok {
				errRes = errors.Errorf("syntax error: %s", err)
				errRes.Locations = []errors.Location{l.Location()}
				return
			}
			panic(err)
		}
	}()

	f()
	return
}

func (l *Lexer) Peek() rune {
	return l.next
}

func (l *Lexer) Consume() {
	l.descComment = ""
	for {
		l.next = l.sc.Scan()
		if l.next == ',' {
			continue
		}
		if l.next == '#' {
			if l.sc.Peek() == ' ' {
				l.sc.Next()
			}
			if l.descComment != "" {
				l.descComment += "\n"
			}
			for {
				next := l.sc.Next()
				if next == '\n' || next == scanner.EOF {
					break
				}
				l.descComment += string(next)
			}
			continue
		}
		break
	}
}

func (l *Lexer) ConsumeIdent() string {
	name := l.sc.TokenText()
	l.ConsumeToken(scanner.Ident)
	return name
}

func (l *Lexer) ConsumeIdentWithLoc() Ident {
	loc := l.Location()
	name := l.sc.TokenText()
	l.ConsumeToken(scanner.Ident)
	return Ident{name, loc}
}

func (l *Lexer) ConsumeKeyword(keyword string) {
	if l.next != scanner.Ident || l.sc.TokenText() != keyword {
		l.SyntaxError(fmt.Sprintf("unexpected %q, expecting %q", l.sc.TokenText(), keyword))
	}
	l.Consume()
}

func (l *Lexer) ConsumeVariable() Variable {
	l.ConsumeToken('$')
	return Variable(l.ConsumeIdent())
}

func (l *Lexer) ConsumeLiteral() interface{} {
	lit := &Literal{Type: l.next, Text: l.sc.TokenText()}
	l.Consume()
	if lit.Type == scanner.Ident && lit.Text == "null" {
		return nil
	}
	return lit
}

func (l *Lexer) ConsumeToken(expected rune) {
	if l.next != expected {
		l.SyntaxError(fmt.Sprintf("unexpected %q, expecting %s", l.sc.TokenText(), scanner.TokenString(expected)))
	}
	l.Consume()
}

func (l *Lexer) DescComment() string {
	return l.descComment
}

func (l *Lexer) SyntaxError(message string) {
	panic(syntaxError(message))
}

func (l *Lexer) Location() errors.Location {
	return errors.Location{
		Line:   l.sc.Line,
		Column: l.sc.Column,
	}
}
