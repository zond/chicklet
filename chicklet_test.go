
package chicklet

import (
	"testing"
	"math/big"
	"reflect"
)

func evalTest(t *testing.T, s string, exp Thing) {
	c := NewContext()
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			if exp != val[0] && !reflect.DeepEqual(exp, val[0]) {
				t.Error(s, "should generate", exp, "but generated", val[0])
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}

func TestIntReturn(t *testing.T) {
	evalTest(t, "func() int { return 1 + 2 }()", 3)
}

func TestStringReturn(t *testing.T) {
	evalTest(t, "\"bla\"", "bla")
}

func TestIdealFloatReturn(t *testing.T) {
	evalTest(t, "1.0 * 4.1", big.NewRat(41, 10))
}

func TestBoolReturn(t *testing.T) {
	evalTest(t, "1 == 1", true)
}

func defineTest(t *testing.T, value Thing) {
 	c := NewContext()
	c.Define("testDef", value)
	s := "testDef"
	code, err := c.Compile("testDef")
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			if value != val[0] && !reflect.DeepEqual(value, val[0]) {
				t.Error(s, "should generate", value, "but generated", val[0])
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
	
}

func TestDefineString(t *testing.T) {
	defineTest(t, "str")
}

func TestDefineInt(t *testing.T) {
	defineTest(t, 14)
}

func TestDefineFloat(t *testing.T) {
	defineTest(t, 0.12)
}

func TestEvalFunc0_0Return(t *testing.T) {
	evalFuncCallTest(t, "func() {}", []Thing{}, []Thing{})
}

func TestEvalFunc2_1Return(t *testing.T) {
	evalFuncCallTest(t, "func(i, j int) int { return i * j }", []Thing{2,5}, []Thing{10})
}

func TestEvalFunc1_1Return(t *testing.T) {
	evalFuncCallTest(t, "func(i int) int { return i * 2 }", []Thing{2}, []Thing{4})
}

func TestEvalFunc0_1Return(t *testing.T) {
	evalFuncCallTest(t, "func() int { return 1 }", []Thing{}, []Thing{1})
}

func TestEvalFunc0_2Return(t *testing.T) {
	evalFuncCallTest(t, "func() (a,b int) { return 1, 2 }", []Thing{}, []Thing{1,2})
}

func evalFuncCallTest(t *testing.T, decl string, args, expect []Thing) {
	c := NewContext()
	code, err := c.Compile(decl)
	if err == nil {
		result, err := code.Call()
		if err == nil {
			rval, err := result[0].(Callable).Call(args...)
			if err == nil {
				if len(rval) != len(expect) {
					t.Error(decl, "should return", len(expect), "values when called with", args, ", returned", len(rval))
				}
				for index, val := range rval {
					if val != expect[index] {
						t.Error(decl, "should return", expect, "when called with", args, "returned, ", rval)
					}
				}
			} else {
				t.Error(decl, "should be callable with", args, ", got", err)
			}
		} else {
			t.Error(decl, "should be callable, got", err)
		}
	} else {
		t.Error(decl, "should compile, got", err)
	}
}