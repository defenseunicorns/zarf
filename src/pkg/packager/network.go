package packager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/errcode"
)

// handlePackagePath If provided package is a URL download it to a temp directory.
func (p *Packager) handlePackagePath() error {
	message.Debug("packager.handlePackagePath()")

	opts := p.cfg.DeployOpts

	// Check if the user gave us a remote package
	providedURL, err := url.Parse(opts.PackagePath)
	if err != nil || providedURL.Scheme == "" || providedURL.Host == "" {
		message.Debug("Provided package path is not a URL, skipping download")
		return nil
	}

	// Handle case where deploying remote package stored in an OCI registry
	if utils.IsOCIURL(opts.PackagePath) {
		ociURL := opts.PackagePath
		p.cfg.DeployOpts.PackagePath = p.tmp.Base
		return p.handleOciPackage(ociURL, p.tmp.Base)
	}

	// Handle case where deploying remote package validated via sget
	if strings.HasPrefix(opts.PackagePath, "sget://") {
		return p.handleSgetPackage()
	}

	spinner := message.NewProgressSpinner("Loading Zarf Package %s", opts.PackagePath)
	defer spinner.Stop()

	if !config.CommonOptions.Insecure && opts.Shasum == "" {
		return fmt.Errorf("remote package provided without a shasum, use --insecure to ignore")
	}

	// Check the extension on the package is what we expect
	if !isValidFileExtension(providedURL.Path) {
		return fmt.Errorf("remote package provided with an invalid extension, must be one of: %s", config.GetValidPackageExtensions())
	}

	// Download the package
	resp, err := http.Get(opts.PackagePath)
	if err != nil {
		return fmt.Errorf("unable to download remote package: %w", err)
	}
	defer resp.Body.Close()

	localPath := p.tmp.Base + providedURL.Path
	message.Debugf("Creating local package with the path: %s", localPath)
	packageFile, _ := os.Create(localPath)
	_, err = io.Copy(packageFile, resp.Body)
	if err != nil {
		return fmt.Errorf("unable to copy the contents of the provided URL into a local file: %w", err)
	}

	// Check the shasum if necessary
	if !config.CommonOptions.Insecure {
		hasher := sha256.New()
		_, err = io.Copy(hasher, packageFile)
		if err != nil {
			return fmt.Errorf("unable to calculate the sha256 of the provided remote package: %w", err)
		}

		value := hex.EncodeToString(hasher.Sum(nil))
		if value != opts.Shasum {
			_ = os.Remove(localPath)
			return fmt.Errorf("shasum of remote package does not match provided shasum, expected %s, got %s", opts.Shasum, value)
		}
	}

	opts.PackagePath = localPath

	spinner.Success()
	return nil
}

func (p *Packager) handleSgetPackage() error {
	message.Debug("packager.handleSgetPackage()")

	opts := p.cfg.DeployOpts

	spinner := message.NewProgressSpinner("Loading Zarf Package %s", opts.PackagePath)
	defer spinner.Stop()

	// Create the local file for the package
	localPath := filepath.Join(p.tmp.Base, "remote.tar.zst")
	destinationFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("unable to create the destination file: %w", err)
	}
	defer destinationFile.Close()

	// If this is a DefenseUnicorns package, use an internal sget public key
	if strings.HasPrefix(opts.PackagePath, "sget://defenseunicorns") {
		os.Setenv("DU_SGET_KEY", config.SGetPublicKey)
		p.cfg.DeployOpts.SGetKeyPath = "env://DU_SGET_KEY"
	}

	// Remove the 'sget://' header for the actual sget call
	remoteBlob := strings.TrimPrefix(opts.PackagePath, "sget://")

	// Sget the package
	err = utils.Sget(context.TODO(), remoteBlob, p.cfg.DeployOpts.SGetKeyPath, destinationFile)
	if err != nil {
		return fmt.Errorf("unable to get the remote package via sget: %w", err)
	}

	p.cfg.DeployOpts.PackagePath = localPath

	spinner.Success()
	return nil
}

