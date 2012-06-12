
package chicklet

import (
	"go/token"
)

type Context struct {
	world *World
	fset *token.FileSet
}
func NewContext() *Context {
	return &Context{NewWorld(), token.NewFileSet()}
}
func (self *Context) Compile(s string) (Code, error) {
	return self.world.Compile(self.fset, s)
}