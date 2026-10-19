package cli

import (
	"testing"
)

type testCommand struct {
	called bool
}

func (c *testCommand) init() error {
	c.called = true
	return nil
}

func TestNewCommandList(t *testing.T) {
	commandList := NewCommandList()
	if commandList == nil {
		t.Fatalf("expected command list, got nil")
	}
	if len(*commandList) != 0 {
		t.Errorf("expected empty command list, got %d commands", len(*commandList))
	}
}

func TestCommandListAddCommand(t *testing.T) {
	commandList := NewCommandList()
	first := &testCommand{}
	second := &testCommand{}
	commandList.AddCommand(first)
	commandList.AddCommand(second)

	if len(*commandList) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(*commandList))
	}
	if (*commandList)[0] != first || (*commandList)[1] != second {
		t.Errorf("expected commands to be added in order")
	}
}

func TestCommandListIsCallable(t *testing.T) {
	commandList := NewCommandList()
	commandList.AddCommand(&testCommand{})

	err := (*commandList)[0].init()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !(*commandList)[0].(*testCommand).called {
		t.Errorf("expected command init to be called")
	}
}