func (p *Packager) handleOciPackage(url string, out string, components ...string) error {
	message.Debugf("packager.handleOciPackage(%s, %s)", url, out)
	ref, err := registry.ParseReference(strings.TrimPrefix(url, "oci://"))
	if err != nil {
		return fmt.Errorf("failed to parse OCI reference: %w", err)
	}
	alwaysPull := []string{
		"zarf.yaml",
		"zarf.yaml.sig",
		"checksums.txt",
	}

	message.Debugf("Pulling %s", ref.String())
	message.Infof("Pulling Zarf package from %s", ref)

	src, err := utils.NewOrasRemote(ref)
	if err != nil {
		return err
	}

	estimatedBytes, err := getOCIPackageSize(src, ref, append(components, alwaysPull...)...)
	if err != nil {
		return err
	}

	dst, err := file.New(out)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Create a thread to update a progress bar as we save the package to disk
	doneSaving := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)
	go utils.RenderProgressBarForLocalDirWrite(out, estimatedBytes, &wg, doneSaving, "Pulling Zarf package data")

	copyOpts := oras.DefaultCopyOptions
	copyOpts.Concurrency = p.cfg.PullOpts.CopyOptions.Concurrency
	copyOpts.PreCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		if utils.SliceContains(alwaysPull, desc.Annotations[ocispec.AnnotationTitle]) {
			return nil
		}
		if len(components) > 0 {
			if !utils.SliceContains(components, desc.Annotations[ocispec.AnnotationTitle]) {
				return errors.New("skip descriptor")
			}
		}
		return nil
	}
	copyOpts.OnCopySkipped = func(ctx context.Context, desc ocispec.Descriptor) error {
		title := desc.Annotations[ocispec.AnnotationTitle]
		var format string
		if title != "" {
			format = fmt.Sprintf("%s %s", desc.Digest.Encoded()[:12], utils.First30last30(title))
		} else {
			format = fmt.Sprintf("%s [%s]", desc.Digest.Encoded()[:12], desc.MediaType)
		}
		message.Successf(format)
		return nil
	}
	copyOpts.PostCopy = copyOpts.OnCopySkipped

	_, err = oras.Copy(src.Context, src.Repository, ref.Reference, dst, ref.Reference, copyOpts)
	if err != nil {
		return err
	}

	// Send a signal to the progress bar that we're done and wait for it to finish
	doneSaving <- 1
	wg.Wait()

	message.Debugf("Pulled %s", ref.String())
	message.Successf("Pulled %s", ref.String())

	return nil
}

// isManifestUnsupported returns true if the error is an unsupported artifact manifest error.
//
// This function was copied verbatim from https://github.com/oras-project/oras/blob/main/cmd/oras/push.go
func isManifestUnsupported(err error) bool {
	var errResp *errcode.ErrorResponse
	if !errors.As(err, &errResp) || errResp.StatusCode != http.StatusBadRequest {
		return false
	}

	var errCode errcode.Error
	if !errors.As(errResp, &errCode) {
		return false
	}

	// As of November 2022, ECR is known to return UNSUPPORTED error when
	// pulling an OCI artifact manifest.
	switch errCode.Code {
	case errcode.ErrorCodeManifestInvalid, errcode.ErrorCodeUnsupported:
		return true
	}
	return false
}

func getOCIPackageSize(src *utils.OrasRemote, ref registry.Reference, layersToPull ...string) (int64, error) {
	var total int64

	layers, err := getLayers(src)
	if err != nil {
		return 0, err
	}

	for _, layer := range layersToPull {
		if !utils.SliceContains(layersToPull, layer) {
			return 0, fmt.Errorf("layer %s not found in package", layer)
		}
	}

	processedLayers := make(map[string]bool)
	for _, layer := range layers {
		// Only include this layer's size if we haven't already processed it
		if !processedLayers[layer.Digest.String()] && (len(layersToPull) == 0 || utils.SliceContains(layersToPull, layer.Annotations[ocispec.AnnotationTitle])) {
			total += layer.Size
			processedLayers[layer.Digest.String()] = true
		}
	}

	return total, nil
}

// getLayers returns the manifest layers of a Zarf OCI package
func getLayers(dst *utils.OrasRemote) ([]ocispec.Descriptor, error) {
	// get the manifest descriptor
	// ref.Reference can be a tag or a digest
	descriptor, err := dst.Resolve(dst.Context, dst.Reference.Reference)
	if err != nil {
		return nil, err
	}

	// get the manifest itself
	pulled, err := content.FetchAll(dst.Context, dst, descriptor)
	if err != nil {
		return nil, err
	}
	manifest := ocispec.Manifest{}
	artifact := ocispec.Artifact{}
	var layers []ocispec.Descriptor
	// if the manifest is an artifact, unmarshal it as an artifact
	// otherwise, unmarshal it as a manifest
	if descriptor.MediaType == ocispec.MediaTypeArtifactManifest {
		if err = json.Unmarshal(pulled, &artifact); err != nil {
			return nil, err
		}
		layers = artifact.Blobs
	} else {
		if err = json.Unmarshal(pulled, &manifest); err != nil {
			return nil, err
		}
		layers = manifest.Layers
	}

	return layers, nil
}

// pullLayer fetches a single layer from a Zarf OCI package
func pullLayer(dst *utils.OrasRemote, desc ocispec.Descriptor, out string) error {
	bytes, err := content.FetchAll(dst.Context, dst, desc)
	if err != nil {
		return err
	}
	err = utils.WriteFile(out, bytes)
	return err
}
