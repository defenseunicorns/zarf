// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package cmd contains the CLI commands for zarf
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/packager"
	"github.com/defenseunicorns/zarf/src/pkg/utils"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Prepares a k8s cluster for the deployment of Zarf packages",
	Long: "Injects a docker registry as well as other optional useful things (such as a git server " +
		"and a logging stack) into a k8s cluster under the 'zarf' namespace " +
		"to support future application deployments. \n" +

		"If you do not have a k8s cluster already configured, this command will give you " +
		"the ability to install a cluster locally.\n\n" +

		"This command looks for a zarf-init package in the local directory that the command was executed " +
		"from. If no package is found in the local directory and the Zarf CLI exists somewhere outside of " +
		"the current directory, Zarf will failover and attempt to find a zarf-init package in the directory " +
		"that the Zarf binary is located in.\n\n\n\n" +

		"Example Usage:\n" +
		"# Initializing without any optional components:\nzarf init\n\n" +
		"# Initializing w/ Zarfs internal git server:\nzarf init --components=git-server\n\n" +
		"# Initializing w/ Zarfs internal git server and PLG stack:\nzarf init --components=git-server,logging\n\n" +
		"# Initializing w/ an internal registry but with a different nodeport:\nzarf init --nodeport=30333\n\n" +
		"# Initializing w/ an external registry:\nzarf init --registry-push-password={PASSWORD} --registry-push-username={USERNAME} --registry-url={URL}\n\n" +
		"# Initializing w/ an external git server:\nzarf init --git-push-password={PASSWORD} --git-push-username={USERNAME} --git-url={URL}\n\n",

	Run: func(cmd *cobra.Command, args []string) {
		zarfLogo := message.GetLogo()
		_, _ = fmt.Fprintln(os.Stderr, zarfLogo)

		if err := validateInitFlags(); err != nil {
			message.Fatal(err, "Invalid command flags were provided.")
		}

		// Continue running package deploy for all components like any other package
		initPackageName := packager.GetInitPackageName("")
		pkgConfig.DeployOpts.PackagePath = initPackageName

		// Try to use an init-package in the executable directory if none exist in current working directory
		var err error
		if pkgConfig.DeployOpts.PackagePath, err = findInitPackage(initPackageName); err != nil {
			message.Fatal(err, err.Error())
		}

		// Run everything
		packager.NewOrDie(&pkgConfig).Deploy()
	},
}

func findInitPackage(initPackageName string) (string, error) {
	// First, look for the init package in the current working directory
	if !utils.InvalidPath(initPackageName) {
		return initPackageName, nil
	}

	// Next, look for the init package in the executable directory
	executablePath, err := utils.GetFinalExecutablePath()
	if err != nil {
		return "", err
	}
	executableDir := path.Dir(executablePath)
	if !utils.InvalidPath(filepath.Join(executableDir, initPackageName)) {
		return filepath.Join(executableDir, initPackageName), nil
	}

	// Next, look in the cache directory
	if !utils.InvalidPath(filepath.Join(config.GetAbsCachePath(), initPackageName)) {
		return filepath.Join(config.GetAbsCachePath(), initPackageName), nil
	}

	// Finally, if the init-package doesn't exist in the cache directory, suggest downloading it
	if err := downloadInitPackage(initPackageName); err != nil {
		if errors.Is(err, lang.ErrInitNotFound) {
			message.Fatal(err, err.Error())
		} else {
			message.Fatalf(err, "Failed to download the init package: %v", err)
		}
	}
	return "", nil
}

func downloadInitPackage(initPackageName string) error {
	if config.CommonOptions.Confirm {
		return lang.ErrInitNotFound
	}

	var confirmDownload bool
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", config.GithubProject, config.CLIVersion, initPackageName)

	// Give the user the choice to download the init-package and note that this does require an internet connection
	message.Question(fmt.Sprintf("It seems the init package could not be found locally, but can be downloaded from %s", url))

	message.Note("Note: This will require an internet connection.")

	// Prompt the user if --confirm not specified
	if !confirmDownload {
		prompt := &survey.Confirm{
			Message: "Do you want to download this init package?",
		}
		if err := survey.AskOne(prompt, &confirmDownload); err != nil {
			message.Fatalf(nil, "Confirm selection canceled: %s", err.Error())
		}
	}

	// If the user wants to download the init-package, download it
	if confirmDownload {
		utils.DownloadToFile(url, pkgConfig.DeployOpts.PackagePath, "")
	} else {
		// Otherwise, exit and tell the user to manually download the init-package
		return fmt.Errorf("you must download the init package manually and place it in the current working directory")
	}

	return nil
}

