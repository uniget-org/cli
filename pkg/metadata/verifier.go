package metadata

import (
	"fmt"

	"gitlab.com/uniget-org/cli/pkg/security"
	"gitlab.com/uniget-org/cli/pkg/source"
)

type MetadataVerifier interface {
	Verify(metadataSource *MetadataSource) error
}

type MetadataVerifierStruct struct {
	SignatureSource *source.Source
}

type SigstoreMetadataVerifier struct {
	MetadataVerifierStruct
	issuer      string
	issuerRegex string
	san         string
	sanRegex    string
}

func NewSigstoreMetadataVerifier(issuer string, issuerRegex string, san string, sanRegex string) *MetadataVerifier {
	var verifier MetadataVerifier = SigstoreMetadataVerifier{
		issuer:      issuer,
		issuerRegex: issuerRegex,
		san:         san,
		sanRegex:    sanRegex,
	}

	return &verifier
}

func (v SigstoreMetadataVerifier) Verify(metadataSource *MetadataSource) error {
	_, err := security.VerifySigstoreBundle(
		metadataSource.Files["metadata.json"],
		metadataSource.Files["metadata.json.sigstore.json"],
		v.issuer,
		v.issuerRegex,
		v.san,
		v.sanRegex,
	)
	if err != nil {
		return fmt.Errorf("error verifying sigstore bundle for metadata: %s", err)
	}

	return nil
}

type NullMetadataVerifier struct {
	MetadataVerifierStruct
}

func NewNullMetadataVerifier() *NullMetadataVerifier {
	return &NullMetadataVerifier{}
}

func (v NullMetadataVerifier) Verify(metadataSource *MetadataSource) error {
	return nil
}
