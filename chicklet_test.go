
package chicklet

import (
	"testing"
	"math/big"
)

func TestIntReturn(t *testing.T) {
	c := NewContext()
	s := "func() int { return 1 + 2 }()"
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			if val[0].(int) != 3 {
				t.Error(s, "should generate 3")
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}

func TestStringReturn(t *testing.T) {
	c := NewContext()
	s := "\"bla\""
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			if val[0].(string) != "bla" {
				t.Error(s, "should generate \"bla\" but generated", val[0])
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}

func TestIdealFloatReturn(t *testing.T) {
	c := NewContext()
	s := "1.0 * 4.1"
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			cmp := big.NewRat(41, 10)
			if cmp.Cmp(val[0].(*big.Rat)) != 0 {
				t.Error(s, "should generate", cmp, "but generated", val[0])
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}

func TestBoolReturn(t *testing.T) {
	c := NewContext()
	s := "1 == 1"
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			cmp := true
			if val[0].(bool) != cmp {
				t.Error(s, "should generate", cmp, "but generated", val[0])
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}

func TestFunc0_0Return(t *testing.T) {
	c := NewContext()
	s := "func() {}"
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			rval, err := val[0].(Callable).Call()
			if err == nil {
				if len(rval) != 0 {
					t.Error(s, "should not return anything when called, returned", rval)
				}
			} else {
				t.Error(s, "should be callable, got", err)
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}

func TestFunc0_1Return(t *testing.T) {
	c := NewContext()
	s := "func() int { return 1 }"
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Call()
		if err == nil {
			if len(val) != 1 {
				t.Error(s, "should generate one value, generated", len(val))
			}
			rval, err := val[0].(Callable).Call()
			if err == nil {
				if len(rval) != 1 {
					t.Error(s, "should not return one value when called, returned", len(rval))
				}
				if rval[0].(int) != 1 {
					t.Error(s, "should return 1 when called, returned", rval[0])
				}
			} else {
				t.Error(s, "should be callable, got", err)
			}
		} else {
			t.Error(s, "should run, got", err)
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}
