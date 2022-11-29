## zarf tools sbom login

Log in to a registry

```
zarf tools sbom login [OPTIONS] [SERVER] [flags]
```

### Options

```
  -h, --help              help for login
  -p, --password string   Password
      --password-stdin    Take the password from stdin
  -u, --username string   Username
```

### Options inherited from parent commands

```
  -a, --architecture string   Set the architecture to use for the package. Valid options are: amd64, arm64.
  -c, --config string         application config file
  -l, --log-level string      Set the log level. Valid options are: warn, info, debug, trace. (default "info")
      --no-log-file           Disable logging to a file.
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
  -q, --quiet                 suppress all logging output
      --tmpdir string         Specify the temporary directory to use for intermediate files
  -v, --verbose count         increase verbosity (-v = info, -vv = debug)
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf tools sbom](zarf_tools_sbom.md)	 - Generates a Software Bill of Materials (SBOM) for the given package

