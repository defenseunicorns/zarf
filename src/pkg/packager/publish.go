// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/mholt/archiver/v3"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
)

// ZarfLayerMediaType<Extension> is the media type for Zarf layers.
const (
	ZarfLayerMediaTypeTarZstd = "application/vnd.zarf.layer.v1.tar+zstd"
	ZarfLayerMediaTypeTarGzip = "application/vnd.zarf.layer.v1.tar+gzip"
	ZarfLayerMediaTypeYaml    = "application/vnd.zarf.layer.v1.yaml"
	ZarfLayerMediaTypeJSON    = "application/vnd.zarf.layer.v1.json"
	ZarfLayerMediaTypeTxt     = "application/vnd.zarf.layer.v1.txt"
	ZarfLayerMediaTypeUnknown = "application/vnd.zarf.layer.v1.unknown"
)

// parseZarfLayerMediaType returns the Zarf layer media type for the given filename.
func parseZarfLayerMediaType(filename string) string {
	// since we are controlling the filenames, we can just use the extension
	switch filepath.Ext(filename) {
	case ".zst":
		return ZarfLayerMediaTypeTarZstd
	case ".gz":
		return ZarfLayerMediaTypeTarGzip
	case ".yaml":
		return ZarfLayerMediaTypeYaml
	case ".json":
		return ZarfLayerMediaTypeJSON
	case ".txt":
		return ZarfLayerMediaTypeTxt
	default:
		return ZarfLayerMediaTypeUnknown
	}
}

// Publish publishes the package to a registry
//
// This is a wrapper around the oras library
// and much of the code was adapted from the oras CLI - https://github.com/oras-project/oras/blob/main/cmd/oras/push.go
//
// Authentication is handled via the Docker config file created w/ `zarf tools registry login`
func (p *Packager) Publish() error {
	p.cfg.DeployOpts.PackagePath = p.cfg.PublishOpts.PackagePath

	// Extract the first layer of the tarball
	if err := archiver.Unarchive(p.cfg.DeployOpts.PackagePath, p.tmp.Base); err != nil {
		return fmt.Errorf("unable to extract the package: %w", err)
	}

	if err := p.readYaml(p.tmp.ZarfYaml, true); err != nil {
		return fmt.Errorf("unable to read the zarf.yaml in %s: %w", p.tmp.Base, err)
	}

	if err := p.validatePackageChecksums(); err != nil {
		return fmt.Errorf("unable to publish package because checksums do not match: %w", err)
	}

	if p.cfg.PublishOpts.SigningKeyPath != "" {
		_, err := utils.CosignSignBlob(p.tmp.ZarfYaml, p.tmp.ZarfSig, p.cfg.PublishOpts.SigningKeyPath)
		if err != nil {
			return fmt.Errorf("unable to sign the package: %w", err)
		}
	}
	paths, err := filepath.Glob(filepath.Join(p.tmp.Base, "*"))
	if err != nil {
		return fmt.Errorf("unable to glob the package: %w", err)
	}

	ref, err := p.ref("")
	if err != nil {
		return err
	}

	message.HeaderInfof("📦 PACKAGE PUBLISH %s:%s", p.cfg.Pkg.Metadata.Name, ref.Reference)
	return p.publish(ref, paths)
}

