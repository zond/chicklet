
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

/*
func TestFunctionReturn(t *testing.T) {
	c := NewContext()
	s := "func() int { return 1 + 2 }"
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Run()
		if err == nil {
			fmt.Printf("%v\n", val)

			 v, ok := val.(*funcV)
			 if ok {
			 t := &Thread{}
			 fun := v.Get(t)
			 f := fun.NewFrame()
			 
			 t.f = f
			 fun.Call(t)
			 fmt.Println(t.f)
						} else {
				t.Error(s, "should generate a function")
			}

		} else {
			t.Error(s, "should run")
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}
*/