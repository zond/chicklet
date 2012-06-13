# chicklet

Go in Go.

## What

This is a slightly modified version of https://bitbucket.org/binet/go-eval that aims towards simplifying evaluating and interacting with Go expressions from Go programs.

## How

    context := chicklet.NewContext()
    context.Eval("func add(i, j int) int { return i + j }")
    fmt.Println(context.Eval("add(2, 3)"))

Look at https://github.com/zond/chicklet/blob/master/example.go for an example of more things you can do.

Look at https://github.com/zond/chicklet/blob/master/chicklet_test.go for a full definition of what I consider working.
