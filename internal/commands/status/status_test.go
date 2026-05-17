package status

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"konfirm/internal/store"
)

func TestRenderReportWithContextAndAllowlist(t *testing.T) {
	report := report{
		Context:       "prod-cluster",
		ConfigPath:    "/tmp/konfirm/config.json",
		Allowed:       []string{"get", "logs"},
		KubectlFound:  true,
		FzfFound:      true,
		DetectedShell: "zsh",
	}

	var out bytes.Buffer
	renderReport(&out, report)

	got := out.String()
	required := []string{
		"Context",
		"  current: prod-cluster",
		"Config",
		"  path: /tmp/konfirm/config.json",
		"  allowed for current context:",
		"    get",
		"    logs",
		"Dependencies",
		"  kubectl: found",
		"  fzf: found",
		"Shell setup",
		"  detected shell: zsh",
		"  completion: add `source <(konfirm completion zsh)` to ~/.zshrc",
		"  recommended alias: alias k=\"konfirm kubectl\"",
	}
	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("renderReport output missing %q\noutput:\n%s", want, got)
		}
	}
}

func TestRenderReportWithNoAllowlist(t *testing.T) {
	report := report{
		Context:       "dev-cluster",
		ConfigPath:    "/tmp/konfirm/config.json",
		Allowed:       nil,
		KubectlFound:  true,
		FzfFound:      false,
		DetectedShell: "fish",
	}

	var out bytes.Buffer
	renderReport(&out, report)

	got := out.String()
	required := []string{
		"  current: dev-cluster",
		"  allowed for current context: (none)",
		"  kubectl: found",
		"  fzf: missing",
		"  detected shell: fish",
		"  completion: add `konfirm completion fish | source` to ~/.config/fish/config.fish",
	}
	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("renderReport output missing %q\noutput:\n%s", want, got)
		}
	}
}

func TestRenderReportWithUnavailableContextAndUnknownShell(t *testing.T) {
	report := report{
		ContextErr:     errors.New("kubectl failed"),
		ConfigPath:     "/tmp/konfirm/config.json",
		KubectlFound:   false,
		FzfFound:       false,
		DetectedShell:  "unknown",
		ContextMissing: true,
	}

	var out bytes.Buffer
	renderReport(&out, report)

	got := out.String()
	required := []string{
		"  current: unavailable (kubectl failed)",
		"  allowed for current context: unavailable because current context could not be resolved",
		"  kubectl: missing",
		"  fzf: missing",
		"  detected shell: unknown",
		"  completion: run `konfirm completion zsh` or `konfirm completion fish` for setup guidance",
	}
	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("renderReport output missing %q\noutput:\n%s", want, got)
		}
	}
}

func TestRenderReportWithConfigError(t *testing.T) {
	report := report{
		Context:       "prod-cluster",
		ConfigPath:    "/tmp/konfirm/config.json",
		ConfigErr:     errors.New("invalid character"),
		KubectlFound:  true,
		FzfFound:      true,
		DetectedShell: "zsh",
	}

	var out bytes.Buffer
	renderReport(&out, report)

	got := out.String()
	required := []string{
		"Config",
		"  path: /tmp/konfirm/config.json",
		"  error: invalid character",
	}
	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("renderReport output missing %q\noutput:\n%s", want, got)
		}
	}
}

func TestRunPrintsDiagnosticsForAvailableStatus(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	restore := stubDeps(t, deps{
		stdout: &stdout,
		stderr: &stderr,
		loadConfig: func() (store.Config, error) {
			return store.Config{
				PermanentAllowKubectlSubcmds: map[string][]string{
					"prod-cluster": {"get", "logs"},
				},
			}, nil
		},
		currentContext: func() (string, error) { return "prod-cluster", nil },
		getenv:         func(name string) string { return "/bin/zsh" },
	})
	defer restore()

	code := Run(nil)

	if code != 0 {
		t.Fatalf("Run(nil) code = %d, want 0", code)
	}
	got := stdout.String()
	required := []string{
		"  current: prod-cluster",
		"    get",
		"    logs",
	}
	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("Run(nil) stdout missing %q\nstdout:\n%s", want, got)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(nil) stderr = %q, want empty", stderr.String())
	}
}

func TestRunContinuesWhenCurrentContextFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	restore := stubDeps(t, deps{
		stdout:         &stdout,
		stderr:         &stderr,
		currentContext: func() (string, error) { return "", errors.New("current-context failed") },
		lookPath:       func(name string) (string, error) { return "", errors.New("missing") },
		getenv:         func(name string) string { return "" },
	})
	defer restore()

	code := Run(nil)

	if code != 0 {
		t.Fatalf("Run(nil) code = %d, want 0", code)
	}
	got := stdout.String()
	required := []string{
		"  current: unavailable (current-context failed)",
		"  allowed for current context: unavailable because current context could not be resolved",
		"  kubectl: missing",
		"  fzf: missing",
		"  detected shell: unknown",
	}
	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("Run(nil) stdout missing %q\nstdout:\n%s", want, got)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(nil) stderr = %q, want empty", stderr.String())
	}
}

func TestRunReturnsOneWhenConfigLoadFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	restore := stubDeps(t, deps{
		stdout:         &stdout,
		stderr:         &stderr,
		loadConfig:     func() (store.Config, error) { return store.Config{}, errors.New("bad json") },
		currentContext: func() (string, error) { return "prod-cluster", nil },
		getenv:         func(name string) string { return "/bin/fish" },
	})
	defer restore()

	code := Run(nil)

	if code != 1 {
		t.Fatalf("Run(nil) code = %d, want 1", code)
	}
	got := stdout.String()
	if !strings.Contains(got, "  error: bad json") {
		t.Fatalf("Run(nil) stdout missing config error\nstdout:\n%s", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(nil) stderr = %q, want empty", stderr.String())
	}
}

func TestRunReturnsOneWhenConfigPathFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	restore := stubDeps(t, deps{
		stdout:     &stdout,
		stderr:     &stderr,
		configPath: func() (string, error) { return "", errors.New("config path failed") },
		loadConfig: func() (store.Config, error) {
			t.Fatal("loadConfig should not be called when configPath fails")
			return store.Config{}, nil
		},
	})
	defer restore()

	code := Run(nil)

	if code != 1 {
		t.Fatalf("Run(nil) code = %d, want 1", code)
	}
	got := stdout.String()
	if !strings.Contains(got, "  error: config path failed") {
		t.Fatalf("Run(nil) stdout missing config path error\nstdout:\n%s", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(nil) stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsArguments(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	restore := stubDeps(t, deps{
		stdout: &stdout,
		stderr: &stderr,
	})
	defer restore()

	code := Run([]string{"--json"})

	if code != 2 {
		t.Fatalf("Run([--json]) code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("Run([--json]) stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "usage: konfirm status") {
		t.Fatalf("Run([--json]) stderr missing usage\nstderr:\n%s", stderr.String())
	}
}

func stubDeps(t *testing.T, replacement deps) func() {
	t.Helper()

	original := commandDeps
	if replacement.stdout == nil {
		replacement.stdout = bytes.NewBuffer(nil)
	}
	if replacement.stderr == nil {
		replacement.stderr = bytes.NewBuffer(nil)
	}
	if replacement.configPath == nil {
		replacement.configPath = func() (string, error) { return "/tmp/konfirm/config.json", nil }
	}
	if replacement.loadConfig == nil {
		replacement.loadConfig = func() (store.Config, error) { return store.Config{}, nil }
	}
	if replacement.currentContext == nil {
		replacement.currentContext = func() (string, error) { return "dev-cluster", nil }
	}
	if replacement.lookPath == nil {
		replacement.lookPath = func(name string) (string, error) { return "/usr/bin/" + name, nil }
	}
	if replacement.getenv == nil {
		replacement.getenv = func(name string) string { return "" }
	}

	commandDeps = replacement
	return func() { commandDeps = original }
}
