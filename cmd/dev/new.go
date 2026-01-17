package main

import (
	"fmt"
	"os"

	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/amenzhinsky/go-memexec"
	"github.com/charmbracelet/huh"
	"github.com/google/go-github/github"
	"github.com/itchyny/gojq"
	"gitlab.com/uniget-org/cli/pkg/tool"
	"gopkg.in/yaml.v2"
)

func initNewCmd() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use: "new",
	Aliases: []string{
		"n",
		"create",
		"c",
	},
	Short: "Create new tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for clean git working directory

		toolName := args[0]
		if unigetTools.Exists(toolName) {
			return fmt.Errorf("tool %s already exists", toolName)
		}

		//fmt.Printf("Creating tool: %s\n", toolName)
		//err := os.Mkdir(fmt.Sprintf("%s/%s", unigetTools.Directory, toolName), 0755)
		//if err != nil {
		//	return fmt.Errorf("error creating tool directory: %w", err)
		//}
		//copyTemplates(toolName)

		new()

		return nil
	},
}

var githubClient = github.NewClient(nil)
var httpClient = &http.Client{}

type Project struct {
	url             string
	platform        string
	owner           string
	repository      string
	name            string
	categoryTag     string
	typeTag         string
	tags            []string
	language        string
	buildDeps       []string
	runtimeDeps     []string
	licenseName     string
	licenseLink     string
	version         string
	tagNameTemplate string
	confirmGeneral  bool
	confirmAssets   bool
	assetAmd64      *Asset
	assetArm64      *Asset
}

var project = Project{}

type Asset struct {
	Name         string
	Version      string
	Template     string
	Type         string
	IsCandidate  bool
	Platform     string
	IsLinux      bool
	Architecture string
	IsAmd64      bool
	IsArm64      bool
	HasSbom      bool
	HasSignature bool
}

func HasAsset(assets []*Asset, name string) bool {
	for _, asset := range assets {
		if asset.Name == name {
			return true
		}
	}
	return false
}

func ExtractTarGzFile(rawStream io.Reader, filename string) ([]byte, error) {
	uncompressedStream, err := gzip.NewReader(rawStream)
	if err != nil {
		return nil, fmt.Errorf("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	var bytes []byte
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		if header.Typeflag == tar.TypeReg && header.Name == filename {
			bytes, err = io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("ExtractTarGz: ReadAll failed: %s", err.Error())
			}
		}

	}

	return bytes, nil
}

func ListTarGzFiles(rawStream io.Reader) ([]string, error) {
	uncompressedStream, err := gzip.NewReader(rawStream)
	if err != nil {
		return nil, fmt.Errorf("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	files := make([]string, 0)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		if header.Typeflag == tar.TypeReg {
			files = append(files, header.Name)
		}

	}

	return files, nil
}

