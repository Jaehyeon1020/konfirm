package kubectl

import (
	"errors"
	"testing"

	"konfirm/internal/store"
)

func TestRunKubecolorExecutesKubecolorWhenSubcommandAllowed(t *testing.T) {
	var executedName string
	var executedArgs []string

	restore := stubDeps(t, deps{
		effectiveContext: func(args []string) (string, error) {
			return "prod-cluster", nil
		},
		loadConfig: func() (store.Config, error) {
			return store.Config{
				PermanentAllowKubectlSubcmds: map[string][]string{
					"prod-cluster": {"get"},
				},
			}, nil
		},
		promptForApproval: func(ctx string) error {
			t.Fatal("promptForApproval should not be called for allowed subcommand")
			return nil
		},
		execCommand: func(name string, args []string) int {
			executedName = name
			executedArgs = append([]string(nil), args...)
			return 0
		},
	})
	defer restore()

	code := RunKubecolor([]string{"get", "pods"})

	if code != 0 {
		t.Fatalf("RunKubecolor code = %d, want 0", code)
	}
	if executedName != "kubecolor" {
		t.Fatalf("executed command = %q, want kubecolor", executedName)
	}
	if len(executedArgs) != 2 || executedArgs[0] != "get" || executedArgs[1] != "pods" {
		t.Fatalf("executed args = %#v, want [get pods]", executedArgs)
	}
}

func TestRunKubectlStillExecutesKubectl(t *testing.T) {
	var executedName string

	restore := stubDeps(t, deps{
		effectiveContext: func(args []string) (string, error) {
			return "prod-cluster", nil
		},
		loadConfig: func() (store.Config, error) {
			return store.Config{
				PermanentAllowKubectlSubcmds: map[string][]string{
					"prod-cluster": {"get"},
				},
			}, nil
		},
		promptForApproval: func(ctx string) error {
			t.Fatal("promptForApproval should not be called for allowed subcommand")
			return nil
		},
		execCommand: func(name string, args []string) int {
			executedName = name
			return 0
		},
	})
	defer restore()

	code := Run([]string{"get", "pods"})

	if code != 0 {
		t.Fatalf("Run code = %d, want 0", code)
	}
	if executedName != "kubectl" {
		t.Fatalf("executed command = %q, want kubectl", executedName)
	}
}

func TestRunKubecolorPromptsBeforeExecutingWhenSubcommandNotAllowed(t *testing.T) {
	prompted := false
	executed := false

	restore := stubDeps(t, deps{
		effectiveContext: func(args []string) (string, error) {
			return "prod-cluster", nil
		},
		loadConfig: func() (store.Config, error) {
			return store.Config{}, nil
		},
		promptForApproval: func(ctx string) error {
			if ctx != "prod-cluster" {
				t.Fatalf("prompt context = %q, want prod-cluster", ctx)
			}
			prompted = true
			return nil
		},
		execCommand: func(name string, args []string) int {
			if name != "kubecolor" {
				t.Fatalf("executed command = %q, want kubecolor", name)
			}
			executed = true
			return 0
		},
	})
	defer restore()

	code := RunKubecolor([]string{"delete", "pod", "nginx"})

	if code != 0 {
		t.Fatalf("RunKubecolor code = %d, want 0", code)
	}
	if !prompted {
		t.Fatal("RunKubecolor did not prompt before executing")
	}
	if !executed {
		t.Fatal("RunKubecolor did not execute after approval")
	}
}

func TestRunKubecolorDoesNotExecuteWhenApprovalFails(t *testing.T) {
	restore := stubDeps(t, deps{
		effectiveContext: func(args []string) (string, error) {
			return "prod-cluster", nil
		},
		loadConfig: func() (store.Config, error) {
			return store.Config{}, nil
		},
		promptForApproval: func(ctx string) error {
			return errors.New("Aborted")
		},
		execCommand: func(name string, args []string) int {
			t.Fatal("execCommand should not be called when approval fails")
			return 0
		},
	})
	defer restore()

	code := RunKubecolor([]string{"delete", "pod", "nginx"})

	if code != 1 {
		t.Fatalf("RunKubecolor code = %d, want 1", code)
	}
}

func stubDeps(t *testing.T, d deps) func() {
	t.Helper()

	previous := commandDeps

	if d.effectiveContext == nil {
		d.effectiveContext = previous.effectiveContext
	}
	if d.loadConfig == nil {
		d.loadConfig = previous.loadConfig
	}
	if d.promptForApproval == nil {
		d.promptForApproval = previous.promptForApproval
	}
	if d.execCommand == nil {
		d.execCommand = previous.execCommand
	}

	commandDeps = d
	return func() {
		commandDeps = previous
	}
}
