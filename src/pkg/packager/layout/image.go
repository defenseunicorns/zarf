// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

package layout

import (
	"path/filepath"

	"slices"
)

type Images struct {
	Base      string
	Index     string
	OCILayout string
	Blobs     []string
}

func (i *Images) AddBlob(blob string) {
	// TODO: verify sha256 hex
	abs := filepath.Join(i.Base, "blobs", "sha256", blob)
	if !slices.Contains(i.Blobs, abs) {
		i.Blobs = append(i.Blobs, abs)
	}
}
