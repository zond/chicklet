// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package eval is the beginning of an interpreter for Go.
// It can run simple Go programs but does not implement
// interface values or packages.
package chicklet

import (
	"bitbucket.org/binet/go-types/pkg/types"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"regexp"
	"strconv"
	"reflect"
)

// track the status of each package we visit (unvisited/visiting/done)
var g_visiting = make(map[string]status)

type status int // status for visiting map
const (
	unvisited status = iota
	visiting
	done
)

type Thing interface{}

type Callable interface {
	Call(things... Thing) ([]Thing, error)
}


type Compiled struct {
	world *World
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
	switch val.Kind() {
	case reflect.Func:
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

type World struct {
	scope *Scope
	frame *Frame
	inits []Code
}

func NewWorld() *World {
	w := new(World)
	w.scope = universe.ChildScope()
	w.scope.global = true // this block's vars allocate directly
	return w
}

type Code interface {
	// The type of the value Run returns, or nil if Run returns nil.
	Type() Type

	// Run runs the code; if the code is a single expression
	// with a value, it returns the value; otherwise it returns nil.
	Run() (Value, error)
}

type pkgCode struct {
	w    *World
	code code
}

func (p *pkgCode) Type() Type { return nil }

func (p *pkgCode) Run() (Value, error) {
	t := new(Thread)
	t.f = p.w.scope.NewFrame(nil)
	return nil, t.Try(func(t *Thread) { p.code.exec(t) })
}

func (w *World) CompilePackage(fset *token.FileSet, files []*ast.File, pkgpath string) (Code, error) {
	pkgFiles := make(map[string]*ast.File)
	for _, f := range files {
		pkgFiles[f.Name.Name] = f
	}
	pkg, err := ast.NewPackage(fset, pkgFiles, srcImporter, types.Universe)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		return nil, errors.New("could not create an AST package out of ast.Files")
	}

	switch g_visiting[pkgpath] {
	case done:
		return &pkgCode{w, make(code, 0)}, nil
	case visiting:
		//fmt.Printf("** package dependency cycle **\n")
		return nil, errors.New("package dependency cycle")
	}
	g_visiting[pkgpath] = visiting
	// create a new scope in which to process this new package
	imports := []*ast.ImportSpec{}
	for _, f := range pkg.Files {
		imports = append(imports, f.Imports...)
	}

	for _, imp := range imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		if _, ok := universe.pkgs[path]; ok {
			// already compiled
			continue
		}
		imp_files, err := findPkgFiles(path)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("could not find files for package [%s]", path))
		}
		code, err := w.CompilePackage(fset, imp_files, path)
		if err != nil {
			return nil, err
		}
		_, err = code.Run()
		if err != nil {
			return nil, err
		}
	}

	prev_scope := w.scope
	w.scope = w.scope.ChildScope()
	w.scope.global = true
	defer func() {
		g_visiting[pkgpath] = done
		// add this scope (the package's scope) to the lookup-table of packages
		universe.pkgs[pkgpath] = w.scope
		// restore the previous scope
		w.scope.exit()
		if pkgpath != "main" {
			w.scope = prev_scope
		}
	}()

	decls := make([]ast.Decl, 0)
	for _, f := range pkg.Files {
		decls = append(decls, f.Decls...)
	}
	code, err := w.CompileDeclList(fset, decls)
	if err != nil {
		return nil, err
	}
	_, err = code.Run()
	if err != nil {
		return nil, err
	}

	//FIXME: make sure all the imports are used at this point.

	{
		// store the init function (if any) for later use
		init_code, init_err := w.Compile(fset, "init()")
		if init_code != nil {
			if init_err == nil || init_err != nil {
				w.inits = append(w.inits, init_code)
			}
		}
	}
	return code, err
}

type stmtCode struct {
	w    *World
	code code
}

func (w *World) CompileStmtList(fset *token.FileSet, stmts []ast.Stmt) (Code, error) {
	if len(stmts) == 1 {
		if s, ok := stmts[0].(*ast.ExprStmt); ok {
			return w.CompileExpr(fset, s.X)
		}
	}
	errors := new(scanner.ErrorList)
	cc := &compiler{fset, errors, 0, 0}
	cb := newCodeBuf()
	fc := &funcCompiler{
		compiler:     cc,
		fnType:       nil,
		outVarsNamed: false,
		codeBuf:      cb,
		flow:         newFlowBuf(cb),
		labels:       make(map[string]*label),
	}
	bc := &blockCompiler{
		funcCompiler: fc,
		block:        w.scope.block,
	}
	nerr := cc.numError()
	for _, stmt := range stmts {
		bc.compileStmt(stmt)
	}
	fc.checkLabels()
	if nerr != cc.numError() {
		errors.Sort()
		return nil, errors.Err()
	}
	return &stmtCode{w, fc.get()}, nil
}

func (w *World) CompileDeclList(fset *token.FileSet, decls []ast.Decl) (Code, error) {
	stmts := make([]ast.Stmt, len(decls))
	for i, d := range decls {
		stmts[i] = &ast.DeclStmt{d}
	}
	return w.CompileStmtList(fset, stmts)
}

func (s *stmtCode) Type() Type { return nil }

func (s *stmtCode) Run() (Value, error) {
	t := new(Thread)
	t.f = s.w.scope.NewFrame(nil)
	return nil, t.Try(func(t *Thread) { s.code.exec(t) })
}

type exprCode struct {
	w    *World
	e    *expr
	eval func(Value, *Thread)
}

