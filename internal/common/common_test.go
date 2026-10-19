package common

import (
	"bytes"
	"io"
	"testing"

	myos "gitlab.com/uniget-org/cli/pkg/os"
)

func TestCreateProgressReader(t *testing.T) {
	isTty := myos.IsTty()
	tests := []struct {
		name          string
		suppress      bool
		expectedQuiet bool
	}{
		{
			name:          "suppressed is always quiet",
			suppress:      true,
			expectedQuiet: true,
		},
		{
			name:          "not suppressed depends on tty",
			suppress:      false,
			expectedQuiet: !isTty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := CreateProgressReader("test", tt.suppress)
			if reader.IsQuiet() != tt.expectedQuiet {
				t.Errorf("expected IsQuiet to be %v, got %v", tt.expectedQuiet, reader.IsQuiet())
			}
		})
	}
}

func TestCreateProgressReaderReadsData(t *testing.T) {
	data := []byte("hello world")
	reader := CreateProgressReader("test", true)
	reader.SetReader(io.NopCloser(bytes.NewReader(data)))

	if reader.IsQuiet() != true {
		t.Errorf("expected suppressed reader to be quiet")
	}

	buf := make([]byte, len(data))
	n, err := io.ReadFull(&reader, buf)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if n != len(data) {
		t.Errorf("expected to read %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(buf, data) {
		t.Errorf("expected %s, got %s", string(data), string(buf))
	}
}
