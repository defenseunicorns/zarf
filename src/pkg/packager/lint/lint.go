// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package lint contains functions for verifying zarf yaml files are valid
package lint

import (
	"context"
	"fmt"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/pkg/packager/composer"
	"github.com/defenseunicorns/zarf/src/pkg/packager/rules"
	"github.com/defenseunicorns/zarf/src/pkg/packager/schema"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
)

// Validate the given Zarf package. The Zarf package should not already be composed when sent to this function.
func Validate(ctx context.Context, pkg types.ZarfPackage, createOpts types.ZarfCreateOptions) ([]types.PackageFinding, error) {
	var findings []types.PackageFinding
	compFindings, err := lintComponents(ctx, pkg, createOpts)
	if err != nil {
		return nil, err
	}
	findings = append(findings, compFindings...)

	schemaFindings, err := schema.Validate()
	if err != nil {
		return nil, err
	}
	findings = append(findings, schemaFindings...)

	return findings, nil
}

func lintComponents(ctx context.Context, pkg types.ZarfPackage, createOpts types.ZarfCreateOptions) ([]types.PackageFinding, error) {
	var findings []types.PackageFinding

	for i, component := range pkg.Components {
		arch := config.GetArch(pkg.Metadata.Architecture)
		if !composer.CompatibleComponent(component, arch, createOpts.Flavor) {
			continue
		}

		chain, err := composer.NewImportChain(ctx, component, i, pkg.Metadata.Name, arch, createOpts.Flavor)

		if err != nil {
			return nil, err
		}

		node := chain.Head()
		for node != nil {
			component := node.ZarfComponent
			compFindings, err := fillComponentTemplate(&component, &createOpts)
			if err != nil {
				return nil, err
			}
			compFindings = append(compFindings, rules.CheckComponentValues(component, node.Index())...)
			for i := range compFindings {
				compFindings[i].PackagePathOverride = node.ImportLocation()
				compFindings[i].PackageNameOverride = node.OriginalPackageName()
			}
			findings = append(findings, compFindings...)
			node = node.Next()
		}
	}
	return findings, nil
}

func fillComponentTemplate(c *types.ZarfComponent, createOpts *types.ZarfCreateOptions) ([]types.PackageFinding, error) {
	var findings []types.PackageFinding
	var templateMap map[string]string

	setVarsAndWarn := func(templatePrefix string, deprecated bool) {
		yamlTemplates, err := utils.FindYamlTemplates(c, templatePrefix, "###")
		if err != nil {
			findings = append(findings, types.PackageFinding{
				Description: err.Error(),
				Severity:    types.SevWarn,
			})
		}

		for key := range yamlTemplates {
			if deprecated {
				findings = append(findings, types.PackageFinding{
					Description: fmt.Sprintf(lang.PkgValidateTemplateDeprecation, key, key, key),
					Severity:    types.SevWarn,
				})
			}
			_, present := createOpts.SetVariables[key]
			if !present {
				findings = append(findings, types.PackageFinding{
					Description: lang.UnsetVarLintWarning,
					Severity:    types.SevWarn,
				})
			}
		}

	}

	setVarsAndWarn(types.ZarfPackageTemplatePrefix, false)

	// [DEPRECATION] Set the Package Variable syntax as well for backward compatibility
	setVarsAndWarn(types.ZarfPackageVariablePrefix, true)

	if err := utils.ReloadYamlTemplate(c, templateMap); err != nil {
		return nil, err
	}
	return findings, nil
}