func (w *World) CompileExpr(fset *token.FileSet, e ast.Expr) (Code, error) {
	errors := new(scanner.ErrorList)
	cc := &compiler{fset, errors, 0, 0}

	ec := cc.compileExpr(w.scope.block, false, e)
	if ec == nil {
		errors.Sort()
		return nil, errors.Err()
	}
	var eval func(Value, *Thread)
	switch t := ec.t.(type) {
	case *idealIntType:
		// nothing
	case *idealFloatType:
		// nothing
	default:
		if tm, ok := t.(*MultiType); ok && len(tm.Elems) == 0 {
			return &stmtCode{w, code{ec.exec}}, nil
		}
		eval = genAssign(ec.t, ec)
	}
	return &exprCode{w, ec, eval}, nil
}

func (e *exprCode) Type() Type { return e.e.t }

func (e *exprCode) Run() (Value, error) {
	t := new(Thread)
	t.f = e.w.scope.NewFrame(nil)
	switch e.e.t.(type) {
	case *idealIntType:
		return &idealIntV{e.e.asIdealInt()()}, nil
	case *idealFloatType:
		return &idealFloatV{e.e.asIdealFloat()()}, nil
	}
	v := e.e.t.Zero()
	eval := e.eval
	err := t.Try(func(t *Thread) { eval(v, t) })
	return v, err
}

func (w *World) run_init() error {
	// run the 'init()' function of all dependent packages
	for _, init_code := range w.inits {
		_, err := init_code.Run()
		if err != nil {
			return err
		}
	}
	// reset
	w.inits = make([]Code, 0)
	return nil
}

// Regexp to match the import keyword
var import_regexp = regexp.MustCompile("[ \t\n]*import[ \t(]")

func (w *World) Compile(fset *token.FileSet, text string) (Code, error) {
	if text == "main()" {
		err := w.run_init()
		if err != nil {
			return nil, err
		}
	}
	if i := import_regexp.FindStringIndex(text); i != nil && i[0] == 0 {
		// special case for import-ing on the command line...
		return w.compileImport(fset, text)
	}

	stmts, err := parseStmtList(fset, text)
	if err == nil {
		return w.CompileStmtList(fset, stmts)
	}

	// Otherwise try as DeclList
	decls, err1 := parseDeclList(fset, text)
	if err1 == nil {
		return w.CompileDeclList(fset, decls)
	}

	// Have to pick an error.
	// Parsing as statement list admits more forms,
	// its error is more likely to be useful.
	return nil, err
}

var defaultFileSet = token.NewFileSet()

func (self *World) Comp(s string) (Callable, error) {
	code, err := self.Compile(defaultFileSet, s)
	if err == nil {
		return &Compiled{self, code}, nil
	}
	return nil, err
}

func (self *World) Define(name string, thing Thing) {
	val, err := convertOne(thing)
	if err != nil {
		panic(fmt.Sprint("Unable to define ", name, " to ", thing, ": ", err))
	}
	v := val.(Value)
	self.DefineVar(name, TypeFromNative(reflect.TypeOf(thing)), v)
}

func (self *World) Eval(s string) []Thing {
	code, err := self.Comp(s)
	if err != nil {
		panic(err)
	}
	result, err := code.Call()
	if err != nil {
		panic(err)
	}
	return result
}

func (w *World) compileImport(fset *token.FileSet, text string) (Code, error) {
	f, err := parser.ParseFile(fset, "input", "package main;"+text, 0)
	if err != nil {
		return nil, err
	}

	imp := f.Imports[0]
	path, _ := strconv.Unquote(imp.Path.Value)
	imp_files, err := findPkgFiles(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not find files for package [%s]", path))
	}
	{
		code, err := w.CompilePackage(fset, imp_files, path)
		if err != nil {
			return nil, err
		}
		_, err = code.Run()
		if err != nil {
			return nil, err
		}
		err = w.run_init()
		if err != nil {
			return nil, err
		}
	}
	return w.CompileDeclList(fset, f.Decls)
}

func parseStmtList(fset *token.FileSet, src string) ([]ast.Stmt, error) {
	f, err := parser.ParseFile(fset, "input", "package p;func _(){"+src+"\n}", 0)
	if err != nil {
		return nil, err
	}
	return f.Decls[0].(*ast.FuncDecl).Body.List, nil
}

func parseDeclList(fset *token.FileSet, src string) ([]ast.Decl, error) {
	f, err := parser.ParseFile(fset, "input", "package p;"+src, 0)
	if err != nil {
		return nil, err
	}
	return f.Decls, nil
}

type RedefinitionError struct {
	Name string
	Prev Def
}

func (e *RedefinitionError) Error() string {
	res := "identifier " + e.Name + " redeclared"
	pos := e.Prev.Pos()
	if pos.IsValid() {
		// TODO: fix this - currently this code is not reached by the tests
		//       need to get a file set (fset) from somewhere
		//res += "; previous declaration at " + fset.Position(pos).String()
		panic(0)
	}
	return res
}

func (w *World) DefineConst(name string, t Type, val Value) error {
	_, prev := w.scope.DefineConst(name, token.NoPos, t, val)
	if prev != nil {
		return &RedefinitionError{name, prev}
	}
	return nil
}

func (w *World) DefineVar(name string, t Type, val Value) error {
	v, prev := w.scope.DefineVar(name, token.NoPos, t)
	if prev != nil {
		return &RedefinitionError{name, prev}
	}
	v.Init = val
	return nil
}
