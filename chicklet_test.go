
package chicklet

import (
	"testing"
)

func TestSimpleEvaluation(t *testing.T) {
	c := NewContext()
	s := "1 + 2"
	code, err := c.Compile(s)
	if err != nil {
		t.Error(s, "should compile")
	}
	val, err := code.Run()
	if err != nil {
		t.Error(s, "should run")
	}
	v, ok := val.(IdealIntValue); 
	if !ok {
		t.Error(s, "should generate an IdealIntValue")
	}
	if v.Get().Int64() != 3 {
		t.Error(s, "should generate 3")
	}
}
