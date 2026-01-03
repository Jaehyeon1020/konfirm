package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type config struct {
	PermanentAllowContexts []string `json:"permanent_allow_contexts"`
	ApprovalPhraseTemplate string   `json:"approval_phrase_template"`
}

type state struct {
	SessionAllowedContext string `json:"session_allowed_context"`
}

const defaultApprovalTemplate = "approve {context}"

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		exitWithCode(2)
	}

	switch os.Args[1] {
	case "kubectl":
		handleKubectl(os.Args[2:])
	case "allow":
		handleAllow(os.Args[2:])
	case "-h", "--help", "help":
		usage(os.Stdout)
	case "version":
		fmt.Fprintln(os.Stdout, "konfirm draft")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage(os.Stderr)
		exitWithCode(2)
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "konfirm - confirm kubectl context before execution")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  konfirm kubectl <kubectl args...>")
	fmt.Fprintln(w, "  konfirm allow add <context>")
	fmt.Fprintln(w, "  konfirm allow remove <context>")
	fmt.Fprintln(w, "  konfirm allow list")
	fmt.Fprintln(w, "  konfirm allow once <context>")
}

func handleKubectl(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "missing kubectl args")
		exitWithCode(2)
	}

	ctx, err := effectiveContext(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve context: %v\n", err)
		exitWithCode(1)
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		exitWithCode(1)
	}

	st, err := loadState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load state: %v\n", err)
		exitWithCode(1)
	}

	if st.SessionAllowedContext != "" && st.SessionAllowedContext != ctx {
		st.SessionAllowedContext = ""
		if err := saveState(st); err != nil {
			fmt.Fprintf(os.Stderr, "failed to update state: %v\n", err)
			exitWithCode(1)
		}
	}

	if contextAllowed(cfg.PermanentAllowContexts, ctx) {
		execKubectl(args)
		return
	}

	if st.SessionAllowedContext == ctx {
		execKubectl(args)
		return
	}

	if err := promptForApproval(cfg, ctx, args); err != nil {
		fmt.Fprintf(os.Stderr, "approval failed: %v\n", err)
		exitWithCode(1)
	}

	execKubectl(args)
}

func handleAllow(args []string) {
	if len(args) < 1 {
		usage(os.Stderr)
		exitWithCode(2)
	}

	sub := args[0]
	switch sub {
	case "add":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow add <context>")
			exitWithCode(2)
		}
		cfg, err := loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			exitWithCode(1)
		}
		ctx := args[1]
		if !contextAllowed(cfg.PermanentAllowContexts, ctx) {
			cfg.PermanentAllowContexts = append(cfg.PermanentAllowContexts, ctx)
		}
		if err := saveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
			exitWithCode(1)
		}
	case "remove":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow remove <context>")
			exitWithCode(2)
		}
		cfg, err := loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			exitWithCode(1)
		}
		ctx := args[1]
		cfg.PermanentAllowContexts = removeContext(cfg.PermanentAllowContexts, ctx)
		if err := saveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
			exitWithCode(1)
		}
	case "list":
		cfg, err := loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			exitWithCode(1)
		}
		for _, ctx := range cfg.PermanentAllowContexts {
			fmt.Fprintln(os.Stdout, ctx)
		}
	case "once":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: konfirm allow once <context>")
			exitWithCode(2)
		}
		current, err := currentContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve context: %v\n", err)
			exitWithCode(1)
		}
		ctx := args[1]
		if ctx != current {
			fmt.Fprintf(os.Stderr, "context mismatch: current is %s\n", current)
			exitWithCode(1)
		}
		st, err := loadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load state: %v\n", err)
			exitWithCode(1)
		}
		st.SessionAllowedContext = ctx
		if err := saveState(st); err != nil {
			fmt.Fprintf(os.Stderr, "failed to update state: %v\n", err)
			exitWithCode(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown allow command: %s\n", sub)
		usage(os.Stderr)
		exitWithCode(2)
	}
}

func effectiveContext(args []string) (string, error) {
	ctx, err := contextFromArgs(args)
	if err != nil {
		return "", err
	}
	if ctx != "" {
		return ctx, nil
	}
	return currentContext()
}

func contextFromArgs(args []string) (string, error) {
	var ctx string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--context" {
			if i+1 >= len(args) {
				return "", errors.New("--context requires a value")
			}
			ctx = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(arg, "--context=") {
			ctx = strings.TrimPrefix(arg, "--context=")
		}
	}
	return ctx, nil
}

func currentContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	cmd.Stdin = os.Stdin
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func promptForApproval(cfg config, ctx string, args []string) error {
	approvalPhrase := approvalPhrase(cfg, ctx)

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return errors.New("no TTY available for approval prompt")
	}
	defer tty.Close()

	fmt.Fprintf(tty, "Context: %s\n", ctx)
	fmt.Fprintf(tty, "Command: kubectl %s\n", strings.Join(args, " "))
	fmt.Fprintf(tty, "Type %q to continue: ", approvalPhrase)

	reader := bufio.NewReader(tty)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	line = strings.TrimSpace(line)
	if line != approvalPhrase {
		return errors.New("approval phrase mismatch")
	}
	return nil
}

func approvalPhrase(cfg config, ctx string) string {
	template := strings.TrimSpace(cfg.ApprovalPhraseTemplate)
	if template == "" {
		template = defaultApprovalTemplate
	}
	return strings.ReplaceAll(template, "{context}", ctx)
}

func contextAllowed(allowed []string, ctx string) bool {
	for _, item := range allowed {
		if item == ctx {
			return true
		}
	}
	return false
}

func removeContext(items []string, target string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item != target {
			result = append(result, item)
		}
	}
	return result
}

func loadConfig() (config, error) {
	path, err := configPath()
	if err != nil {
		return config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config{}, nil
		}
		return config{}, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func saveConfig(cfg config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func loadState() (state, error) {
	path, err := statePath()
	if err != nil {
		return state{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state{}, nil
		}
		return state{}, err
	}
	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return state{}, err
	}
	return st, nil
}

func saveState(st state) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func configPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "konfirm", "config.json"), nil
}

func statePath() (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "konfirm", "state.json"), nil
}

func execKubectl(args []string) {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		exitWithCommandError(err)
	}
}

func exitWithCommandError(err error) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ProcessState != nil {
			exitWithCode(exitErr.ProcessState.ExitCode())
		}
	}
	fmt.Fprintf(os.Stderr, "failed to run kubectl: %v\n", err)
	exitWithCode(1)
}

func exitWithCode(code int) {
	os.Exit(code)
}
