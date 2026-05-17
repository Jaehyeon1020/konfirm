package constants

const ASCII_LOGO = `
██╗  ██╗ ██████╗ ███╗   ██╗███████╗██╗██████╗ ███╗   ███╗
██║ ██╔╝██╔═══██╗████╗  ██║██╔════╝██║██╔══██╗████╗ ████║
█████╔╝ ██║   ██║██╔██╗ ██║█████╗  ██║██████╔╝██╔████╔██║
██╔═██╗ ██║   ██║██║╚██╗██║██╔══╝  ██║██╔══██╗██║╚██╔╝██║
██║  ██╗╚██████╔╝██║ ╚████║██║     ██║██║  ██║██║ ╚═╝ ██║     

`

const VERSION = "v0.8.3"

const (
	ANSI_BOLD_RED  = "\x1b[1;31m"
	ANSI_BOLD_BLUE = "\x1b[1;34m"
	ANSI_RESET     = "\x1b[0m"
)

func GetKubectlSubcommands() []string {
	return []string{
		"get", "describe", "create", "delete", "apply",
		"edit", "patch", "replace", "rollout", "scale",
		"autoscale", "logs", "exec", "port-forward", "proxy",
		"cp", "attach", "run", "expose", "set", "explain",
		"diff", "wait", "kustomize", "label", "annotate",
		"certificate", "cluster-info", "top", "cordon",
		"uncordon", "drain", "taint", "auth", "debug",
		"events", "api-resources", "api-versions", "config",
		"plugin", "version", "completion", "alpha",
	}
}
