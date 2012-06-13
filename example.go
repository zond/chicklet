package main

import (
	"github.com/zond/chicklet"
	"fmt"
)

func nativeTest(i int, s string) int {
	fmt.Println("nativeTest called with", i, s)
	return i + len(s)
}

func main() {
	context := chicklet.NewContext()
	context.Define("myNative", nativeTest)
	context.Eval("func callNative(s string, i int) int { return myNative(i, s) }")
	result := context.Eval("func(s string, i int) (int, string) { return callNative(s, i), s }")
	r, _ := result[0].(chicklet.Callable).Call("hello world", 17)
	fmt.Printf("result is %v of type %T\n", r, r)
}