package test

import (
	"net/http"
	"testing"

	"github.com/defenseunicorns/zarf/src/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	t.Log("E2E: Logging")
	e2e.setup(t)
	defer e2e.teardown(t)

	tunnel := k8s.NewZarfTunnel()
	tunnel.Connect(k8s.ZarfLogging, false)
	defer tunnel.Close()

	// Make sure Grafana comes up cleanly
	resp, err := http.Get(tunnel.HttpEndpoint())
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	stdOut, stdErr, err := e2e.execZarfCommand("package", "remove", "init", "--components=logging", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}
