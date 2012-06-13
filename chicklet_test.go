
package chicklet

import (
	"testing"
	"math/big"
	"reflect"
	"fmt"
)

func evalTest(t *testing.T, c *Context, s string, exp Thing) {
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

func evalTestReturn(t *testing.T, s string, exp Thing) {
	c := NewContext()
	evalTest(t, c, s, exp)
}

func TestIntReturn(t *testing.T) {
	evalTestReturn(t, "func() int { return 1 + 2 }()", 3)
}

func TestStringReturn(t *testing.T) {
	evalTestReturn(t, "\"bla\"", "bla")
}

func TestIdealFloatReturn(t *testing.T) {
	evalTestReturn(t, "1.0 * 4.1", big.NewRat(41, 10))
}

func TestBoolReturn(t *testing.T) {
	evalTestReturn(t, "1 == 1", true)
}

func defineTest(t *testing.T, value Thing) {
 	c := NewContext()
	c.Define("testDef", value)
	evalTest(t, c, "testDef", value)
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

func TestDefineBool(t *testing.T) {
	defineTest(t, false)
	defineTest(t, true)
}

func nativeFuncCallTest(t *testing.T, some Thing, params string, exp Thing) {
	c := NewContext()
	c.Define("testFun", some)
	evalTest(t, c, fmt.Sprint("testFun(", params, ")"), exp)
}

func TestDefineFunc0_0(t *testing.T) {
	nativeFuncCallTest(t, func() {}, "", nil)
}

func TestDefineFunc0_1(t *testing.T) {
	nativeFuncCallTest(t, func() int { return 1 }, "", 1)
}

func TestDefineFunc1_1(t *testing.T) {
	nativeFuncCallTest(t, func(i int) int { return i - 1 }, "3", 2)
}

func TestDefineFunc2_1(t *testing.T) {
	nativeFuncCallTest(t, func(i int, f float64) float64 { return float64(i) + f }, "3, 0.4", 3.4)
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