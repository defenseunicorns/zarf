//go:build !alt_language

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package lang contains the language strings for english used by Zarf
// Alternative languages can be created by duplicating this file and changing the build tag to "//go:build alt_language && <language>".
package lang

import "errors"

// All language strings should be in the form of a constant
// The constants should be grouped by the top level package they are used in (or common)
// The format should be <PathName><Err/Info><ShortDescription>
// Debug messages will not be a part of the language strings since they are not intended to be user facing
// Include sprintf formatting directives in the string if needed.
const (
	ErrLoadingConfig       = "failed to load config: %w"
	ErrLoadState           = "Failed to load the Zarf State from the Kubernetes cluster."
	ErrLoadPackageSecret   = "Failed to load %s's secret from the Kubernetes cluster"
	ErrMarshal             = "failed to marshal file: %w"
	ErrNoClusterConnection = "Failed to connect to the Kubernetes cluster."
	ErrTunnelFailed        = "Failed to create a tunnel to the Kubernetes cluster."
	ErrUnmarshal           = "failed to unmarshal file: %w"
	ErrWritingFile         = "failed to write the file %s: %s"
	ErrDownloading         = "failed to download %s: %s"
	ErrCreatingDir         = "failed to create directory %s: %s"
)

// Zarf CLI commands.
const (
	// root zarf command
	RootCmdShort = "DevSecOps for Airgap"
	RootCmdLong  = "Zarf eliminates the complexity of air gap software delivery for Kubernetes clusters and cloud native workloads\n" +
		"using a declarative packaging strategy to support DevSecOps in offline and semi-connected environments."

	RootCmdFlagLogLevel    = "Log level when running Zarf. Valid options are: warn, info, debug, trace"
	RootCmdFlagArch        = "Architecture for OCI images and Zarf packages"
	RootCmdFlagSkipLogFile = "Disable log file creation"
	RootCmdFlagNoProgress  = "Disable fancy UI progress bars, spinners, logos, etc"
	RootCmdFlagCachePath   = "Specify the location of the Zarf cache directory"
	RootCmdFlagTempDir     = "Specify the temporary directory to use for intermediate files"
	RootCmdFlagInsecure    = "Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture."

	RootCmdDeprecatedDeploy = "Please use \"zarf package deploy %s\" to deploy this package."
	RootCmdDeprecatedCreate = "Please use \"zarf package create\" to create this package."

	RootCmdErrInvalidLogLevel = "Invalid log level. Valid options are: warn, info, debug, trace."

	// zarf connect
	CmdConnectShort = "Access services or pods deployed in the cluster."
	CmdConnectLong  = "Uses a k8s port-forward to connect to resources within the cluster referenced by your kube-context.\n" +
		"Three default options for this command are <REGISTRY|LOGGING|GIT>. These will connect to the Zarf created resources " +
		"(assuming they were selected when performing the `zarf init` command).\n\n" +
		"Packages can provide service manifests that define their own shortcut connection options. These options will be " +
		"printed to the terminal when the package finishes deploying.\n If you don't remember what connection shortcuts your deployed " +
		"package offers, you can search your cluster for services that have the 'zarf.dev/connect-name' label. The value of that label is " +
		"the name you will pass into the 'zarf connect' command.\n\n" +
		"Even if the packages you deploy don't define their own shortcut connection options, you can use the command flags " +
		"to connect into specific resources. You can read the command flag descriptions below to get a better idea how to connect " +
		"to whatever resource you are trying to connect to."

	// zarf connect list
	CmdConnectListShort = "List all available connection shortcuts."

	CmdConnectFlagName       = "Specify the resource name.  E.g. name=unicorns or name=unicorn-pod-7448499f4d-b5bk6"
	CmdConnectFlagNamespace  = "Specify the namespace.  E.g. namespace=default"
	CmdConnectFlagType       = "Specify the resource type.  E.g. type=svc or type=pod"
	CmdConnectFlagLocalPort  = "(Optional, autogenerated if not provided) Specify the local port to bind to.  E.g. local-port=42000"
	CmdConnectFlagRemotePort = "Specify the remote port of the resource to bind to.  E.g. remote-port=8080"
	CmdConnectFlagCliOnly    = "Disable browser auto-open"

	// zarf destroy
	CmdDestroyShort = "Tear it all down, we'll miss you Zarf..."
	CmdDestroyLong  = "Tear down Zarf.\n\n" +
		"Deletes everything in the 'zarf' namespace within your connected k8s cluster.\n\n" +
		"If Zarf deployed your k8s cluster, this command will also tear your cluster down by " +
		"searching through /opt/zarf for any scripts that start with 'zarf-clean-' and executing them. " +
		"Since this is a cleanup operation, Zarf will not stop the teardown if one of the scripts produce " +
		"an error.\n\n" +
		"If Zarf did not deploy your k8s cluster, this command will delete the Zarf namespace, delete secrets " +
		"and labels that only Zarf cares about, and optionally uninstall components that Zarf deployed onto " +
		"the cluster. Since this is a cleanup operation, Zarf will not stop the uninstalls if one of the " +
		"resources produce an error while being deleted."

	CmdDestroyFlagConfirm          = "REQUIRED. Confirm the destroy action to prevent accidental deletions"
	CmdDestroyFlagRemoveComponents = "Also remove any installed components outside the zarf namespace"

	CmdDestroyErrNoScriptPath           = "Unable to find the folder (%s) which has the scripts to cleanup the cluster. Please double-check you have the right kube-context"
	CmdDestroyErrScriptPermissionDenied = "Received 'permission denied' when trying to execute the script (%s). Please double-check you have the correct kube-context."

	// zarf init
	CmdInitShort = "Prepares a k8s cluster for the deployment of Zarf packages"
	CmdInitLong  = "Injects a docker registry as well as other optional useful things (such as a git server " +
		"and a logging stack) into a k8s cluster under the 'zarf' namespace " +
		"to support future application deployments.\n" +
		"If you do not have a k8s cluster already configured, this command will give you " +
		"the ability to install a cluster locally.\n\n" +
		"This command looks for a zarf-init package in the local directory that the command was executed " +
		"from. If no package is found in the local directory and the Zarf CLI exists somewhere outside of " +
		"the current directory, Zarf will failover and attempt to find a zarf-init package in the directory " +
		"that the Zarf binary is located in.\n\n\n\n"

	CmdInitExample = `# Initializing without any optional components:
zarf init

# Initializing w/ Zarfs internal git server:
zarf init --components=git-server

# Initializing w/ Zarfs internal git server and PLG stack:
zarf init --components=git-server,logging

# Initializing w/ an internal registry but with a different nodeport:
zarf init --nodeport=30333

# Initializing w/ an external registry:
zarf init --registry-push-password={PASSWORD} --registry-push-username={USERNAME} --registry-url={URL}

# Initializing w/ an external git server:
zarf init --git-push-password={PASSWORD} --git-push-username={USERNAME} --git-url={URL}
`

	CmdInitErrFlags             = "Invalid command flags were provided."
	CmdInitErrDownload          = "failed to download the init package: %s"
	CmdInitErrValidateGit       = "the 'git-push-username' and 'git-push-password' flags must be provided if the 'git-url' flag is provided"
	CmdInitErrValidateRegistry  = "the 'registry-push-username' and 'registry-push-password' flags must be provided if the 'registry-url' flag is provided"
	CmdInitErrValidateArtifact  = "the 'artifact-push-username' and 'artifact-push-token' flags must be provided if the 'artifact-url' flag is provided"
	CmdInitErrUnableCreateCache = "Unable to create the cache directory: %s"

	CmdInitDownloadAsk       = "It seems the init package could not be found locally, but can be downloaded from %s"
	CmdInitDownloadNote      = "Note: This will require an internet connection."
	CmdInitDownloadConfirm   = "Do you want to download this init package?"
	CmdInitDownloadErrCancel = "confirm selection canceled: %s"
	CmdInitDownloadErrManual = "download the init package manually and place it in the current working directory"

	CmdInitFlagSet = "Specify deployment variables to set on the command line (KEY=value)"

	CmdInitFlagConfirm      = "Confirms package deployment without prompting. ONLY use with packages you trust. Skips prompts to review SBOM, configure variables, select optional components and review potential breaking changes."
	CmdInitFlagComponents   = "Specify which optional components to install.  E.g. --components=git-server,logging"
	CmdInitFlagStorageClass = "Specify the storage class to use for the registry and git server.  E.g. --storage-class=standard"

	CmdInitFlagGitURL      = "External git server url to use for this Zarf cluster"
	CmdInitFlagGitPushUser = "Username to access to the git server Zarf is configured to use. User must be able to create repositories via 'git push'"
	CmdInitFlagGitPushPass = "Password for the push-user to access the git server"
	CmdInitFlagGitPullUser = "Username for pull-only access to the git server"
	CmdInitFlagGitPullPass = "Password for the pull-only user to access the git server"

	CmdInitFlagRegURL      = "External registry url address to use for this Zarf cluster"
	CmdInitFlagRegNodePort = "Nodeport to access a registry internal to the k8s cluster. Between [30000-32767]"
	CmdInitFlagRegPushUser = "Username to access to the registry Zarf is configured to use"
	CmdInitFlagRegPushPass = "Password for the push-user to connect to the registry"
	CmdInitFlagRegPullUser = "Username for pull-only access to the registry"
	CmdInitFlagRegPullPass = "Password for the pull-only user to access the registry"
	CmdInitFlagRegSecret   = "Registry secret value"

	CmdInitFlagArtifactURL       = "External artifact registry url to use for this Zarf cluster"
	CmdInitFlagArtifactPushUser  = "Username to access to the artifact registry Zarf is configured to use. User must be able to upload package artifacts."
	CmdInitFlagArtifactPushToken = "API Token for the push-user to access the artifact registry"

	// zarf internal
	CmdInternalShort = "Internal tools used by zarf"

	CmdInternalAgentShort = "Runs the zarf agent"
	CmdInternalAgentLong  = "NOTE: This command is a hidden command and generally shouldn't be run by a human.\n" +
		"This command starts up a http webhook that Zarf deployments use to mutate pods to conform " +
		"with the Zarf container registry and Gitea server URLs."

	CmdInternalGenerateCliDocsShort   = "Creates auto-generated markdown of all the commands for the CLI"
	CmdInternalGenerateCliDocsSuccess = "Successfully created the CLI documentation"

	CmdInternalConfigSchemaShort = "Generates a JSON schema for the zarf.yaml configuration"
	CmdInternalConfigSchemaErr   = "Unable to generate the zarf config schema"

	CmdInternalAPISchemaShort       = "Generates a JSON schema from the API types"
	CmdInternalAPISchemaGenerateErr = "Unable to generate the zarf api schema"

	CmdInternalCreateReadOnlyGiteaUserShort = "Creates a read-only user in Gitea"
	CmdInternalCreateReadOnlyGiteaUserLong  = "Creates a read-only user in Gitea by using the Gitea API. " +
		"This is called internally by the supported Gitea package component."
	CmdInternalCreateReadOnlyGiteaUserLoadErr = "Unable to load the Zarf state"
	CmdInternalCreateReadOnlyGiteaUserErr     = "Unable to create a read-only user in the Gitea service."

	CmdInternalUIShort = "Launch the experimental Zarf UI"

	CmdInternalIsValidHostnameShort = "Checks if the current machine's hostname is RFC1123 compliant"

	// zarf package
	CmdPackageShort           = "Zarf package commands for creating, deploying, and inspecting packages"
	CmdPackageFlagConcurrency = "Number of concurrent layer operations to perform when interacting with a remote package."

	CmdPackageCreateShort = "Use to create a Zarf package from a given directory or the current directory"
	CmdPackageCreateLong  = "Builds an archive of resources and dependencies defined by the 'zarf.yaml' in the active directory.\n" +
		"Private registries and repositories are accessed via credentials in your local '~/.docker/config.json', " +
		"'~/.git-credentials' and '~/.netrc'.\n"

	CmdPackageDeployShort = "Use to deploy a Zarf package from a local file or URL (runs offline)"
	CmdPackageDeployLong  = "Uses current kubecontext to deploy the packaged tarball onto a k8s cluster."

	CmdPackageInspectShort = "Lists the payload of a Zarf package (runs offline)"
	CmdPackageInspectLong  = "Lists the payload of a compiled package file (runs offline)\n" +
		"Unpacks the package tarball into a temp directory and displays the " +
		"contents of the archive."

	CmdPackageListShort         = "List out all of the packages that have been deployed to the cluster"
	CmdPackageListNoPackageWarn = "Unable to get the packages deployed to the cluster"

	CmdPackageRemoveShort       = "Use to remove a Zarf package that has been deployed already"
	CmdPackageRemoveTarballErr  = "Invalid tarball path provided"
	CmdPackageRemoveExtractErr  = "Unable to extract the package contents"
	CmdPackageRemoveReadZarfErr = "Unable to read zarf.yaml"

	CmdPackageCreateFlagConfirm            = "Confirm package creation without prompting"
	CmdPackageCreateFlagSet                = "Specify package variables to set on the command line (KEY=value)"
	CmdPackageCreateFlagOutput             = "Specify the output (either a directory or an oci:// URL) for the created Zarf package"
	CmdPackageCreateFlagSbom               = "View SBOM contents after creating the package"
	CmdPackageCreateFlagSbomOut            = "Specify an output directory for the SBOMs from the created Zarf package"
	CmdPackageCreateFlagSkipSbom           = "Skip generating SBOM for this package"
	CmdPackageCreateFlagMaxPackageSize     = "Specify the maximum size of the package in megabytes, packages larger than this will be split into multiple parts. Use 0 to disable splitting."
	CmdPackageCreateFlagSigningKey         = "Path to private key file for signing packages"
	CmdPackageCreateFlagSigningKeyPassword = "Password to the private key file used for signing packages"
	CmdPackageCreateFlagDifferential       = "Build a package that only contains the differential changes from local resources and differing remote resources from the specified previously built package"
	CmdPackageCreateFlagRegistryOverride   = "Specify a map of domains to override on package create when pulling images (e.g. --registry-override docker.io=dockerio-reg.enterprise.intranet)"

	CmdPackageDeployFlagConfirm                = "Confirms package deployment without prompting. ONLY use with packages you trust. Skips prompts to review SBOM, configure variables, select optional components and review potential breaking changes."
	CmdPackageDeployFlagAdoptExistingResources = "Adopts any pre-existing K8s resources into the Helm charts managed by Zarf. ONLY use when you have existing deployments you want Zarf to takeover."
	CmdPackageDeployFlagSet                    = "Specify deployment variables to set on the command line (KEY=value)"
	CmdPackageDeployFlagComponents             = "Comma-separated list of components to install.  Adding this flag will skip the init prompts for which components to install"
	CmdPackageDeployFlagShasum                 = "Shasum of the package to deploy. Required if deploying a remote package and \"--insecure\" is not provided"
	CmdPackageDeployFlagSget                   = "Path to public sget key file for remote packages signed via cosign"
	CmdPackageDeployFlagPublicKey              = "Path to public key file for validating signed packages"
	CmdPackageDeployValidateArchitectureErr    = "this package architecture is %s, but the target cluster has the %s architecture. These architectures must be the same"

	CmdPackageInspectFlagSbom      = "View SBOM contents while inspecting the package"
	CmdPackageInspectFlagSbomOut   = "Specify an output directory for the SBOMs from the inspected Zarf package"
	CmdPackageInspectFlagValidate  = "Validate any checksums and signatures while inspecting the package"
	CmdPackageInspectFlagPublicKey = "Path to a public key file that will be used to validate a signed package"

	CmdPackageRemoveFlagConfirm    = "REQUIRED. Confirm the removal action to prevent accidental deletions"
	CmdPackageRemoveFlagComponents = "Comma-separated list of components to uninstall"

	CmdPackagePublishFlagSigningKey         = "Path to private key file for signing packages"
	CmdPackagePublishFlagSigningKeyPassword = "Password to the private key file used for publishing packages"

	CmdPackagePullFlagOutputDirectory = "Specify the output directory for the pulled Zarf package"
	CmdPackagePullFlagPublicKey       = "Path to public key file for validating signed packages"

	// zarf bundle
	CmdBundleShort           = "Zarf commands for creating, deploying, removing, pulling, and inspecting bundles"
	CmdBundleFlagConcurrency = "Number of concurrent layer operations to perform when interacting with a remote bundle."

	CmdBundleCreateShort = "Create a Zarf bundle from a given directory or the current directory"

	CmdBundleDeployShort = "Deploy a Zarf bundle from a local file or URL (runs offline)"

	CmdBundleInspectShort = "Display the zarf.yaml of a compiled Zarf bundle (runs offline)"

	CmdBundleRemoveShort = "Remove a Zarf bundle or sub-packages that have been deployed already"

	CmdBundlePullShort = "Pull a Zarf bundle from a remote reigstry and save to the local file system"

	// zarf prepare
	CmdPrepareShort = "Tools to help prepare assets for packaging"

	CmdPreparePatchGitShort = "Converts all .git URLs to the specified Zarf HOST and with the Zarf URL pattern in a given FILE.  NOTE:\n" +
		"This should only be used for manifests that are not mutated by the Zarf Agent Mutating Webhook."
	CmdPreparePatchGitFileWriteErr = "Unable to write the changes back to the file"

	CmdPrepareSha256sumShort   = "Generate a SHA256SUM for the given file"
	CmdPrepareSha256sumHashErr = "Unable to compute the hash"

	CmdPrepareFindImagesShort = "Evaluates components in a zarf file to identify images specified in their helm charts and manifests"
	CmdPrepareFindImagesLong  = "Evaluates components in a zarf file to identify images specified in their helm charts and manifests.\n\n" +
		"Components that have repos that host helm charts can be processed by providing the --repo-chart-path."

	CmdPrepareGenerateConfigShort = "Generates a config file for Zarf"
	CmdPrepareGenerateConfigLong  = "Generates a Zarf config file for controlling how the Zarf CLI operates. Optionally accepts a filename to write the config to.\n\n" +
		"The extension will determine the format of the config file, e.g. env-1.yaml, env-2.json, env-3.toml etc.\n" +
		"Accepted extensions are json, toml, yaml.\n\n" +
		"NOTE: This file must not already exist. If no filename is provided, the config will be written to the current working directory as zarf-config.toml."

	CmdPrepareFlagSet           = "Specify package variables to set on the command line (KEY=value). Note, if using a config file, this will be set by [package.create.set]."
	CmdPrepareFlagRepoChartPath = `If git repos hold helm charts, often found with gitops tools, specify the chart path, e.g. "/" or "/chart"`
	CmdPrepareFlagGitAccount    = "User or organization name for the git account that the repos are created under."
	CmdPrepareFlagKubeVersion   = "Override the default helm template KubeVersion when performing a package chart template"

	// zarf tools
	CmdToolsShort = "Collection of additional tools to make airgap easier"

	CmdToolsArchiverShort           = "Compress/Decompress generic archives, including Zarf packages."
	CmdToolsArchiverCompressShort   = "Compress a collection of sources based off of the destination file extension."
	CmdToolsArchiverCompressErr     = "Unable to perform compression"
	CmdToolsArchiverDecompressShort = "Decompress an archive or Zarf package based off of the source file extension."
	CmdToolsArchiverDecompressErr   = "Unable to perform decompression: %s"
	CmdToolsArchiverUnarchiveAllErr = "Unable to unarchive all nested tarballs"

	CmdToolsRegistryShort = "Tools for working with container registries using go-containertools."

	CmdToolGetGitDeprecation  = "Deprecated: This command has been replaced by 'zarf tools get-creds git' and will be removed in a future release."
	CmdToolsGetGitPasswdShort = "Returns the push user's password for the Git server"
	CmdToolsGetGitPasswdLong  = "Reads the password for a user with push access to the configured Git server from the zarf-state secret in the zarf namespace"
	CmdToolsGetGitPasswdInfo  = "Git Server Push Password: "

	CmdToolsMonitorShort = "Launch a terminal UI to monitor the connected cluster using K9s."

	CmdToolsClearCacheShort         = "Clears the configured git and image cache directory."
	CmdToolsClearCacheErr           = "Unable to clear the cache directory %s"
	CmdToolsClearCacheSuccess       = "Successfully cleared the cache from %s"
	CmdToolsClearCacheFlagCachePath = "Specify the location of the Zarf artifact cache (images and git repositories)"

	CmdToolsDownloadInitShort               = "Download the init package for the current Zarf version into the specified directory."
	CmdToolsDownloadInitFlagOutputDirectory = "Specify a directory to place the init package in."
	CmdToolsDownloadInitErr                 = "Unable to download the init package: %s"

	CmdToolsGenPkiShort       = "Generates a Certificate Authority and PKI chain of trust for the given host"
	CmdToolsGenPkiSuccess     = "Successfully created a chain of trust for %s"
	CmdToolsGenPkiFlagAltName = "Specify Subject Alternative Names for the certificate"

	CmdToolsGenKeyShort                 = "Generates a cosign public/private keypair that can be used to sign packages"
	CmdToolsGenKeyPrompt                = "Private key password (empty for no password): "
	CmdToolsGenKeyPromptAgain           = "Private key password again (empty for no password): "
	CmdToolsGenKeyPromptExists          = "File %s already exists. Overwrite? "
	CmdToolsGenKeyErrUnableGetPassword  = "unable to get password for private key: %s"
	CmdToolsGenKeyErrPasswordsNotMatch  = "passwords do not match"
	CmdToolsGenKeyErrUnableToGenKeypair = "unable to generate key pair: %s"
	CmdToolsGenKeyErrNoConfirmOverwrite = "did not receive confirmation for overwriting key file(s)"
	CmdToolsGenKeySuccess               = "Generated key pair and written to %s and %s"

	CmdToolsSbomShort = "Generates a Software Bill of Materials (SBOM) for the given package"
	CmdToolsSbomErr   = "Unable to create sbom (syft) CLI"

	CmdToolsWaitForShort = "Waits for a given Kubernetes resource to be ready"
	CmdToolsWaitForLong  = "By default Zarf will wait for all Kubernetes resources to be ready before completion of a component during a deployment.\n" +
		"This command can be used to wait for a Kubernetes resources to exist and be ready that may be created by a Gitops tool or a Kubernetes operator.\n" +
		"You can also wait for arbitrary network endpoints using REST or TCP checks.\n\n"
	CmdToolsWaitForFlagTimeout        = "Specify the timeout duration for the wait command."
	CmdToolsWaitForErrTimeoutString   = "Invalid timeout duration. Please use a valid duration string (e.g. 1s, 2m, 3h)."
	CmdToolsWaitForErrTimeout         = "Wait timed out."
	CmdToolsWaitForErrConditionString = "Invalid HTTP status code. Please use a valid HTTP status code (e.g. 200, 404, 500)."
	CmdToolsWaitForErrZarfPath        = "Could not locate the current Zarf binary path."
	CmdToolsWaitForFlagNamespace      = "Specify the namespace of the resources to wait for."

	CmdToolsKubectlDocs = "Kubectl command. See https://kubernetes.io/docs/reference/kubectl/overview/ for more information."

	CmdToolsGetCredsShort = "Display a Table of credentials for deployed components. Pass a component name to get a single credential."
	CmdToolsGetCredsLong  = "Display a Table of credentials for deployed components. Pass a component name to get a single credential. i.e. 'zarf tools get-creds registry'"

	// zarf version
	CmdVersionShort = "Version of the Zarf binary"
	CmdVersionLong  = "Displays the version of the Zarf release that the Zarf binary was built from."

	// cmd viper setup
	CmdViperErrLoadingConfigFile = "failed to load config file: %s"
	CmdViperInfoUsingConfigFile  = "Using config file %s"
)

