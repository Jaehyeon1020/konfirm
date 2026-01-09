package completion

import (
	"fmt"
	"os"
)

const zshScript = `# konfirm kubectl completion for zsh
_konfirm() {
  if [[ $words[2] == "kubectl" ]]; then
    if (( ! $+functions[_kubectl] )); then
      if command -v kubectl >/dev/null 2>&1; then
        source <(kubectl completion zsh 2>/dev/null)
      fi
    fi

    if (( ! $+functions[_kubectl] )); then
      return
    fi

    local -a kwords
    kwords=($words[2,-1])
    words=($kwords)
    CURRENT=$((CURRENT-1))
    _kubectl
  fi
}

compdef _konfirm konfirm
`

func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "missing shell (zsh)")
		return 2
	}

	switch args[0] {
	case "zsh":
		fmt.Print(zshScript)
	default:
		fmt.Fprintf(os.Stderr, "unsupported shell: %s\n", args[0])
		return 2
	}

	return 0
}