func validateInitFlags() error {
	// If 'git-url' is provided, make sure they provided values for the username and password of the push user
	if pkgConfig.InitOpts.GitServer.Address != "" {
		if pkgConfig.InitOpts.GitServer.PushUsername == "" || pkgConfig.InitOpts.GitServer.PushPassword == "" {
			return fmt.Errorf("the 'git-push-username' and 'git-push-password' flags must be provided if the 'git-url' flag is provided")
		}
	}

	//If 'registry-url' is provided, make sure they provided values for the username and password of the push user
	if pkgConfig.InitOpts.RegistryInfo.Address != "" {
		if pkgConfig.InitOpts.RegistryInfo.PushUsername == "" || pkgConfig.InitOpts.RegistryInfo.PushPassword == "" {
			return fmt.Errorf("the 'registry-push-username' and 'registry-push-password' flags must be provided if the 'registry-url' flag is provided ")
		}
	}
	return nil
}

func init() {
	initViper()

	rootCmd.AddCommand(initCmd)

	v.SetDefault(V_INIT_COMPONENTS, "")
	v.SetDefault(V_INIT_STORAGE_CLASS, "")

	v.SetDefault(V_INIT_GIT_URL, "")
	v.SetDefault(V_INIT_GIT_PUSH_USER, config.ZarfGitPushUser)
	v.SetDefault(V_INIT_GIT_PUSH_PASS, "")
	v.SetDefault(V_INIT_GIT_PULL_USER, "")
	v.SetDefault(V_INIT_GIT_PULL_PASS, "")

	v.SetDefault(V_INIT_REGISTRY_URL, "")
	v.SetDefault(V_INIT_REGISTRY_NODEPORT, 0)
	v.SetDefault(V_INIT_REGISTRY_SECRET, "")
	v.SetDefault(V_INIT_REGISTRY_PUSH_USER, config.ZarfRegistryPushUser)
	v.SetDefault(V_INIT_REGISTRY_PUSH_PASS, "")
	v.SetDefault(V_INIT_REGISTRY_PULL_USER, "")
	v.SetDefault(V_INIT_REGISTRY_PULL_PASS, "")

	// Continue to require --confirm flag for init command to avoid accidental deployments
	initCmd.Flags().BoolVar(&config.CommonOptions.Confirm, "confirm", false, "Confirm the install without prompting")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.Components, "components", v.GetString(V_INIT_COMPONENTS), "Comma-separated list of components to install.")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.StorageClass, "storage-class", v.GetString(V_INIT_STORAGE_CLASS), "Describe the StorageClass to be used")

	// Flags for using an external Git server
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.GitServer.Address, "git-url", v.GetString(V_INIT_GIT_URL), "External git server url to use for this Zarf cluster")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.GitServer.PushUsername, "git-push-username", v.GetString(V_INIT_GIT_PUSH_USER), "Username to access to the git server Zarf is configured to use. User must be able to create repositories via 'git push'")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.GitServer.PushPassword, "git-push-password", v.GetString(V_INIT_GIT_PUSH_PASS), "Password for the push-user to access the git server")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.GitServer.PullUsername, "git-pull-username", v.GetString(V_INIT_GIT_PULL_USER), "Username for pull-only access to the git server")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.GitServer.PullPassword, "git-pull-password", v.GetString(V_INIT_GIT_PULL_PASS), "Password for the pull-only user to access the git server")

	// Flags for using an external registry
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.RegistryInfo.Address, "registry-url", v.GetString(V_INIT_REGISTRY_URL), "External registry url address to use for this Zarf cluster")
	initCmd.Flags().IntVar(&pkgConfig.InitOpts.RegistryInfo.NodePort, "nodeport", v.GetInt(V_INIT_REGISTRY_NODEPORT), "Nodeport to access a registry internal to the k8s cluster. Between [30000-32767]")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.RegistryInfo.PushUsername, "registry-push-username", v.GetString(V_INIT_REGISTRY_PUSH_USER), "Username to access to the registry Zarf is configured to use")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.RegistryInfo.PushPassword, "registry-push-password", v.GetString(V_INIT_REGISTRY_PUSH_PASS), "Password for the push-user to connect to the registry")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.RegistryInfo.PullUsername, "registry-pull-username", v.GetString(V_INIT_REGISTRY_PULL_USER), "Username for pull-only access to the registry")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.RegistryInfo.PullPassword, "registry-pull-password", v.GetString(V_INIT_REGISTRY_PULL_PASS), "Password for the pull-only user to access the registry")
	initCmd.Flags().StringVar(&pkgConfig.InitOpts.RegistryInfo.Secret, "registry-secret", v.GetString(V_INIT_REGISTRY_SECRET), "Registry secret value")

	initCmd.Flags().SortFlags = true
}
