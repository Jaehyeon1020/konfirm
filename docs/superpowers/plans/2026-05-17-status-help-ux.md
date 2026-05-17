# Status Help UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Improve `konfirm status` diagnostics and make `konfirm -h` a clearer conventional CLI help page without adding new commands or `--json`.

**Architecture:** Keep the command surface unchanged. Add testable status report rendering behind `status.Run`, and keep help rendering in `internal/commands/support` as the single source used by help-like invocations.

**Tech Stack:** Go 1.21, standard library only, existing package layout under `internal/commands`.

---

## File Structure

- Modify: `internal/commands/support/support.go`
  - Owns `Usage(w io.Writer)` and `Version(w io.Writer)`.
  - `Usage` should write the full help output to the provided writer.

- Create: `internal/commands/support/support_test.go`
  - Tests the help output contains conventional CLI sections and examples.

- Modify: `internal/commands/status/status.go`
  - Keeps `Run(args []string) int`.
  - Adds dependency seams for tests.
  - Collects config path, config, current context, dependency status, shell, and renders the report.

- Create: `internal/commands/status/status_test.go`
  - Tests status rendering and command behavior without requiring real `kubectl`, `fzf`, or a Kubernetes cluster.

- No changes: `cmd/konfirm/main.go`
  - Existing help dispatch already calls `support.Usage`.

- No changes: `internal/store/store.go`
  - Existing `ConfigPath` and `LoadConfig` are sufficient.

---

### Task 1: Expand Help Output

**Files:**
- Modify: `internal/commands/support/support.go`
- Create: `internal/commands/support/support_test.go`

- [ ] **Step 1: Write the failing help test**

Create `internal/commands/support/support_test.go`:

```go
package support

import (
	"bytes"
	"strings"
	"testing"
)

func TestUsageIncludesConventionalCLIHelp(t *testing.T) {
	var out bytes.Buffer

	Usage(&out)

	got := out.String()
	required := []string{
		"konfirm - confirm the effective kubectl context before execution",
		"Usage:",
		"  konfirm kubectl <kubectl args...>",
		"  konfirm <command> [args...]",
		"Commands:",
		"  kubectl",
		"  config",
		"  add",
		"  remove",
		"  status",
		"  completion",
		"  version",
		"Options:",
		"  -h, --help",
		"Examples:",
		"  konfirm kubectl get pods",
		"  konfirm kubectl --context prod-cluster get deploy",
		"  konfirm add get logs",
		"  konfirm config",
		"  konfirm status",
		"Run `konfirm status` to inspect your current context, allowlist, and local setup.",
	}

	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("Usage() output missing %q\noutput:\n%s", want, got)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./internal/commands/support
```

Expected: FAIL because the current usage output does not include the expanded sections and examples.

- [ ] **Step 3: Implement the help output**

Replace `Usage` in `internal/commands/support/support.go` with:

```go
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
```

Also update `Version` to respect the provided writer for the logo:

```go
func Version(w io.Writer) {
	fmt.Fprint(w, constants.ASCII_LOGO)
	fmt.Fprintln(w, "konfirm", constants.VERSION)
}
```

- [ ] **Step 4: Run the help test to verify it passes**

Run:

```bash
go test ./internal/commands/support
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/commands/support/support.go internal/commands/support/support_test.go
git commit -m "feat: expand help output"
```

---

### Task 2: Add Status Report Rendering Tests

**Files:**
- Modify: `internal/commands/status/status.go`
- Create: `internal/commands/status/status_test.go`

- [ ] **Step 1: Write failing render tests**

Create `internal/commands/status/status_test.go`:

```go
package status

import (
	"bytes"
	"errors"
	"strings"
	"testing"
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
```

- [ ] **Step 2: Run the status tests to verify they fail**

Run:

```bash
go test ./internal/commands/status
```

Expected: FAIL because `report` and `renderReport` do not exist yet.

- [ ] **Step 3: Add the report model and renderer**

