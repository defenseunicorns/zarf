## zarf tools sbom attest

Generate a package SBOM as an attestation for the given [SOURCE] container image

### Synopsis

Generate a packaged-based Software Bill Of Materials (SBOM) from a container image as the predicate of an in-toto attestation

```
zarf tools sbom attest --output [FORMAT] --key [KEY] [SOURCE] [flags]
```

### Options

```
      --catalogers stringArray     enable one or more package catalogers
      --cert string                path to the x.509 certificate in PEM format to include in the OCI Signature
      --exclude stringArray        exclude paths from being scanned using a glob expression
      --file string                file to write the default report output to (default is STDOUT)
      --force                      skip warnings and confirmations
      --fulcio-url string          address of sigstore PKI server (default "https://fulcio.sigstore.dev")
  -h, --help                       help for attest
      --identity-token string      identity token to use for certificate from fulcio
      --insecure-skip-verify       skip verifying fulcio certificat and the SCT (Signed Certificate Timestamp) (this should only be used for testing).
      --key string                 path to the private key file to use for attestation (default "cosign.key")
      --name string                set the name of the target being analyzed
      --no-upload                  do not upload the generated attestation
      --oidc-client-id string      OIDC client ID for application (default "sigstore")
      --oidc-issuer string         OIDC provider to be used to issue ID token (default "https://oauth2.sigstore.dev/auth")
      --oidc-redirect-url string   OIDC redirect URL (Optional)
  -o, --output stringArray         report output format, options=[syft-json cyclonedx-xml cyclonedx-json github github-json spdx-tag-value spdx-json table text template] (default [table])
      --platform string            an optional platform specifier for container image sources (e.g. 'linux/arm64', 'linux/arm64/v8', 'arm64', 'linux')
      --recursive                  if a multi-arch image is specified, additionally sign each discrete image
      --rekor-url string           address of rekor STL server (default "https://rekor.sigstore.dev")
  -s, --scope string               selection of layers to catalog, options=[Squashed AllLayers] (default "Squashed")
  -t, --template string            specify the path to a Go template file
```

### Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images
  -c, --config string         application config file
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
  -q, --quiet                 suppress all logging output
      --tmpdir string         Specify the temporary directory to use for intermediate files
  -v, --verbose count         increase verbosity (-v = info, -vv = debug)
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf tools sbom](zarf_tools_sbom.md)	 - Generates a Software Bill of Materials (SBOM) for the given package

