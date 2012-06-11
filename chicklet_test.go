
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
	p := satisfy(unicode.IsSpace)
	parserTest(t, p, "h", false, "")
	parserTest(t, p, " ", true, " ")
	parserTest(t, p, "\n", true, "\n")
	parserTest(t, p, "\r", true, "\r")
}
	
func TestOneLineComment(t *testing.T) {
	p := oneLineComment()
	parserTest(t, p, "// kommentar", true, "// kommentar")
	parserTest(t, p, "kod // kommentar", false, "")
}

func TestUntil(t *testing.T) {
	p := until([]rune("foo"))
	parserTest(t, p, "foo", true, "")
	parserTest(t, p, "bar", false, "")
	parserTest(t, p, "1foo", true, "1")
	parserTest(t, p, "apabapa hej\n\rgnu åäöfoo", true, "apabapa hej\n\rgnu åäö")
}

func TestMultiLineComment(t *testing.T) {
	p := multiLineComment()
	parserTest(t, p, "/* kommentar\n\n*/", true, "/* kommentar\n\n*/")
	parserTest(t, p, "/* kommentar\n\n/* nested broken comment\n  \n \r*/", true, "/* kommentar\n\n/* nested broken comment\n  \n \r*/")
	parserTest(t, p, "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/", true, "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/")
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