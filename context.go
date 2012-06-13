
package chicklet

import (
	"go/token"
	"fmt"
)

type Thing interface{}

type Callable interface {
	Call(things... Thing) ([]Thing, error)
}

type Context struct {
	world *World
	fset *token.FileSet
}
func NewContext() *Context {
	return &Context{NewWorld(), token.NewFileSet()}
}
func (self *Context) Compile(s string) (Callable, error) {
	code, err := self.world.Compile(self.fset, s)
	if err == nil {
		return &Compiled{code}, nil
	}
	return nil, err
}

type Compiled struct {
	code Code
}
func (self *Compiled) Call(things... Thing) (rval []Thing, err error) {
	r, err := self.code.Run()
	if err == nil {
		rval, err := convert(r)
		if err == nil {
			return rval, nil
		} else {
			return nil, err
		}
	}
	return nil, err
}

type ConvertError struct {
	Message string
}
func (self *ConvertError) Error() string {
	return self.Message
}

func convert(things... Thing) (rval []Thing, err error) {
	for _, t := range things {
		c, err := convertOne(t)
		if err == nil {
			rval = append(rval, c)
		} else {
			return nil, err
		}
	}
	return rval, nil
}

func convertOne(t Thing) (rval Thing, err error) {
	switch t.(type) {
	case *intV: return int(*(t.(*intV))), nil
	case *stringV: return string(*(t.(*stringV))), nil
	case *idealFloatV: return (*(t.(*idealFloatV))).Get(), nil
	case *boolV: return bool(*(t.(*boolV))), nil
	}
	return nil, &ConvertError{fmt.Sprintf("Unable to convert %v of type %T", t, t)}
}