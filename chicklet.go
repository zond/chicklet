
package chicklet

import (
	"strings"
	"bytes"
	"fmt"
	"unicode"
)

type Vessel interface {
	GetPosition() position
	SetPosition(position)

	Next() (rune, bool)
	Pop(int)
	Push(int)
}

type Value interface{}

type Context interface {
	GetSub(string) Context
	Get(string) Value
	Set(string, Value) Value
}

var SLASH2 = []rune("//")
var SLASHS = []rune("/*")
var SSLASH = []rune("*/")
var BACKSLASH = []rune("\\")
var NL = []rune("\n")
var QUOT = []rune("\"")
var LEGAL_ESCAPES = []rune("abfnrtv\\\"")
var UNI2 = []rune("x")
var UNI4 = []rune("u")
var UNI8 = []rune("U")

type parser func(Vessel) *Output

type input interface{}

type Output struct {
	matched bool
	match []rune
	children []*Output
	content []rune
	eval func(context Context) Value
}
func (self *Output) Eval(context Context) Value {
	if self.eval == nil {
		return nil
	}
	return self.eval(context)
}
func (self *Output) String() string {
	return fmt.Sprint(self.matched, " content:", string(self.content), " match:", string(self.match), " children:", self.children)
}
func (self *Output) concatMatch(o *Output) {
	for _, r := range o.match {
		self.match = append(self.match, r)
	}
}
func (self *Output) concatContent(o *Output) {
	for _, r := range o.content {
		self.content = append(self.content, r)
	}
}
func (self *Output) concat(o *Output) {
	self.concatMatch(o)
	self.concatContent(o)
	self.children = append(self.children, o)
}

func rary(r rune) []rune {
	var ary []rune
	return append(ary, r)
}

type position struct {
	offset int
}

func satisfy(check func(c rune) bool) parser {
	return func(in Vessel) *Output {
		target, ok := in.Next()
		if ok && check(target) {
			in.Pop(1)
			return &Output{matched: true, match: rary(target), content: rary(target)}
		}

		return &Output{}
	}
}

func escapeUnicode() parser {
	return func(in Vessel) *Output {
		out := any(collect(static(BACKSLASH), static(UNI2), count(hex(), 2)),
			collect(static(BACKSLASH), static(UNI4), count(hex(), 4)),
			collect(static(BACKSLASH), static(UNI8), count(hex(), 8)))(in)
		if out.matched {
			buffer := bytes.NewBufferString("0x")
			fmt.Fprint(buffer, string(out.children[2].match))
			var r rune
			fmt.Fscanf(buffer, "%v", &r)
			out.match = rary(r)
			return out
		}		
		return &Output{}
	}
}

func escapeSingle() parser {
	return func(in Vessel) *Output {
		out := collect(static(BACKSLASH), oneOf(LEGAL_ESCAPES))(in)
		if out.matched {
			switch string(out.children[1].match) {
			case "a": out.match = []rune("\a")
			case "b": out.match = []rune("\b")
			case "f": out.match = []rune("\f")
			case "n": out.match = []rune("\n")
			case "r": out.match = []rune("\r")
			case "t": out.match = []rune("\t")
			case "v": out.match = []rune("\v")
			case "\\": out.match = []rune("\\")
			case "\"": out.match = []rune("\"")
			}
			return out
		} 
		return &Output{}
	}
}

func count(p parser, c int) parser {
	return func(in Vessel) *Output {
		out := &Output{matched: true}
		for i := 0; i < c; i++ {
			sub := p(in)
			if !sub.matched {
				in.Push(len(out.match))
				return &Output{}
			}
			out.concat(sub)
		}
		sub := p(in)
		if sub.matched {
			out.concat(sub)
			in.Push(len(out.match))
			return &Output{}
		}
		return out
	}
	
}

func replace(str []rune, replacement []rune) parser {
	return func(in Vessel) *Output {
		out := static(str)(in)
		if out.matched {
			return &Output{matched: true, match: replacement, content: replacement}
		}
		return &Output{}
	}
}

