// Package process manages the Node.js child process lifecycle for `nego start`.
package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ProcessRunner defines the interface for managing the child process.
type ProcessRunner interface {
	Start() error
	Kill()
	Wait() int
}

// Options configures the Node.js child process.
type Options struct {
	// EntryFile is the name of the entry file without extension (e.g. "main").
	EntryFile string
	// SourceRoot is the source root as used in the output path (e.g. "src").
	SourceRoot string
	// OutDir is the compiled output directory (e.g. "dist").
	OutDir string
	// Binary is the executable to run (default "node").
	Binary string
	// Debug enables --inspect mode.
	Debug string
	// EnvFiles are paths to .env files to pass via --env-file=.
	EnvFiles []string
	// ExtraArgs are args passed after -- on the command line.
	ExtraArgs []string
	// Shell spawns the process inside a shell.
	Shell bool
	// Cwd is the working directory.
	Cwd string
}

// Runner manages a single Node.js child process and supports hot-reload.
type Runner struct {
	opts    Options
	mu      sync.Mutex
	current *exec.Cmd
}

// New creates a Runner with the given options.
func New(opts Options) ProcessRunner {
	if opts.Binary == "" {
		opts.Binary = "node"
	}
	if opts.Cwd == "" {
		opts.Cwd, _ = os.Getwd()
	}
	return &Runner{opts: opts}
}

// Start spawns the Node.js process.
func (r *Runner) Start() error {
	cmd, err := r.build()
	if err != nil {
		return err
	}
	
	r.mu.Lock()
	r.current = cmd
	r.mu.Unlock()
	
	return cmd.Start()
}

// Wait blocks until the child process exits and returns its exit code.
func (r *Runner) Wait() int {
	r.mu.Lock()
	cmd := r.current
	r.mu.Unlock()

	if cmd == nil {
		return 0
	}
	err := cmd.Wait()
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return 1
}

// Kill terminates the process tree and waits for it to fully exit.
func (r *Runner) Kill() {
	r.mu.Lock()
	cmd := r.current
	r.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		killTree(cmd.Process.Pid)
		_ = cmd.Wait() // Reap the main process
		
		r.mu.Lock()
		r.current = nil
		r.mu.Unlock()
	}
}

// build constructs the exec.Cmd for the node process.
func (r *Runner) build() (*exec.Cmd, error) {
	outputFile := r.resolveOutputFile()
	if outputFile == "" {
		return nil, fmt.Errorf(
			"could not find entry file. Looked for %s/%s/%s.js and %s/%s.js",
			r.opts.OutDir, r.opts.SourceRoot, r.opts.EntryFile,
			r.opts.OutDir, r.opts.EntryFile,
		)
	}

	nodeArgs := []string{}
	if r.opts.Debug != "" {
		if r.opts.Debug == "true" {
			nodeArgs = append(nodeArgs, "--inspect")
		} else {
			nodeArgs = append(nodeArgs, "--inspect="+r.opts.Debug)
		}
	}

	for _, ef := range r.opts.EnvFiles {
		nodeArgs = append(nodeArgs, "--env-file="+ef)
	}

	nodeArgs = append(nodeArgs, "--enable-source-maps", outputFile)
	nodeArgs = append(nodeArgs, r.opts.ExtraArgs...)

	var cmd *exec.Cmd
	if r.opts.Shell {
		allParts := append([]string{r.opts.Binary}, nodeArgs...)
		// Safely quote arguments for the shell
		var quotedParts []string
		for _, p := range allParts {
			quotedParts = append(quotedParts, quoteShellArg(p))
		}
		shellCmd := strings.Join(quotedParts, " ")
		cmd = exec.Command("sh", "-c", shellCmd)
	} else {
		cmd = exec.Command(r.opts.Binary, nodeArgs...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = r.opts.Cwd
	setSysProcAttr(cmd)

	return cmd, nil
}

func (r *Runner) resolveOutputFile() string {
	candidates := []string{
		filepath.Join(r.opts.Cwd, r.opts.OutDir, r.opts.SourceRoot, r.opts.EntryFile+".js"),
		filepath.Join(r.opts.Cwd, r.opts.OutDir, r.opts.EntryFile+".js"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}



func quoteShellArg(arg string) string {
	if !strings.ContainsAny(arg, " \t\n\r\"'\\$;<>|&()[]*?!#~") {
		return arg
	}
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}
