package logging

import (
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
)

var (
	OutputWriter io.Writer      = os.Stdout
	ErrorWriter  io.Writer      = os.Stderr
	Level        pterm.LogLevel = pterm.LogLevelInfo
	Description  pterm.PrefixPrinter
	Info         pterm.PrefixPrinter
	Success      pterm.PrefixPrinter
	Error        pterm.PrefixPrinter
	Fatal        pterm.PrefixPrinter
	Warning      pterm.PrefixPrinter
	Skip         pterm.PrefixPrinter
)

func Init() {
	Description = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.DescriptionMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.DescriptionPrefixStyle,
			Text:  "DESCRIPTION",
		},
		Writer: OutputWriter,
	}

	Info = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.InfoMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoPrefixStyle,
			Text:  "INFO",
		},
		Writer: OutputWriter,
	}

	Success = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessPrefixStyle,
			Text:  "SUCCESS",
		},
		Writer: OutputWriter,
	}

	Error = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorPrefixStyle,
			Text:  " ERROR ",
		},
		Writer: ErrorWriter,
	}

	Fatal = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.FatalMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.FatalPrefixStyle,
			Text:  " FATAL ",
		},
		Writer: ErrorWriter,
		Fatal:  true,
	}

	Warning = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningPrefixStyle,
			Text:  "WARNING",
		},
		Writer: OutputWriter,
	}

	Skip = pterm.PrefixPrinter{
		MessageStyle: pterm.NewStyle(pterm.FgDarkGray),
		Prefix: pterm.Prefix{
			Style: pterm.NewStyle(pterm.FgBlack, pterm.BgGray),
			Text:  "SKIP",
		},
		Writer: OutputWriter,
	}
}

func Debug(message string) {
	pterm.DefaultLogger.
		WithLevel(Level).
		WithTime(false).
		WithMaxWidth(1000).
		WithWriter(OutputWriter).
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
		WithWriter(OutputWriter).
		Trace(message)
}

func Tracef(message string, args ...interface{}) {
	Trace(
		fmt.Sprintf(message, args...),
	)
}