Add these imports to `internal/commands/status/status.go`:

```go
import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"konfirm/internal/context"
	"konfirm/internal/store"
)
```

Add this type and rendering code to `internal/commands/status/status.go`:

```go
type report struct {
	Context        string
	ContextErr     error
	ContextMissing bool
	ConfigPath     string
	ConfigErr      error
	Allowed        []string
	KubectlFound   bool
	FzfFound       bool
	DetectedShell  string
}

func renderReport(w io.Writer, r report) {
	fmt.Fprintln(w, "Context")
	if r.ContextErr != nil {
		fmt.Fprintf(w, "  current: unavailable (%v)\n", r.ContextErr)
	} else {
		fmt.Fprintf(w, "  current: %s\n", r.Context)
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Config")
	if r.ConfigPath != "" {
		fmt.Fprintf(w, "  path: %s\n", r.ConfigPath)
	}
	if r.ConfigErr != nil {
		fmt.Fprintf(w, "  error: %v\n", r.ConfigErr)
	} else if r.ContextMissing {
		fmt.Fprintln(w, "  allowed for current context: unavailable because current context could not be resolved")
	} else if len(r.Allowed) == 0 {
		fmt.Fprintln(w, "  allowed for current context: (none)")
	} else {
		fmt.Fprintln(w, "  allowed for current context:")
		for _, subcommand := range r.Allowed {
			fmt.Fprintf(w, "    %s\n", subcommand)
		}
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Dependencies")
	fmt.Fprintf(w, "  kubectl: %s\n", foundLabel(r.KubectlFound))
	fmt.Fprintf(w, "  fzf: %s\n", foundLabel(r.FzfFound))

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Shell setup")
	fmt.Fprintf(w, "  detected shell: %s\n", r.DetectedShell)
	fmt.Fprintf(w, "  completion: %s\n", completionHint(r.DetectedShell))
	fmt.Fprintln(w, "  recommended alias: alias k=\"konfirm kubectl\"")
}

func foundLabel(found bool) string {
	if found {
		return "found"
	}
	return "missing"
}

func completionHint(shell string) string {
	switch shell {
	case "zsh":
		return "add `source <(konfirm completion zsh)` to ~/.zshrc"
	case "fish":
		return "add `konfirm completion fish | source` to ~/.config/fish/config.fish"
	default:
		return "run `konfirm completion zsh` or `konfirm completion fish` for setup guidance"
	}
}
```

At this stage, do not fully rewrite `Run` yet. Remove the unused `constants` import if it remains.

- [ ] **Step 4: Run the status render tests**

Run:

```bash
go test ./internal/commands/status
```

Expected: PASS for the render tests, unless `Run` still has compile errors from unused imports. Fix only compile errors caused by imports.

- [ ] **Step 5: Commit**

```bash
git add internal/commands/status/status.go internal/commands/status/status_test.go
git commit -m "feat: add status report renderer"
```

---

### Task 3: Wire Status Diagnostics Into `Run`

**Files:**
- Modify: `internal/commands/status/status.go`
- Modify: `internal/commands/status/status_test.go`

- [ ] **Step 1: Add failing command behavior tests**

Append these tests to `internal/commands/status/status_test.go`:

