package main

import (
	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/protobuf/pkg/gen"
)

func main() {
	g := NewSwaggerGenerator()
	gen.Main(g)
}