func new() {
	var err error

	metadata, err := tool.LoadMetadataFromRegistry(registryHost, repositoryPrefix, metadataTag)
	if err != nil {
		panic(fmt.Sprintf("error loading metadata: %s", err))
	}

	categoryTagValues, err := extractFromJsonUsingJq(metadata, `[ .tools[].tags[] ] | unique | .[] | select(. | startswith("category/"))`)
	if err != nil {
		panic(fmt.Sprintf("error extracting category tags: %s", err))
	}
	fmt.Printf("Found %d category tags\n", len(categoryTagValues))

	typeTagValues, err := extractFromJsonUsingJq(metadata, `[ .tools[].tags[] ] | unique | .[] | select(. | startswith("type/"))`)
	if err != nil {
		panic(fmt.Sprintf("error extracting type tags: %s", err))
	}
	fmt.Printf("Found %d type tags\n", len(typeTagValues))

	tagValues, err := extractFromJsonUsingJq(metadata, `[ .tools[].tags[] ] | unique | .[] | select(. | contains("/") | not)`)
	if err != nil {
		panic(fmt.Sprintf("error extracting tags: %s", err))
	}
	fmt.Printf("Found %d free tags\n", len(tagValues))

	allTools, err := extractFromJsonUsingJq(metadata, `.tools[].name`)
	if err != nil {
		panic(fmt.Sprintf("error extracting tool names: %s", err))
	}
	fmt.Printf("Found %d tools\n", len(allTools))

	err = huh.NewForm(
		huh.NewGroup(

			huh.NewInput().
				Title("What's the url?").
				Value(&project.url).
				Validate(func(str string) error {
					if str[0:19] == "https://github.com/" {
						project.platform = "github"

					} else {
						return errors.New("sorry, we do not support a platform other than GitHub right now")
					}

					return nil
				}),
		),
	).Run()
	if err != nil {
		log.Fatal(err)
	}

	regex := regexp.MustCompile(`^https://(www\.)?github.com/([^/]+)/([^/]+)(/.*)?$`)
	matches := regex.FindStringSubmatch(project.url)
	if len(matches) == 0 {
		panic("Invalid URL")
	}
	project.owner = matches[2]
	project.repository = matches[3]
	project.url = fmt.Sprintf("https://github.com/%s/%s", project.owner, project.repository)
	project.name = project.repository

	//client = github.NewClient(nil)
	githubProject, _, err := githubClient.Repositories.Get(context.Background(), project.owner, project.repository)
	if err != nil {
		panic(err)
	}
	project.language = strings.ToLower(*githubProject.Language)
	if githubProject.License != nil {
		if githubProject.License.SPDXID != nil {
			project.licenseName = *githubProject.License.SPDXID
		}
		if githubProject.License.URL != nil {
			project.licenseLink = *githubProject.License.URL
		}
	}

	releases, _, err := githubClient.Repositories.ListReleases(context.Background(), project.owner, project.repository, nil)
	if err != nil {
		panic(err)
	}
	var release *github.RepositoryRelease
	if len(releases) > 0 {
		release = releases[0]
		tag_re := regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)
		matches2 := tag_re.FindStringSubmatch(*release.TagName)
		if len(matches2) == 0 {
			panic(fmt.Sprintf("Unable to extract version from tag %s", *release.TagName))
		}
		project.version = matches2[1]
		project.tagNameTemplate = strings.ReplaceAll(*release.TagName, project.version, "{{ .Version }}")
	}

	err = huh.NewForm(
		huh.NewGroup(

			huh.NewText().
				Title("Url").
				CharLimit(50).
				Lines(1).
				Value(&project.url),

			huh.NewText().
				Title("Platform").
				CharLimit(50).
				Lines(1).
				Value(&project.platform),

			huh.NewText().
				Title("Owner").
				CharLimit(50).
				Lines(1).
				Value(&project.owner),

			huh.NewText().
				Title("Repository").
				CharLimit(50).
				Lines(1).
				Value(&project.repository),

			huh.NewInput().
				Title("Name").
				Value(&project.name),

			huh.NewInput().
				Title("License Name").
				Value(&project.licenseName),

			huh.NewInput().
				Title("License Link").
				Value(&project.licenseLink),

			huh.NewInput().
				Title("Latest version").
				Value(&project.version),

			huh.NewConfirm().
				Title("Is this correct?").
				Affirmative("Yes!").
				Negative("No.").
				Value(&project.confirmGeneral).
				Validate(func(b bool) error {
					if !b {
						return errors.New("please correct the information")
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		log.Fatal(err)
	}

	project.categoryTag = "?"
	project.typeTag = "?"

	err = huh.NewForm(
		huh.NewGroup(

			huh.NewSelect[string]().
				Title("Choose a category").
				Options(huh.NewOptions(categoryTagValues...)...).
				Value(&project.categoryTag),

			huh.NewSelect[string]().
				Title("Choose a type").
				Options(huh.NewOptions(typeTagValues...)...).
				Value(&project.typeTag),

			huh.NewMultiSelect[string]().
				Title("Choose more tags").
				Options(huh.NewOptions(tagValues...)...).
				Value(&project.tags).
				Height(10),
		),
	).Run()
	if err != nil {
		log.Fatal(err)
	}

	err = huh.NewForm(
		huh.NewGroup(

			huh.NewMultiSelect[string]().
				Title("Choose build dependencies").
				Options(huh.NewOptions(allTools...)...).
				Value(&project.buildDeps).
				Height(10),

			huh.NewMultiSelect[string]().
				Title("Choose runtime dependencies").
				Options(huh.NewOptions(allTools...)...).
				Value(&project.runtimeDeps).
				Height(10),
		),
	).Run()
	if err != nil {
		log.Fatal(err)
	}

	fullTags := []string{
		"org/?",
		fmt.Sprintf("category/%s", project.categoryTag),
		fmt.Sprintf("lang/%s", project.language),
		fmt.Sprintf("type/%s", project.typeTag),
	}
	fullTags = append(fullTags, project.tags...)
	tool := tool.Tool{
		Name:                project.name,
		License:             tool.License{Name: project.licenseName, Link: project.licenseLink},
		Version:             project.version,
		Check:               "",
		Platforms:           []string{},
		BuildDependencies:   project.buildDeps,
		RuntimeDependencies: project.runtimeDeps,
		Tags:                fullTags,
		Homepage:            *githubProject.HTMLURL,
		Repository:          *githubProject.HTMLURL,
		Description:         *githubProject.Description,
		Renovate: tool.Renovate{
			Datasource:     "github-releases",
			Package:        fmt.Sprintf("%s/%s", project.owner, project.repository),
			ExtractVersion: "^v?(?<version>.+)$",
		},
	}

	err = os.MkdirAll("tools/"+tool.Name, 0750)
	if err != nil {
		panic(fmt.Sprintf("unable to create directory: %v", err))
	}

	file, err := os.OpenFile("tools/"+tool.Name+"/manifest.yaml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Sprintf("error opening/creating file: %v", err))
	}
	//nolint:errcheck
	defer file.Close()

	enc := yaml.NewEncoder(file)
	err = enc.Encode(tool)
	if err != nil {
		panic(fmt.Sprintf("error encoding: %v", err))
	}

	var assets []*Asset
	for _, asset := range release.Assets {
		assets = append(assets, &Asset{
			Name:     *asset.Name,
			Version:  project.version,
			Template: strings.ReplaceAll(*asset.Name, project.version, "{{ .Version }}"),
		})
	}

	archive_re := regexp.MustCompile(`\.t(ar\.)?(gz|xz|bz2)$`)
	package_re := regexp.MustCompile(`\.(deb|rpm)$`)
	checksum_re := regexp.MustCompile(`(checksums?|sha\d+)`)
	sbom_re := regexp.MustCompile(`(s?bom|cyclonedx|spdx)`)
	signature_re := regexp.MustCompile(`\.(sig|pem)$`)
	platform_re := regexp.MustCompile(`(?i)(Linux|Darwin|MacOS|FreeBSD|Windows|\.exe$)`)
	amd64_re := regexp.MustCompile(`(?i)(x86_64|amd64|64bit)`)
	arm64_re := regexp.MustCompile(`(?i)(aarch64|arm64)`)
	arch_exlude_re := regexp.MustCompile(`(?i)(386|32bit|arm|armhf|s390x|ppc64le)`)
	var amd64_assets []*Asset
	var arm64_assets []*Asset
	for _, asset := range assets {
		if archive_re.MatchString(asset.Name) {
			asset.Type = "archive"
			asset.IsCandidate = true

		} else if package_re.MatchString(asset.Name) {
			asset.Type = "package"

		} else if checksum_re.MatchString(asset.Name) {
			asset.Type = "checksum"

		} else if sbom_re.MatchString(asset.Name) {
			asset.Type = "sbom"

		} else if signature_re.MatchString(asset.Name) {
			asset.Type = "signature"

		} else {
			asset.Type = "binary"
			asset.IsCandidate = true
		}

		arch_matches := platform_re.FindStringSubmatch(asset.Name)
		if len(arch_matches) > 0 {
			asset.Platform = arch_matches[0]
			if asset.Platform == "Linux" || asset.Platform == "linux" {
				asset.IsLinux = true
			}
		}
		if len(asset.Platform) == 0 {
			asset.IsLinux = true
		}

		amd64_matches := amd64_re.FindStringSubmatch(asset.Name)
		arm64_matches := arm64_re.FindStringSubmatch(asset.Name)
		if len(amd64_matches) > 0 {
			asset.IsAmd64 = true
			asset.Architecture = amd64_matches[0]

		} else if len(arm64_matches) > 0 {
			asset.IsArm64 = true
			asset.Architecture = arm64_matches[0]

		} else if arch_exlude_re.MatchString(asset.Name) {
			asset.Architecture = "unsupported"
		}
		if len(asset.Architecture) == 0 {
			asset.IsAmd64 = true
		}
		if len(asset.Architecture) > 0 {
			asset.Template = strings.ReplaceAll(asset.Template, asset.Architecture, "{{ .Arch }}")
		}

		if HasAsset(assets, asset.Name+".sig") && HasAsset(assets, asset.Name+".pem") {
			asset.HasSignature = true
		}

		if asset.IsCandidate && asset.IsLinux && asset.IsAmd64 {
			amd64_assets = append(amd64_assets, asset)
		}
		if asset.IsCandidate && asset.IsLinux && asset.IsArm64 {
			arm64_assets = append(arm64_assets, asset)
		}
	}

	if len(amd64_assets) == 0 {
		fmt.Printf("No assets. Falling back to candidates with architecture\n")

		amd64_assets = make([]*Asset, 0)
		arm64_assets = make([]*Asset, 0)

		for _, asset := range assets {
			if asset.IsCandidate && asset.IsAmd64 {
				amd64_assets = append(amd64_assets, asset)
			}
			if asset.IsCandidate && asset.IsArm64 {
				arm64_assets = append(arm64_assets, asset)
			}
		}
	}

	if len(amd64_assets) == 0 {
		fmt.Printf("No assets. Falling back to all candidates\n")

		amd64_assets = make([]*Asset, 0)
		arm64_assets = make([]*Asset, 0)

		for _, asset := range assets {
			if asset.IsCandidate {
				amd64_assets = append(amd64_assets, asset)
			}
		}
	}

	if len(amd64_assets) == 0 {
		// @TODO: Build based on language
		panic("unable to find assets for amd64")
	}

	project.assetAmd64 = amd64_assets[0]
	project.assetArm64 = arm64_assets[0]

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Asset for x86_64/amd64").
				OptionsFunc(func() []huh.Option[string] {
					opts := make([]huh.Option[string], len(amd64_assets))
					for i, asset := range amd64_assets {
						opts[i] = huh.NewOption(asset.Name, asset.Name)
					}
					return opts
				}, &project.assetAmd64.Name).
				Value(&project.assetAmd64.Name),

			huh.NewSelect[string]().
				Title("Asset for aarch64/arm64").
				OptionsFunc(func() []huh.Option[string] {
					opts := make([]huh.Option[string], len(arm64_assets))
					for i, asset := range arm64_assets {
						opts[i] = huh.NewOption(asset.Name, asset.Name)
					}
					return opts
				}, &project.assetArm64.Name).
				Value(&project.assetArm64.Name),

			huh.NewConfirm().
				Title("Is this correct?").
				Affirmative("Yes!").
				Negative("No.").
				Value(&project.confirmAssets).
				Validate(func(b bool) error {
					if !b {
						return errors.New("please correct the information")
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		log.Fatal(err)
	}

	switch project.assetAmd64.Type {
	case "archive":
		url := githubBuildAssetUrl(&project)
		binaryBytes, err := slurpHttpDownload(url)
		if err != nil {
			panic(fmt.Sprintf("error downloading archive: %s", err))
		}
		binaryBytesReader := bytes.NewReader(binaryBytes)
		files, err := ListTarGzFiles(binaryBytesReader)
		if err != nil {
			panic(fmt.Sprintf("error listing files: %s", err))
		}
		fmt.Printf("Files of %s\n", project.assetAmd64.Name)
		binaryRegex := regexp.MustCompile(fmt.Sprintf(`(^|/)%s$`, project.name))
		manpageRegex := regexp.MustCompile(`/.+\.\d(\.gz)?$`)
		completionRegex := regexp.MustCompile(`completions?/`)
		binary := ""
		manpages := make([]string, 0)
		completions := make([]string, 0)
		stripComponents := 0
		for _, file := range files {
			fmt.Printf("File: %s\n", file)

			if binaryRegex.MatchString(file) {
				fmt.Println("  BINARY")
				binary = file
				stripComponents = len(strings.Split(file, "/")) - 1

			} else if manpageRegex.MatchString(file) {
				fmt.Println("  MANPAGE")
				manpages = append(manpages, file)

			} else if completionRegex.MatchString(file) {
				fmt.Println("  COMPLETION")
				completions = append(completions, file)
			}
		}
		fmt.Printf("Binary: %s\n", binary)
		fmt.Printf("Strip components: %d\n", stripComponents)
		fmt.Printf("Manpages: %v\n", manpages)
		fmt.Printf("Completions: %v\n", completions)

		// @TODO: Extract binary and check for version

		fmt.Println("XXX GENERATE Dockerfile.template")

	case "binary":
		url := githubBuildAssetUrl(&project)
		binaryBytes, err := slurpHttpDownload(url)
		if err != nil {
			panic(fmt.Sprintf("error downloading binary: %s", err))
		}
		exe, err := memexec.New(binaryBytes)
		if err != nil {
			panic(fmt.Sprintf("error creating executable: %s", err))
		}
		//nolint:errcheck
		defer exe.Close()
		cmd := exe.Command("--help")
		_, err = cmd.Output()
		if err != nil {
			panic(fmt.Sprintf("failed to run command: %s", err))
		}

		// @TODO: List archive contents in comments
		fmt.Println("XXX GENERATE Dockerfile.template")

	default:
		panic(fmt.Sprintf("unsupported asset type %s", project.assetAmd64.Type))
	}

	// @TODO: If checksum file is present, generate code for checksum check
	// @TODO: Generate code for signature check
}

func githubBuildAssetUrl(project *Project) string {
	tmpl := template.Must(template.New("project_tag_name_template").Parse(project.tagNameTemplate))
	var b bytes.Buffer
	err := tmpl.Execute(&b, map[string]interface{}{
		"Version": project.version,
	})
	if err != nil {
		panic(fmt.Sprintf("error executing template: %s", err))
	}
	tagName := b.String()
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", project.owner, project.repository, tagName, project.assetAmd64.Name)
}

func slurpHttpDownload(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", fmt.Sprintf("uniget-auto-adder/%s", version))
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %s", url, err)
	}
	//nolint:errcheck
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download %s: %s", url, resp.Status)
	}
	binaryBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading binary: %s", err)
	}
	return binaryBytes, nil
}

func extractFromJsonUsingJq(data []byte, query string) ([]string, error) {
	result := make([]string, 0)

	var input any
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, fmt.Errorf("error parsing json: %s", err)
	}
	jqQuery, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("error parsing query: %s", err)
	}
	iter := jqQuery.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			log.Fatalln(err)
		}
		if strings.Contains(v.(string), "/") {
			result = append(result, strings.Split(v.(string), "/")[1])
		} else {
			result = append(result, v.(string))
		}
	}

	return result, nil
}
