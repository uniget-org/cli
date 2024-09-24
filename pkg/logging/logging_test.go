package logging

import (
	"bytes"
	"io"
	"testing"
)

func TestInit(t *testing.T) {
	var outWriter io.Writer
	var errWriter io.Writer

	OutputWriter = outWriter
	ErrorWriter = errWriter

	Init()

	if Description.Writer != outWriter {
		t.Errorf("Description.Writer = %v; want %v", Description.Writer, outWriter)
	}
	if Info.Writer != outWriter {
		t.Errorf("Info.Writer = %v; want %v", Info.Writer, outWriter)
	}
	if Success.Writer != outWriter {
		t.Errorf("Success.Writer = %v; want %v", Success.Writer, outWriter)
	}
	if Error.Writer != errWriter {
		t.Errorf("Error.Writer = %v; want %v", Error.Writer, errWriter)
	}
	if Fatal.Writer != errWriter {
		t.Errorf("Fatal.Writer = %v; want %v", Fatal.Writer, errWriter)
	}
	if Warning.Writer != outWriter {
		t.Errorf("Warning.Writer = %v; want %v", Warning.Writer, outWriter)
	}
	if Skip.Writer != outWriter {
		t.Errorf("Skip.Writer = %v; want %v", Skip.Writer, outWriter)
	}
}

func TestPrefixWriters(t *testing.T) {
	var outBuffer bytes.Buffer
	OutputWriter = &outBuffer

	var errBuffer bytes.Buffer
	ErrorWriter = &errBuffer

	Init()

	Description.Println("description")
	Info.Println("info")
	Success.Println("success")
	Error.Println("error")
	Warning.Println("warning")
	Skip.Println("skip")

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}
	if len(errBuffer.String()) == 0 {
		t.Errorf("errBuffer is empty")
	}

	expectedOut := "" +
		" DESCRIPTION  description" + "\n" +
		" INFO  info" + "\n" +
		" SUCCESS  success" + "\n" +
		" WARNING  warning" + "\n" +
		" SKIP  skip" + "\n"
	expectedErr := "" +
		"  ERROR   error" + "\n"

	strippedOut := StripAnsi(outBuffer.String())
	strippedErr := StripAnsi(errBuffer.String())

	if strippedOut != expectedOut {
		t.Errorf("outBuffer = %s; want %s", strippedOut, expectedOut)
	}
	if strippedErr != expectedErr {
		t.Errorf("errBuffer = %s; want %s", strippedErr, expectedErr)
	}
}

func TestLoggers(t *testing.T) {
	var outBuffer bytes.Buffer
	OutputWriter = &outBuffer

	Init()

	Debug("foo")
	Trace("bar")

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		"DEBUG foo" + "\n" +
		"TRACE bar" + "\n"
	strippedOut := StripAnsi(outBuffer.String())
	if strippedOut != expectedOut {
		t.Errorf("outBuffer = <%s>; want <%s>", strippedOut, expectedOut)
	}
}
