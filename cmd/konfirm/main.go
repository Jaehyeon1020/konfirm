package main

import (
	"fmt"
	"os"

	"konfirm/internal/commands/allow"
	"konfirm/internal/commands/completion"
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
	case "add":
		exitWithCode(allow.Run(append([]string{"add"}, os.Args[2:]...)))
	case "remove":
		exitWithCode(allow.Run(append([]string{"remove"}, os.Args[2:]...)))
	case "status":
		exitWithCode(status.Run(os.Args[2:]))
	case "completion":
		exitWithCode(completion.Run(os.Args[2:]))
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
