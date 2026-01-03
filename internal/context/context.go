package context

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func GetEffectiveContext(args []string) (string, error) {
	ctx, err := GetContextFromArgs(args)
	if err != nil {
		return "", err
	}
	if ctx != "" {
		return ctx, nil
	}
	return GetCurrentContext()
}

func GetContextFromArgs(args []string) (string, error) {
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

func GetCurrentContext() (string, error) {
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
