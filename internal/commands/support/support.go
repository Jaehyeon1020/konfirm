package support

import (
	"fmt"
	"io"
)

const VersionText = "v0.1.0"

func Usage(w io.Writer) {
	fmt.Fprintln(w, "konfirm - confirm kubectl context before execution")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  konfirm kubectl <kubectl args...>")
	fmt.Fprintln(w, "  konfirm allow add")
	fmt.Fprintln(w, "  konfirm allow remove")
	fmt.Fprintln(w, "  konfirm allow list")
	fmt.Fprintln(w, "  konfirm allow kubectl add <subcommand>")
	fmt.Fprintln(w, "  konfirm allow kubectl remove <subcommand>")
	fmt.Fprintln(w, "  konfirm allow kubectl list")
}

func Version(w io.Writer) {
	fmt.Fprintln(w, VersionText)
}
