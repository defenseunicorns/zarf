// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package test provides e2e tests for Zarf.
package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/defenseunicorns/zarf/src/internal/cluster"
	"github.com/stretchr/testify/require"
)

func TestPrometheus(t *testing.T) {
	t.Log("E2E: Prometheus")
	e2e.SetupWithCluster(t)

	path := fmt.Sprintf("build/zarf-package-scrape-zarf-agent-%s.tar.zst", e2e.Arch)

	// Deploy Prometheus
	stdOut, stdErr, err := e2e.Zarf("package", "deploy", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// tunnel, err := cluster.NewTunnel("monitoring", "svc", "prometheus-operator", 8080, 8080)
	tunnel, err := cluster.NewTunnel("monitoring", "svc", "prometheus-operator", 8080, 8080)

	require.NoError(t, err)
	err = tunnel.Connect("", false)
	require.NoError(t, err)
	defer tunnel.Close()

	// Check that 'curl' returns something.
	resp, err := http.Get(tunnel.HTTPEndpoint() + "/healthz")
	require.NoError(t, err, resp)
	require.Equal(t, 200, resp.StatusCode)

	stdOut, stdErr, err = e2e.Zarf("package", "remove", "scrape-zarf-agent", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

}
func TestDosGames(t *testing.T) {
	t.Log("E2E: Dos games")
	e2e.SetupWithCluster(t)

	path := fmt.Sprintf("build/zarf-package-dos-games-%s.tar.zst", e2e.Arch)

	// Deploy the game
	stdOut, stdErr, err := e2e.Zarf("package", "deploy", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	tunnel, err := cluster.NewZarfTunnel()
	require.NoError(t, err)
	err = tunnel.Connect("doom", false)
	require.NoError(t, err)
	defer tunnel.Close()

	// Check that 'curl' returns something.
	resp, err := http.Get(tunnel.HTTPEndpoint())
	require.NoError(t, err, resp)
	require.Equal(t, 200, resp.StatusCode)

	stdOut, stdErr, err = e2e.Zarf("package", "remove", "dos-games", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

func TestManifests(t *testing.T) {
	t.Log("E2E: Local, Remote, and Kustomize Manifests")
	e2e.SetupWithCluster(t)

	path := fmt.Sprintf("build/zarf-package-manifests-%s-0.0.1.tar.zst", e2e.Arch)

	// Deploy the package
	stdOut, stdErr, err := e2e.Zarf("package", "deploy", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Remove the package
	stdOut, stdErr, err = e2e.Zarf("package", "remove", "manifests", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}
