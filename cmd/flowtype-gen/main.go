package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/brian14708/go-flow/flowtype/codegen"
)

var (
	flagBuiltin = flag.Bool("builtin", false, "generate code for builtin types")
	flagTest    = flag.Bool("test", false, "generate test code")
	flagPkgName = flag.String("pkg", "", "package name")
	flagOutput  = flag.String("output", "", "output path")
)

func main() {
	flag.Parse()

	if *flagPkgName == "" || *flagOutput == "" {
		panic("missing arguments")
	}

	src, err := codegen.Generate(codegen.GenerateOptions{
		PkgName:   *flagPkgName,
		Builtin:   *flagBuiltin,
		TypeNames: flag.Args(),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to generate code: %v", err))
	}

	err = os.WriteFile(*flagOutput, src, 0o644)
	if err != nil {
		panic(fmt.Sprintf("failed to write file: %v", err))
	}

	if *flagTest {
		test, err := codegen.GenerateTest(codegen.GenerateOptions{
			PkgName:   *flagPkgName,
			Builtin:   *flagBuiltin,
			TypeNames: flag.Args(),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to generate test code: %v", err))
		}
		fn := *flagOutput
		fn = fn[:len(fn)-3] + "_test.go"
		err = os.WriteFile(fn, test, 0o644)
		if err != nil {
			panic(fmt.Sprintf("failed to write file: %v", err))
		}
	}
}
