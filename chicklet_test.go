
package chicklet

import (
	"testing"
	"math/big"
	"reflect"
	"fmt"
)

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

func TestDefineString(t *testing.T) {
	defineTest(t, "str")
}

func TestDefineInt(t *testing.T) {
	defineTest(t, 14)
}

func TestDefineFloat(t *testing.T) {
	defineTest(t, 0.12)
}
/*
func TestDefineStruct(t *testing.T) {
	defineTest(t, &testStruct{1, "hello"})
}
*/
func TestDefineBool(t *testing.T) {
	defineTest(t, false)
	defineTest(t, true)
}

func nativeFuncCallTest(t *testing.T, some Thing, params string, exp Thing) {
	c := NewWorld()
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

func TestEvalFuncEval(t *testing.T) {
	c := NewWorld()
	c.Eval("func testFunc() int { return 11 }")
	s := "testFunc()"
	result := c.Eval(s)
	if result[0] != 11 {
		t.Error(s, "should return 11 when called, returned", result[0])
	}
	s = "func() int { return testFunc() }"
	result = c.Eval(s)
	rval, err := result[0].(Callable).Call()
	if err == nil {
		if rval[0] != 11 {
			t.Error(s, "should return 11 when called, returned", result[0])
		}
	} else {
		t.Error(s, "should be callable, got", err)
	}
}

func defineTest(t *testing.T, value Thing) {
 	c := NewWorld()
	c.Define("testDef", value)
	evalTest(t, c, "testDef", value)
}

func evalFuncCallTest(t *testing.T, decl string, args, expect []Thing) {
	c := NewWorld()
	result := c.Eval(decl)
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
}

func evalTest(t *testing.T, c *World, s string, exp Thing) {
	val := c.Eval(s)
	if len(val) != 1 {
		t.Error(s, "should generate one value, generated", len(val))
	}
	if exp != val[0] && !reflect.DeepEqual(exp, val[0]) {
		t.Error(s, "should generate", exp, "but generated", val[0])
	}
}

func evalTestReturn(t *testing.T, s string, exp Thing) {
	c := NewWorld()
	evalTest(t, c, s, exp)
}

type testStruct struct {
	i int
	s string
}

type testStruct2 struct {
	i int
	f float64
	s *testStruct
}