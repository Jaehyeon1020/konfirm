package status

import (
	"fmt"
	"io"
	"os"

	"konfirm/internal/constants"
	"konfirm/internal/context"
	"konfirm/internal/store"
)

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
		fmt.Fprintln(os.Stderr, "usage: konfirm status")
		return 2
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	currentCtx, err := context.GetCurrentContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve context: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "Context: %s%s%s\n", constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
	subcommands := cfg.PermanentAllowKubectlSubcmds[currentCtx]
	if len(subcommands) == 0 {
		fmt.Fprint(os.Stdout, "Allowed kubectl subcommands: (none)\n")
		return 0
	}

	fmt.Fprintln(os.Stdout, "Allowed kubectl subcommands:")
	for _, subcommand := range subcommands {
		fmt.Fprintf(os.Stdout, " • %s\n", subcommand)
	}
	return 0
}
