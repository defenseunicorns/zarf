// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package rules verifies that Zarf packages are following best practices.
package rules

import (
	"errors"
	"testing"

	"github.com/defenseunicorns/zarf/src/types"
	"github.com/stretchr/testify/require"
)

func TestValidateComponent(t *testing.T) {
	t.Parallel()

	t.Run("Unpinnned repo warning", func(t *testing.T) {
		t.Parallel()
		unpinnedRepo := "https://github.com/defenseunicorns/zarf-public-test.git"
		component := types.ZarfComponent{Repos: []string{
			unpinnedRepo,
			"https://dev.azure.com/defenseunicorns/zarf-public-test/_git/zarf-public-test@v0.0.1",
		}}
		findings := checkForUnpinnedRepos(component, 0)
		expected := []types.PackageFinding{
			{
				Item:        unpinnedRepo,
				Description: "Unpinned repository",
				Severity:    types.SevWarn,
				YqPath:      ".components.[0].repos.[0]",
			},
		}
		require.Equal(t, expected, findings)
	})

	t.Run("Unpinnned image warning", func(t *testing.T) {
		t.Parallel()
		unpinnedImage := "registry.com:9001/whatever/image:1.0.0"
		badImage := "badimage:badimage@@sha256:3fbc632167424a6d997e74f5"
		component := types.ZarfComponent{Images: []string{
			unpinnedImage,
			"busybox:latest@sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79",
			badImage,
		}}
		findings := checkForUnpinnedImages(component, 0)
		expected := []types.PackageFinding{
			{
				Item:        unpinnedImage,
				Description: "Image not pinned with digest",
				Severity:    types.SevWarn,
				YqPath:      ".components.[0].images.[0]",
			},
			{
				Item:        badImage,
				Description: "Failed to parse image reference",
				Severity:    types.SevWarn,
				YqPath:      ".components.[0].images.[2]",
			},
		}
		require.Equal(t, expected, findings)
	})

	t.Run("Unpinnned file warning", func(t *testing.T) {
		t.Parallel()
		fileURL := "http://example.com/file.zip"
		localFile := "local.txt"
		zarfFiles := []types.ZarfFile{
			{
				Source: fileURL,
			},
			{
				Source: localFile,
			},
			{
				Source: fileURL,
				Shasum: "fake-shasum",
			},
		}
		component := types.ZarfComponent{Files: zarfFiles}
		findings := checkForUnpinnedFiles(component, 0)
		expectedErr := []types.PackageFinding{
			{
				Item:        fileURL,
				Description: "No shasum for remote file",
				Severity:    types.SevWarn,
				YqPath:      ".components.[0].files.[0]",
			},
		}
		require.Equal(t, expectedErr, findings)
		require.Len(t, findings, 1)
	})

	t.Run("isImagePinned", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			input    string
			expected bool
			err      error
		}{
			{
				input:    "registry.com:8080/defenseunicorns/whatever",
				expected: false,
				err:      nil,
			},
			{
				input:    "ghcr.io/defenseunicorns/pepr/controller:v0.15.0",
				expected: false,
				err:      nil,
			},
			{
				input:    "busybox:latest@sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79",
				expected: true,
				err:      nil,
			},
			{
				input:    "busybox:bad/image",
				expected: false,
				err:      errors.New("invalid reference format"),
			},
			{
				input:    "busybox:###ZARF_PKG_TMPL_BUSYBOX_IMAGE###",
				expected: true,
				err:      nil,
			},
		}
		for _, tc := range tests {
			t.Run(tc.input, func(t *testing.T) {
				actual, err := isPinnedImage(tc.input)
				if err != nil {
					require.EqualError(t, err, tc.err.Error())
				}
				require.Equal(t, tc.expected, actual)
			})
		}
	})
}
