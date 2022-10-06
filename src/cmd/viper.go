package cmd

import (
	"os"

	"github.com/defenseunicorns/zarf/src/internal/message"
	"github.com/spf13/viper"
)

const (
	// Root config keys
	V_LOG_LEVEL    = "log_level"
	V_ARCHITECTURE = "architecture"
	V_NO_LOG_FILE  = "no_log_file"
	V_NO_PROGRESS  = "no_progress"
	V_TMP_DIR      = "tmp_dir"

	// Init config keys
	V_INIT_COMPONENTS    = "init.components"
	V_INIT_STORAGE_CLASS = "init.storage_class"

	// Init Git config keys
	V_INIT_GIT_URL       = "init.git.url"
	V_INIT_GIT_PUSH_USER = "init.git.push_username"
	V_INIT_GIT_PUSH_PASS = "init.git.push_password"
	V_INIT_GIT_PULL_USER = "init.git.pull_username"
	V_INIT_GIT_PULL_PASS = "init.git.pull_password"

	// Init Registry config keys
	V_INIT_REGISTRY_URL       = "init.registry.url"
	V_INIT_REGISTRY_NODEPORT  = "init.registry.nodeport"
	V_INIT_REGISTRY_SECRET    = "init.registry.secret"
	V_INIT_REGISTRY_PUSH_USER = "init.registry.push_username"
	V_INIT_REGISTRY_PUSH_PASS = "init.registry.push_password"
	V_INIT_REGISTRY_PULL_USER = "init.registry.pull_username"
	V_INIT_REGISTRY_PULL_PASS = "init.registry.pull_password"

	// Package create config keys
	V_PKG_CREATE_SET        = "package.create.set"
	V_PKG_CREATE_ZARF_CACHE = "package.create.zarf_cache"
	V_PKG_CREATE_OUTPUT_DIR = "package.create.output_directory"
	V_PKG_CREATE_SKIP_SBOM  = "package.create.skip_sbom"
	V_PKG_CREATE_INSECURE   = "package.create.insecure"
	V_PKG_CREATE_CHUNK_SIZE = "package.create.chunk_size"

	// Package deploy config keys
	V_PKG_DEPLOY_SET        = "package.deploy.set"
	V_PKG_DEPLOY_COMPONENTS = "package.deploy.components"
	V_PKG_DEPLOY_INSECURE   = "package.deploy.insecure"
	V_PKG_DEPLOY_SHASUM     = "package.deploy.shasum"
	V_PKG_DEPLOY_SGET       = "package.deploy.sget"
)

func initViper() {
	// Already initializedby some other command
	if v != nil {
		return
	}

	v = viper.New()
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

	v.SetEnvPrefix("zarf")
	v.AutomaticEnv()

	// E.g. ZARF_LOG_LEVEL=debug
	v.SetEnvPrefix("zarf")
	v.AutomaticEnv()

	// Optional, so ignore errors
	err := v.ReadInConfig()

	if err != nil {
		// Config file not found; ignore
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			message.Error(err, "Failed to read config file")
		}
	} else {
		message.Notef("Using config file %s", v.ConfigFileUsed())
	}
}
