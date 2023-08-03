// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package common handles command configuration across all commands
package common

import (
	"os"
	"strings"

	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/spf13/viper"
)

// Constants for use when loading configurations from viper config files
const (

	// Root config keys

	VLogLevel     = "log_level"
	VArchitecture = "architecture"
	VNoLogFile    = "no_log_file"
	VNoProgress   = "no_progress"
	VZarfCache    = "zarf_cache"
	VTmpDir       = "tmp_dir"
	VInsecure     = "insecure"

	// Init config keys

	VInitComponents   = "init.components"
	VInitStorageClass = "init.storage_class"

	// Init Git config keys

	VInitGitURL      = "init.git.url"
	VInitGitPushUser = "init.git.push_username"
	VInitGitPushPass = "init.git.push_password"
	VInitGitPullUser = "init.git.pull_username"
	VInitGitPullPass = "init.git.pull_password"

	// Init Registry config keys

	VInitRegistryURL      = "init.registry.url"
	VInitRegistryNodeport = "init.registry.nodeport"
	VInitRegistrySecret   = "init.registry.secret"
	VInitRegistryPushUser = "init.registry.push_username"
	VInitRegistryPushPass = "init.registry.push_password"
	VInitRegistryPullUser = "init.registry.pull_username"
	VInitRegistryPullPass = "init.registry.pull_password"

	// Init Package config keys

	VInitArtifactURL       = "init.artifact.url"
	VInitArtifactPushUser  = "init.artifact.push_username"
	VInitArtifactPushToken = "init.artifact.push_token"

	// Package config keys

	VPkgOCIConcurrency = "package.oci_concurrency"

	// Package create config keys

	VPkgCreateSet                = "package.create.set"
	VPkgCreateOutput             = "package.create.output"
	VPkgCreateSbom               = "package.create.sbom"
	VPkgCreateSbomOutput         = "package.create.sbom_output"
	VPkgCreateSkipSbom           = "package.create.skip_sbom"
	VPkgCreateMaxPackageSize     = "package.create.max_package_size"
	VPkgCreateSigningKey         = "package.create.signing_key"
	VPkgCreateSigningKeyPassword = "package.create.signing_key_password"
	VPkgCreateDifferential       = "package.create.differential"
	VPkgCreateRegistryOverride   = "package.create.registry_override"

	// Package deploy config keys

	VPkgDeploySet        = "package.deploy.set"
	VPkgDeployComponents = "package.deploy.components"
	VPkgDeployShasum     = "package.deploy.shasum"
	VPkgDeploySget       = "package.deploy.sget"
	VPkgDeployPublicKey  = "package.deploy.public_key"

	// Package publish config keys

	VPkgPublishSigningKey         = "package.publish.signing_key"
	VPkgPublishSigningKeyPassword = "package.publish.signing_key_password"

	// Package pull config keys

	VPkgPullOutputDir = "package.pull.output_directory"
	VPkgPullPublicKey = "package.pull.public_key"
)

// Viper instance used by commands
var v *viper.Viper

func InitViper() *viper.Viper {
	// Already initialized by some other command
	if v != nil {
		return v
	}

	v = viper.New()

	// Skip for vendor-only commands
	if CheckVendorOnlyFromArgs() {
		return v
	}

	// Skip for the version command
	if isVersionCmd() {
		return v
	}

	// Specify an alternate config file
	cfgFile := os.Getenv("ZARF_CONFIG")

	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Search config paths in the current directory and $HOME/.zarf.
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.zarf")
		v.SetConfigName("zarf-config")
	}

	// E.g. ZARF_LOG_LEVEL=debug
	v.SetEnvPrefix("zarf")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Optional, so ignore errors
	err := v.ReadInConfig()

	if err != nil {
		// Config file not found; ignore
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			message.WarnErrorf(err, lang.CmdViperErrLoadingConfigFile, err.Error())
		}
	} else {
		message.Notef(lang.CmdViperInfoUsingConfigFile, v.ConfigFileUsed())
	}

	return v
}

func GetViper() *viper.Viper {
	return v
}

func isVersionCmd() bool {
	args := os.Args
	return len(args) > 1 && (args[1] == "version" || args[1] == "v")
}
