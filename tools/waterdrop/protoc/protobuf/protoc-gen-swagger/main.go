package main

import (
	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/protoc/protobuf/pkg/gen"
)

func main() {
	g := NewSwaggerGenerator()
	gen.Main(g)
}
