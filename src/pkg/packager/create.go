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
	"github.com/defenseunicorns/zarf/src/types"
)

// Creator is an interface for creating Zarf packages.
type Creator interface {
	Loader
	Assembler
}

// NewCreator returns a new Creator based on the provided create options.
func NewCreator(createOpts *types.ZarfCreateOptions) Creator {
	if createOpts.IsSkeleton {
		return &SkeletonCreator{}
	}
	return &PackageCreator{}
}

// SkeletonAssembler is used to create skeleton Zarf packages.
type SkeletonCreator struct {
	*SkeletonLoader
	*SkeletonAssembler
}

// PackageCreator is used to create normal (not skeleton) Zarf packages.
type PackageCreator struct {
	*PackageLoader
	*PackageAssembler
}

// Create generates a Zarf package tarball for a given PackageConfig and optional base directory.
func (p *Packager) Create() (err error) {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := cdToBaseDir(&p.cfg.CreateOpts, cwd); err != nil {
		return err
	}

	c := NewCreator(&p.cfg.CreateOpts)

	if err := c.LoadPackageDefinition(p); err != nil {
		return err
	}

	// Perform early package validation.
	if err := validate.Run(p.cfg.Pkg); err != nil {
		return fmt.Errorf("unable to validate package: %w", err)
	}

	if !p.confirmAction(config.ZarfCreateStage) {
		return fmt.Errorf("package creation canceled")
	}

	if err := c.Assemble(p); err != nil {
		return err
	}

	// cd back for output
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	return p.output()
}

func cdToBaseDir(createOpts *types.ZarfCreateOptions, cwd string) error {
	if err := os.Chdir(createOpts.BaseDir); err != nil {
		return fmt.Errorf("unable to access directory %q: %w", createOpts.BaseDir, err)
	}
	message.Note(fmt.Sprintf("Using build directory %s", createOpts.BaseDir))

	// differentials are relative to the current working directory
	if createOpts.DifferentialData.DifferentialPackagePath != "" {
		createOpts.DifferentialData.DifferentialPackagePath = filepath.Join(cwd, createOpts.DifferentialData.DifferentialPackagePath)
	}
	return nil
}
