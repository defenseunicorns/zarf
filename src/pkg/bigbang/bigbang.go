package bigbang

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/defenseunicorns/zarf/src/internal/packager/helm"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"

	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type bbImageYaml struct {
	ChartList map[string]bbImageYamlChartDef `json:"package-image-list"`
}

type bbImageYamlChartDef struct {
	Version string
	Images  []string
}

// Default location for pulling BigBang
const DEFAULT_BIGBANG_REPO = "https://repo1.dso.mil/big-bang/bigbang.git"

func CreateFluxComponent(bbComponent types.ZarfComponent, bbCount int) (fluxComponent types.ZarfComponent, err error) {
	fluxComponent.Name = "flux"

	fluxComponent.Required = bbComponent.Required

	fluxManifest := GetFluxManifest(bbComponent.BigBang.Version)
	fluxComponent.Manifests = []types.ZarfManifest{fluxManifest}
	repo := DEFAULT_BIGBANG_REPO
	if bbComponent.BigBang.Repo != "" {
		repo = bbComponent.BigBang.Repo
	}
	images, err := helm.FindFluxImages(repo, bbComponent.BigBang.Version)
	fluxComponent.Images = images

	return fluxComponent, nil
}

// MutateBigbangComponent Mutates a component that should deploy BigBang to a set of manifests
// that contain the flux deployment of BigBang
func MutateBigbangComponent(componentPath types.ComponentPaths, component types.ZarfComponent) (types.ZarfComponent, error) {
	_ = utils.CreateDirectory(componentPath.Charts, 0700)
	_ = utils.CreateDirectory(componentPath.Manifests, 0700)

	repos := make([]string, 0)

	// use the default repo unless overridden
	if component.BigBang.Repo == "" {
		repos = append(repos, DEFAULT_BIGBANG_REPO)
		component.BigBang.Repo = repos[0]
	} else {
		repos = append(repos, fmt.Sprintf("%s@%s", component.BigBang.Repo, component.BigBang.Version))
	}

	// download bigbang so we can peek inside
	chart := types.ZarfChart{
		Name:        "bigbang",
		Namespace:   "bigbang",
		URL:         repos[0],
		Version:     component.BigBang.Version,
		ValuesFiles: component.BigBang.ValuesFrom,
		GitPath:     "./chart",
	}

	zarfHelmInstance := helm.Helm{
		Chart: chart,
		Cfg: &types.PackagerConfig{
			State: types.ZarfState{},
		},
		BasePath: componentPath.Charts,
	}

	bb := zarfHelmInstance.DownloadChartFromGit("bigbang")

	zarfHelmInstance.ChartLoadOverride = bb

	// Template the chart so we can see what GitRepositories are being referenced in the
	// manifests created with the provided Helm
	template, err := zarfHelmInstance.TemplateChart()
	if err != nil {
		return component, fmt.Errorf("unable to template BigBang Chart: %w", err)
	}

	subPackageURLS := findURLs(template)
	repos[0] = fmt.Sprintf("%s@%s", repos[0], component.BigBang.Version)
	repos = append(repos, subPackageURLS...)

	// Save the list of repos to be pulled down by Zarf
	component.Repos = repos

	// just select the images needed to suppor the repos this configuration of BigBang will need
	images, err := GetImages(repos)
	if err != nil {
		return component, fmt.Errorf("unable to get bb images: %w", err)
	}

	// dedupe the list o fimages
	uniqueList := utils.Unique(images)

	// add the images to the component for Zarf to download
	component.Images = append(component.Images, uniqueList...)

	//Create the flux wrapper around BigBang for deployment
	manifest, err := GetBigBangManifests(componentPath.Manifests, component)
	if err != nil {
		return component, err
	}

	component.Manifests = []types.ZarfManifest{manifest}
	return component, nil
}

