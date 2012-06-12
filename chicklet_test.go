
package chicklet

import (
	"testing"
	"unicode"
)

func vessel(s string) *StringVessel {
	return &StringVessel{[]rune(s), position{}}
}

func parserTest(t *testing.T, p parser, in string, m bool, match, content string, eval Value) {
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
	
	if r := o.Eval(NewMapContext()); r != eval {
		t.Error("expected",p,"parsing",in,"to eval to",eval,"but it evaled to",r)
	}
}

func TestSatisfy(t *testing.T) {
	p := satisfy(unicode.IsSpace)
	parserTest(t, p, "h", false, "", "", nil)
	parserTest(t, p, " ", true, " ", " ", nil)
	parserTest(t, p, "\n", true, "\n", "\n", nil)
	parserTest(t, p, "\r", true, "\r", "\r", nil)
}
	
func TestOneLineComment(t *testing.T) {
	p := oneLineComment()
	parserTest(t, p, "// kommentar", true, "// kommentar", "// kommentar", nil)
	parserTest(t, p, "kod // kommentar", false, "", "", nil)
}

func TestUntil(t *testing.T) {
	p := until([]rune("foo"))
	parserTest(t, p, "foo", true, "", "", nil)
	parserTest(t, p, "bar", false, "", "", nil)
	parserTest(t, p, "1foo", true, "1", "1", nil)
	parserTest(t, p, "apabapa hej\n\rgnu åäöfoo", true, "apabapa hej\n\rgnu åäö", "apabapa hej\n\rgnu åäö", nil)
}

func TestMultiLineComment(t *testing.T) {
	p := multiLineComment()
	parserTest(t, p, "/* kommentar\n\n*/", true, "/* kommentar\n\n*/", "/* kommentar\n\n*/", nil)
	parserTest(t, p, "/* kommentar\n\n/* nested broken comment\n  \n \r*/", true, "/* kommentar\n\n/* nested broken comment\n  \n \r*/", "/* kommentar\n\n/* nested broken comment\n  \n \r*/", nil)
	parserTest(t, p, "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/", true, "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/", "/* kommentar\n\n/* nested complete comment\n*/  \n \r*/", nil)
}

func TestNumber(t *testing.T) {
	p := number()
	parserTest(t, p, "0", true, "0", "0", 0)
	parserTest(t, p, "01", true, "01", "01", 1)
	parserTest(t, p, "20", true, "20", "20", 20)
	parserTest(t, p, "344", true, "344", "344", 344)
	parserTest(t, p, "34234.01", true, "34234.01", "34234.01", 34234.01)
	parserTest(t, p, "2134.11", true, "2134.11", "2134.11", 2134.11)
	parserTest(t, p, "0.131", true, "0.131", "0.131", 0.131)
	parserTest(t, p, ".2", false, "", "", nil)
	parserTest(t, p, "f", false, "", "", nil)
	parserTest(t, p, "f2", false, "", "", nil)
	parserTest(t, p, "f.2", false, "", "", nil)
	parserTest(t, p, "2.f", true, "2", "2", 2)
	parserTest(t, p, "2.01x", true, "2.01", "2.01", 2.01)
	parserTest(t, p, "2.01.1", true, "2.01", "2.01", 2.01)
}

func TestCount(t *testing.T) {
	p := count(digit(), 3)
	parserTest(t, p, "0", false, "", "", nil)
	parserTest(t, p, "aaa", false, "", "", nil)
	parserTest(t, p, "012", true, "012", "012", nil)
	parserTest(t, p, "0123", false, "", "", nil)
}

func TestStringLiteral(t *testing.T) {
	p := stringLiteral()
	parserTest(t, p, "\"a\"", true, "a", "\"a\"", "a")
	parserTest(t, p, "\"asdf asdf a sdf\"", true, "asdf asdf a sdf", "\"asdf asdf a sdf\"", "asdf asdf a sdf")
	parserTest(t, p, "\"\\x42\"", true, "B", "\"\\x42\"", "B")
	parserTest(t, p, "\"\\u4142\"", true, "\u4142", "\"\\u4142\"", "\u4142")
	parserTest(t, p, "\"\\U0002070E\"", true, "\U0002070E", "\"\\U0002070E\"", "\U0002070E")
	parserTest(t, p, "\"\\a\\b\\f\\n\\r\\t\\v\\\\\\\"\\x42\\u4142\\U0002070Ehej\"", true, "\a\b\f\n\r\t\v\\\"B\u4142\U0002070Ehej", "\"\\a\\b\\f\\n\\r\\t\\v\\\\\\\"\\x42\\u4142\\U0002070Ehej\"", "\a\b\f\n\r\t\v\\\"B\u4142\U0002070Ehej")
}
