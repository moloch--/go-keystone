package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/moloch--/go-keystone"
	"github.com/spf13/cobra"
)

var (
	archS   string
	modeS   string
	syntaxS string
	address uint64
	srcPath string
	output  string
)

const supportedOptions = `
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

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go-keystone",
		Short: "Assemble source files with the Keystone engine",
		Long:  "Assemble source files with the Keystone engine." + supportedOptions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			if srcPath == "" || output == "" {
				return errors.New("both --src and --out must be specified")
			}
			return assemble()
		},
	}

	cmd.Flags().StringVar(&archS, "arch", "x86", "set the target architecture")
	cmd.Flags().StringVar(&modeS, "mode", "32", "set the target mode")
	cmd.Flags().StringVar(&syntaxS, "syntax", "intel", "set the assembly syntax")
	cmd.Flags().Uint64Var(&address, "addr", 0, "set the base address")
	cmd.Flags().StringVar(&srcPath, "src", "", "set the path to the source file")
	cmd.Flags().StringVar(&output, "out", "", "set the output file path")

	if err := cmd.MarkFlagRequired("src"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("out"); err != nil {
		panic(err)
	}

	completionCmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion scripts",
		Long:      "Generate shell completion scripts for supported shells.",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
	cmd.AddCommand(completionCmd)

	return cmd
}

func assemble() error {
	arch := keystone.StringToArch(archS)
	mode := keystone.StringToMode(modeS)
	syntax := keystone.StringToSyntax(syntaxS)

	engine, err := keystone.NewEngine(arch, mode)
	if err != nil {
		return err
	}
	defer func() { _ = engine.Close() }()

	if err := engine.Option(keystone.OPT_SYNTAX, syntax); err != nil {
		return err
	}

	src, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	inst, err := engine.Assemble(string(src), address)
	if err != nil {
		return err
	}

	if err := os.WriteFile(output, inst, 0644); err != nil {
		return err
	}

	return nil
}
