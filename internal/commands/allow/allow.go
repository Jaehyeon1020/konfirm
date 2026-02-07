package allow

import (
	"fmt"
	"os"
	"strings"

	"konfirm/internal/commands/support"
	"konfirm/internal/constants"
	"konfirm/internal/context"
	"konfirm/internal/store"
)

func Run(args []string) int {
	if len(args) < 1 {
		support.Usage(os.Stderr)
		return 2
	}

	command := args[0]
	switch command {
	case "add":
		return handleCommandAdd(args)
	case "remove":
		return handleCommandRemove(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		support.Usage(os.Stderr)
		return 2
	}
}

func handleCommandAdd(args []string) int {
	allFlagEnabled, subcommands, err := parseAllowArgs(args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	if !allFlagEnabled && len(subcommands) == 0 {
		fmt.Fprintln(os.Stderr, "usage: konfirm add <subcommand>... | konfirm add --all")
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

	if cfg.PermanentAllowKubectlSubcmds == nil {
		cfg.PermanentAllowKubectlSubcmds = make(map[string][]string)
	}

	if allFlagEnabled {
		subcommands = constants.GetKubectlSubcommands()
	}

	added := 0
	for _, subcommand := range subcommands {
		if !store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, currentCtx, subcommand) {
			cfg.PermanentAllowKubectlSubcmds[currentCtx] = append(cfg.PermanentAllowKubectlSubcmds[currentCtx], subcommand)
			if allFlagEnabled {
				added++
			} else {
				fmt.Fprintf(os.Stdout, "kubectl subcommand added to allow list: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
			}
		} else if !allFlagEnabled {
			fmt.Fprintf(os.Stdout, "kubectl subcommand already allowed: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
		}
	}

	if allFlagEnabled {
		fmt.Fprintf(os.Stdout, "%d kubectl subcommand(s) added to allow list (context %s%s%s)\n", added, constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
	}

	if err := store.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	return 0
}

func handleCommandRemove(args []string) int {
	allFlagEnabled, subcommands, err := parseAllowArgs(args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	if !allFlagEnabled && len(subcommands) == 0 {
		fmt.Fprintln(os.Stderr, "usage: konfirm remove <subcommand>... | konfirm remove --all")
		return 2
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	ctx, err := context.GetCurrentContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve context: %v\n", err)
		return 1
	}

	if allFlagEnabled {
		if existing := cfg.PermanentAllowKubectlSubcmds[ctx]; len(existing) > 0 {
			delete(cfg.PermanentAllowKubectlSubcmds, ctx)
			fmt.Fprintf(os.Stdout, "all kubectl subcommands removed from allow list (context %s%s%s)\n", constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
		} else {
			fmt.Fprintf(os.Stdout, "no allowed subcommands for context: %s%s%s\n", constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
		}
	} else {
		for _, subcommand := range subcommands {
			if store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, ctx, subcommand) {
				cfg.PermanentAllowKubectlSubcmds[ctx] = store.RemoveKubectlSubcommand(cfg.PermanentAllowKubectlSubcmds[ctx], subcommand)
				fmt.Fprintf(os.Stdout, "kubectl subcommand removed from allow list: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
			} else {
				fmt.Fprintf(os.Stdout, "kubectl subcommand not in allow list: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
			}
		}
	}
	if err := store.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	return 0
}

func parseAllowArgs(args []string) (bool, []string, error) {
	allFlagEnabled := false
	var subcommands []string
	for _, arg := range args {
		switch {
		case arg == "--all":
			allFlagEnabled = true
		case strings.HasPrefix(arg, "-"):
			return false, nil, fmt.Errorf("unknown flag: %s", arg)
		default:
			subcommands = append(subcommands, arg)
		}
	}

	if allFlagEnabled && len(subcommands) > 0 {
		return false, nil, fmt.Errorf("cannot combine --all with a subcommand")
	}
	return allFlagEnabled, subcommands, nil
}
