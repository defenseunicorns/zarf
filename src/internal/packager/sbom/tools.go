// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package sbom contains tools for generating SBOMs
package sbom

import (
	"os"
	"path/filepath"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
)

// WriteSBOMFiles writes the SBOM viewer files to the config.ZarfSBOMDir
func WriteSBOMFiles(sbomViewFiles []string) error {
	// Check if we even have any SBOM files to process
	if len(sbomViewFiles) == 0 {
		return nil
	}

	// Cleanup any failed prior removals
	_ = os.RemoveAll(config.ZarfSBOMDir)

	// Create the directory again
	err := utils.CreateDirectory(config.ZarfSBOMDir, 0755)
	if err != nil {
		return err
	}

	// Write each of the sbom files
	for _, file := range sbomViewFiles {
		// Our file copy lib explodes on these files for some reason...
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		dst := filepath.Join(config.ZarfSBOMDir, filepath.Base(file))
		err = os.WriteFile(dst, data, 0644)
		if err != nil {
			message.Debugf("Unable to write the sbom-viewer file %s", dst)
			return err
		}
	}

	return nil
}
