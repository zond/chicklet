
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
	if until([]rune("bajs"))(vessel("bajs")).matched {
		t.Error("\"bajs\" is bajs")
	}
	if !until([]rune("bajs"))(vessel("1bajs")).matched {
		t.Error("\"1bajs\" is more than bajs")
	}
	if string(until([]rune("bajs"))(vessel("1bajs")).match) != "1" {
		t.Error("\"1bajs\" is 1 and bajs")
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