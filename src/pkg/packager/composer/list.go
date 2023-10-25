// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package composer contains functions for composing components within Zarf packages.
package composer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/internal/packager/validate"
	"github.com/defenseunicorns/zarf/src/pkg/layout"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/oci"
	"github.com/defenseunicorns/zarf/src/pkg/packager/deprecated"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/mholt/archiver/v3"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	ocistore "oras.land/oras-go/v2/content/oci"
)

// Node is a node in the import chain
type Node struct {
	types.ZarfComponent

	vars   []types.ZarfPackageVariable
	consts []types.ZarfPackageConstant

	relativeToHead string

	prev *Node
	next *Node
}

// ImportName returns the name of the component to import
//
// If the component import has a ComponentName defined, that will be used
// otherwise the name of the component will be used
func (n *Node) ImportName() string {
	name := n.ZarfComponent.Name
	if n.Import.ComponentName != "" {
		name = n.Import.ComponentName
	}
	return name
}

// ImportChain is a doubly linked list of component import definitions
type ImportChain struct {
	head *Node
	tail *Node

	remote *oci.OrasRemote
}

func (ic *ImportChain) append(c types.ZarfComponent, relativeToHead string, vars []types.ZarfPackageVariable, consts []types.ZarfPackageConstant) {
	node := &Node{
		ZarfComponent:  c,
		relativeToHead: relativeToHead,
		vars:           vars,
		consts:         consts,
		prev:           nil,
		next:           nil,
	}
	if ic.head == nil {
		ic.head = node
		ic.tail = node
	} else {
		p := ic.head
		for p.next != nil {
			p = p.next
		}
		node.prev = p

		p.next = node
		ic.tail = node
	}
}

// NewImportChain creates a new import chain from a component
func NewImportChain(head types.ZarfComponent, arch string) (*ImportChain, error) {
	if arch == "" {
		return nil, fmt.Errorf("cannot build import chain: architecture must be provided")
	}

	ic := &ImportChain{}

	ic.append(head, ".", nil, nil)

	history := []string{}

	node := ic.head
	for node != nil {
		isLocal := node.Import.Path != ""
		isRemote := node.Import.URL != ""

		if !isLocal && !isRemote {
			// This is the end of the import chain,
			// as the current node/component is not importing anything
			return ic, nil
		}

		// TODO: stuff like this should also happen in linting
		if err := validate.ImportDefinition(&node.ZarfComponent); err != nil {
			return ic, err
		}

		// ensure that remote components are not importing other remote components
		if node.prev != nil && node.prev.Import.URL != "" && isRemote {
			return ic, fmt.Errorf("detected malformed import chain, cannot import remote components from remote components")
		}
		// ensure that remote components are not importing local components
		if node.prev != nil && node.prev.Import.URL != "" && isLocal {
			return ic, fmt.Errorf("detected malformed import chain, cannot import local components from remote components")
		}

		var pkg types.ZarfPackage

		if isLocal {
			history = append(history, node.Import.Path)
			relativeToHead := filepath.Join(history...)

			// prevent circular imports (including self-imports)
			// this is O(n^2) but the import chain should be small
			prev := node.prev
			for prev != nil {
				if prev.relativeToHead == relativeToHead {
					return ic, fmt.Errorf("detected circular import chain: %s", strings.Join(history, " -> "))
				}
				prev = prev.prev
			}

			// this assumes the composed package is following the zarf layout
			if err := utils.ReadYaml(filepath.Join(relativeToHead, layout.ZarfYAML), &pkg); err != nil {
				return ic, err
			}
		} else if isRemote {
			remote, err := ic.getRemote(node.Import.URL)
			if err != nil {
				return ic, err
			}
			pkg, err = remote.FetchZarfYAML()
			if err != nil {
				return ic, err
			}
		}

		name := node.ImportName()

		found := helpers.Filter(pkg.Components, func(c types.ZarfComponent) bool {
			matchesName := c.Name == name
			satisfiesArch := c.Only.Cluster.Architecture == "" || c.Only.Cluster.Architecture == arch
			return matchesName && satisfiesArch
		})

		if len(found) == 0 {
			if isLocal {
				return ic, fmt.Errorf("component %q not found in %q", name, filepath.Join(history...))
			} else if isRemote {
				return ic, fmt.Errorf("component %q not found in %q", name, node.Import.URL)
			}
		} else if len(found) > 1 {
			if isLocal {
				return ic, fmt.Errorf("multiple components named %q found in %q satisfying %q", name, filepath.Join(history...), arch)
			} else if isRemote {
				return ic, fmt.Errorf("multiple components named %q found in %q satisfying %q", name, node.Import.URL, arch)
			}
		}

		ic.append(found[0], filepath.Join(history...), pkg.Variables, pkg.Constants)
		node = node.next
	}
	return ic, nil
}

