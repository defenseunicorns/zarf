## zarf completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(zarf completion zsh); compdef _zarf zarf

To load completions for every new session, execute once:

#### Linux:

	zarf completion zsh > "${fpath[1]}/_zarf"

#### macOS:

	zarf completion zsh > $(brew --prefix)/share/zsh/site-functions/_zarf

You will need to start a new shell for this setup to take effect.


```
zarf completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -a, --architecture string   Set the architecture to use for the package. Valid options are: amd64, arm64.
  -l, --log-level string      Set the log level. Valid options are: warn, info, debug, trace. (default "info")
      --no-log-file           Disable logging to a file.
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf completion](zarf_completion.md)	 - Generate the autocompletion script for the specified shell

