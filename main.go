package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal"
)

func main() {
	var gen internal.Generator
	var flags flag.FlagSet

	flags.BoolVar(&gen.Verbose, "verbose", false, "enable verbose logging")

	opt := protogen.Options{ParamFunc: flags.Set}
	opt.Run(gen.Run)
}
