
package chicklet

import (
	"testing"
	"unicode"
//	"fmt"
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
	if until([]rune("foo"))(vessel("foo")).matched {
		t.Error("\"foo\" is foo")
	}
	if !until([]rune("foo"))(vessel("baj")).matched {
		t.Error("\"baj\" is not foo")
	}
	if string(until([]rune("foo"))(vessel("baj")).match) != "baj" {
		t.Error("\"baj\" is baj")
	}
	if !until([]rune("foo"))(vessel("1foo")).matched {
		t.Error("\"1foo\" is more than foo")
	}
	if string(until([]rune("foo"))(vessel("1foo")).match) != "1" {
		t.Error("\"1foo\" is 1 and foo")
	}
	if string(until([]rune("foo"))(vessel("apabapa hej\n\rgnu åäöfoo")).match) != "apabapa hej\n\rgnu åäö" {
		t.Error("\"apabapa hej\n\rgnu åäöfoo\" is 1 and foo")
	}
}

/*
func TestMultiLineComment(t *testing.T) {
        s := fmt.Sprint("/","* kommentar\n\n*", "/")
	if !multiLineComment()(vessel(s)).matched {
		t.Error("\"",s,"\" is comment!")
	}
}
*/