// findURLs takes a list of yaml objects (as a string) and
// parses it for GitRepository objects that it then parses
// to return the list of git repos and tags needed.
func findURLs(t string) []string {

	// Break the template into separate resources
	urls := make([]string, 0)
	yamls, _ := utils.SplitYAML([]byte(t))

	for _, y := range yamls {
		// see if its a GitRepository
		if y.GetKind() == "GitRepository" {
			url := y.Object["spec"].(map[string]interface{})["url"].(string)
			var ref string
			ref, ok := y.Object["spec"].(map[string]interface{})["ref"].(map[string]interface{})["commit"].(string)
			if !ok {
				ref, ok = y.Object["spec"].(map[string]interface{})["ref"].(map[string]interface{})["semver"].(string)
			}
			if !ok {
				ref, ok = y.Object["spec"].(map[string]interface{})["ref"].(map[string]interface{})["tag"].(string)
			}
			if !ok {
				ref, ok = y.Object["spec"].(map[string]interface{})["ref"].(map[string]interface{})["branch"].(string)
			}

			urls = append(urls, fmt.Sprintf("%v@%v", url, ref))
		}
	}

	return urls
}

// GetImages identifies the list of images needed for the list
// of repos provided.
func GetImages(repos []string) ([]string, error) {

	images := make([]string, 0)
	for _, r := range repos {
		is, err := helm.FindImagesForChartRepo(r, "chart")
		if err != nil {
			message.Warn(fmt.Sprintf("Could not pull images for chart %s: %s", r, err))
			continue
		}
		images = append(images, is...)
	}

	return images, nil
}

func printObject(filename string, o runtime.Object) error {
	return nil
}

// GetFluxManifest creates the manifests for deploying the particular version of BigBang
func GetFluxManifest(version string) types.ZarfManifest {
	return types.ZarfManifest{
		Name:      "flux-system",
		Namespace: "flux-system",
		Kustomizations: []string{
			fmt.Sprintf("https://repo1.dso.mil/platform-one/big-bang/bigbang.git//base/flux?ref=%v", version),
		},
	}
}

