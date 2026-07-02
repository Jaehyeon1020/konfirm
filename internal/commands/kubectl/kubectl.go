package kubectl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"konfirm/internal/constants"
	"konfirm/internal/context"
	"konfirm/internal/store"
)

type deps struct {
	effectiveContext  func([]string) (string, error)
	loadConfig        func() (store.Config, error)
	promptForApproval func(string) error
	execCommand       func(string, []string) int
}

var commandDeps = deps{
	effectiveContext:  context.GetEffectiveContext,
	loadConfig:        store.LoadConfig,
	promptForApproval: promptForApproval,
	execCommand:       execKubectlCommand,
}

func Run(args []string) int {
	return run("kubectl", args)
}

func RunKubecolor(args []string) int {
	return run("kubecolor", args)
}

func run(commandName string, args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "missing %s args\n", commandName)
		return 2
	}

	ctx, err := commandDeps.effectiveContext(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve context: %v\n", err)
		return 1
	}

	cfg, err := commandDeps.loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	// Skip approval if stored as an allowed kubectl subcommand for the current context.
	subcommand := getKubectlSubcommand(args)
	if store.IsKubectlSubcommandAllowed(cfg.PermanentAllowKubectlSubcmds, ctx, subcommand) {
		return commandDeps.execCommand(commandName, args)
	}

	if err := commandDeps.promptForApproval(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return commandDeps.execCommand(commandName, args)
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

	fmt.Fprintf(tty, "%skonfirm%s is waiting for your confirmation\n", constants.ANSI_BOLD_BLUE, constants.ANSI_RESET)
	fmt.Fprintf(tty, "🔒 Context: %s%s%s 🔒\n", constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET)
	fmt.Fprintf(tty, "Continue? [y/N]: ")

	reader := bufio.NewReader(tty)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	line = strings.TrimSpace(line)
	if line == "" || strings.EqualFold(line, "n") {
		return errors.New("Aborted")
	}
	if !strings.EqualFold(line, "y") {
		return errors.New("Invalid input")
	}
	fmt.Fprintln(tty, "==================")
	return nil
}

func execKubectl(args []string) int {
	return execKubectlCommand("kubectl", args)
}

func execKubectlCommand(name string, args []string) int {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return commandExitCode(name, err)
	}
	return 0
}

func commandExitCode(name string, err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ProcessState != nil {
			return exitErr.ProcessState.ExitCode()
		}
	}
	fmt.Fprintf(os.Stderr, "failed to run %s: %v\n", name, err)
	return 1
}
