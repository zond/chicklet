package main

import (
	"github.com/zond/chicklet"
	"fmt"
)

func nativeTest(i int64, s string) int {
	fmt.Println("nativeTest called with", i, s)
	return int(i) + len(s)
}

func main() {
	context := chicklet.NewWorld()
	context.Define("myNative", nativeTest)
	context.Eval("func callNative(s string, i int64) int { return myNative(i, s) }")
	evalFunc := context.Eval("func(s string, i int64) (int, string) { return callNative(s, i), s }").(chicklet.Executable)
	r, _ := evalFunc.Execute("hello world", 17)
	fmt.Printf("result is %v of type %T\n", r, r)
}