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
		"  konfirm kubecolor <kubectl args...>",
		"  konfirm <command> [args...]",
		"Commands:",
		"  kubectl",
		"  kubecolor",
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
		"  konfirm kubecolor get pods",
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