```go
func TestRunPrintsDiagnosticsAndReturnsZero(t *testing.T) {
	restore := stubDeps(t, deps{
		stdout: bytes.NewBuffer(nil),
		stderr: bytes.NewBuffer(nil),
		configPath: func() (string, error) {
			return "/tmp/konfirm/config.json", nil
		},
		loadConfig: func() (store.Config, error) {
			return store.Config{
				PermanentAllowKubectlSubcmds: map[string][]string{
					"prod-cluster": []string{"get", "logs"},
				},
			}, nil
		},
		currentContext: func() (string, error) {
			return "prod-cluster", nil
		},
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(name string) string {
			if name == "SHELL" {
				return "/bin/zsh"
			}
			return ""
		},
	})
	defer restore()

	code := Run(nil)
	if code != 0 {
		t.Fatalf("Run(nil) code = %d, want 0", code)
	}

	got := commandDeps.stdout.(*bytes.Buffer).String()
	if !strings.Contains(got, "  current: prod-cluster") {
		t.Fatalf("stdout missing current context:\n%s", got)
	}
	if !strings.Contains(got, "    get") || !strings.Contains(got, "    logs") {
		t.Fatalf("stdout missing allowlist:\n%s", got)
	}
}

func TestRunContinuesWhenContextFails(t *testing.T) {
	restore := stubDeps(t, deps{
		stdout: bytes.NewBuffer(nil),
		stderr: bytes.NewBuffer(nil),
		configPath: func() (string, error) {
			return "/tmp/konfirm/config.json", nil
		},
		loadConfig: func() (store.Config, error) {
			return store.Config{}, nil
		},
		currentContext: func() (string, error) {
			return "", errors.New("current-context failed")
		},
		lookPath: func(name string) (string, error) {
			return "", errors.New("missing")
		},
		getenv: func(name string) string {
			return ""
		},
	})
	defer restore()

	code := Run(nil)
	if code != 0 {
		t.Fatalf("Run(nil) code = %d, want 0", code)
	}

	got := commandDeps.stdout.(*bytes.Buffer).String()
	required := []string{
		"  current: unavailable (current-context failed)",
		"  allowed for current context: unavailable because current context could not be resolved",
		"  kubectl: missing",
		"  fzf: missing",
		"  detected shell: unknown",
	}
	for _, want := range required {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout missing %q\noutput:\n%s", want, got)
		}
	}
}

func TestRunReturnsOneWhenConfigLoadFails(t *testing.T) {
	restore := stubDeps(t, deps{
		stdout: bytes.NewBuffer(nil),
		stderr: bytes.NewBuffer(nil),
		configPath: func() (string, error) {
			return "/tmp/konfirm/config.json", nil
		},
		loadConfig: func() (store.Config, error) {
			return store.Config{}, errors.New("bad json")
		},
		currentContext: func() (string, error) {
			return "prod-cluster", nil
		},
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(name string) string {
			return "/bin/fish"
		},
	})
	defer restore()

	code := Run(nil)
	if code != 1 {
		t.Fatalf("Run(nil) code = %d, want 1", code)
	}

	got := commandDeps.stdout.(*bytes.Buffer).String()
	if !strings.Contains(got, "  error: bad json") {
		t.Fatalf("stdout missing config error:\n%s", got)
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
		t.Fatalf("Run with args code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: konfirm status") {
		t.Fatalf("stderr missing usage:\n%s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}
```

Add these imports to `status_test.go` if they are not already present:

```go
import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"konfirm/internal/store"
)
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```bash
go test ./internal/commands/status
```

Expected: FAIL because `deps`, `commandDeps`, and `stubDeps` do not exist, and `Run` still uses hard-coded dependencies.

- [ ] **Step 3: Add dependency seams and status collection**

Update `internal/commands/status/status.go` so it contains these dependency definitions:

```go
type deps struct {
	stdout         io.Writer
	stderr         io.Writer
	configPath     func() (string, error)
	loadConfig     func() (store.Config, error)
	currentContext func() (string, error)
	lookPath       func(string) (string, error)
	getenv         func(string) string
}