func whitespace() parser {
	return many1(any(satisfy(unicode.IsSpace), oneLineComment(), multiLineComment()))
}

func oneLineComment() parser {
	return collect(static(SLASH2), many(noneOf(NL)))
}

func multiLineComment() parser {
	return collect(static(SLASHS), inMulti())
}

func inMulti() parser {
	return func(in Vessel) *Output {
		return any(collect(until(SLASHS), multiLineComment(), inMulti()),
			collect(until(SSLASH),static(SSLASH)))(in)
	}
}

/*
 * Will consume until cs is found. Will match if cs is found, not otherwise.
 */
func until(cs []rune) parser {
	return func(in Vessel) *Output {
		out := &Output{}
		for {
			next, ok := in.Next()
			if ok {
				in.Pop(1)
				out.match = append(out.match, next)
				out.content = append(out.content, next)
				if strings.Index(string(out.match), string(cs)) != -1 {
					out.match = out.match[0:len(out.match) - len(cs)]
					out.content = out.content[0:len(out.content) - len(cs)]
					out.matched = true
					in.Push(len(cs))
					return out
				}
			} else {
				break
			}
		}
		in.Push(len(out.match))
		out.match = nil
		out.content = nil
		return out
	}
}

func digit() parser {
	return oneOf([]rune("0123456789"))
}

func hex() parser {
	return oneOf([]rune("0123456789abcdefABCDEF"))
}

func number() parser {
	return func(in Vessel) *Output {
		out := lexeme(any(collect(many1(digit()), static([]rune(".")), many1(digit())),
			many1(digit())))(in)
		if out.matched {
			m := string(out.match)
			out.eval = func(context Context) Value {
				if strings.Index(m, ".") != -1 {
					buffer := bytes.NewBufferString(m)
					var f float64
					fmt.Fscanf(buffer, "%v", &f)
					return f
				}
				buffer := bytes.NewBufferString(m)
				var f int
				fmt.Fscanf(buffer, "%v", &f)
				return f
			}
		}
		return out
	}
}

func stringLiteral() parser {
	return func(in Vessel) *Output {
		out := lexeme(between(static(QUOT), static(QUOT), many(any(escapeUnicode(), escapeSingle(), noneOf(QUOT)))))(in)
		if out.matched {
			out.eval = func(context Context) Value {
				return string(out.match)
			}
		}
		return out
	}
}

func oneOf(cs []rune) parser {
	return func(in Vessel) *Output {
		next, ok := in.Next()
		if ok {
			if strings.IndexRune(string(cs), next) != -1 {
				in.Pop(1)
				return &Output{matched: true, match: rary(next), content: rary(next)}
			}
		}
		return &Output{}
	}
}

func noneOf(cs []rune) parser {
	return func(in Vessel) *Output {
		next, ok := in.Next()
		if ok {
			if strings.IndexRune(string(cs), next) == -1 {
				in.Pop(1)
				return &Output{matched: true, match: rary(next), content: rary(next)}
			}
		}
		return &Output{}
	}
}

// Match a parser and skip whitespace
func lexeme(match parser) parser {
	return func(in Vessel) *Output {
		out := collect(match, many(whitespace()))(in)
		if out.matched {
			out.match = out.children[0].match
		}
		return out
	}
}

// Match a parser 0 or more times.
func many(match parser) parser {
	return func(in Vessel) *Output {
		out := &Output{matched: true}
		for {
			sub := match(in)
			if !sub.matched {
				break
			}

			out.concat(sub)
		}

		return out
	}
}

func many1(match parser) parser {
	return func(in Vessel) *Output {
		out := match(in)
		if !out.matched {
			return out
		}

		sub := many(match)(in)
		
		out.concat(sub)

		return out
	}
}

