package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal"
)

func main() {
	var gen internal.Plugin
	var flags flag.FlagSet

	flags.BoolVar(&gen.Verbose, "verbose", false, "enable verbose logging")
	flags.StringVar(&gen.PrivatePackageName, "private_package", "private", "name of private service package")

	opt := protogen.Options{ParamFunc: flags.Set}
	opt.Run(gen.Run)
}
