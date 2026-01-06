package kubectl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"konfirm/internal/context"
	"konfirm/internal/store"
)

func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "missing kubectl args")
		return 2
	}

	ctx, err := context.GetEffectiveContext(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve context: %v\n", err)
		return 1
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	// Skip approval if stored as an allowed context.
	if store.IsContextAllowed(cfg.PermanentAllowContexts, ctx) {
		return execKubectl(args)
	}

	// Skip approval if stored as an allowed kubectl subcommand for the current context.
	subcommand := getKubectlSubcommand(args)
	if store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, ctx, subcommand) {
		return execKubectl(args)
	}

	if err := promptForApproval(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "approval failed: %v\n", err)
		return 1
	}

	return execKubectl(args)
}

func getKubectlSubcommand(args []string) string {
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--" {
			continue
		}
		if arg == "--context" || arg == "--namespace" || arg == "-n" || arg == "--kubeconfig" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(arg, "--context=") || strings.HasPrefix(arg, "--namespace=") || strings.HasPrefix(arg, "--kubeconfig=") {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return arg
	}
	return ""
}

func promptForApproval(ctx string) error {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return errors.New("no TTY available for approval prompt")
	}
	defer tty.Close()

	const ansiBoldRed = "\x1b[1;31m"
	const ansiBoldBlue = "\x1b[1;34m"
	const ansiReset = "\x1b[0m"

	fmt.Fprintf(tty, "%skonfirm%s is waiting for your confirmation\n", ansiBoldBlue, ansiReset)
	fmt.Fprintf(tty, "🔒 Context: %s%s%s 🔒\n", ansiBoldRed, ctx, ansiReset)
	fmt.Fprintf(tty, "Type [Y/y] to continue: ")

	reader := bufio.NewReader(tty)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	line = strings.TrimSpace(line)
	if !strings.EqualFold(line, "y") {
		return errors.New("approval phrase mismatch")
	}
	fmt.Fprintln(tty, "==================")
	return nil
}

func execKubectl(args []string) int {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return commandExitCode(err)
	}
	return 0
}

func commandExitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ProcessState != nil {
			return exitErr.ProcessState.ExitCode()
		}
	}
	fmt.Fprintf(os.Stderr, "failed to run kubectl: %v\n", err)
	return 1
}
