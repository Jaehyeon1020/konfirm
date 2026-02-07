package status

import (
	"fmt"
	"os"

	"konfirm/internal/constants"
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
