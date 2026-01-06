package allow

import (
	"fmt"
	"os"

	"konfirm/internal/commands/support"
	"konfirm/internal/context"
	"konfirm/internal/store"
)

func Run(args []string) int {
	if len(args) < 1 {
		support.Usage(os.Stderr)
		return 2
	}

	const ansiBoldRed = "\x1b[1;31m"
	const ansiBoldBlue = "\x1b[1;34m"
	const ansiReset = "\x1b[0m"

	sub := args[0]
	switch sub {
	case "add":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow add")
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

		if !store.IsContextAllowed(cfg.PermanentAllowContexts, currentCtx) {
			cfg.PermanentAllowContexts = append(cfg.PermanentAllowContexts, currentCtx)
			fmt.Fprintf(os.Stdout, "context added to allow list: %s%s%s\n", ansiBoldRed, currentCtx, ansiReset)
		} else {
			fmt.Fprintf(os.Stdout, "context already allowed: %s%s%s\n", ansiBoldRed, currentCtx, ansiReset)
		}

		if err := store.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
			return 1
		}
	case "remove":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow remove")
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

		if store.IsContextAllowed(cfg.PermanentAllowContexts, ctx) {
			cfg.PermanentAllowContexts = store.RemoveContext(cfg.PermanentAllowContexts, ctx)
			fmt.Fprintf(os.Stdout, "context removed from allow list: %s%s%s\n", ansiBoldRed, ctx, ansiReset)
		} else {
			fmt.Fprintf(os.Stdout, "context not in allow list: %s%s%s\n", ansiBoldRed, ctx, ansiReset)
		}
		if err := store.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
			return 1
		}
	case "list":
		cfg, err := store.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			return 1
		}

		for _, ctx := range cfg.PermanentAllowContexts {
			fmt.Fprintln(os.Stdout, ctx)
		}
	case "kubectl":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow kubectl <add|remove|list> <subcommand>")
			return 2
		}

		action := args[1]
		switch action {
		case "add":
			if len(args) != 3 {
				fmt.Fprintln(os.Stderr, "usage: konfirm allow kubectl add <subcommand>")
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

			subcommand := args[2]
			if cfg.PermanentAllowKubectlSubcmds == nil {
				cfg.PermanentAllowKubectlSubcmds = make(map[string][]string)
			}

			if !store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, currentCtx, subcommand) {
				cfg.PermanentAllowKubectlSubcmds[currentCtx] = append(cfg.PermanentAllowKubectlSubcmds[currentCtx], subcommand)
				fmt.Fprintf(os.Stdout, "kubectl subcommand added to allow list: %s%s%s (context %s%s%s)\n", ansiBoldBlue, subcommand, ansiReset, ansiBoldRed, currentCtx, ansiReset)
			} else {
				fmt.Fprintf(os.Stdout, "kubectl subcommand already allowed: %s%s%s (context %s%s%s)\n", ansiBoldBlue, subcommand, ansiReset, ansiBoldRed, currentCtx, ansiReset)
			}

			if err := store.SaveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
				return 1
			}
		case "remove":
			if len(args) != 3 {
				fmt.Fprintln(os.Stderr, "usage: konfirm allow kubectl remove <subcommand>")
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

			subcommand := args[2]
			if store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, currentCtx, subcommand) {
				cfg.PermanentAllowKubectlSubcmds[currentCtx] = store.RemoveKubectlSubcommand(cfg.PermanentAllowKubectlSubcmds[currentCtx], subcommand)
				fmt.Fprintf(os.Stdout, "kubectl subcommand removed from allow list: %s%s%s (context %s%s%s)\n", ansiBoldBlue, subcommand, ansiReset, ansiBoldRed, currentCtx, ansiReset)
			} else {
				fmt.Fprintf(os.Stdout, "kubectl subcommand not in allow list: %s%s%s (context %s%s%s)\n", ansiBoldBlue, subcommand, ansiReset, ansiBoldRed, currentCtx, ansiReset)
			}

			if err := store.SaveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
				return 1
			}
		case "list":
			if len(args) != 2 {
				fmt.Fprintln(os.Stderr, "usage: konfirm allow kubectl list")
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

			for _, subcommand := range cfg.PermanentAllowKubectlSubcmds[currentCtx] {
				fmt.Fprintln(os.Stdout, subcommand)
			}
		default:
			fmt.Fprintf(os.Stderr, "unknown allow kubectl command: %s\n", action)
			return 2
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown allow command: %s\n", sub)
		support.Usage(os.Stderr)
		return 2
	}
	return 0
}
