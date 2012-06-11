
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
/*
func TestMultiLineComment(t *testing.T) {
        s := fmt.Sprint("/","* kommentar\n\n*", "/")
	if !multiLineComment()(vessel(s)).matched {
		t.Error("\"",s,"\" is comment!")
	}
}
*/