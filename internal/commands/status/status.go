package status

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"konfirm/internal/context"
	"konfirm/internal/store"
)

type deps struct {
	stdout         io.Writer
	stderr         io.Writer
	configPath     func() (string, error)
	loadConfig     func() (store.Config, error)
	currentContext func() (string, error)
	lookPath       func(string) (string, error)
	getenv         func(string) string
}

var commandDeps = deps{
	stdout:         os.Stdout,
	stderr:         os.Stderr,
	configPath:     store.ConfigPath,
	loadConfig:     store.LoadConfig,
	currentContext: context.GetCurrentContext,
	lookPath:       exec.LookPath,
	getenv:         os.Getenv,
}

type report struct {
	Context        string
	ContextErr     error
	ContextMissing bool
	ConfigPath     string
	ConfigErr      error
	Allowed        []string
	KubectlFound   bool
	FzfFound       bool
	DetectedShell  string
}

func renderReport(w io.Writer, r report) {
	fmt.Fprintln(w, "Context")
	if r.ContextErr != nil {
		fmt.Fprintf(w, "  current: unavailable (%v)\n", r.ContextErr)
	} else {
		fmt.Fprintf(w, "  current: %s\n", r.Context)
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Config")
	if r.ConfigPath != "" {
		fmt.Fprintf(w, "  path: %s\n", r.ConfigPath)
	}
	if r.ConfigErr != nil {
		fmt.Fprintf(w, "  error: %v\n", r.ConfigErr)
	} else if r.ContextMissing {
		fmt.Fprintln(w, "  allowed for current context: unavailable because current context could not be resolved")
	} else if len(r.Allowed) == 0 {
		fmt.Fprintln(w, "  allowed for current context: (none)")
	} else {
		fmt.Fprintln(w, "  allowed for current context:")
		for _, subcommand := range r.Allowed {
			fmt.Fprintf(w, "    %s\n", subcommand)
		}
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Dependencies")
	fmt.Fprintf(w, "  kubectl: %s\n", foundLabel(r.KubectlFound))
	fmt.Fprintf(w, "  fzf: %s\n", foundLabel(r.FzfFound))

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Shell setup")
	fmt.Fprintf(w, "  detected shell: %s\n", r.DetectedShell)
	fmt.Fprintf(w, "  completion: %s\n", completionHint(r.DetectedShell))
	fmt.Fprintln(w, "  recommended alias: alias k=\"konfirm kubectl\"")
}

func foundLabel(found bool) string {
	if found {
		return "found"
	}
	return "missing"
}

func completionHint(shell string) string {
	switch shell {
	case "zsh":
		return "add `source <(konfirm completion zsh)` to ~/.zshrc"
	case "fish":
		return "add `konfirm completion fish | source` to ~/.config/fish/config.fish"
	default:
		return "run `konfirm completion zsh` or `konfirm completion fish` for setup guidance"
	}
}

func Run(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(commandDeps.stderr, "usage: konfirm status")
		return 2
	}

	r, exitCode := collectReport(commandDeps)
	renderReport(commandDeps.stdout, r)
	return exitCode
}

func collectReport(d deps) (report, int) {
	r := report{
		KubectlFound:  hasCommand(d, "kubectl"),
		FzfFound:      hasCommand(d, "fzf"),
		DetectedShell: detectShell(d.getenv("SHELL")),
	}

	ctx, ctxErr := d.currentContext()
	if ctxErr != nil {
		r.ContextErr = ctxErr
		r.ContextMissing = true
	} else {
		r.Context = ctx
	}

	path, err := d.configPath()
	if err != nil {
		r.ConfigErr = err
		return r, 1
	}
	r.ConfigPath = path

	cfg, err := d.loadConfig()
	if err != nil {
		r.ConfigErr = err
		return r, 1
	}

	if !r.ContextMissing {
		r.Allowed = cfg.PermanentAllowKubectlSubcmds[r.Context]
	}

	return r, 0
}

func hasCommand(d deps, name string) bool {
	_, err := d.lookPath(name)
	return err == nil
}

func detectShell(shellPath string) string {
	shell := filepath.Base(shellPath)
	switch shell {
	case "zsh", "fish":
		return shell
	default:
		return "unknown"
	}
}
