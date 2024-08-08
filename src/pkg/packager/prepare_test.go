// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

package packager

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/zarf-dev/zarf/src/pkg/lint"
	"github.com/zarf-dev/zarf/src/test/testutil"
	"github.com/zarf-dev/zarf/src/types"
)

func TestFindImages(t *testing.T) {
	t.Parallel()

	ctx := testutil.TestContext(t)

	lint.ZarfSchema = testutil.LoadSchema(t, "../../../zarf.schema.json")

	cfg := &types.PackagerConfig{
		CreateOpts: types.ZarfCreateOptions{
			BaseDir: "../../../examples/dos-games/",
		},
	}
	p, err := New(cfg)
	require.NoError(t, err)
	images, err := p.FindImages(ctx)
	require.NoError(t, err)
	expectedImages := map[string][]string{
		"baseline": {
			"ghcr.io/zarf-dev/doom-game:0.0.1",
			"ghcr.io/zarf-dev/doom-game:sha256-7464ecc8a7172fce5c2ad631fc2a1b8572c686f4bf15c4bd51d7d6c9f0c460a7.sig",
		},
	}
	require.Equal(t, len(expectedImages), len(images))
	for k, v := range expectedImages {
		require.ElementsMatch(t, v, images[k])
	}

	cfg = &types.PackagerConfig{
		CreateOpts: types.ZarfCreateOptions{
			BaseDir: "../../../examples/dos-games/",
		},
		FindImagesOpts: types.ZarfFindImagesOptions{
			Why: "foobar",
		},
	}
	p, err = New(cfg)
	require.NoError(t, err)
	_, err = p.FindImages(ctx)
	require.EqualError(t, err, "image foobar not found in any charts or manifests")
}

func TestBuildImageMap(t *testing.T) {
	t.Parallel()

	podSpec := corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Image: "init-image",
			},
			{
				Image: "duplicate-image",
			},
		},
		Containers: []corev1.Container{

			{
				Image: "container-image",
			},
			{
				Image: "alpine:latest",
			},
		},
		EphemeralContainers: []corev1.EphemeralContainer{
			{
				EphemeralContainerCommon: corev1.EphemeralContainerCommon{
					Image: "ephemeral-image",
				},
			},
			{
				EphemeralContainerCommon: corev1.EphemeralContainerCommon{
					Image: "duplicate-image",
				},
			},
		},
	}
	imgMap := appendToImageMap(map[string]bool{}, podSpec)
	expectedImgMap := map[string]bool{
		"init-image":      true,
		"duplicate-image": true,
		"container-image": true,
		"alpine:latest":   true,
		"ephemeral-image": true,
	}
	require.Equal(t, expectedImgMap, imgMap)
}

func TestGetSortedImages(t *testing.T) {
	t.Parallel()

	matchedImages := map[string]bool{
		"C": true,
		"A": true,
		"E": true,
		"D": true,
	}
	maybeImages := map[string]bool{
		"Z": true,
		"A": true,
		"B": true,
	}
	sortedMatchedImages, sortedMaybeImages := getSortedImages(matchedImages, maybeImages)
	expectedSortedMatchedImages := []string{"A", "C", "D", "E"}
	require.Equal(t, expectedSortedMatchedImages, sortedMatchedImages)
	expectedSortedMaybeImages := []string{"B", "Z"}
	require.Equal(t, expectedSortedMaybeImages, sortedMaybeImages)
}
