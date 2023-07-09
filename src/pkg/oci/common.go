// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package oci contains functions for interacting with Zarf packages stored in OCI registries.
package oci

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	zarfconfig "github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	// ZarfLayerMediaTypeBlob is the media type for all Zarf layers due to the range of possible content
	ZarfLayerMediaTypeBlob = "application/vnd.zarf.layer.v1.blob"
	// SkeletonSuffix is the reference suffix used for skeleton packages
	SkeletonSuffix = "skeleton"
)

// OrasRemote is a wrapper around the Oras remote repository that includes a progress bar for interactive feedback.
type OrasRemote struct {
	repo           *remote.Repository
	ctx            context.Context
	Transport      *utils.Transport
	CopyOpts       oras.CopyOptions
	client         *auth.Client
	hasCredentials bool
}

// NewOrasRemote returns an oras remote repository client and context for the given url.
//
// Registry auth is handled by the Docker CLI's credential store and checked before returning the client
func NewOrasRemote(url string) (*OrasRemote, error) {
	ref, err := registry.ParseReference(strings.TrimPrefix(url, utils.OCIURLPrefix))
	if err != nil {
		return nil, fmt.Errorf("failed to parse OCI reference: %w", err)
	}
	o := &OrasRemote{}
	o.ctx = context.TODO()

	err = o.WithRepository(ref)
	if err != nil {
		return nil, err
	}

	copyOpts := oras.DefaultCopyOptions
	copyOpts.OnCopySkipped = o.printLayerSuccess
	copyOpts.PostCopy = o.printLayerSuccess
	o.CopyOpts = copyOpts

	return o, nil
}

// WithRepository sets the repository for the remote as well as the auth client.
func (o *OrasRemote) WithRepository(ref registry.Reference) error {
	o.hasCredentials = false
	// patch docker.io to registry-1.docker.io
	// this allows end users to use docker.io as an alias for registry-1.docker.io
	if ref.Registry == "docker.io" {
		ref.Registry = "registry-1.docker.io"
	}
	client, err := o.withAuthClient(ref)
	if err != nil {
		return err
	}
	o.client = client

	repo, err := remote.NewRepository(ref.String())
	if err != nil {
		return err
	}
	repo.PlainHTTP = zarfconfig.CommonOptions.Insecure
	repo.Client = o.client
	o.repo = repo
	return nil
}

// withAuthClient returns an auth client for the given reference.
//
// The credentials are pulled using Docker's default credential store.
func (o *OrasRemote) withAuthClient(ref registry.Reference) (*auth.Client, error) {
	message.Debugf("Loading docker config file from default config location: %s", config.Dir())
	cfg, err := config.Load(config.Dir())
	if err != nil {
		return nil, err
	}
	if !cfg.ContainsAuth() {
		message.Debug("no docker config file found, run 'zarf tools registry login --help'")
		return nil, nil
	}

	configs := []*configfile.ConfigFile{cfg}

	var key = ref.Registry
	if key == "registry-1.docker.io" {
		// Docker stores its credentials under the following key, otherwise credentials use the registry URL
		key = "https://index.docker.io/v1/"
	}

	authConf, err := configs[0].GetCredentialsStore(key).Get(key)
	if err != nil {
		return nil, fmt.Errorf("unable to get credentials for %s: %w", key, err)
	}

	if authConf.ServerAddress != "" {
		o.hasCredentials = true
	}

	cred := auth.Credential{
		Username:     authConf.Username,
		Password:     authConf.Password,
		AccessToken:  authConf.RegistryToken,
		RefreshToken: authConf.IdentityToken,
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = zarfconfig.CommonOptions.Insecure

	o.Transport = utils.NewTransport(transport, nil)

	client := &auth.Client{
		Credential: auth.StaticCredential(ref.Registry, cred),
		Cache:      auth.NewCache(),
		Client: &http.Client{
			Transport: o.Transport,
		},
	}
	client.SetUserAgent("zarf/" + zarfconfig.CLIVersion)

	return client, nil
}

// CheckAuth checks if the user is authenticated to the remote registry.
func (o *OrasRemote) CheckAuth(scopes ...string) error {
	if scopes == nil && o.hasCredentials {
		return fmt.Errorf("%s requires authentication but no request scopes were provided", o.repo.Reference)
	}
	// check if we've already checked the scopes
	currentScopes := auth.GetScopes(o.ctx)
	equal := reflect.DeepEqual(currentScopes, scopes)
	// if we've already checked the scopes return
	if equal {
		return nil
	}

	// if we have credentials, add the scopes to the context
	if o.hasCredentials {
		o.ctx = auth.WithScopes(o.ctx, scopes...)
	}
	reg, err := remote.NewRegistry(o.repo.Reference.Registry)
	if err != nil {
		return err
	}
	reg.PlainHTTP = zarfconfig.CommonOptions.Insecure
	reg.Client = o.client
	err = reg.Ping(o.ctx)
	if err != nil {
		return fmt.Errorf("unable to authenticate to %s: %s", o.repo.Reference.Registry, err.Error())
	}
	return nil
}
