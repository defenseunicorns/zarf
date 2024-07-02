// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package creator contains functions for creating Zarf packages.
package creator

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/layout"
	"github.com/defenseunicorns/zarf/src/pkg/packager/deprecated"
	"github.com/defenseunicorns/zarf/src/pkg/packager/lint"
	"github.com/defenseunicorns/zarf/src/pkg/packager/schema"
	"github.com/defenseunicorns/zarf/src/types"
)

func validate(createOpts types.ZarfCreateOptions, pkg types.ZarfPackage) error {
	if err := pkg.Validate(); err != nil {
		return fmt.Errorf("package validation failed: %w", err)
	}

	findings, err := schema.Validate()
	if err != nil {
		return fmt.Errorf("unable to check schema: %w", err)
	}

	lint.PrintFindings(findings, types.SevErr, createOpts.BaseDir, pkg.Metadata.Name)
	if lint.HasSeverity(findings, types.SevErr) {
		return fmt.Errorf("found errors in package")
	}

	return nil
}

func loadWithValidate(ctx context.Context, c Creator, src *layout.PackagePaths) (types.ZarfPackage, []types.PackageFinding, error) {
	pkg, warnings, err := c.LoadPackageDefinition(ctx, src)
	if err != nil {
		return types.ZarfPackage{}, nil, err
	}
	err = c.Validate(ctx, pkg)
	if err != nil {
		return types.ZarfPackage{}, nil, err
	}
	return pkg, warnings, err
}

// recordPackageMetadata records various package metadata during package create.
func recordPackageMetadata(pkg *types.ZarfPackage, createOpts types.ZarfCreateOptions) error {
	now := time.Now()
	// Just use $USER env variable to avoid CGO issue.
	// https://groups.google.com/g/golang-dev/c/ZFDDX3ZiJ84.
	// Record the name of the user creating the package.
	if runtime.GOOS == "windows" {
		pkg.Build.User = os.Getenv("USERNAME")
	} else {
		pkg.Build.User = os.Getenv("USER")
	}

	// Record the hostname of the package creation terminal.
	// The error here is ignored because the hostname is not critical to the package creation.
	hostname, _ := os.Hostname()
	pkg.Build.Terminal = hostname

	if pkg.IsInitConfig() {
		pkg.Metadata.Version = config.CLIVersion
	}

	pkg.Build.Architecture = pkg.Metadata.Architecture

	// Record the Zarf Version the CLI was built with.
	pkg.Build.Version = config.CLIVersion

	// Record the time of package creation.
	pkg.Build.Timestamp = now.Format(time.RFC1123Z)

	// Record the migrations that will be ran on the package.
	pkg.Build.Migrations = []string{
		deprecated.ScriptsToActionsMigrated,
		deprecated.PluralizeSetVariable,
	}

	// Record the flavor of Zarf used to build this package (if any).
	pkg.Build.Flavor = createOpts.Flavor

	pkg.Build.RegistryOverrides = createOpts.RegistryOverrides

	// Record the latest version of Zarf without breaking changes to the package structure.
	pkg.Build.LastNonBreakingVersion = deprecated.LastNonBreakingVersion

	return nil
}
