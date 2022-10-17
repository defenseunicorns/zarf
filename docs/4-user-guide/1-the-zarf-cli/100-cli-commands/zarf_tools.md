## zarf tools

Collection of additional tools to make airgap easier

### Options

```
  -h, --help   help for tools
```

### Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf](zarf.md)	 - DevSecOps Airgap Toolkit
* [zarf tools archiver](zarf_tools_archiver.md)	 - Compress/Decompress tools for Zarf packages
* [zarf tools clear-cache](zarf_tools_clear-cache.md)	 - Clears the configured git and image cache directory
* [zarf tools gen-pki](zarf_tools_gen-pki.md)	 - Generates a Certificate Authority and PKI chain of trust for the given host
* [zarf tools get-git-password](zarf_tools_get-git-password.md)	 - Returns the push user's password for the Git server
* [zarf tools monitor](zarf_tools_monitor.md)	 - Launch K9s tool for managing K8s clusters
* [zarf tools registry](zarf_tools_registry.md)	 - Collection of registry commands provided by Crane
* [zarf tools sbom](zarf_tools_sbom.md)	 - SBOM tools provided by Anchore Syft
