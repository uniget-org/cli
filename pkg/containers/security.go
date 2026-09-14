package containers

import (
	"fmt"

	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/security"
)

func VerifyContainerImageSignature(ref *ToolRef, expectedOIDIssuer, expectedOIDIssuerRegex, expectedSAN, expectedSANRegex string) error {
	logging.Debugf("Getting image digest for reference %s", ref)
	digest, err := GetImageDigest(ref)
	if err != nil {
		return fmt.Errorf("failed to get image digest: %w", err)
	}
	imageWithDigest := fmt.Sprintf("%s@%s", ref.RepositoryString(), digest)

	logging.Debugf("Verifying signature for image %s", imageWithDigest)

	_, err = security.VerifySignstoreBundleForContainerImage(
		imageWithDigest,
		expectedOIDIssuer,
		expectedOIDIssuerRegex,
		expectedSAN,
		expectedSANRegex,
	)
	if err != nil {
		return fmt.Errorf("error verifying tool signature: %w", err)
	}

	return nil
}
