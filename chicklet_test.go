
package chicklet

import (
	"testing"
	"unicode"
	"fmt"
)

func vessel(s string) *StringVessel {
	return &StringVessel{[]rune(s), position{}}
}

func parserTest(t *testing.T, p parser, in string, m bool, exp string) {
	o := p(vessel(in))
	if o.matched != m {
		if m {
			t.Error("expected",in,"to match",p,"but it didn't")
		} else {
			t.Error("expected",in,"to NOT match",p,"but it DID")
		}
	}
	if string(o.match) != exp {
		t.Error("expected",p,"parsing",in,"to generate",exp,"but it generated",string(o.match))
	} 
}

func TestSatisfy(t *testing.T) {
	parserTest(t, satisfy(unicode.IsSpace), "h", false, "")
	parserTest(t, satisfy(unicode.IsSpace), " ", true, " ")
	parserTest(t, satisfy(unicode.IsSpace), "\n", true, "\n")
	parserTest(t, satisfy(unicode.IsSpace), "\r", true, "\r")
}
	
func TestOneLineComment(t *testing.T) {
	parserTest(t, oneLineComment(), "// kommentar", true, "// kommentar")
	parserTest(t, oneLineComment(), "kod // kommentar", false, "")
}

func TestUntil(t *testing.T) {
	if !until([]rune("foo"))(vessel("foo")).matched {
		t.Error("\"foo\" has foo")
	}
	if string(until([]rune("foo"))(vessel("foo")).match) != "" {
		t.Error("\"foo\" has \"\" before foo")
	}
	if until([]rune("foo"))(vessel("bar")).matched {
		t.Error("\"baj\" is not foo")
	}
	if !until([]rune("foo"))(vessel("1foo")).matched {
		t.Error("\"1foo\" has foo")
	}
	if string(until([]rune("foo"))(vessel("1foo")).match) != "1" {
		t.Error("\"1foo\" has \"1\" before \"foo\"")
	}
	if string(until([]rune("foo"))(vessel("apabapa hej\n\rgnu åäöfoo")).match) != "apabapa hej\n\rgnu åäö" {
		t.Error("\"apabapa hej\n\rgnu åäöfoo\" has \"apabapa hej\n\rgnu åäö\" before \"foo\"")
	}
}

func TestMultiLineComment(t *testing.T) {
        s := "/* kommentar\n\n*/"
	if !multiLineComment()(vessel(s)).matched {
		t.Error(s, "is comment!")
	}
	if string(multiLineComment()(vessel(s)).match) != s {
		t.Error(s, "is", s)
	}
	s = "/* kommentar\n\n/* nested broken comment\n  \n \r*/"
	if !multiLineComment()(vessel(s)).matched {
		t.Error(s, "is comment!")
	}
	if string(multiLineComment()(vessel(s)).match) != s {
		t.Error(s, "is", s)
	}
	s = "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/"
	if !multiLineComment()(vessel(s)).matched {
		t.Error(s, "is comment!")
	}
	if string(multiLineComment()(vessel(s)).match) != s {
		t.Error(s, "is", s)
	}
}

func numberTester(t *testing.T, test string) {
	if !number()(vessel(test)).matched {
		t.Error(test, "is number")
	}
	if string(number()(vessel(test)).match) != test {
		t.Error(test,"is",test)
	}
}

func badNumberTester(t *testing.T, test string) {
	if number()(vessel(test)).matched {
		t.Error(test, "is not number")
	}
	if string(number()(vessel(test)).match) != "" {
		t.Error(test,"should not consume")
	}
}

func TestNumber(t *testing.T) {
	numberTester(t, "0")
	numberTester(t, "01")
	numberTester(t, "20")
	numberTester(t, "3434")
	numberTester(t, "3434.0")
	numberTester(t, "3434.131")
	numberTester(t, "0.131")
	badNumberTester(t, ".2")
	badNumberTester(t, "f")
	badNumberTester(t, "f2")
	badNumberTester(t, "f.2")
}

func stringLiteralTester(t *testing.T, s string) {
	if !stringLiteral()(vessel(s)).matched {
		t.Error(s,"is a string literal!")
	}
	m := string(stringLiteral()(vessel(s)).match)
	if fmt.Sprint("\"", m, "\"") != s {
		t.Error(s,"is",s,"but got",m)
	}
}

func TestStringLiteral(t *testing.T) {
	stringLiteralTester(t, "\"a\"")
	stringLiteralTester(t, "\"aasdfasd  fasfd\"")
}