func (p *Packager) publish(ref registry.Reference, paths []string) error {
	message.Infof("Publishing package to %s", ref)
	spinner := message.NewProgressSpinner("")
	defer spinner.Stop()

	// destination remote
	dst, err := utils.NewOrasRemote(ref)
	if err != nil {
		return err
	}
	ctx := dst.Context

	// source file store
	src, err := file.New(p.tmp.Base)
	if err != nil {
		return err
	}
	defer src.Close()

	var descs []ocispec.Descriptor

	for idx, path := range paths {
		name, err := filepath.Rel(p.tmp.Base, path)
		if err != nil {
			return err
		}
		spinner.Updatef("Preparing layer %d/%d: %s", idx+1, len(paths), name)

		mediaType := parseZarfLayerMediaType(name)

		desc, err := src.Add(ctx, name, mediaType, path)
		if err != nil {
			return err
		}
		descs = append(descs, desc)
	}
	spinner.Successf("Prepared %d layers", len(descs))

	copyOpts := oras.DefaultCopyOptions
	copyOpts.Concurrency = p.cfg.PublishOpts.CopyOptions.Concurrency
	copyOpts.OnCopySkipped = utils.PrintLayerExists
	copyOpts.PostCopy = utils.PrintLayerExists

	var root ocispec.Descriptor

	// try to push an ArtifactManifest first
	// not every registry supports ArtifactManifests, so fallback to an ImageManifest if the push fails
	// see https://oras.land/implementors/#registries-supporting-oci-artifacts
	root, err = p.publishArtifact(dst, src, descs, copyOpts)
	if err != nil {
		// reset the progress bar between attempts
		dst.Transport.ProgressBar.Stop()

		// log the error, the expected error is a 400 manifest invalid
		message.Debug("ArtifactManifest push failed with the following error, falling back to an ImageManifest push:", err)

		// if the error returned from the push is not an expected error, then return the error
		if !isManifestUnsupported(err) {
			return err
		}
		// fallback to an ImageManifest push
		root, err = p.publishImage(dst, src, descs, copyOpts)
		if err != nil {
			return err
		}
	}
	dst.Transport.ProgressBar.Successf("Published %s [%s]", ref, root.MediaType)
	fmt.Println()
	flags := ""
	if config.CommonOptions.Insecure {
		flags = "--insecure"
	}
	message.Info("To inspect/deploy/pull:")
	message.Infof("zarf package inspect oci://%s %s", ref, flags)
	message.Infof("zarf package deploy oci://%s %s", ref, flags)
	message.Infof("zarf package pull oci://%s %s", ref, flags)

	return nil
}

func (p *Packager) publishArtifact(dst *utils.OrasRemote, src *file.Store, descs []ocispec.Descriptor, copyOpts oras.CopyOptions) (root ocispec.Descriptor, err error) {
	var total int64
	for _, desc := range descs {
		total += desc.Size
	}
	packOpts := p.cfg.PublishOpts.PackOptions

	// first attempt to do a ArtifactManifest push
	root, err = p.pack(dst.Context, ocispec.MediaTypeArtifactManifest, descs, src, packOpts)
	if err != nil {
		return root, err
	}
	total += root.Size

	dst.Transport.ProgressBar = message.NewProgressBar(total, fmt.Sprintf("Publishing %s:%s", dst.Reference.Repository, dst.Reference.Reference))
	defer dst.Transport.ProgressBar.Stop()

	// attempt to push the artifact manifest
	_, err = oras.Copy(dst.Context, src, root.Digest.String(), dst, dst.Reference.Reference, copyOpts)
	return root, err
}

func (p *Packager) publishImage(dst *utils.OrasRemote, src *file.Store, descs []ocispec.Descriptor, copyOpts oras.CopyOptions) (root ocispec.Descriptor, err error) {
	var total int64
	for _, desc := range descs {
		total += desc.Size
	}
	// assumes referrers API is not supported since OCI artifact
	// media type is not supported
	dst.SetReferrersCapability(false)

	// fallback to an ImageManifest push
	manifestConfigDesc, manifestConfigContent, err := p.generateManifestConfigFile()
	if err != nil {
		return root, err
	}
	// push the manifest config
	// since this config is so tiny, and the content is not used again
	// it is not logged to the progress, but will error if it fails
	err = dst.Push(dst.Context, manifestConfigDesc, bytes.NewReader(manifestConfigContent))
	if err != nil {
		return root, err
	}
	packOpts := p.cfg.PublishOpts.PackOptions
	packOpts.ConfigDescriptor = &manifestConfigDesc
	packOpts.PackImageManifest = true
	root, err = p.pack(dst.Context, ocispec.MediaTypeImageManifest, descs, src, packOpts)
	if err != nil {
		return root, err
	}
	total += root.Size + manifestConfigDesc.Size

	dst.Transport.ProgressBar = message.NewProgressBar(total, fmt.Sprintf("Publishing %s:%s", dst.Reference.Repository, dst.Reference.Reference))
	defer dst.Transport.ProgressBar.Stop()
	// attempt to push the image manifest
	_, err = oras.Copy(dst.Context, src, root.Digest.String(), dst, dst.Reference.Reference, copyOpts)
	if err != nil {
		return root, err
	}

	return root, nil
}

