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
	allFlagEnabled, subcommand, err := parseAllowArgs(args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	if !allFlagEnabled && subcommand == "" {
		fmt.Fprintln(os.Stderr, "usage: konfirm add <subcommand> | konfirm add --all")
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

	if allFlagEnabled {
		if !store.IsContextAllowed(cfg.PermanentAllowContexts, currentCtx) {
			cfg.PermanentAllowContexts = append(cfg.PermanentAllowContexts, currentCtx)
			fmt.Fprintf(os.Stdout, "context added to allow list: %s%s%s\n", constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
		} else {
			fmt.Fprintf(os.Stdout, "context already allowed: %s%s%s\n", constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
		}
	} else {
		if cfg.PermanentAllowKubectlSubcmds == nil {
			cfg.PermanentAllowKubectlSubcmds = make(map[string][]string)
		}
		if !store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, currentCtx, subcommand) {
			cfg.PermanentAllowKubectlSubcmds[currentCtx] = append(cfg.PermanentAllowKubectlSubcmds[currentCtx], subcommand)
			fmt.Fprintf(os.Stdout, "kubectl subcommand added to allow list: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
		} else {
			fmt.Fprintf(os.Stdout, "kubectl subcommand already allowed: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
		}
	}

	if err := store.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	return 0
}

func handleCommandRemove(args []string) int {
	allFlagEnabled, subcommand, err := parseAllowArgs(args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	if !allFlagEnabled && subcommand == "" {
		fmt.Fprintln(os.Stderr, "usage: konfirm remove <subcommand> | konfirm remove --all")
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
		if store.IsContextAllowed(cfg.PermanentAllowContexts, ctx) {
			cfg.PermanentAllowContexts = store.RemoveContext(cfg.PermanentAllowContexts, ctx)
			fmt.Fprintf(os.Stdout, "context removed from allow list: %s%s%s\n", constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
		} else {
			fmt.Fprintf(os.Stdout, "context not in allow list: %s%s%s\n", constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
		}
	} else {
		if store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, ctx, subcommand) {
			cfg.PermanentAllowKubectlSubcmds[ctx] = store.RemoveKubectlSubcommand(cfg.PermanentAllowKubectlSubcmds[ctx], subcommand)
			fmt.Fprintf(os.Stdout, "kubectl subcommand removed from allow list: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
		} else {
			fmt.Fprintf(os.Stdout, "kubectl subcommand not in allow list: %s%s%s (context %s%s%s)\n", constants.ANSI_BOLD_BLUE, subcommand, constants.ANSI_RESET, constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
		}
	}
	if err := store.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	return 0
}

func parseAllowArgs(args []string) (bool, string, error) {
	allFlagEnabled := false
	subcommand := ""
	for _, arg := range args {
		switch {
		case arg == "--all":
			allFlagEnabled = true
		case strings.HasPrefix(arg, "-"):
			return false, "", fmt.Errorf("unknown flag: %s", arg)
		default:
			if subcommand != "" {
				return false, "", fmt.Errorf("usage: konfirm add <subcommand> | konfirm remove <subcommand>")
			}
			subcommand = arg
		}
	}

	if allFlagEnabled && subcommand != "" {
		return false, "", fmt.Errorf("cannot combine --all with a subcommand")
	}
	return allFlagEnabled, subcommand, nil
}
