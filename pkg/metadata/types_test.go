package metadata

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/source"
	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type fakeDownloader struct {
	err    error
	called bool
	src    *source.Source
	data   string
}

func (d *fakeDownloader) Get(src *source.Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	d.called = true
	d.src = src
	if d.err != nil {
		return d.err
	}

	return callback(io.NopCloser(strings.NewReader(d.data)))
}

type fakeUnpacker struct {
	err     error
	called  bool
	payload string
}

func (u *fakeUnpacker) Unpack(reader io.ReadCloser) error {
	u.called = true
	defer func() {
		_ = reader.Close()
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	u.payload = string(data)

	return u.err
}

type fakeVerifier struct {
	err            error
	called         bool
	metadataSource *MetadataSource
}

func (v *fakeVerifier) Verify(metadataSource *MetadataSource) error {
	v.called = true
	v.metadataSource = metadataSource

	return v.err
}

func TestNewMetadataSource(t *testing.T) {
	t.Run("creates source for file url", func(t *testing.T) {
		metadataSource, err := NewMetadataSource(
			"file:///tmp/metadata.tgz",
			t.TempDir(),
			cache.CacheNone,
			nil,
			nil,
			nil,
			nil,
		)
		if err != nil {
			t.Fatalf("NewMetadataSource() unexpected error: %v", err)
		}
		if metadataSource == nil {
			t.Fatal("NewMetadataSource() returned nil metadata source")
		}
		if metadataSource.Source == nil {
			t.Fatal("Source is nil")
		}
		if got, want := metadataSource.Source.Url, "file:///tmp/metadata.tgz"; got != want {
			t.Errorf("Source.Url = %q, want %q", got, want)
		}
		if metadataSource.Downloader == nil {
			t.Fatal("Downloader is nil")
		}
	})

	t.Run("unsupported scheme returns error", func(t *testing.T) {
		_, err := NewMetadataSource(
			"unsupported://metadata.tgz",
			t.TempDir(),
			cache.CacheNone,
			nil,
			nil,
			nil,
			nil,
		)
		if err == nil {
			t.Fatal("NewMetadataSource() expected error for unsupported scheme, got nil")
		}
	})
}

func TestMetadataSourceDownload(t *testing.T) {
	t.Run("successfully downloads unpacks and verifies", func(t *testing.T) {
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() unexpected error: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(currentDir)
		})

		downloaderImpl := &fakeDownloader{data: "payload"}
		var downloader source.Downloader = downloaderImpl

		unpackerImpl := &fakeUnpacker{}
		var unpacker Unpacker = unpackerImpl

		verifierImpl := &fakeVerifier{}
		var verifier MetadataVerifier = verifierImpl

		directory := t.TempDir()
		metadataSource := &MetadataSource{
			Source:     &source.Source{Url: "file:///tmp/metadata.tgz"},
			Downloader: &downloader,
			Directory:  directory,
			Unpacker:   &unpacker,
			Verifier:   &verifier,
		}

		err = metadataSource.Download(tui.NewQuietProgressReader())
		if err != nil {
			t.Fatalf("Download() unexpected error: %v", err)
		}
		if !downloaderImpl.called {
			t.Fatal("expected downloader to be called")
		}
		if downloaderImpl.src != metadataSource.Source {
			t.Fatal("expected downloader to receive metadata source Source")
		}
		if !unpackerImpl.called {
			t.Fatal("expected unpacker to be called")
		}
		if got, want := unpackerImpl.payload, "payload"; got != want {
			t.Errorf("unpacker payload = %q, want %q", got, want)
		}
		if !verifierImpl.called {
			t.Fatal("expected verifier to be called")
		}
		if verifierImpl.metadataSource != metadataSource {
			t.Fatal("expected verifier to receive metadata source")
		}
		if got, want := mustGetwd(t), directory; got != want {
			t.Errorf("working directory = %q, want %q", got, want)
		}
	})

	t.Run("change directory failure returns error", func(t *testing.T) {
		downloaderImpl := &fakeDownloader{}
		var downloader source.Downloader = downloaderImpl

		unpackerImpl := &fakeUnpacker{}
		var unpacker Unpacker = unpackerImpl

		verifierImpl := &fakeVerifier{}
		var verifier MetadataVerifier = verifierImpl

		metadataSource := &MetadataSource{
			Source:     &source.Source{Url: "file:///tmp/metadata.tgz"},
			Downloader: &downloader,
			Directory:  filepath.Join(t.TempDir(), "does", "not", "exist"),
			Unpacker:   &unpacker,
			Verifier:   &verifier,
		}

		err := metadataSource.Download(tui.NewQuietProgressReader())
		if err == nil {
			t.Fatal("Download() expected error when directory does not exist, got nil")
		}
		if !strings.Contains(err.Error(), "error changing directory") {
			t.Fatalf("Download() error = %q, want changing directory error", err)
		}
		if downloaderImpl.called {
			t.Fatal("expected downloader not to be called")
		}
	})

	t.Run("downloader failure is returned", func(t *testing.T) {
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() unexpected error: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(currentDir)
		})

		downloaderImpl := &fakeDownloader{err: errors.New("download failed")}
		var downloader source.Downloader = downloaderImpl

		unpackerImpl := &fakeUnpacker{}
		var unpacker Unpacker = unpackerImpl

		verifierImpl := &fakeVerifier{}
		var verifier MetadataVerifier = verifierImpl

		metadataSource := &MetadataSource{
			Source:     &source.Source{Url: "file:///tmp/metadata.tgz"},
			Downloader: &downloader,
			Directory:  t.TempDir(),
			Unpacker:   &unpacker,
			Verifier:   &verifier,
		}

		err = metadataSource.Download(tui.NewQuietProgressReader())
		if err == nil {
			t.Fatal("Download() expected downloader error, got nil")
		}
		if !strings.Contains(err.Error(), "download failed") {
			t.Fatalf("Download() error = %q, want downloader error", err)
		}
		if unpackerImpl.called {
			t.Fatal("expected unpacker not to be called")
		}
		if verifierImpl.called {
			t.Fatal("expected verifier not to be called")
		}
	})

	t.Run("verifier failure is wrapped", func(t *testing.T) {
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() unexpected error: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(currentDir)
		})

		downloaderImpl := &fakeDownloader{data: "payload"}
		var downloader source.Downloader = downloaderImpl

		unpackerImpl := &fakeUnpacker{}
		var unpacker Unpacker = unpackerImpl

		verifierImpl := &fakeVerifier{err: errors.New("verify failed")}
		var verifier MetadataVerifier = verifierImpl

		metadataSource := &MetadataSource{
			Source:     &source.Source{Url: "file:///tmp/metadata.tgz"},
			Downloader: &downloader,
			Directory:  t.TempDir(),
			Unpacker:   &unpacker,
			Verifier:   &verifier,
		}

		err = metadataSource.Download(tui.NewQuietProgressReader())
		if err == nil {
			t.Fatal("Download() expected verifier error, got nil")
		}
		if !strings.Contains(err.Error(), "error verifying metadata: verify failed") {
			t.Fatalf("Download() error = %q, want wrapped verifier error", err)
		}
		if !unpackerImpl.called {
			t.Fatal("expected unpacker to be called before verifier error")
		}
	})
}

func mustGetwd(t *testing.T) string {
	t.Helper()

	directory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() unexpected error: %v", err)
	}

	return directory
}