// GetBigBangManifests creates the manifests for deploying BigBang
func GetBigBangManifests(manifestDir string, component types.ZarfComponent) (types.ZarfManifest, error) {
	// here or in

	// componentPath, err := p.createOrGetComponentPaths(baseComponent)
	manifest := types.ZarfManifest{
		Name:      "bigbang",
		Namespace: "bigbang",
		Files:     []string{},
		// Kustomizations: []string{
		// 	destination,
		// },
	}

	gitIgnore := `# exclude file extensions
/**/*.md
/**/*.txt
/**/*.sh
`

	source := sourcev1beta2.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitRepository",
			APIVersion: "source.toolkit.fluxcd.io/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bigbang",
			Namespace: "bigbang",
		},
		Spec: sourcev1beta2.GitRepositorySpec{
			URL:    component.BigBang.Repo,
			Ignore: &gitIgnore,
			Interval: metav1.Duration{
				Duration: time.Minute,
			},
			Reference: &sourcev1beta2.GitRepositoryRef{
				Tag: component.BigBang.Version,
			},
		},
	}
	// printObject(fmt.Sprintf("%s/gitrepository.yaml", kDir), &source)
	// utils.WriteYaml(fmt.Sprintf("%s/gitrepository.yaml", kDir), source, 0644)
	data, _ := yaml.Marshal(source)
	utils.WriteFile(fmt.Sprintf("%s/gitrepository.yaml", manifestDir), data)
	// manifest.Files = append(manifest.Files, fmt.Sprintf("%s/gitrepository.yaml", manifestDir))
	manifest.Files = append(manifest.Files, fmt.Sprintf("%s/gitrepository.yaml", manifestDir))

	//imagepull secret
	creds := `
registryCredentials:
  registry: "###ZARF_REGISTRY###"
  username: "zarf-pull"
  password: "###ZARF_REGISTRY_AUTH_PULL###"
git:
  existingSecret: "private-git-server"	# -- Chart created secrets with user defined values
  credentials:
  # -- HTTP git credentials, both username and password must be provided
    username: "###ZARF_GIT_PUSH###"
    password: "###ZARF_GIT_AUTH_PUSH###"
kyvernopolicies:
  values:
    exclude:
      any:
      - resources:
          namespaces: 
          - zarf
`
	secretData := make(map[string]string)
	secretData["values.yaml"] = creds
	zarfSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "bigbang",
			Name:      "zarf-credentials",
		},
		StringData: secretData,
	}
	// credentialsSecretName := foobar
	data, _ = yaml.Marshal(zarfSecret)
	ioutil.WriteFile(fmt.Sprintf("%s/zarf-credentials.yaml", manifestDir), []byte(data), 0644)
	// manifest.Files = append(manifest.Files, fmt.Sprintf("%s/zarf-credentials.yaml", manifestDir))
	manifest.Files = append(manifest.Files, fmt.Sprintf("%s/zarf-credentials.yaml", manifestDir))

	hrValues := make([]helmv2beta1.ValuesReference, len(component.BigBang.ValuesFrom)+1)
	hrValues[0] = helmv2beta1.ValuesReference{
		Kind: "Secret",
		Name: "zarf-credentials",
	}

	// zarf magic for the value files to put them in the right directory
	// This might only assume that values files are in the same folder as the zarf.yaml.  Maybe that's okay.
	for idx, path := range component.BigBang.ValuesFrom {
		//load the values file.
		file, err := os.ReadFile(path)
		if err != nil {
			return manifest, err
		}
		//make a secret
		secretData["values.yaml"] = string(file)
		zarfSecret := corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "bigbang",
				Name:      fmt.Sprintf("bigbang-values-%s", strconv.Itoa(idx)),
			},
			StringData: secretData,
		}
		// write the secet down
		data, _ = yaml.Marshal(zarfSecret)
		ioutil.WriteFile(fmt.Sprintf("%s/bigbang-values-%s.yaml", manifestDir, strconv.Itoa(idx)), []byte(data), 0644)
		//add it to the manifests
		// manifest.Files = append(manifest.Files, fmt.Sprintf("%s/bigbang-values-%s.yaml", manifestDir))
		manifest.Files = append(manifest.Files, fmt.Sprintf("%s/bigbang-values-%s.yaml", manifestDir, strconv.Itoa(idx)))

		// Add it to the list of valuesFrom for the HelmRelease
		hrValues[idx+1] = helmv2beta1.ValuesReference{
			Kind: "Secret",
			Name: zarfSecret.Name,
		}
	}

	namespace := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "bigbang",
		},
	}
	utils.WriteYaml(fmt.Sprintf("%s/namespace.yaml", manifestDir), namespace, 0644)
	manifest.Files = append(manifest.Files, fmt.Sprintf("%s/namespace.yaml", manifestDir))

	t := true
	release := helmv2beta1.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HelmRelease",
			APIVersion: "helm.toolkit.fluxcd.io/v2beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bigbang",
			Namespace: "bigbang",
		},
		Spec: helmv2beta1.HelmReleaseSpec{
			Chart: helmv2beta1.HelmChartTemplate{
				Spec: helmv2beta1.HelmChartTemplateSpec{
					Chart: "./chart",
					SourceRef: helmv2beta1.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "bigbang",
					},
				},
			},
			Install: &helmv2beta1.Install{
				Remediation: &helmv2beta1.InstallRemediation{
					Retries: -1,
				},
			},
			Upgrade: &helmv2beta1.Upgrade{
				Remediation: &helmv2beta1.UpgradeRemediation{
					Retries:              5,
					RemediateLastFailure: &t,
				},
				CleanupOnFail: true,
			},
			Rollback: &helmv2beta1.Rollback{
				Timeout: &metav1.Duration{
					10 * time.Minute,
				},
				CleanupOnFail: true,
			},

			ValuesFrom: hrValues,
		},
	}

	// printObject(fmt.Sprintf("%s/helmrepository.yaml", kDir), &release)
	// utils.WriteYaml(fmt.Sprintf("%s/helmrepository.yaml", kDir), release, 0644)
	data, _ = yaml.Marshal(release)
	utils.WriteFile(fmt.Sprintf("%s/helmrepository.yaml", manifestDir), data)
	manifest.Files = append(manifest.Files, fmt.Sprintf("%s/helmrepository.yaml", manifestDir))

	return manifest, nil
}
