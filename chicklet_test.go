
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

func TestDefineStruct(t *testing.T) {
	defineTest(t, testStruct{1, "hello"})
}

func TestDefineStructPtr(t *testing.T) {
	defineTest(t, &testStruct{1, "hello"})
}

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
	result, err := c.Eval(s)
	if err == nil {
		if result != 11 {
			t.Error(s, "should return 11 when called, returned", result)
		}
		s = "func() int { return testFunc() }"
		result, err = c.Eval(s)
		if err == nil {
			rval, err := result.(Executable).Execute()
			if err == nil {
				if len(rval) == 1 {
					if rval[0] != 11 {
						t.Error(s, "should return 11 when called, returned", result)
					}
				} else {
					t.Error(s, "should return one arg when called, returned", len(rval))
				}
			} else {
				t.Error(s, "should be executable, but got", err)
			}
		} else {
			t.Error(s, "should be evaluable, got", err)
		}
	} else {
		t.Error(s, "should be evaluable, got", err)
	}
}

func defineTest(t *testing.T, value Thing) {
 	c := NewWorld()
	c.Define("testDef", value)
	evalTest(t, c, "testDef", value)
}

func evalFuncCallTest(t *testing.T, decl string, args, expect []Thing) {
	c := NewWorld()
	result, err := c.Eval(decl)
	if err == nil {
		rval, err := result.(Executable).Execute(args...)
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
			t.Error(decl, "should be executable, but got", err)
		}
	} else {
		t.Error(decl, "should be evaluable, got", err)
	}
}

func evalTest(t *testing.T, c *World, s string, exp Thing) {
	val, err := c.Eval(s)
	if err == nil {
		if val == nil {
			if exp != nil {
				t.Error(fmt.Sprintf("%v should generate %v of type %T but generated %v of type %T\n", s, exp, exp, val, val))
			}
		} else {
			v := reflect.ValueOf(exp)
			typ := reflect.TypeOf(exp)
			if typ.Kind() == reflect.Ptr && v.Elem().Type().Name() != "Rat" {
				checkEquality(t, s, v.Elem().Interface(), reflect.ValueOf(val).Elem().Interface())
			} else {
				checkEquality(t, s, exp, val)
			}
		}
	} else {
		t.Error(s, "should be evaluable, got", err)
	}
}

func checkEquality(t *testing.T, s string, exp, val Thing) {
	if exp != val && !reflect.DeepEqual(exp, val) {
		t.Error(fmt.Sprintf("%s should generate %v of type %T but generated %v of type %T\n", 
			s, exp, exp, val, val))
	}
}

func evalTestReturn(t *testing.T, s string, exp Thing) {
	c := NewWorld()
	evalTest(t, c, s, exp)
}

type testStruct struct {
	I int
	S string
}

type testStruct2 struct {
	I int
	F float64
	S *testStruct
}