package main

import (
	"flag"
	"fmt"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	if i == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	relativeRoot := flag.String("r", "", "path to project absoluteRoot")
	compileProtos := flag.Bool("gogo", true, "compile normal gogo protos")
	genDocs := flag.Bool("docs", true, "generate docs as well")
	var customImports, skipDirs arrayFlags
	flag.Var(&customImports, "i", "import additional directories as proto roots")
	flag.Var(&skipDirs, "s", "skip generating for this directory")
	flag.Parse()

	if err := cmd.Run(*relativeRoot, *compileProtos, *genDocs, customImports, skipDirs); err != nil {
		log.Fatalf("%v", err)
	}
}
