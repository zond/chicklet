
package chicklet

import (
	"testing"
	"unicode"
)

func vessel(s string) *StringVessel {
	return &StringVessel{[]rune(s), position{}}
}

func parserTest(t *testing.T, p parser, in string, m bool, match, content string) {
	v := vessel(in)
	pos := v.position.offset
	o := p(v)
	if o.matched != m {
		if m {
			t.Error("expected",in,"to match",p,"but it didn't")
		} else {
			t.Error("expected",in,"to NOT match",p,"but it DID")
		}
	}
	if string(o.match) != match {
		t.Error("expected",p,"parsing",in,"to have match",match,"but it generated",string(o.match))
	} 
	if string(o.content) != content {
		t.Error("expected",p,"parsing",in,"to have content",content,"but it generated",string(o.content))
	}
	if len(o.content) + pos != v.position.offset {
		t.Error("expected",p,"parsing",in,"to consume",content,"but it consumed",v.position.offset - pos)
	}
}

func TestSatisfy(t *testing.T) {
	p := satisfy(unicode.IsSpace)
	parserTest(t, p, "h", false, "", "")
	parserTest(t, p, " ", true, " ", " ")
	parserTest(t, p, "\n", true, "\n", "\n")
	parserTest(t, p, "\r", true, "\r", "\r")
}
	
func TestOneLineComment(t *testing.T) {
	p := oneLineComment()
	parserTest(t, p, "// kommentar", true, "// kommentar", "// kommentar")
	parserTest(t, p, "kod // kommentar", false, "", "")
}

func TestUntil(t *testing.T) {
	p := until([]rune("foo"))
	parserTest(t, p, "foo", true, "", "")
	parserTest(t, p, "bar", false, "", "")
	parserTest(t, p, "1foo", true, "1", "1")
	parserTest(t, p, "apabapa hej\n\rgnu åäöfoo", true, "apabapa hej\n\rgnu åäö", "apabapa hej\n\rgnu åäö")
}

func TestMultiLineComment(t *testing.T) {
	p := multiLineComment()
	parserTest(t, p, "/* kommentar\n\n*/", true, "/* kommentar\n\n*/", "/* kommentar\n\n*/")
	parserTest(t, p, "/* kommentar\n\n/* nested broken comment\n  \n \r*/", true, "/* kommentar\n\n/* nested broken comment\n  \n \r*/", "/* kommentar\n\n/* nested broken comment\n  \n \r*/")
	parserTest(t, p, "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/", true, "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/", "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/")
}

func TestNumber(t *testing.T) {
	p := number()
	parserTest(t, p, "0", true, "0", "0")
	parserTest(t, p, "01", true, "01", "01")
	parserTest(t, p, "20", true, "20", "20")
	parserTest(t, p, "344", true, "344", "344")
	parserTest(t, p, "34234.01", true, "34234.01", "34234.01")
	parserTest(t, p, "2134.11", true, "2134.11", "2134.11")
	parserTest(t, p, "0.131", true, "0.131", "0.131")
	parserTest(t, p, ".2", false, "", "")
	parserTest(t, p, "f", false, "", "")
	parserTest(t, p, "f2", false, "", "")
	parserTest(t, p, "f.2", false, "", "")
	parserTest(t, p, "2.f", true, "2", "2")
	parserTest(t, p, "2.01x", true, "2.01", "2.01")
	parserTest(t, p, "2.01.1", true, "2.01", "2.01")
}

func TestCount(t *testing.T) {
	p := count(digit(), 3)
	parserTest(t, p, "0", false, "", "")
	parserTest(t, p, "aaa", false, "", "")
	parserTest(t, p, "012", true, "012", "012")
	parserTest(t, p, "0123", false, "", "")
}

func TestStringLiteral(t *testing.T) {
	p := stringLiteral()
	parserTest(t, p, "\"a\"", true, "a", "\"a\"")
	parserTest(t, p, "\"asdf asdf a sdf\"", true, "asdf asdf a sdf", "\"asdf asdf a sdf\"")
	parserTest(t, p, "\"\\x42\"", true, "B", "\"\\x42\"")
	parserTest(t, p, "\"\\u4142\"", true, "\u4142", "\"\\u4142\"")
	parserTest(t, p, "\"\\U0002070E\"", true, "\U0002070E", "\"\\U0002070E\"")
	parserTest(t, p, "\"\\a\\b\\f\\n\\r\\t\\v\\\\\\\"\\x42\\u4142\\U0002070Ehej\"", true, "\a\b\f\n\r\t\v\\\"B\u4142\U0002070Ehej", "\"\\a\\b\\f\\n\\r\\t\\v\\\\\\\"\\x42\\u4142\\U0002070Ehej\"")
}
