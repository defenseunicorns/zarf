## zarf package deploy

Use to deploy a Zarf package from a local file or URL (runs offline)

### Synopsis

Uses current kubecontext to deploy the packaged tarball onto a k8s cluster.

```
zarf package deploy [PACKAGE] [flags]
```

### Options

```
      --components string    Comma-separated list of components to install.  Adding this flag will skip the init prompts for which components to install
      --confirm              Confirm package deployment without prompting
  -h, --help                 help for deploy
      --set stringToString   Specify deployment variables to set on the command line (KEY=value) (default [])
      --sget string          Path to public sget key file for remote packages signed via cosign
      --shasum --insecure    Shasum of the package to deploy. Required if deploying a remote package and --insecure is not provided
```

### Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images
      --insecure              Allow access to insecure registries and disable other recommended security enforcements. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf package](zarf_package.md)	 - Zarf package commands for creating, deploying, and inspecting packages