var commandDeps = deps{
	stdout:         os.Stdout,
	stderr:         os.Stderr,
	configPath:     store.ConfigPath,
	loadConfig:     store.LoadConfig,
	currentContext: context.GetCurrentContext,
	lookPath:       exec.LookPath,
	getenv:         os.Getenv,
}
```

Replace `Run` in `internal/commands/status/status.go` with:

```go
func Run(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(commandDeps.stderr, "usage: konfirm status")
		return 2
	}

	r, exitCode := collectReport(commandDeps)
	renderReport(commandDeps.stdout, r)
	return exitCode
}
```

Add these helper functions:

```go
func collectReport(d deps) (report, int) {
	r := report{
		KubectlFound:  hasCommand(d, "kubectl"),
		FzfFound:      hasCommand(d, "fzf"),
		DetectedShell: detectShell(d.getenv("SHELL")),
	}

	ctx, ctxErr := d.currentContext()
	if ctxErr != nil {
		r.ContextErr = ctxErr
		r.ContextMissing = true
	} else {
		r.Context = ctx
	}

	path, err := d.configPath()
	if err != nil {
		r.ConfigErr = err
		return r, 1
	}
	r.ConfigPath = path

	cfg, err := d.loadConfig()
	if err != nil {
		r.ConfigErr = err
		return r, 1
	}

	if !r.ContextMissing {
		r.Allowed = cfg.PermanentAllowKubectlSubcmds[r.Context]
	}

	return r, 0
}

func hasCommand(d deps, name string) bool {
	_, err := d.lookPath(name)
	return err == nil
}

func detectShell(shellPath string) string {
	shell := filepath.Base(shellPath)
	switch shell {
	case "zsh", "fish":
		return shell
	default:
		return "unknown"
	}
}
```

- [ ] **Step 4: Add test-only dependency reset helper**

Append this helper to `internal/commands/status/status_test.go`:

```go
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
		replacement.configPath = func() (string, error) {
			return "/tmp/konfirm/config.json", nil
		}
	}
	if replacement.loadConfig == nil {
		replacement.loadConfig = func() (store.Config, error) {
			return store.Config{}, nil
		}
	}
	if replacement.currentContext == nil {
		replacement.currentContext = func() (string, error) {
			return "dev-cluster", nil
		}
	}
	if replacement.lookPath == nil {
		replacement.lookPath = func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		}
	}
	if replacement.getenv == nil {
		replacement.getenv = func(name string) string {
			return ""
		}
	}

	commandDeps = replacement
	return func() {
		commandDeps = original
	}
}
```

- [ ] **Step 5: Run the status tests**

Run:

```bash
go test ./internal/commands/status
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/commands/status/status.go internal/commands/status/status_test.go
git commit -m "feat: improve status diagnostics"
```

---

### Task 4: Full Verification

**Files:**
- Read: `docs/superpowers/specs/2026-05-17-status-help-ux-design.md`
- Verify: all changed Go files

- [ ] **Step 1: Run all tests**

Run:

```bash
go test ./...
```

Expected: PASS for all packages.

- [ ] **Step 2: Run full build**

Run:

```bash
go build ./...
```

Expected: PASS with no output.

- [ ] **Step 3: Smoke-test help output**

Run:

```bash
go run ./cmd/konfirm -h
```

Expected output includes:

```text
Usage:
Commands:
Options:
Examples:
Run `konfirm status` to inspect your current context, allowlist, and local setup.
```

- [ ] **Step 4: Smoke-test status argument rejection**

Run:

```bash
go run ./cmd/konfirm status --json
```

Expected: exits with code `2` and prints:

```text
usage: konfirm status
```

- [ ] **Step 5: Smoke-test status output shape**

Run:

```bash
go run ./cmd/konfirm status
```

Expected output includes these section headers, regardless of local Kubernetes availability:

```text
Context
Config
Dependencies
Shell setup
```

If local `kubectl config current-context` fails, expected exit code is `0` and the context line says `unavailable (...)`.

- [ ] **Step 6: Review git diff**

Run:

```bash
git diff --stat
git diff
```

Expected:

- Changes are limited to `internal/commands/support`, `internal/commands/status`, and their tests.
- No `--json` support has been added.
- No new commands have been added.

- [ ] **Step 7: Commit any final fixes**

If Task 4 required fixes, commit them:

```bash
git add internal/commands/support internal/commands/status
git commit -m "test: verify status help ux"
```

If Task 4 required no fixes, do not create an empty commit.
