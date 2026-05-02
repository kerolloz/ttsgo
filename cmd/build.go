package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"nego/internal/logger"
	"nego/internal/orchestrator"
)

// BuildFlags holds all flags for the build command.
type BuildFlags struct {
	ConfigPath   string
	TsConfigPath string
	Watch        bool
	WatchAssets  bool
}

func NewBuildCmd() *cobra.Command {
	var flags BuildFlags

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the Nest application using tsgo",
		Long: `Compiles your NestJS TypeScript project using tsgo (TypeScript 7.0's native
Go-based compiler), rewrites path aliases, and copies assets to the output directory.

This command is a drop-in replacement for 'nest build' (tsc builder only).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "", "Path to nest-cli.json (auto-detected if omitted)")
	cmd.Flags().StringVarP(&flags.TsConfigPath, "path", "p", "", "Path to tsconfig file (overrides nest-cli.json)")
	cmd.Flags().BoolVarP(&flags.Watch, "watch", "w", false, "Watch mode – rebuild on source changes")
	cmd.Flags().BoolVar(&flags.WatchAssets, "watchAssets", false, "Also watch non-TypeScript asset files")

	return cmd
}

func runBuild(flags BuildFlags) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	o, err := orchestrator.New(orchestrator.Options{
		Cwd:          cwd,
		ConfigPath:   flags.ConfigPath,
		TsConfigPath: flags.TsConfigPath,
		Watch:        flags.Watch,
		WatchAssets:  flags.WatchAssets,
	})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-done:
			return
		case <-ctx.Done():
		}
		logger.Step("Shutting down...")

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		select {
		case <-done:
		case <-sigCh:
			os.Exit(1)
		}
	}()

	if flags.Watch {
		return o.Watch(ctx, flags.WatchAssets)
	}
	return o.Build(ctx, false)
}
