package support

import (
	"fmt"
	"io"
	"konfirm/internal/constants"
)

func Usage(w io.Writer) {
	fmt.Fprint(w, constants.ASCII_LOGO)
	fmt.Fprintln(w, "konfirm - confirm the effective kubectl context before execution")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  konfirm kubectl <kubectl args...>")
	fmt.Fprintln(w, "  konfirm <command> [args...]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  kubectl      Run kubectl after confirming the effective context")
	fmt.Fprintln(w, "  config       Interactively manage allowed kubectl subcommands")
	fmt.Fprintln(w, "  add          Allow kubectl subcommands for the current context")
	fmt.Fprintln(w, "  remove       Remove allowed kubectl subcommands for the current context")
	fmt.Fprintln(w, "  status       Show current context, allowlist, and local setup")
	fmt.Fprintln(w, "  completion   Generate shell completion script")
	fmt.Fprintln(w, "  version      Show konfirm version")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  -h, --help   Show help")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  konfirm kubectl get pods")
	fmt.Fprintln(w, "  konfirm kubectl --context prod-cluster get deploy")
	fmt.Fprintln(w, "  konfirm add get logs")
	fmt.Fprintln(w, "  konfirm config")
	fmt.Fprintln(w, "  konfirm status")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Run `konfirm status` to inspect your current context, allowlist, and local setup.")
}

func Version(w io.Writer) {
	fmt.Fprint(w, constants.ASCII_LOGO)
	fmt.Fprintln(w, "konfirm", constants.VERSION)
}
