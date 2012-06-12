
package chicklet

import (
	"testing"
)

func TestSimpleEvaluation(t *testing.T) {
	c := NewContext()
	s := "func() int { return 1 + 2 }()"
	code, err := c.Compile(s)
	if err == nil {
		val, err := code.Run()
		if err == nil {
			v, ok := val.(*intV)
			if ok {
				if *v != 3 {
					t.Error(s, "should generate 3")
				}
			} else {
				t.Error(s, "should generate an int")
			}
		} else {
			t.Error(s, "should run")
		}
	} else {
		t.Error(s, "should compile, got", err)
	}
}
