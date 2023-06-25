// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package tools contains the CLI commands for Zarf.
package tools

import (
	"os"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/internal/cluster"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	craneCmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/logs"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
)

func init() {
	verbose := false
	insecure := false
	ndlayers := false
	platform := "all"

	// No package information is available so do not pass in a list of architectures
	craneOptions := []crane.Option{}

	registryCmd := &cobra.Command{
		Use:     "registry",
		Aliases: []string{"r", "crane"},
		Short:   lang.CmdToolsRegistryShort,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// The crane options loading here comes from the rootCmd of crane
			craneOptions = append(craneOptions, crane.WithContext(cmd.Context()))
			// TODO(jonjohnsonjr): crane.Verbose option?
			if verbose {
				logs.Debug.SetOutput(os.Stderr)
			}
			if insecure {
				craneOptions = append(craneOptions, crane.Insecure)
			}
			if ndlayers {
				craneOptions = append(craneOptions, crane.WithNondistributable())
			}

			var err error
			var v1Platform *v1.Platform
			if platform != "all" {
				v1Platform, err = v1.ParsePlatform(platform)
				if err != nil {
					message.Fatalf(err, lang.CmdToolsRegistryInvalidPlatformErr, err.Error())
				}
			}

			craneOptions = append(craneOptions, crane.WithPlatform(v1Platform))
		},
	}

	craneLogin := craneCmd.NewCmdAuthLogin()
	craneLogin.Example = ""

	registryCmd.AddCommand(craneLogin)
	registryCmd.AddCommand(craneCmd.NewCmdPull(&craneOptions))
	registryCmd.AddCommand(craneCmd.NewCmdPush(&craneOptions))

	craneCopy := craneCmd.NewCmdCopy(&craneOptions)

	registryCmd.AddCommand(craneCopy)
	registryCmd.AddCommand(zarfCraneCatalog(&craneOptions))

	registryCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, lang.CmdToolsRegistryFlagVerbose)
	registryCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, lang.CmdToolsRegistryFlagInsecure)
	registryCmd.PersistentFlags().BoolVar(&ndlayers, "allow-nondistributable-artifacts", false, lang.CmdToolsRegistryFlagNonDist)
	registryCmd.PersistentFlags().StringVar(&platform, "platform", "all", lang.CmdToolsRegistryFlagPlatform)

	toolsCmd.AddCommand(registryCmd)
}

// Wrap the original crane catalog with a zarf specific version
func zarfCraneCatalog(cranePlatformOptions *[]crane.Option) *cobra.Command {
	craneCatalog := craneCmd.NewCmdCatalog(cranePlatformOptions)

	craneCatalog.Example = lang.CmdToolsRegistryCatalogExample
	craneCatalog.Args = nil

	originalCatalogFn := craneCatalog.RunE

	craneCatalog.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return originalCatalogFn(cmd, args)
		}

		// Load Zarf state
		zarfState, err := cluster.NewClusterOrDie().LoadZarfState()
		if err != nil {
			return err
		}

		// Open a tunnel to the Zarf registry
		tunnelReg, err := cluster.NewZarfTunnel()
		if err != nil {
			return err
		}
		err = tunnelReg.Connect(cluster.ZarfRegistry, false)
		if err != nil {
			return err
		}

		// Add the correct authentication to the crane command options
		authOption := config.GetCraneAuthOption(zarfState.RegistryInfo.PullUsername, zarfState.RegistryInfo.PullPassword)
		*cranePlatformOptions = append(*cranePlatformOptions, authOption)
		registryEndpoint := tunnelReg.Endpoint()

		return originalCatalogFn(cmd, []string{registryEndpoint})
	}

	return craneCatalog
}
