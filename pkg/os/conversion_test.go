package os

import "testing"

func TestConvertBytesToHumanReadable(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{1, "1 B"},
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1 KB"},
		{2048, "2 KB"},
		{1048576, "1 MB"},
		{1073741824, "1 GB"},
		{1099511627776, "1 TB"},
	}

	for _, test := range tests {
		result := ConvertBytesToHumanReadable(test.size)
		if result != test.expected {
			t.Errorf("ConvertBytesToHumanReadable(%d) = %s; want %s", test.size, result, test.expected)
		}
	}
}