func (p *Packager) generateAnnotations(artifactType string) map[string]string {
	annotations := map[string]string{
		ocispec.AnnotationDescription: p.cfg.Pkg.Metadata.Description,
	}

	if artifactType == ocispec.MediaTypeArtifactManifest {
		annotations[ocispec.AnnotationTitle] = p.cfg.Pkg.Metadata.Name
	}

	if url := p.cfg.Pkg.Metadata.URL; url != "" {
		annotations[ocispec.AnnotationURL] = url
	}
	if authors := p.cfg.Pkg.Metadata.Authors; authors != "" {
		annotations[ocispec.AnnotationAuthors] = authors
	}
	if documentation := p.cfg.Pkg.Metadata.Documentation; documentation != "" {
		annotations[ocispec.AnnotationDocumentation] = documentation
	}
	if source := p.cfg.Pkg.Metadata.Source; source != "" {
		annotations[ocispec.AnnotationSource] = source
	}
	if vendor := p.cfg.Pkg.Metadata.Vendor; vendor != "" {
		annotations[ocispec.AnnotationVendor] = vendor
	}

	return annotations
}

func (p *Packager) generateManifestConfigFile() (ocispec.Descriptor, []byte, error) {
	// Unless specified, an empty manifest config will be used: `{}`
	// which causes an error on Google Artifact Registry
	// to negate this, we create a simple manifest config with some build metadata
	// the contents of this file are not used by Zarf
	type OCIConfigPartial struct {
		Architecture string            `json:"architecture"`
		OCIVersion   string            `json:"ociVersion"`
		Annotations  map[string]string `json:"annotations,omitempty"`
	}

	annotations := map[string]string{
		ocispec.AnnotationTitle:       p.cfg.Pkg.Metadata.Name,
		ocispec.AnnotationDescription: p.cfg.Pkg.Metadata.Description,
	}

	manifestConfig := OCIConfigPartial{
		Architecture: p.cfg.Pkg.Build.Architecture,
		OCIVersion:   "1.0.1",
		Annotations:  annotations,
	}
	manifestConfigBytes, err := json.Marshal(manifestConfig)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}
	manifestConfigDesc := content.NewDescriptorFromBytes("application/vnd.unknown.config.v1+json", manifestConfigBytes)

	return manifestConfigDesc, manifestConfigBytes, nil
}

// pack creates an artifact/image manifest from the provided descriptors and pushes it to the store
func (p *Packager) pack(ctx context.Context, artifactType string, descs []ocispec.Descriptor, src *file.Store, packOpts oras.PackOptions) (ocispec.Descriptor, error) {
	packOpts.ManifestAnnotations = p.generateAnnotations(artifactType)
	root, err := oras.Pack(ctx, src, artifactType, descs, packOpts)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	if err = src.Tag(ctx, root, root.Digest.String()); err != nil {
		return ocispec.Descriptor{}, err
	}

	return root, nil
}

// ref returns a registry.Reference using metadata from the package's build config and the PublishOpts
//
// if skeleton is not empty, the architecture will be replaced with the skeleton string (e.g. "skeleton")
func (p *Packager) ref(skeleton string) (registry.Reference, error) {
	ver := p.cfg.Pkg.Metadata.Version
	if len(ver) == 0 {
		return registry.Reference{}, errors.New("version is required for publishing")
	}
	arch := p.cfg.Pkg.Build.Architecture
	// changes package ref from "name:version-arch" to "name:version-skeleton"
	if len(skeleton) > 0 {
		arch = skeleton
	}
	ref := registry.Reference{
		Registry:   p.cfg.PublishOpts.Reference.Registry,
		Repository: fmt.Sprintf("%s/%s", p.cfg.PublishOpts.Reference.Repository, p.cfg.Pkg.Metadata.Name),
		Reference:  fmt.Sprintf("%s-%s", ver, arch),
	}
	if len(p.cfg.PublishOpts.Reference.Repository) == 0 {
		ref.Repository = p.cfg.Pkg.Metadata.Name
	}
	err := ref.Validate()
	if err != nil {
		return registry.Reference{}, err
	}
	return ref, nil
}
