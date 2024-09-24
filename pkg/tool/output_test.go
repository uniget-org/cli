package tool

import (
	"bytes"
	"testing"
)

func TestListOne(t *testing.T) {
	var outBuffer bytes.Buffer

	tool := Tool{
		Name:    "test",
		Version: "1.2.3",
	}
	tool.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		"+---+------+---------+" + "\n" +
		"| # | NAME | VERSION |" + "\n" +
		"+---+------+---------+" + "\n" +
		"| 1 | test | 1.2.3   |" + "\n" +
		"+---+------+---------+" + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected %s, got %s", expectedOut, outBuffer.String())
	}
}

func TestListNone(t *testing.T) {
	var outBuffer bytes.Buffer

	tool := Tool{}
	tool.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		"+---+------+---------+" + "\n" +
		"| # | NAME | VERSION |" + "\n" +
		"+---+------+---------+" + "\n" +
		"| 1 |      |         |" + "\n" +
		"+---+------+---------+" + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected %s, got %s", expectedOut, outBuffer.String())
	}
}

func TestToolsListOne(t *testing.T) {
	var outBuffer bytes.Buffer

	tools := Tools{}
	tools.Tools = append(tools.Tools, Tool{
		Name:        "foo",
		Version:     "1.2.3",
		Description: "bar",
	})
	tools.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	t.Logf("%s", outBuffer.String())

	expectedOut := "" +
		" #  NAME  VERSION  DESCRIPTION" + "\n" +
		" 1  foo   1.2.3    bar" + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected <%s>, got <%s>", expectedOut, outBuffer.String())
	}
}

func TestToolsListNone(t *testing.T) {
	var outBuffer bytes.Buffer

	tools := Tools{}
	tools.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		" #  NAME  VERSION  DESCRIPTION" + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected <%s>, got <%s>", expectedOut, outBuffer.String())
	}
}

func TestToolsListMultiple(t *testing.T) {
	var outBuffer bytes.Buffer

	tools := Tools{}
	tools.Tools = append(tools.Tools, Tool{
		Name:        "foo",
		Version:     "1.2.3",
		Description: "bar",
	})
	tools.Tools = append(tools.Tools, Tool{
		Name:        "baz",
		Version:     "1.2.3",
		Description: "blarg",
	})
	tools.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		" #  NAME  VERSION  DESCRIPTION" + "\n" +
		" 1  foo   1.2.3    bar" + "\n" +
		" 2  baz   1.2.3    blarg" + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected <%s>, got <%s>", expectedOut, outBuffer.String())
	}
}
