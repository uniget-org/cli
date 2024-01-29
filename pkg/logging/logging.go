package logging

import (
	"fmt"

	"github.com/pterm/pterm"
)

var (
	Description = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.DescriptionMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.DescriptionPrefixStyle,
			Text:  "Description",
		},
	}

	Info = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.InfoMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoPrefixStyle,
			Text:  "INFO",
		},
	}

	Success = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessPrefixStyle,
			Text:  "SUCCESS",
		},
	}

	Error = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorPrefixStyle,
			Text:  " ERROR ",
		},
	}

	Fatal = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.FatalMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.FatalPrefixStyle,
			Text:  " FATAL ",
		},
		Fatal: true,
	}

	Warning = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningPrefixStyle,
			Text:  "WARNING",
		},
	}

	Skip = pterm.PrefixPrinter{
		MessageStyle: pterm.NewStyle(pterm.FgDarkGray),
		Prefix: pterm.Prefix{
			Style: pterm.NewStyle(pterm.FgBlack, pterm.BgGray),
			Text:  "SKIP",
		},
	}
)

var Level pterm.LogLevel = pterm.LogLevelInfo

func Debug(message string) {
	pterm.DefaultLogger.
		WithLevel(Level).
		WithTime(false).
		WithMaxWidth(1000).
		Debug(message)
}

func Debugf(message string, args ...interface{}) {
	Debug(
		fmt.Sprintf(message, args...),
	)
}

func Trace(message string) {
	pterm.DefaultLogger.
		WithLevel(Level).
		WithTime(false).
		WithMaxWidth(1000).
		Trace(message)
}

func Tracef(message string, args ...interface{}) {
	Trace(
		fmt.Sprintf(message, args...),
	)
}
