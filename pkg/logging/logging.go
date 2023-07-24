package logging

import "github.com/pterm/pterm"

var (
	Info    = pterm.Info
	Error   = pterm.Error
	Debug   = pterm.Debug
	Warning = pterm.Warning
	Skip    = pterm.PrefixPrinter{
		MessageStyle: pterm.NewStyle(pterm.FgDarkGray),
		Prefix: pterm.Prefix{
			Style: pterm.NewStyle(pterm.FgBlack, pterm.BgGray),
			Text:  "SKIP",
		},
	}
)
