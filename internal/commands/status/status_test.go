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
