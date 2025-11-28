package os

import (
	"os"

	"github.com/charmbracelet/x/term"
)

func IsTty() bool {
	return term.IsTerminal(os.Stdout.Fd())
}
