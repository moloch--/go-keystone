package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/For-ACGN/go-keystone"
)

var (
	arch    keystone.Arch
	mode    keystone.Mode
	syntax  keystone.OptionValue
	address uint64
	srcPath string
	output  string
)

func init() {
	flag.UintVar(&arch, "arch", keystone.ARCH_X86, "set the target architecture")
	flag.UintVar(&mode, "mode", keystone.MODE_32, "set the target mode")
	flag.UintVar(&syntax, "syntax", keystone.OPT_SYNTAX_INTEL, "set the assembly syntax")
	flag.Uint64Var(&address, "addr", 0, "set the base address")
	flag.StringVar(&srcPath, "src", "", "set the path to the source file")
	flag.StringVar(&output, "out", "", "set the output file path")
	flag.Parse()
}

func main() {
	engine, err := keystone.NewEngine(arch, mode)
	checkError(err)
	defer func() { _ = engine.Close() }()
	err = engine.Option(keystone.OPT_SYNTAX, syntax)
	checkError(err)

	src, err := os.ReadFile(srcPath)
	checkError(err)
	inst, err := engine.Assemble(string(src), address)
	checkError(err)
	err = os.WriteFile(output, inst, 0644)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
