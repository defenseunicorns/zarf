package packager

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/defenseunicorns/zarf/cli/config"
	"github.com/defenseunicorns/zarf/cli/internal/git"
	"github.com/defenseunicorns/zarf/cli/internal/helm"
	"github.com/defenseunicorns/zarf/cli/internal/images"
	"github.com/defenseunicorns/zarf/cli/internal/log"
	"github.com/defenseunicorns/zarf/cli/internal/utils"
	"github.com/mholt/archiver/v3"
)

func Create(confirm bool) {

	config.Load("zarf.yaml")

	tempPath := createPaths()
	packageName := config.GetPackageName()
	dataInjections := config.GetDataInjections()
	components := config.GetComponents()
	configFile := tempPath.base + "/zarf.yaml"

	// Save the transformed config
	config.WriteConfig(configFile)

	confirm = confirmAction(configFile, confirm, "Create")

	if !confirm {
		os.Exit(0)
	}

	for _, component := range components {
		log.Logger.WithField("component", component.Name).Info("Loading component assets")
		componentPath := createComponentPaths(tempPath.components, component)
		addLocalAssets(componentPath, component)
	}

	if config.IsZarfInitConfig() {
		// Override the package name for init packages
		packageName = config.PackageInitName
	} else {
		// Init packages do not use data or utilityCluster keys
		if len(dataInjections) > 0 {
			log.Logger.Info("Loading data injections")
			for _, data := range dataInjections {
				destinationFile := tempPath.dataInjections + "/" + filepath.Base(data.Target.Path)
				utils.CreatePathAndCopy(data.Source, destinationFile)
			}
		}
	}
	_ = os.RemoveAll(packageName)
	err := archiver.Archive([]string{tempPath.base + "/"}, packageName)
	if err != nil {
		log.Logger.Fatal("Unable to create the package archive")
	}

	log.Logger.WithField("name", packageName).Info("Package creation complete")

	cleanup(tempPath)
}

func addLocalAssets(tempPath componentPaths, assets config.ZarfComponent) {
	if len(assets.Charts) > 0 {
		log.Logger.Info("Loading static helm charts")
		utils.CreateDirectory(tempPath.charts, 0700)
		for _, chart := range assets.Charts {
			isGitURL, _ := regexp.MatchString("\\.git$", chart.Url)
			if isGitURL {
				helm.DownloadChartFromGit(chart, tempPath.charts)
			} else {
				helm.DownloadPublishedChart(chart, tempPath.charts)
			}
		}
	}

	if len(assets.Files) > 0 {
		log.Logger.Info("Downloading files for local install")
		_ = utils.CreateDirectory(tempPath.files, 0700)
		for index, file := range assets.Files {
			destinationFile := tempPath.files + "/" + strconv.Itoa(index)
			if utils.IsUrl(file.Source) {
				utils.DownloadToFile(file.Source, destinationFile)
			} else {
				utils.CreatePathAndCopy(file.Source, destinationFile)
			}

			// Abort packaging on invalid shasum (if one is specified)
			if file.Shasum != "" {
				utils.ValidateSha256Sum(file.Shasum, destinationFile)
			}

			if file.Executable {
				_ = os.Chmod(destinationFile, 0700)
			} else {
				_ = os.Chmod(destinationFile, 0600)
			}
		}
	}

	if len(assets.Images) > 0 {
		log.Logger.Info("Loading container images")
		images.PullAll(assets.Images, tempPath.images)
	}

	if assets.Manifests != "" {
		log.Logger.WithField("path", assets.Manifests).Info("Loading manifests for local install")
		utils.CreatePathAndCopy(assets.Manifests, tempPath.manifests)
	}

	if len(assets.Repos) > 0 {
		log.Logger.Info("loading git repos for gitops service transfer")
		// Load all specified git repos
		for _, url := range assets.Repos {
			matches := strings.Split(url, "@")
			if len(matches) < 2 {
				log.Logger.WithField("remote", url).Fatal("Unable to parse git url. Ensure you use the format url.git@tag")
			}
			git.Pull(matches[0], tempPath.repos)
		}
	}
}
