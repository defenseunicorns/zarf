// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package lint contains functions for verifying zarf yaml files are valid
package lint

import (
	"fmt"
	"path/filepath"

	"github.com/defenseunicorns/pkg/helpers/v2"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/fatih/color"
)

func itemizedDescription(description string, item string) string {
	if item == "" {
		return description
	}
	return fmt.Sprintf("%s - %s", description, item)
}

func colorWrapSev(s types.Severity) string {
	if s == types.SevErr {
		return message.ColorWrap("Error", color.FgRed)
	} else if s == types.SevWarn {
		return message.ColorWrap("Warning", color.FgYellow)
	}
	return "unknown"
}

// PrintFindings prints the findings of the given severity in a table
func PrintFindings(findings []types.PackageFinding, severity types.Severity, baseDir string, packageName string) {
	// TODO add filter sev function
	mapOfFindingsByPath := GroupFindingsByPath(findings, severity, packageName)
	if len(mapOfFindingsByPath) == 0 {
		return
	}

	header := []string{"Type", "Path", "Message"}

	for _, findings := range mapOfFindingsByPath {
		lintData := [][]string{}
		for _, finding := range findings {
			lintData = append(lintData, []string{
				colorWrapSev(finding.Severity),
				message.ColorWrap(finding.YqPath, color.FgCyan),
				itemizedDescription(finding.Description, finding.Item),
			})
		}
		var packagePathFromUser string
		if helpers.IsOCIURL(findings[0].PackagePathOverride) {
			packagePathFromUser = findings[0].PackagePathOverride
		} else {
			packagePathFromUser = filepath.Join(baseDir, findings[0].PackagePathOverride)
		}
		message.Notef("Linting package %q at %s", findings[0].PackageNameOverride, packagePathFromUser)
		message.Table(header, lintData)
	}
}

// GroupFindingsByPath groups findings by their package path
func GroupFindingsByPath(findings []types.PackageFinding, severity types.Severity, packageName string) map[string][]types.PackageFinding {
	findings = helpers.RemoveMatches(findings, func(finding types.PackageFinding) bool {
		return finding.Severity > severity
	})
	for i := range findings {
		if findings[i].PackageNameOverride == "" {
			findings[i].PackageNameOverride = packageName
		}
		if findings[i].PackagePathOverride == "" {
			findings[i].PackagePathOverride = "."
		}
	}

	mapOfFindingsByPath := make(map[string][]types.PackageFinding)
	for _, finding := range findings {
		mapOfFindingsByPath[finding.PackagePathOverride] = append(mapOfFindingsByPath[finding.PackagePathOverride], finding)
	}
	return mapOfFindingsByPath
}

// HasSeverity returns true if the findings contain a severity equal to or greater than the given severity
func HasSeverity(findings []types.PackageFinding, severity types.Severity) bool {
	for _, finding := range findings {
		if finding.Severity <= severity {
			return true
		}
	}
	return false
}
