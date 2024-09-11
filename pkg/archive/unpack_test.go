package archive

import (
	"bytes"
	"compress/gzip"
	"testing"
)

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

func TestGunzip(t *testing.T) {
	data := "test"
	dataBytes := []byte(data)

	var gzDataBytes bytes.Buffer
	w := gzip.NewWriter(&gzDataBytes)
	_, err := w.Write(dataBytes)
	if err != nil {
		t.Errorf("gzip failed: %v", err)
	}

	gunzipDataBytes, err := Gunzip(gzDataBytes.Bytes())
	if err != nil {
		t.Errorf("gunzip failed: %v", err)
	}

	if string(gunzipDataBytes) != data {
		t.Errorf("expected %s, got %s", data, string(gunzipDataBytes))
	}
}

func TestExtractTarGz(t *testing.T) {
	//
}
