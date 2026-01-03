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

	st, err := store.LoadState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load state: %v\n", err)
		return 1
	}

	if st.SessionAllowedContext != "" && st.SessionAllowedContext != ctx {
		st.SessionAllowedContext = ""
		if err := store.SaveState(st); err != nil {
			fmt.Fprintf(os.Stderr, "failed to update state: %v\n", err)
			return 1
		}
	}

	if store.IsContextAllowed(cfg.PermanentAllowContexts, ctx) {
		return execKubectl(args)
	}

	if st.SessionAllowedContext == ctx {
		return execKubectl(args)
	}

	if err := promptForApproval(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "approval failed: %v\n", err)
		return 1
	}

	return execKubectl(args)
}

func promptForApproval(ctx string) error {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return errors.New("no TTY available for approval prompt")
	}
	defer tty.Close()

	const ansiBoldRed = "\x1b[1;31m"
	const ansiReset = "\x1b[0m"

	fmt.Fprintf(tty, "Context: %s%s%s\n", ansiBoldRed, ctx, ansiReset)
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
