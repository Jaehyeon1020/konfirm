package support

import (
	"fmt"
	"io"
	"konfirm/internal/constants"
)

func Usage(w io.Writer) {
	fmt.Print(constants.ASCII_LOGO)
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
	fmt.Print(constants.ASCII_LOGO)
	fmt.Fprintln(w, "konfirm", constants.VERSION)
}