// String returns a string representation of the import chain
func (ic *ImportChain) String() string {
	if ic.head.next == nil {
		return fmt.Sprintf("component %q imports nothing", ic.head.Name)
	}

	s := strings.Builder{}

	name := ic.head.ImportName()

	if ic.head.Import.Path != "" {
		s.WriteString(fmt.Sprintf("component %q imports %q in %s", ic.head.Name, name, ic.head.Import.Path))
	} else {
		s.WriteString(fmt.Sprintf("component %q imports %q in %s", ic.head.Name, name, ic.head.Import.URL))
	}

	node := ic.head.next
	for node != ic.tail {
		name := node.ImportName()
		s.WriteString(", which imports ")
		if node.Import.Path != "" {
			s.WriteString(fmt.Sprintf("%q in %s", name, node.Import.Path))
		} else {
			s.WriteString(fmt.Sprintf("%q in %s", name, node.Import.URL))
		}

		node = node.next
	}

	return s.String()
}

// Migrate performs migrations on the import chain
func (ic *ImportChain) Migrate(build types.ZarfBuildData) (warnings []string) {
	node := ic.head
	for node != nil {
		migrated, w := deprecated.MigrateComponent(build, node.ZarfComponent)
		node.ZarfComponent = migrated
		warnings = append(warnings, w...)
		node = node.next
	}
	if len(warnings) > 0 {
		final := fmt.Sprintf("migrations were performed on the import chain of: %q", ic.head.Name)
		warnings = append(warnings, final)
	}
	return warnings
}

func (ic *ImportChain) getRemote(url string) (*oci.OrasRemote, error) {
	if ic.remote != nil {
		return ic.remote, nil
	}
	var err error
	ic.remote, err = oci.NewOrasRemote(url)
	if err != nil {
		return nil, err
	}
	return ic.remote, nil
}

// ContainsOCIImport returns true if the import chain contains a remote import
func (ic *ImportChain) ContainsOCIImport() bool {
	// only the 2nd to last node may have a remote import
	return ic.tail.prev != nil && ic.tail.prev.Import.URL != ""
}

// OCIImportDefinition returns the url and name of the remote import
func (ic *ImportChain) OCIImportDefinition() (string, string) {
	if !ic.ContainsOCIImport() {
		return "", ""
	}
	return ic.tail.prev.Import.URL, ic.tail.prev.ImportName()
}

