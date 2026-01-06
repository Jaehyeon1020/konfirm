package main

import (
	"fmt"
	"os"

	"konfirm/internal/commands/allow"
	"konfirm/internal/commands/kubectl"
	"konfirm/internal/commands/status"
	"konfirm/internal/commands/support"
)

func main() {
	if len(os.Args) < 2 {
		support.Usage(os.Stderr)
		exitWithCode(2)
	}

	switch os.Args[1] {
	case "kubectl", "k":
		exitWithCode(kubectl.Run(os.Args[2:]))
	case "allow":
		exitWithCode(allow.Run(os.Args[2:]))
	case "status":
		exitWithCode(status.Run(os.Args[2:]))
	case "-h", "--help", "help":
		support.Usage(os.Stdout)
	case "version":
		support.Version(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		support.Usage(os.Stderr)
		exitWithCode(2)
	}
}

func exitWithCode(code int) {
	os.Exit(code)
}
