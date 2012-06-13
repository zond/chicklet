
package chicklet

import (
	"go/token"
	"fmt"
	"reflect"
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
func (self *Context) Define(name string, thing Thing) {
	val, err := convertOne(thing)
	if err != nil {
		panic(fmt.Sprint("Unable to define ", name, " to ", thing, ": ", err))
	}
	v := val.(Value)
	self.world.DefineVar(name, TypeFromNative(reflect.TypeOf(thing)), v)
}
func (self *Context) Eval(s string) []Thing {
	code, err := self.Compile(s)
	if err != nil {
		panic(err)
	}
	result, err := code.Call()
	if err != nil {
		panic(err)
	}
	return result
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

type CallError struct {
	Message string
}
func (self *CallError) Error() string {
	return self.Message
}

type EvalFuncWrapper struct {
	target *evalFunc
}
func (self *EvalFuncWrapper) Call(things... Thing) (rval []Thing, err error) {
	if len(things) != len(self.target.inTypes) {
		return nil, &CallError{fmt.Sprint("Wrong number of arguments. Wanted ", len(self.target.inTypes), " but got ", len(things))}
	}
	frame := self.target.NewFrame()
	for index, thing := range things {
		val, err := convertOne(thing)
		if err != nil {
			return nil, err
		}
		frame.Vars[index] = val.(Value)
	}
	for index, t := range self.target.outTypes {
		frame.Vars[len(self.target.inTypes) + index] = t.(*NamedType).Zero()
	}
	thread := &Thread{}
	thread.f = frame
	self.target.Call(thread)
	for index, _ := range self.target.outTypes {
		val := frame.Vars[len(self.target.inTypes) + index]
		converted, err := convertOne(val)
		if (err != nil) {
			return nil, err
		}
		rval = append(rval, converted)
	}
	return rval, nil
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
	case int: 
		val := IntType.Zero()
		*(val.(*intV)) = intV(t.(int))
		return val, nil
	case string: 
		val := StringType.Zero()
		*(val.(*stringV)) = stringV(t.(string))
		return val, nil
	case float64: 
		val := Float64Type.Zero()
		*(val.(*float64V)) = float64V(t.(float64))
		return val, nil
	case bool: 
		val := BoolType.Zero()
		*(val.(*boolV)) = boolV(t.(bool))
		return val, nil
	case *intV: return int(*(t.(*intV))), nil
	case *stringV: return string(*(t.(*stringV))), nil
	case *idealFloatV: return (*(t.(*idealFloatV))).Get(), nil
	case *boolV: return bool(*(t.(*boolV))), nil
	case *float64V: return float64(*(t.(*float64V))), nil
	case *funcV: 
		switch t.(*funcV).target.(type) {
		case *evalFunc: return &EvalFuncWrapper{t.(*funcV).target.(*evalFunc)}, nil
		}
	case nil: return nil, nil
	}
	val := reflect.ValueOf(t)
	if val.Kind() == reflect.Func {
		_, fval := FuncFromNativeTyped(func(thread *Thread, in, out []Value) {
			var reflect_in []reflect.Value
			for _, inv := range in {
				converted, err := convertOne(inv)
				if err != nil {
					panic(fmt.Sprint("Unable to call ", t, "(", in, "): ", err))
				}
				reflect_in = append(reflect_in, reflect.ValueOf(converted))
			}
			reflect_out := val.Call(reflect_in)
			for index, outv := range reflect_out {
				converted, err := convertOne(outv.Interface())
				if err != nil {
					panic(fmt.Sprint("Unable to respond from call to ", t, "(", in, ") with ", reflect_out, ": ", err))
				}
				out[index] = converted.(Value)
			}
		}, t)
		return fval, nil
	}
	return nil, &ConvertError{fmt.Sprintf("Unable to convert %v of type %T", t, t)}
}