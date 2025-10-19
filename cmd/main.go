package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/moloch--/go-keystone"
)

var (
	archS   string
	modeS   string
	syntaxS string
	address uint64
	srcPath string
	output  string
)

func init() {
	flag.StringVar(&archS, "arch", "x86", "set the target architecture")
	flag.StringVar(&modeS, "mode", "32", "set the target mode")
	flag.StringVar(&syntaxS, "syntax", "intel", "set the assembly syntax")
	flag.Uint64Var(&address, "addr", 0, "set the base address")
	flag.StringVar(&srcPath, "src", "", "set the path to the source file")
	flag.StringVar(&output, "out", "", "set the output file path")

	usage := flag.Usage
	flag.Usage = func() {
		args := `
  +---------+----------+---------+
  |  arch   |   mode   | syntax  |
  +---------+----------+---------+
  | arm     | le       | intel   |
  | arm64   | be       | att     |
  | mips    | arm      | nasm    |
  | x86     | thumb    | masm    |
  | ppc     | v8       | gas     |
  | sparc   | micro    | radix16 |
  | systemz | mips3    |         |
  | hexagon | mips32r6 |         |
  | evm     | mips32   |         |
  | riscv   | mips64   |         |
  | max     | 16       |         |
  |         | 32       |         |
  |         | 64       |         |
  |         | ppc32    |         |
  |         | ppc64    |         |
  |         | qpx      |         |
  |         | riscv32  |         |
  |         | riscv64  |         |
  |         | sparc32  |         |
  |         | sparc64  |         |
  |         | v9       |         |
  +---------+----------+---------+

`
		fmt.Print(args)
		usage()
	}
	flag.Parse()
}

func main() {
	if srcPath == "" || output == "" {
		flag.Usage()
		return
	}

	arch := keystone.StringToArch(archS)
	mode := keystone.StringToMode(modeS)
	syntax := keystone.StringToSyntax(syntaxS)

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
