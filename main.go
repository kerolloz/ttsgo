package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"nego/cmd"
)

var version = "0.1.0"

func main() {
	root := &cobra.Command{
		Use:   "nego",
		Short: "⚡ Ultra-fast drop-in replacement for nest build & nest start",
		Long: `nego is a drop-in replacement for 'nest build' and 'nest start'.

It uses tsgo — TypeScript 7.0's native Go-based compiler — instead of tsc,
giving you ~10x faster builds. It reads your existing nest-cli.json and
tsconfig.json, so no configuration changes are required.

Supported nest-cli.json fields:
  sourceRoot, entryFile, exec
  compilerOptions.tsConfigPath
  compilerOptions.deleteOutDir
  compilerOptions.assets
  compilerOptions.watchAssets

Not supported (out of scope):
  swc / webpack builders, plugins, generate, new, info, add`,
		Version: version,
	}

	root.AddCommand(cmd.NewBuildCmd())
	root.AddCommand(cmd.NewStartCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
