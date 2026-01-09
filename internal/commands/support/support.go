package support

import (
	"fmt"
	"io"
)

const VersionText = "v0.1.0"

func Usage(w io.Writer) {
	fmt.Fprintln(w, "konfirm - confirm kubectl before execution")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  konfirm kubectl <kubectl args...>")
	fmt.Fprintln(w, "  konfirm add <subcommand>")
	fmt.Fprintln(w, "  konfirm add --all")
	fmt.Fprintln(w, "  konfirm remove <subcommand>")
	fmt.Fprintln(w, "  konfirm remove --all")
	fmt.Fprintln(w, "  konfirm status")
}

func Version(w io.Writer) {
	fmt.Fprintln(w, VersionText)
}
