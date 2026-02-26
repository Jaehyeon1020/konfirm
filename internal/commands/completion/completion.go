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
    kwords=("${(@)words[2,-1]}")
    words=("${kwords[@]}")
    if (( CURRENT > 1 )); then
      CURRENT=$((CURRENT-1))
    fi
    _kubectl
  fi
}

compdef _konfirm konfirm
`

const fishScript = `# konfirm kubectl completion for fish
function __konfirm_kubectl_complete
    set -l tokens (commandline -opc)
    set -e tokens[1]
    set -l cur (commandline -ct)
    complete -C (string join ' ' $tokens $cur)
end

complete -c konfirm -f -n "not __fish_seen_subcommand_from kubectl add remove config status completion help version" -a "kubectl add remove config status completion help version"
complete -c konfirm -f -n "__fish_seen_subcommand_from kubectl" -a "(__konfirm_kubectl_complete)"
`

func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "missing shell (zsh)")
		return 2
	}

	switch args[0] {
	case "zsh":
		fmt.Print(zshScript)
	case "fish":
		fmt.Print(fishScript)
	default:
		fmt.Fprintf(os.Stderr, "unsupported shell: %s (supported: zsh, fish)\n", args[0])
		return 2
	}

	return 0
}
