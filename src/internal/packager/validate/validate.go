// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package validate provides zarf package validation functions
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
)

// Run performs config validations
func Run(pkg types.ZarfPackage) error {
	if err := validatePackageName(pkg.Metadata.Name); err != nil {
		return fmt.Errorf("invalid package name: %w", err)
	}

	for _, variable := range pkg.Variables {
		if err := validatePackageVariable(variable); err != nil {
			return fmt.Errorf("invalid package variable: %w", err)
		}
	}

	for _, constant := range pkg.Constants {
		if err := validatePackageConstant(constant); err != nil {
			return fmt.Errorf("invalid package constant: %w", err)
		}
	}

	uniqueNames := make(map[string]bool)

	for _, component := range pkg.Components {
		// ensure component name is unique
		if _, ok := uniqueNames[component.Name]; ok {
			return fmt.Errorf("component name '%s' is not unique", component.Name)
		}
		uniqueNames[component.Name] = true

		validateComponent(component)
	}

	return nil
}

// ImportPackage validates the package trying to be imported.
func ImportPackage(composedComponent *types.ZarfComponent) error {
	intro := fmt.Sprintf("imported package %s", composedComponent.Name)
	path := composedComponent.Import.Path

	// ensure path exists
	if !(len(path) > 0) {
		return fmt.Errorf("%s must include a path", intro)
	}

	// remove zarf.yaml from path if path has zarf.yaml suffix
	if strings.HasSuffix(path, config.ZarfYAML) {
		path = strings.Split(path, config.ZarfYAML)[0]
	}

	// add a forward slash to end of path if it does not have one
	if !strings.HasSuffix(path, "/") {
		path = filepath.Clean(path) + string(os.PathSeparator)
	}

	// ensure there is a zarf.yaml in provided path
	if utils.InvalidPath(path + config.ZarfYAML) {
		return fmt.Errorf("invalid file path \"%s\" provided directory must contain a valid zarf.yaml file", composedComponent.Import.Path)
	}

	return nil
}

func oneIfNotEmpty(testString string) int {
	if testString == "" {
		return 0
	}

	return 1
}

func validateComponent(component types.ZarfComponent) error {
	if component.Required {
		if component.Default {
			return fmt.Errorf("component %s cannot be both required and default", component.Name)
		}
		if component.Group != "" {
			return fmt.Errorf("component %s cannot be both required and grouped", component.Name)
		}
	}

	for _, chart := range component.Charts {
		if err := validateChart(chart); err != nil {
			return fmt.Errorf("invalid chart definition: %w", err)
		}
	}
	for _, manifest := range component.Manifests {
		if err := validateManifest(manifest); err != nil {
			return fmt.Errorf("invalid manifest definition: %w", err)
		}
	}

	return nil
}

func validatePackageName(subject string) error {
	// https://regex101.com/r/vpi8a8/1
	isValid := regexp.MustCompile(`^[a-z0-9\-]+$`).MatchString

	if !isValid(subject) {
		return fmt.Errorf("package name '%s' must be all lowercase and contain no special characters except -", subject)
	}

	return nil
}

func validatePackageVariable(subject types.ZarfPackageVariable) error {
	isAllCapsUnderscore := regexp.MustCompile(`^[A-Z_]+$`).MatchString

	// ensure the variable name is only capitals and underscores
	if !isAllCapsUnderscore(subject.Name) {
		return fmt.Errorf("variable name '%s' must be all uppercase and contain no special characters except _", subject.Name)
	}

	return nil
}

func validatePackageConstant(subject types.ZarfPackageConstant) error {
	isAllCapsUnderscore := regexp.MustCompile(`^[A-Z_]+$`).MatchString

	// ensure the constant name is only capitals and underscores
	if !isAllCapsUnderscore(subject.Name) {
		return fmt.Errorf("constant name '%s' must be all uppercase and contain no special characters except _", subject.Name)
	}

	return nil
}

func validateChart(chart types.ZarfChart) error {
	intro := fmt.Sprintf("chart %s", chart.Name)

	// Don't allow empty names
	if chart.Name == "" {
		return fmt.Errorf("%s must include a name", intro)
	}

	// Helm max release name
	if len(chart.Name) > config.ZarfMaxChartNameLength {
		return fmt.Errorf("%s exceed the maximum length of %d characters",
			intro,
			config.ZarfMaxChartNameLength)
	}

	// Must have a namespace
	if chart.Namespace == "" {
		return fmt.Errorf("%s must include a namespace", intro)
	}

	// Must only have a url or localPath
	count := oneIfNotEmpty(chart.Url) + oneIfNotEmpty(chart.LocalPath)
	if count != 1 {
		return fmt.Errorf("%s must only have a url or localPath", intro)
	}

	// Must have a version
	if chart.Version == "" {
		return fmt.Errorf("%s must include a chart version", intro)
	}

	return nil
}

func validateManifest(manifest types.ZarfManifest) error {
	intro := fmt.Sprintf("chart %s", manifest.Name)

	// Don't allow empty names
	if manifest.Name == "" {
		return fmt.Errorf("%s must include a name", intro)
	}

	// Helm max release name
	if len(manifest.Name) > config.ZarfMaxChartNameLength {
		return fmt.Errorf("%s exceed the maximum length of %d characters",
			intro,
			config.ZarfMaxChartNameLength)
	}

	// Require files in manifest
	if len(manifest.Files) < 1 && len(manifest.Kustomizations) < 1 {
		return fmt.Errorf("%s must have at least one file or kustomization", intro)
	}

	return nil
}