func (ic *ImportChain) fetchOCISkeleton() error {
	if !ic.ContainsOCIImport() {
		return nil
	}
	node := ic.tail.prev
	remote, err := ic.getRemote(node.Import.URL)
	if err != nil {
		return err
	}

	manifest, err := remote.FetchRoot()
	if err != nil {
		return err
	}

	name := node.ImportName()

	componentDesc := manifest.Locate(filepath.Join(layout.ComponentsDir, fmt.Sprintf("%s.tar", name)))

	cache := filepath.Join(config.GetAbsCachePath(), "oci")
	if err := utils.CreateDirectory(cache, 0700); err != nil {
		return err
	}

	var tb, dir string

	// if there is not a tarball to fetch, create a directory named based upon
	// the import url and the component name
	if oci.IsEmptyDescriptor(componentDesc) {
		h := sha256.New()
		h.Write([]byte(node.Import.URL + name))
		id := fmt.Sprintf("%x", h.Sum(nil))

		dir = filepath.Join(cache, "dirs", id)

		message.Debug("creating empty directory for remote component:", filepath.Join("<zarf-cache>", "oci", "dirs", id))
	} else {
		tb = filepath.Join(cache, "blobs", "sha256", componentDesc.Digest.Encoded())
		dir = filepath.Join(cache, "dirs", componentDesc.Digest.Encoded())

		store, err := ocistore.New(cache)
		if err != nil {
			return err
		}

		ctx := context.TODO()
		// ensure the tarball is in the cache
		exists, err := store.Exists(ctx, componentDesc)
		if err != nil {
			return err
		} else if !exists {
			copyOpts := remote.CopyOpts
			// TODO: investigate why the default FindSuccessors in CopyWithProgress is not working
			copyOpts.FindSuccessors = content.Successors
			if err := remote.CopyWithProgress([]ocispec.Descriptor{componentDesc}, store, copyOpts, cache); err != nil {
				return err
			}
			exists, err := store.Exists(ctx, componentDesc)
			if err != nil {
				return err
			} else if !exists {
				return fmt.Errorf("failed to fetch remote component: %+v", componentDesc)
			}
		}
	}

	if err := utils.CreateDirectory(dir, 0700); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(cwd, dir)
	if err != nil {
		return err
	}
	// the tail node is the only node whose relativeToHead is based solely upon cwd<->cache
	// contrary to the other nodes, which are based upon the previous node
	ic.tail.relativeToHead = rel

	if oci.IsEmptyDescriptor(componentDesc) {
		// nothing was fetched, nothing to extract
		return nil
	}

	tu := archiver.Tar{
		OverwriteExisting: true,
		// removes /<name>/ from the paths
		StripComponents: 1,
	}
	return tu.Unarchive(tb, dir)
}

// Compose merges the import chain into a single component
// fixing paths, overriding metadata, etc
func (ic *ImportChain) Compose() (composed types.ZarfComponent, err error) {
	composed = ic.tail.ZarfComponent

	if ic.tail.prev == nil {
		// only had one component in the import chain
		return composed, nil
	}

	if err := ic.fetchOCISkeleton(); err != nil {
		return composed, err
	}

	// start with an empty component to compose into
	composed = types.ZarfComponent{}

	// start overriding with the tail node
	node := ic.tail
	for node != nil {
		fixPaths(&node.ZarfComponent, node.relativeToHead)

		// perform overrides here
		overrideMetadata(&composed, node.ZarfComponent)
		overrideDeprecated(&composed, node.ZarfComponent)
		overrideResources(&composed, node.ZarfComponent)
		overrideActions(&composed, node.ZarfComponent)

		composeExtensions(&composed, node.ZarfComponent, node.relativeToHead)

		node = node.prev
	}

		node = node.prev
	}

	return composed, nil
}

// MergeVariables merges variables from the import chain
func (ic *ImportChain) MergeVariables(existing []types.ZarfPackageVariable) (merged []types.ZarfPackageVariable) {
	exists := func(v1 types.ZarfPackageVariable, v2 types.ZarfPackageVariable) bool {
		return v1.Name == v2.Name
	}

	merged = helpers.MergeSlices(existing, merged, exists)
	node := ic.head
	for node != nil {
		// merge the vars
		merged = helpers.MergeSlices(node.vars, merged, exists)
		node = node.next
	}
	return merged
}

// MergeConstants merges constants from the import chain
func (ic *ImportChain) MergeConstants(existing []types.ZarfPackageConstant) (merged []types.ZarfPackageConstant) {
	exists := func(c1 types.ZarfPackageConstant, c2 types.ZarfPackageConstant) bool {
		return c1.Name == c2.Name
	}

	merged = helpers.MergeSlices(existing, merged, exists)
	node := ic.head
	for node != nil {
		// merge the consts
		merged = helpers.MergeSlices(node.consts, merged, exists)
		node = node.next
	}
	return merged
}
