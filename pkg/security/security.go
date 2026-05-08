package security

import (
	"fmt"
	"os"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/util"
	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/theupdateframework/go-tuf/v2/metadata/fetcher"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

func GetSigstoreTrustedRoot() (*root.TrustedRoot, error) {
	var trustedRootJSON []byte
	opts := tuf.DefaultOptions()
	opts.RepositoryBaseURL = "https://tuf-repo-cdn.sigstore.dev"
	fetcher := fetcher.NewDefaultFetcher()
	fetcher.SetHTTPUserAgent(util.ConstructUserAgent())
	opts.Fetcher = fetcher
	client, err := tuf.New(opts)
	if err != nil {
		return nil, fmt.Errorf("error creating TUF client: %s", err)
	}
	trustedRootJSON, err = client.GetTarget("trusted_root.json")
	if err != nil {
		return nil, fmt.Errorf("error getting trusted root: %s", err)
	}
	trustedRoot, err := root.NewTrustedRootFromJSON(trustedRootJSON)
	if err != nil {
		return nil, fmt.Errorf("error creating trusted root from JSON: %s", err)
	}
	return trustedRoot, nil
}

func VerifySigstoreBundle(artifactPath string, bundlePath string, expectedOIDIssuer, expectedOIDIssuerRegex, expectedSAN, expectedSANRegex string) (bool, error) {
	logging.Tracef("Verifying cosign bundle with artifact path %s and bundle path %s", artifactPath, bundlePath)

	b, err := bundle.LoadJSONFromPath(bundlePath)
	if err != nil {
		return false, fmt.Errorf("error loading bundle from path %s: %s", bundlePath, err)
	}

	file, err := os.Open(artifactPath) // #nosec G304 -- We need to open the file to verify it, and the path is controlled by user input
	if err != nil {
		return false, fmt.Errorf("error opening artifact file %s: %s", artifactPath, err)
	}

	var trustedMaterial = make(root.TrustedMaterialCollection, 0)
	trustedRoot, err := GetSigstoreTrustedRoot()
	if err != nil {
		return false, fmt.Errorf("error getting Sigstore trusted root: %s", err)
	}
	trustedMaterial = append(trustedMaterial, trustedRoot)

	verifierConfig := []verify.VerifierOption{}
	verifierConfig = append(verifierConfig, verify.WithSignedCertificateTimestamps(1))
	verifierConfig = append(verifierConfig, verify.WithObserverTimestamps(1))
	verifierConfig = append(verifierConfig, verify.WithTransparencyLog(1))

	identityPolicies := []verify.PolicyOption{}
	certID, err := verify.NewShortCertificateIdentity(expectedOIDIssuer, expectedOIDIssuerRegex, expectedSAN, expectedSANRegex)
	if err != nil {
		return false, fmt.Errorf("error creating short certificate identity: %s", err)
	}
	identityPolicies = append(identityPolicies, verify.WithCertificateIdentity(certID))

	sev, err := verify.NewVerifier(trustedMaterial, verifierConfig...)
	if err != nil {
		return false, fmt.Errorf("error creating verifier: %s", err)
	}

	artifactPolicy := verify.WithArtifact(file)
	_, err = sev.Verify(b, verify.NewPolicy(artifactPolicy, identityPolicies...))
	if err != nil {
		return false, fmt.Errorf("error verifying bundle: %s", err)
	}

	return true, nil
}