func sepBy(delim parser, match parser) parser {
	return func(in Vessel) *Output {
		out := &Output{}		
		for {
			sub := match(in)
			if sub.matched {
				out.matched = true
				out.concat(sub)
			} else {
				break
			}

			sub = delim(in)
			if sub.matched {
				out.concatMatch(sub)
			} else {
				break
			}
		}
		return out
	}
}

// Go through the parsers until one matches.
func any(parsers... parser) parser {
	return func(in Vessel) *Output {
		for _, parser := range parsers {
			sub := parser(in)
			if sub.matched {
				return sub
			}
		}

		return &Output{}
	}
}

// Match all parsers, returning the final result. If one fails, it stops.
func all(parsers... parser) parser {
	return try(func(in Vessel) *Output {
		var out *Output
		for _, parser := range parsers {
			out = parser(in)
			if !out.matched {
				return &Output{}
			}
		}
		return out
	})
}

// Match all parsers, collecting their Outputs
// If one parser fails, the whole thing fails.
func collect(parsers... parser) parser {
	return try(func(in Vessel) *Output {
		out := &Output{matched: true}
		for _, parser := range parsers {
			sub := parser(in)
			if sub.matched {
				out.concat(sub)
			} else {
				out = &Output{}
				break
			}
		}

		return out
	})
}

// Try matching begin, match, and then end.
func between(begin parser, end parser, match parser) parser {
	return try(func(in Vessel) *Output {
		out := collect(begin, match, end)(in)
		if out.matched {
			out.match = out.children[1].match
		}
		return out
	})
}

// Lexeme parser for `match' wrapped in parens.
func parens(match parser) parser { 
	return lexeme(between(symbol([]rune("(")), symbol([]rune(")")), match)) 
}

// Match a string and consume any following whitespace.
func symbol(str []rune) parser { 
	return lexeme(static(str)) 
}

// Match a string and pop the string's length from the input.
func static(str []rune) parser {
	return func(in Vessel) *Output {
		out := &Output{matched: true}
		for _, v := range str {
			next, ok := in.Next()
			if ok && next == v {
				out.match = append(out.match, next)
				out.content = append(out.content, next)
				in.Pop(1)
			} else {
				out.matched = false
				in.Push(len(out.match))
				out.match = nil
				out.content = nil
				return out
			}
		}
		return out
	}
}

// Try a parse and revert the state and position if it fails.
func try(match parser) parser {
	return func(in Vessel) *Output {
		pos := in.GetPosition()
		out := match(in)
		if !out.matched {
			in.SetPosition(pos)
			return out
		}

		return out
	}
}

type MapContext struct {
	data map[string]Value
	subs map[string]Context
}
func NewMapContext() *MapContext {
	return &MapContext{make(map[string]Value), make(map[string]Context)}
}
func (self *MapContext) GetSub(name string) Context {
	if sub, ok := self.subs[name]; ok {
		return sub
	}
	sub := NewMapContext()
	self.subs[name] = sub
	return sub
}
func (self *MapContext) Get(key string) Value {
	return self.data[key]
}
func (self *MapContext) Set(key string, value Value) Value {
	self.data[key] = value
	return value
}

// Basic string Vessel for parsing over a string input.
type StringVessel struct {
	input    []rune
	position position
}
func (self *StringVessel) String() string {
	return fmt.Sprint(self.position, "@", string(self.input))
}

func (self *StringVessel) Next() (rune, bool) {
	if len(self.input) < self.position.offset + 1 && self.position.offset >= 0 {
		return 0, false
	}
	return self.input[self.position.offset], true
}

func (self *StringVessel) Pop(i int) { 
	self.position.offset += i
}

func (self *StringVessel) Push(i int) { 
	self.position.offset -= i 
}

func (self *StringVessel) SetInput(in string) { 
	self.input = []rune(in) 
}

func (self *StringVessel) GetPosition() position {
	return self.position
}

func (self *StringVessel) SetPosition(pos position) {
	self.position = pos
}

