package archive

import "testing"

func TestPathIsInsideTarget(t *testing.T) {
	var err error

	err = pathIsInsideTarget("/tmp", "/tmp")
	if err != nil {
		t.Errorf("expected /tmp to be inside /tmp")
	}

	err = pathIsInsideTarget("/tmp", "/tmp/foo")
	if err != nil {
		t.Errorf("expected /tmp/foo to be inside /tmp")
	}

	err = pathIsInsideTarget("/tmp/foo", "/tmp")
	if err != nil {
		t.Errorf("expected /tmp/foo not to be inside /tmp")
	}
}
