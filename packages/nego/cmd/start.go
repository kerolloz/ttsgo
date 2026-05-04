package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kerollosmagdy/nego/internal/logger"
	"github.com/kerollosmagdy/nego/internal/orchestrator"
	"github.com/kerollosmagdy/nego/internal/process"
)

// StartFlags holds all flags for the start command.
type StartFlags struct {
	BuildFlags
	Debug      string
	Exec       string
	EntryFile  string
	SourceRoot string
	NoShell    bool
	EnvFiles   []string
}

func NewStartCmd() *cobra.Command {
	var flags StartFlags

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Nest application (build + run)",
		Long: `Compiles your NestJS TypeScript project with tsgo and starts it with Node.js.

Supports watch mode (--watch) for hot-reload during development.
This command is a drop-in replacement for 'nest start' (tsc builder only).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "", "Path to nest-cli.json")
	cmd.Flags().StringVarP(&flags.TsConfigPath, "path", "p", "", "Path to tsconfig file")
	cmd.Flags().BoolVarP(&flags.Watch, "watch", "w", false, "Watch mode (live-reload)")
	cmd.Flags().BoolVar(&flags.WatchAssets, "watchAssets", false, "Watch non-TypeScript asset files")
	cmd.Flags().StringVarP(&flags.Debug, "debug", "d", "", "Run with --inspect[=host:port]")
	cmd.Flags().Lookup("debug").NoOptDefVal = "true"
	cmd.Flags().StringVarP(&flags.Exec, "exec", "e", "", "Binary to run (default: node)")
	cmd.Flags().StringVar(&flags.EntryFile, "entryFile", "", "Entry file name (overrides config)")
	cmd.Flags().StringVar(&flags.SourceRoot, "sourceRoot", "", "Source root (overrides config)")
	cmd.Flags().BoolVar(&flags.NoShell, "no-shell", false, "Do not spawn child process within a shell")
	cmd.Flags().StringArrayVar(&flags.EnvFiles, "env-file", nil, "Path to an .env file (repeatable)")

	return cmd
}

func runStart(flags StartFlags) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	runnerOpts := &process.Options{
		EntryFile:  flags.EntryFile,
		SourceRoot: flags.SourceRoot,
		Binary:     flags.Exec,
		Debug:      flags.Debug,
		EnvFiles:   flags.EnvFiles,
		Shell:      !flags.NoShell,
		ExtraArgs:  extraArgsFromCLI(),
	}

	o, err := orchestrator.New(orchestrator.Options{
		Cwd:          cwd,
		ConfigPath:   flags.ConfigPath,
		TsConfigPath: flags.TsConfigPath,
		Watch:        flags.Watch,
		WatchAssets:  flags.WatchAssets,
		RunnerOpts:   runnerOpts,
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
		o.KillRunner()

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

	if err := o.Build(ctx, false); err != nil {
		return err
	}

	if code := o.WaitRunner(); code != 0 {
		os.Exit(code)
	}
	return nil
}

func extraArgsFromCLI() []string {
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			return os.Args[i+1:]
		}
	}
	return nil
}
