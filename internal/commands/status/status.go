package status

import (
	"fmt"
	"os"

	"konfirm/internal/context"
	"konfirm/internal/store"
)

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

	fmt.Fprintf(os.Stdout, "context: %s\n", currentCtx)
	if store.IsContextAllowed(cfg.PermanentAllowContexts, currentCtx) {
		fmt.Fprintln(os.Stdout, "context allowed: yes")
		return 0
	}

	fmt.Fprintln(os.Stdout, "context allowed: no")
	subcommands := cfg.PermanentAllowKubectlSubcmds[currentCtx]
	if len(subcommands) == 0 {
		fmt.Fprintln(os.Stdout, "allowed kubectl subcommands: (none)")
		return 0
	}

	fmt.Fprintln(os.Stdout, "allowed kubectl subcommands:")
	for _, subcommand := range subcommands {
		fmt.Fprintln(os.Stdout, subcommand)
	}
	return 0
}
