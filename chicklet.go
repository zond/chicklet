
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

type parser func(Vessel) *output

type input interface{}

type output struct {
	matched bool
	match []rune
	children []*output
}
func (self *output) body() string {
	if self.match != nil {
		return string(self.match)
	}
	buffer := bytes.NewBufferString("")
	for _, o := range self.children {
		fmt.Fprint(buffer, o.body())
	}
	return string(buffer.Bytes())
}
func (self *output) String() string {
	return fmt.Sprint(string(self.match), self.children)
}
func (self *output) concatMatch(o *output) {
	for _, r := range o.match {
		self.match = append(self.match, r)
	}
}
func (self *output) concatChildren(o *output) {
	for _, c := range o.children {
		self.children = append(self.children, c)
	}
}
func (self *output) concat(o *output) {
	self.concatMatch(o)
	self.concatChildren(o)
}

func FALSE() *output {
	return &output{false, nil, nil}
}
func rary(r rune) []rune {
	var ary []rune
	return append(ary, r)
}

type position struct {
	offset int
}

func satisfy(check func(c rune) bool) parser {
	return func(in Vessel) *output {
		target, ok := in.Next()
		if ok && check(target) {
			in.Pop(1)
			return &output{true, rary(target), nil}
		}

		return FALSE()
	}
}

func whitespace() parser {
	return many(any(satisfy(unicode.IsSpace), oneLineComment(), multiLineComment()))
}

func oneLineComment() parser {
	return all(try(static("//")), many(noneOf("\n")))
}

func multiLineComment() parser {
	return all(static("/*"), inMulti())
}

func inMulti() parser {
	return any(static("*/"),
  		all(multiLineComment(), inMulti()),
		all(many1(noneOf("/*")), inMulti()),
	all(oneOf("/*"), inMulti()))
}

func oneOf(cs string) parser {
	return func(in Vessel) *output {
		next, ok := in.Next()
		if ok {
			if strings.IndexRune(cs, next) != -1 {
				in.Pop(1)
				return &output{true, rary(next), nil}
			}
		}
		return FALSE()
	}
}

func noneOf(cs string) parser {
	return func(in Vessel) *output {
		next, ok := in.Next()
		if ok {
			if strings.IndexRune(cs, next) == -1 {
				in.Pop(1)
				return &output{true, rary(next), nil}
			}
		}
		return FALSE()
	}
}

// Match a parser and skip whitespace
func lexeme(match parser) parser {
	return try(func(in Vessel) *output {
		out := match(in)
		if out.matched {
			whitespace()(in)
			return out
		}
		return FALSE()
	})
}

// Match a parser 0 or more times.
func many(match parser) parser {
	return func(in Vessel) *output {
		out := &output{true, nil, nil}
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
	return func(in Vessel) *output {
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
	return func(in Vessel) *output {
		out := FALSE()		
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
	return func(in Vessel) *output {
		for _, parser := range parsers {
			sub := parser(in)
			if sub.matched {
				return sub
			}
		}

		return FALSE()
	}
}

// Match all parsers, returning the final result. If one fails, it stops.
func all(parsers... parser) parser {
	return try(func(in Vessel) *output {
		var out *output
		for _, parser := range parsers {
			out = parser(in)
			if !out.matched {
				return FALSE()
			}
		}
		return out
	})
}

// Match all parsers, collecting their outputs into a slice.
// If one parser fails, the whole thing fails.
func collect(parsers... parser) parser {
	return try(func(in Vessel) *output {
		out := &output{true, nil, nil}
		for _, parser := range parsers {
			sub := parser(in)
			if sub.matched {
				out.concat(sub)
			} else {
				out = FALSE()
				break
			}
		}

		return out
	})
}

// Try matching begin, match, and then end.
func between(begin parser, end parser, match parser) parser {
	return func(in Vessel) *output {
		out := try(collect(begin, match, end))(in)
		if out.matched {
			return out.children[1]
		}
		return out
	}
}

// Lexeme parser for `match' wrapped in parens.
func parens(match parser) parser { 
	return lexeme(between(symbol("("), symbol(")"), match)) 
}

// Match a string and consume any following whitespace.
func symbol(str string) parser { 
	return lexeme(static(str)) 
}

// Match a string and pop the string's length from the input.
func static(str string) parser {
	return func(in Vessel) *output {
		for _, v := range str {
			next, ok := in.Next()
			if !ok || next != v {
				return FALSE()
			}
		}
		in.Pop(len(str))
		return &output{true, []rune(str), nil}
	}
}

// Try a parse and revert the state and position if it fails.
func try(match parser) parser {
	return func(in Vessel) *output {
		pos := in.GetPosition()
		out := match(in)
		if !out.matched {
			in.SetPosition(pos)
			return out
		}

		return out
	}
}

// Basic string Vessel for parsing over a string input.
type StringVessel struct {
	input    string
	position position
}

func (self *StringVessel) Next() (rune, bool) {
	if len(self.input) < self.position.offset + 1 && self.position.offset >= 0 {
		return 0, false
	}
	for index, r := range self.input {
		if index == self.position.offset {
			return r, true
		}
	}
	return 0, false
}

func (self *StringVessel) Pop(i int) { 
	self.position.offset += i
}

func (self *StringVessel) Push(i int) { 
	self.position.offset -= i 
}

func (self *StringVessel) SetInput(in string) { self.input = in }

func (self *StringVessel) GetPosition() position {
	return self.position
}

func (self *StringVessel) SetPosition(pos position) {
	self.position = pos
}

