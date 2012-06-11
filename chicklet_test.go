
package chicklet

import (
	"testing"
	"unicode"
	"fmt"
)

func vessel(s string) *StringVessel {
	return &StringVessel{[]rune(s), position{}}
}

func TestSatisfy(t *testing.T) {
	if satisfy(unicode.IsSpace)(vessel("h")).matched {
		t.Error("\"h\" is not space!")
	}
	if !satisfy(unicode.IsSpace)(vessel(" ")).matched {
		t.Error("\" \" is space!")
	}
	if !satisfy(unicode.IsSpace)(vessel("\n")).matched {
		t.Error("\"\\n\" is space!")
	}
	if !satisfy(unicode.IsSpace)(vessel("\r")).matched {
		t.Error("\"\\r\" is space!")
	}
}
	
func TestOneLineComment(t *testing.T) {
	if !oneLineComment()(vessel("// kommentar")).matched {
		t.Error("\"// kommentar\" is comment!")
	}
	if oneLineComment()(vessel("kod // kommentar")).matched {
		t.Error("\"kod // kommentar\" is not comment!")
	}
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
	stringLiteralTester(t, "\"aasdfasdfasfd\"")
}