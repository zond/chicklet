
package chicklet

import (
	"testing"
	"unicode"
)

func TestSatisfy(t *testing.T) {
	if satisfy(unicode.IsSpace)(&StringVessel{"h", position{}}).matched {
		t.Error("\"h\" is not space!")
	}
	if !satisfy(unicode.IsSpace)(&StringVessel{" ", position{}}).matched {
		t.Error("\" \" is space!")
	}
	if !satisfy(unicode.IsSpace)(&StringVessel{"\n", position{}}).matched {
		t.Error("\"\\n\" is space!")
	}
	if !satisfy(unicode.IsSpace)(&StringVessel{"\r", position{}}).matched {
		t.Error("\"\\r\" is space!")
	}
}

func TestOneLineComment(t *testing.T) {
	if !oneLineComment()(&StringVessel{"// kommentar", position{}}).matched {
		t.Error("\"// kommentar\" is comment!")
	}
	if oneLineComment()(&StringVessel{"kod // kommentar", position{}}).matched {
		t.Error("\"kod // kommentar\" is not comment!")
	}
}

func TestMultiLineComment(t *testing.T) {
	if !oneLineComment()(&StringVessel{"// kommentar", position{}}).matched {
		t.Error("\"// kommentar\" is comment!")
	}
}
