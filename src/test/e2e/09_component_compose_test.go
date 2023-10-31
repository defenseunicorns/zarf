// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package test provides e2e tests for Zarf.
package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CompositionSuite struct {
	suite.Suite
	*require.Assertions
}

var (
	composeExample     = filepath.Join("examples", "composable-packages")
	composeExamplePath string
	composeTest        = filepath.Join("src", "test", "packages", "09-composable-packages")
	composeTestPath    string
)

func (suite *CompositionSuite) SetupSuite() {
	suite.Assertions = require.New(suite.T())

	// Setup the package paths after e2e has been initialized
	composeExamplePath = filepath.Join("build", fmt.Sprintf("zarf-package-composable-packages-%s.tar.zst", e2e.Arch))
	composeTestPath = filepath.Join("build", fmt.Sprintf("zarf-package-test-compose-package-%s.tar.zst", e2e.Arch))

}

func (suite *CompositionSuite) TearDownSuite() {
	err := os.RemoveAll(composeExamplePath)
	suite.NoError(err)
	err = os.RemoveAll(composeTestPath)
	suite.NoError(err)
}

func (suite *CompositionSuite) Test_0_ComposabilityExample() {
	suite.T().Log("E2E: Package Compose Example")

	_, stdErr, err := e2e.Zarf("package", "create", composeExample, "-o", "build", "--insecure", "--no-color", "--confirm")
	suite.NoError(err)

	// Ensure that common names merge
	suite.Contains(stdErr, `
  manifests:
  - name: multi-games
    namespace: dos-games
    files:
    - ../dos-games/manifests/deployment.yaml
    - ../dos-games/manifests/service.yaml
    - quake-service.yaml`)

	// Ensure that the action was appended
	suite.Contains(stdErr, `
  - defenseunicorns/zarf-game:multi-tile-dark
  actions:
    onDeploy:
      before:
      - cmd: ./zarf tools kubectl get -n dos-games deployment -o jsonpath={.items[0].metadata.creationTimestamp}`)
}

func (suite *CompositionSuite) Test_1_FullComposability() {
	suite.T().Log("E2E: Full Package Compose")

	_, stdErr, err := e2e.Zarf("package", "create", composeTest, "-o", "build", "--insecure", "--no-color", "--confirm")
	suite.NoError(err)

	// Ensure that names merge and that composition is added appropriately

	// Check metadata
	suite.Contains(stdErr, `
- name: test-compose-package
  description: A contrived example for podinfo using many Zarf primitives for compose testing
  required: true
`)

	// Check files
	suite.Contains(stdErr, `
  files:
  - source: files/coffee-ipsum.txt
    target: coffee-ipsum.txt
  - source: files/coffee-ipsum.txt
    target: coffee-ipsum.txt
`)

	// Check charts
	suite.Contains(stdErr, `
  charts:
  - name: podinfo-compose
    releaseName: podinfo-override
    url: oci://ghcr.io/stefanprodan/charts/podinfo
    version: 6.4.0
    namespace: podinfo-override
    valuesFiles:
    - files/test-values.yaml
    - files/test-values.yaml
  - name: podinfo-compose-two
    releaseName: podinfo-compose-two
    url: oci://ghcr.io/stefanprodan/charts/podinfo
    version: 6.4.0
    namespace: podinfo-compose-two
    valuesFiles:
    - files/test-values.yaml
`)

	// Check manifests
	suite.Contains(stdErr, `
  manifests:
  - name: connect-service
    namespace: podinfo-override
    files:
    - files/service.yaml
    - files/service.yaml
    kustomizations:
    - files
    - files
  - name: connect-service-two
    namespace: podinfo-compose-two
    files:
    - files/service.yaml
    kustomizations:
    - files
`)

	// Check images + repos
	suite.Contains(stdErr, `
  images:
  - ghcr.io/stefanprodan/podinfo:6.4.0
  - ghcr.io/stefanprodan/podinfo:6.4.1
  repos:
  - https://github.com/defenseunicorns/zarf-public-test.git
  - https://github.com/defenseunicorns/zarf-public-test.git@refs/heads/dragons
`)

	// Check dataInjections
	suite.Contains(stdErr, `
  dataInjections:
  - source: files
    target:
      namespace: podinfo-compose
      selector: app.kubernetes.io/name=podinfo-compose
      container: podinfo
      path: /home/app/service.yaml
  - source: files
    target:
      namespace: podinfo-compose
      selector: app.kubernetes.io/name=podinfo-compose
      container: podinfo
      path: /home/app/service.yaml
`)

	// Check actions
	suite.Contains(stdErr, `
  actions:
    onCreate:
      before:
      - dir: sub-package
        cmd: ls
      - dir: .
        cmd: ls
    onDeploy:
      after:
      - cmd: cat coffee-ipsum.txt
      - wait:
          cluster:
            kind: deployment
            name: podinfo-compose-two
            namespace: podinfo-compose-two
            condition: available`)
}

func TestCompositionSuite(t *testing.T) {
	e2e.SetupWithCluster(t)

	suite.Run(t, new(CompositionSuite))
}