package parse

import (
	"fmt"

	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tool"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func LoadKubernetesManifest(file []byte) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	manifest, _, err := decode(file, nil, nil)
	if err != nil {
		panic(err)
	}

	return manifest, nil
}

func LoadKubernetesManifestFromFile(filename string) (runtime.Object, error) {
	file, err := myos.SlurpFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	manifest, err := LoadKubernetesManifest(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes manifest: %w", err)
	}

	return manifest, nil
}

func ExtractImageReferencesFromKubernetesManifest(manifest runtime.Object) (ImageRefs, error) {
	var imageRefs ImageRefs

	var podSpec corev1.PodSpec
	switch o := manifest.(type) {
	case *corev1.Pod:
		podSpec = o.Spec
	case *appsv1.Deployment:
		podSpec = o.Spec.Template.Spec
	default:
		return imageRefs, fmt.Errorf("unsupported manifest: %+v", o)
	}

	images := map[string]struct{}{}
	for _, container := range podSpec.Containers {
		images[container.Image] = struct{}{}
	}
	for _, initContainer := range podSpec.InitContainers {
		images[initContainer.Image] = struct{}{}
	}
	for _, volumes := range podSpec.Volumes {
		if volumes.Image != nil {
			images[volumes.Image.Reference] = struct{}{}
		}
	}

	for image := range images {
		imageRef, err := ref.New(image)
		if err != nil {
			logging.Debugf("Failed to create image reference from %s: %v", image, err)
			continue
		}
		imageRefs.Add(imageRef)
	}

	return imageRefs, nil
}

func BumpKubernetesFile(filename string, tools *tool.Tools) error {
	manifest, err := LoadKubernetesManifestFromFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load kubernetes file: %w", err)
	}

	imageRefs, err := ExtractImageReferencesFromKubernetesManifest(manifest)
	if err != nil {
		return fmt.Errorf("failed to extract image references: %w", err)
	}
	if len(imageRefs.Refs) == 0 {
		logging.Warning.Printfln("No image references found in kubernetes file %s", filename)
		return nil
	}

	err = ReplaceInFile(filename, &imageRefs, tools)
	if err != nil {
		return fmt.Errorf("failed to bump image references: %w", err)
	}

	return nil
}
