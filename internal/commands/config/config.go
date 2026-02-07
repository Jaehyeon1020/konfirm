package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"konfirm/internal/constants"
	"konfirm/internal/context"
	"konfirm/internal/store"
)

var ErrUserCancelled = errors.New("user cancelled")

func Run(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, "usage: konfirm config")
		return 2
	}

	if err := checkDependencies(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	currentCtx, err := context.GetCurrentContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get current context: %v\n", err)
		return 1
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	allSubcommands := constants.GetKubectlSubcommands()

	allowedSubcommands := cfg.PermanentAllowKubectlSubcmds[currentCtx]
	if allowedSubcommands == nil {
		allowedSubcommands = []string{}
	}

	mode, err := promptForMode()
	if err != nil {
		if errors.Is(err, ErrUserCancelled) {
			fmt.Fprintln(os.Stderr, "Cancelled")
			return 130
		}
		fmt.Fprintf(os.Stderr, "mode selection error: %v\n", err)
		return 1
	}

	var selectedSubcommands []string
	if mode == "add" {
		selectedSubcommands, err = selectForAdd(allSubcommands, allowedSubcommands, currentCtx)
	} else {
		selectedSubcommands, err = selectForRemove(allowedSubcommands, currentCtx)
	}

	if err != nil {
		if errors.Is(err, ErrUserCancelled) {
			fmt.Fprintln(os.Stderr, "Cancelled")
			return 130
		}
		fmt.Fprintf(os.Stderr, "fzf error: %v\n", err)
		return 1
	}

	if cfg.PermanentAllowKubectlSubcmds == nil {
		cfg.PermanentAllowKubectlSubcmds = make(map[string][]string)
	}

	if mode == "add" {
		allowedSubcommands = append(allowedSubcommands, selectedSubcommands...)
		cfg.PermanentAllowKubectlSubcmds[currentCtx] = allowedSubcommands
	} else {
		cfg.PermanentAllowKubectlSubcmds[currentCtx] = removeFromList(allowedSubcommands, selectedSubcommands)
	}

	if err := store.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	finalList := cfg.PermanentAllowKubectlSubcmds[currentCtx]
	if mode == "add" {
		fmt.Println("==================")
		fmt.Fprintf(os.Stdout, "Added %d subcommand(s) for context %s%s%s\n",
			len(selectedSubcommands), constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
	} else {
		fmt.Println("==================")
		fmt.Fprintf(os.Stdout, "Removed %d subcommand(s) from context %s%s%s\n",
			len(selectedSubcommands), constants.ANSI_BOLD_RED, currentCtx, constants.ANSI_RESET)
	}

	fmt.Fprintln(os.Stdout, "\nCurrent allowed kubectl subcommands:")
	if len(finalList) == 0 {
		fmt.Fprintln(os.Stdout, "  (none)")
	} else {
		for _, sub := range finalList {
			fmt.Fprintf(os.Stdout, "  • %s\n", sub)
		}
	}

	return 0
}

func checkDependencies() error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return errors.New("kubectl not found in PATH")
	}

	if _, err := exec.LookPath("fzf"); err != nil {
		return fmt.Errorf(`fzf is not installed. To use 'konfirm config', please install fzf:
  https://github.com/junegunn/fzf

Alternatively, you can manage allowed subcommands using:
  konfirm add <subcommand>    # Allow a specific subcommand
  konfirm remove <subcommand> # Remove a specific subcommand
  konfirm status              # Show current configuration`)
	}

	return nil
}

func promptForMode() (string, error) {
	options := []string{
		"Add subcommands to allowlist",
		"Remove subcommands from allowlist",
	}
	input := strings.Join(options, "\n")

	cmd := exec.Command("fzf",
		"--ansi",
		"--prompt", "What would you like to do? > ",
		"--no-info",
		"--reverse",
	)

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", errors.New("no TTY available for mode prompt")
	}
	defer tty.Close()

	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = tty

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 130 {
			return "", ErrUserCancelled
		}
		return "", fmt.Errorf("fzf failed: %w", err)
	}

	selected := strings.TrimSpace(string(output))
	if strings.Contains(selected, "Add") {
		return "add", nil
	} else if strings.Contains(selected, "Remove") {
		return "remove", nil
	}

	return "", ErrUserCancelled
}

func selectForAdd(allSubcommands, currentlyAllowed []string, ctx string) ([]string, error) {
	allowedSet := make(map[string]bool)
	for _, s := range currentlyAllowed {
		allowedSet[s] = true
	}

	var notAllowed []string
	for _, sub := range allSubcommands {
		if !allowedSet[sub] {
			notAllowed = append(notAllowed, sub)
		}
	}

	if len(notAllowed) == 0 {
		return nil, errors.New("all subcommands are already allowed")
	}

	return selectWithFzf(notAllowed, ctx, "add")
}

func selectForRemove(currentlyAllowed []string, ctx string) ([]string, error) {
	if len(currentlyAllowed) == 0 {
		return nil, errors.New("no subcommands to remove")
	}

	return selectWithFzf(currentlyAllowed, ctx, "remove")
}

func removeFromList(list, toRemove []string) []string {
	removeSet := make(map[string]bool)
	for _, item := range toRemove {
		removeSet[item] = true
	}

	var result []string
	for _, item := range list {
		if !removeSet[item] {
			result = append(result, item)
		}
	}

	return result
}

func selectWithFzf(subcommands []string, ctx string, mode string) ([]string, error) {
	input := strings.Join(subcommands, "\n")

	var promptMsg string
	if mode == "add" {
		promptMsg = "Select subcommands to ADD > "
	} else {
		promptMsg = "Select subcommands to REMOVE > "
	}

	cmd := exec.Command("fzf",
		"--multi",
		"--ansi",
		"--prompt", promptMsg,
		"--header", fmt.Sprintf("Context: %s%s%s | TAB: toggle | ENTER: confirm | ESC: cancel", constants.ANSI_BOLD_RED, ctx, constants.ANSI_RESET),
		"--reverse",
		"--marker", "▶ ",
		"--no-info",
	)

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, errors.New("no TTY available for fzf")
	}
	defer tty.Close()

	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = tty

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 130 {
			return nil, ErrUserCancelled
		}
		return nil, fmt.Errorf("fzf failed: %w", err)
	}

	var selected []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			selected = append(selected, line)
		}
	}

	return selected, nil
}
