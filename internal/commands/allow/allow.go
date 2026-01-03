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
		}

		if err := store.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
			return 1
		}
	case "remove":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow remove <context>")
			return 2
		}

		cfg, err := store.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			return 1
		}

		ctx := args[1]
		cfg.PermanentAllowContexts = store.RemoveContext(cfg.PermanentAllowContexts, ctx)
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
	case "once":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow once")
			return 2
		}

		currentCtx, err := context.GetCurrentContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve context: %v\n", err)
			return 1
		}

		st, err := store.LoadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load state: %v\n", err)
			return 1
		}
		st.SessionAllowedContext = currentCtx
		if err := store.SaveState(st); err != nil {
			fmt.Fprintf(os.Stderr, "failed to update state: %v\n", err)
			return 1
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown allow command: %s\n", sub)
		support.Usage(os.Stderr)
		return 2
	}
	return 0
}
