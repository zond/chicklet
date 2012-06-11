
package chicklet

import (
	"testing"
	"unicode"
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

func TestNumber(t *testing.T) {
	p := number()
	parserTest(t, p, "0", true, "0")
	parserTest(t, p, "01", true, "01")
	parserTest(t, p, "20", true, "20")
	parserTest(t, p, "344", true, "344")
	parserTest(t, p, "34234.01", true, "34234.01")
	parserTest(t, p, "2134.11", true, "2134.11")
	parserTest(t, p, "0.131", true, "0.131")
	parserTest(t, p, ".2", false, "")
	parserTest(t, p, "f", false, "")
	parserTest(t, p, "f2", false, "")
	parserTest(t, p, "f.2", false, "")
	parserTest(t, p, "2.f", true, "2")
	parserTest(t, p, "2.01x", true, "2.01")
	parserTest(t, p, "2.01.1", true, "2.01")
}

func TestStringLiteral(t *testing.T) {
	p := stringLiteral()
	parserTest(t, p, "\"a\"", true, "a")
	parserTest(t, p, "\"asdf asdf a sdf\"", true, "asdf asdf a sdf")
	parserTest(t, p, "\"hej \\\"kompis\\\"!\"", true, "hej \"kompis\"!")
	parserTest(t, p, "\"hej\\nkompis\"", true, "hej\nkompis")
	parserTest(t, p, "\"hej\\rkompis\"", true, "hej\rkompis")
	parserTest(t, p, "\"hej\\tkompis\"", true, "hej\tkompis")
}