// Zarf Agent messages
// These are only seen in the Kubernetes logs.
const (
	AgentInfoWebhookAllowed = "Webhook [%s - %s] - Allowed: %t"
	AgentInfoShutdown       = "Shutdown gracefully..."
	AgentInfoPort           = "Server running in port: %s"

	AgentErrBadRequest             = "could not read request body: %s"
	AgentErrBindHandler            = "Unable to bind the webhook handler"
	AgentErrCouldNotDeserializeReq = "could not deserialize request: %s"
	AgentErrGetState               = "failed to load zarf state from file: %w"
	AgentErrHostnameMatch          = "failed to complete hostname matching: %w"
	AgentErrImageSwap              = "Unable to swap the host for (%s)"
	AgentErrInvalidMethod          = "invalid method only POST requests are allowed"
	AgentErrInvalidOp              = "invalid operation: %s"
	AgentErrInvalidType            = "only content type 'application/json' is supported"
	AgentErrMarshallJSONPatch      = "unable to marshall the json patch"
	AgentErrMarshalResponse        = "unable to marshal the response"
	AgentErrNilReq                 = "malformed admission review: request is nil"
	AgentErrShutdown               = "unable to properly shutdown the web server"
	AgentErrStart                  = "Failed to start the web server"
	AgentErrUnableTransform        = "unable to transform the provided request; see zarf http proxy logs for more details"
)

