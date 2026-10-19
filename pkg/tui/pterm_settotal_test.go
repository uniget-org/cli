package tui

import (
	"testing"
)

func TestPtermProgressReaderSetTotal(t *testing.T) {
	// Previously, the total update callback called SetTotal again which
	// recursed infinitely.
	reader, err := NewPtermProgressReader("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	reader.SetTotal(42)

	if reader.progressPrinter.Total != 42 {
		t.Errorf("expected progressbar total to be 42, got %d", reader.progressPrinter.Total)
	}
}
