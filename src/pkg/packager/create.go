// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/internal/packager/validate"
	"github.com/defenseunicorns/zarf/src/pkg/message"
)

func (p *Packager) cdToBaseDir(base string, cwd string) error {
	if err := os.Chdir(base); err != nil {
		return fmt.Errorf("unable to access directory %q: %w", base, err)
	}
	message.Note(fmt.Sprintf("Using build directory %s", base))

	// differentials are relative to the current working directory
	if p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath != "" {
		p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath = filepath.Join(cwd, p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath)
	}
	return nil
}

// Create generates a Zarf package tarball for a given PackageConfig and optional base directory.
func (p *Packager) Create() (err error) {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := p.cdToBaseDir(p.cfg.CreateOpts.BaseDir, cwd); err != nil {
		return err
	}

	if err := p.load(); err != nil {
		return err
	}

	// Perform early package validation.
	if err := validate.Run(p.cfg.Pkg); err != nil {
		return fmt.Errorf("unable to validate package: %w", err)
	}

	if !p.confirmAction(config.ZarfCreateStage) {
		return fmt.Errorf("package creation canceled")
	}

	if err := p.assemble(); err != nil {
		return err
	}

	// cd back
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	return p.output()
}