// src/internal/packager/create
const (
	PkgCreateErrDifferentialSameVersion = "unable to create a differential package with the same version as the package you are using as a reference; the package version must be incremented"
)

// src/internal/packager/validate.
const (
	PkgValidateTemplateDeprecation        = "Package template '%s' is using the deprecated syntax ###ZARF_PKG_VAR_%s###.  This will be removed in a future Zarf version.  Please update to ###ZARF_PKG_TMPL_%s###."
	PkgValidateMustBeUppercase            = "variable name '%s' must be all uppercase and contain no special characters except _"
	PkgValidateErrAction                  = "invalid action: %w"
	PkgValidateErrActionVariables         = "component %s cannot contain setVariables outside of onDeploy in actions"
	PkgValidateErrActionCmdWait           = "action %s cannot be both a command and wait action"
	PkgValidateErrActionClusterNetwork    = "a single wait action must contain only one of cluster or network"
	PkgValidateErrChart                   = "invalid chart definition: %w"
	PkgValidateErrChartName               = "chart %s exceed the maximum length of %d characters"
	PkgValidateErrChartNameMissing        = "chart %s must include a name"
	PkgValidateErrChartNamespaceMissing   = "chart %s must include a namespace"
	PkgValidateErrChartURLOrPath          = "chart %s must only have a url or localPath"
	PkgValidateErrChartVersion            = "chart %s must include a chart version"
	PkgValidateErrComponentNameNotUnique  = "component name '%s' is not unique"
	PkgValidateErrComponent               = "invalid component: %w"
	PkgValidateErrComponentReqDefault     = "component %s cannot be both required and default"
	PkgValidateErrComponentReqGrouped     = "component %s cannot be both required and grouped"
	PkgValidateErrComponentYOLO           = "component %s incompatible with the online-only package flag (metadata.yolo): %w"
	PkgValidateErrConstant                = "invalid package constant: %w"
	PkgValidateErrImportPathInvalid       = "invalid file path '%s' provided directory must contain a valid zarf.yaml file"
	PkgValidateErrImportURLInvalid        = "invalid url '%s' provided"
	PkgValidateErrImportOptions           = "imported package %s must have either a url or a path"
	PkgValidateErrImportPathMissing       = "imported package %s must include a path"
	PkgValidateErrInitNoYOLO              = "sorry, you can't YOLO an init package"
	PkgValidateErrManifest                = "invalid manifest definition: %w"
	PkgValidateErrManifestFileOrKustomize = "manifest %s must have at least one file or kustomization"
	PkgValidateErrManifestNameLength      = "manifest %s exceed the maximum length of %d characters"
	PkgValidateErrManifestNameMissing     = "manifest %s must include a name"
	PkgValidateErrName                    = "invalid package name: %w"
	PkgValidateErrPkgConstantName         = "constant name '%s' must be all uppercase and contain no special characters except _"
	PkgValidateErrPkgName                 = "package name '%s' must be all lowercase and contain no special characters except -"
	PkgValidateErrVariable                = "invalid package variable: %w"
	PkgValidateErrYOLONoArch              = "cluster architecture not allowed"
	PkgValidateErrYOLONoDistro            = "cluster distros not allowed"
	PkgValidateErrYOLONoGit               = "git repos not allowed"
	PkgValidateErrYOLONoOCI               = "OCI images not allowed"
)

// Collection of reusable error messages.
var (
	ErrInitNotFound      = errors.New("this command requires a zarf-init package, but one was not found on the local system. Re-run the last command again without '--confirm' to download the package")
	ErrUnableToCheckArch = errors.New("unable to get the configured cluster's architecture")
)
