package metadata

import (
	"fmt"
	"io"
	"os"

	"gitlab.com/uniget-org/cli/pkg/source"
	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tool"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type MetadataSource struct {
	Source             *source.Source
	Downloader         *source.Downloader
	CacheType          cache.CacheType
	CacheConfiguration source.CacheConfiguration
	Unpacker           *Unpacker
	Directory          string
	Files              map[string]string
	Verifier           *MetadataVerifier
}

func NewMetadataSource(
	url string,
	directory string,
	cacheType cache.CacheType,
	cacheConfiguration source.CacheConfiguration,
	unpacker *Unpacker,
	files map[string]string,
	verifier *MetadataVerifier,
) (*MetadataSource, error) {
	metadataSource := &source.Source{
		Url: url,
	}
	downloader, err := source.NewBackendFromScheme(metadataSource, cacheType, cacheConfiguration)
	if err != nil {
		return nil, err
	}

	return &MetadataSource{
		Source:             metadataSource,
		Downloader:         &downloader,
		CacheType:          cacheType,
		CacheConfiguration: cacheConfiguration,
		Directory:          directory,
		Unpacker:           unpacker,
		Files:              files,
		Verifier:           verifier,
	}, nil
}

func (m *MetadataSource) Download(p tui.ProgressReader) error {
	err := os.Chdir(m.Directory)
	if err != nil {
		return fmt.Errorf("error changing directory to %s: %s", m.Directory, err)
	}

	err = (*m.Downloader).Get(m.Source, p, func(reader io.ReadCloser) error {
		return (*m.Unpacker).Unpack(reader)
	})
	if err != nil {
		return fmt.Errorf("error downloading metadata: %s", err)
	}

	err = (*m.Verifier).Verify(m)
	if err != nil {
		return fmt.Errorf("error verifying metadata: %s", err)
	}

	return nil
}

func (m *MetadataSource) Load() (*tool.Tools, error) {
	tools, err := tool.LoadFromFile(m.Files["metadata.json"])
	if err != nil {
		return nil, fmt.Errorf("error loading tools from metadata.json: %s", err)
	}

	return tools, nil
}
