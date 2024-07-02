// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package lint contains functions for verifying zarf yaml files are valid
package lint

import (
	"context"
	"testing"

	"github.com/defenseunicorns/zarf/src/types"
	"github.com/stretchr/testify/require"
)

func TestLintComponents(t *testing.T) {
	t.Run("Test composable components with bad path", func(t *testing.T) {
		t.Parallel()
		zarfPackage := types.ZarfPackage{
			Components: []types.ZarfComponent{
				{
					Import: types.ZarfComponentImport{Path: "bad-path"},
				},
			},
			Metadata: types.ZarfMetadata{Name: "test-zarf-package"},
		}

		createOpts := types.ZarfCreateOptions{Flavor: "", BaseDir: "."}
		_, err := lintComponents(context.Background(), zarfPackage, createOpts)
		require.Error(t, err)
	})
}
