// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package test provides e2e tests for Zarf.
package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/exec"
	"github.com/defenseunicorns/zarf/src/types"
	dconfig "github.com/docker/cli/cli/config"
	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/suite"
	"oras.land/oras-go/v2/registry"
)

/*
HOW TO TEST:
1. Publish a standard zarf package to the registry
*/
type OCIDifferentialSuite struct {
	suite.Suite
	Remote    *utils.OrasRemote
	Reference registry.Reference
}

func (suite *OCIDifferentialSuite) SetupSuite() {

	suite.Reference.Registry = "localhost:555"

	// spin up a local registry
	registryImage := "registry:2.8.1"
	err := exec.CmdWithPrint("docker", "run", "-d", "--restart=always", "-p", "555:5000", "--name", "registry", registryImage)
	suite.NoError(err)

	// docker config folder
	cfg, err := dconfig.Load(dconfig.Dir())
	suite.NoError(err)
	if !cfg.ContainsAuth() {
		// make a docker config file w/ some blank creds
		_, _, err := e2e.ExecZarfCommand("tools", "registry", "login", "--username", "zarf", "-p", "zarf", "localhost:6000")
		suite.NoError(err)
	}
	// publish one of the example packages to the registry
	examplePackagePath := filepath.Join("examples", "helm-oci-chart")
	stdOut, stdErr, err := e2e.ExecZarfCommand("package", "publish", examplePackagePath, "oci://"+suite.Reference.String(), "--insecure")
	suite.NoError(err, stdOut, stdErr)

	// build the package that we are going to publish
	anotherPackagePath := "src/test/test-packages/oci-differential"
	stdOut, stdErr, err = e2e.ExecZarfCommand("package", "create", anotherPackagePath, "--insecure", "--set=PACKAGE_VERSION=v0.24.0", "--confirm")
	suite.NoError(err, stdOut, stdErr)

	// publish the package that we just built
	packageName := "zarf-package-podinfo-with-oci-flux-amd64-v0.24.0.tar.zst"
	stdOut, stdErr, err = e2e.ExecZarfCommand("package", "publish", packageName, "oci://"+suite.Reference.String(), "--insecure")
	suite.NoError(err, stdOut, stdErr)

}

func (suite *OCIDifferentialSuite) TearDownSuite() {
	_, _, err := exec.Cmd("docker", "rm", "-f", "registry")
	suite.NoError(err)
}

func (suite *OCIDifferentialSuite) Test_0_Publish_SkeletonsXXX() {
	suite.T().Log("E2E: Skeleton Package Publish oci://")
	differentialPackageName := fmt.Sprintf("zarf-package-podinfo-with-oci-flux-%s-v0.24.0-differential-v0.25.0.tar.zst", e2e.Arch)
	normalPackageName := fmt.Sprintf("zarf-package-podinfo-with-oci-flux-%s-v0.24.0.tar.zst", e2e.Arch)
	tmpPath, _ := utils.MakeTempDir("")

	// Build without differential
	anotherPackagePath := "src/test/test-packages/oci-differential"
	stdOut, stdErr, err := e2e.ExecZarfCommand("package", "create", anotherPackagePath, "--insecure", "--set=PACKAGE_VERSION=v0.25.0", "--confirm")
	suite.NoError(err, stdOut, stdErr)

	// Extract and load the zarf.yaml config for the normally built package
	err = archiver.Extract(normalPackageName, "zarf.yaml", tmpPath)
	suite.NoError(err, "unable to extract zarf.yaml from the differential git package")
	var normalZarfConfig types.ZarfPackage
	err = utils.ReadYaml(filepath.Join(tmpPath, "zarf.yaml"), &normalZarfConfig)
	suite.NoError(err, "unable to read zarf.yaml from the differential git package")
	os.Remove(filepath.Join(tmpPath, "zarf.yaml"))

	stdOut, stdErr, err = e2e.ExecZarfCommand("package", "create", anotherPackagePath, "--differential", "oci://"+suite.Reference.String()+"/podinfo-with-oci-flux:v0.24.0-amd64", "--insecure", "--set=PACKAGE_VERSION=v0.25.0", "--confirm")
	suite.NoError(err, stdOut, stdErr)

	// Extract and load the zarf.yaml config for the differentially built package
	err = archiver.Extract(differentialPackageName, "zarf.yaml", tmpPath)
	suite.NoError(err, "unable to extract zarf.yaml from the differential git package")
	var differentialZarfConfig types.ZarfPackage
	err = utils.ReadYaml(filepath.Join(tmpPath, "zarf.yaml"), &differentialZarfConfig)
	suite.NoError(err, "unable to read zarf.yaml from the differential git package")

	// Perform a bunch of asserts around the non-differential package
	suite.Equal(normalZarfConfig.Metadata.Version, "v0.24.0")
	suite.False(normalZarfConfig.Build.Differential)
	suite.Len(normalZarfConfig.Build.OCIImportedComponents, 1)
	suite.Equal(normalZarfConfig.Build.OCIImportedComponents["oci://127.0.0.1:555/helm-oci-chart:0.0.1-skeleton"], "helm-oci-chart")

	suite.Len(normalZarfConfig.Components, 3)
	suite.Equal(normalZarfConfig.Components[0].Name, "helm-oci-chart")
	suite.Equal(normalZarfConfig.Components[0].Charts[0].URL, "oci://ghcr.io/stefanprodan/charts/podinfo")
	suite.Equal(normalZarfConfig.Components[0].Images[0], "ghcr.io/stefanprodan/podinfo:6.3.3")
	suite.Len(normalZarfConfig.Components[1].Images, 2)
	suite.Len(normalZarfConfig.Components[1].Repos, 4)
	suite.Len(normalZarfConfig.Components[2].Images, 1)
	suite.Len(normalZarfConfig.Components[2].Repos, 3)

	// Perform a bunch of asserts around the differential package
	suite.Equal(differentialZarfConfig.Metadata.Version, "v0.25.0")
	suite.True(differentialZarfConfig.Build.Differential)
	suite.Len(differentialZarfConfig.Build.DifferentialMissing, 1)
	suite.Equal(differentialZarfConfig.Build.DifferentialMissing[0], "helm-oci-chart")
	suite.Len(differentialZarfConfig.Build.OCIImportedComponents, 0)

	suite.Len(differentialZarfConfig.Components, 2)
	suite.Equal(differentialZarfConfig.Components[0].Name, "versioned-assets")
	suite.Len(differentialZarfConfig.Components[0].Images, 1)
	suite.Equal(differentialZarfConfig.Components[0].Images[0], "ghcr.io/defenseunicorns/zarf/agent:v0.25.0")
	suite.Len(differentialZarfConfig.Components[0].Repos, 1)
	suite.Equal(differentialZarfConfig.Components[0].Repos[0], "https://github.com/defenseunicorns/zarf.git@refs/tags/v0.25.0")

	suite.Len(differentialZarfConfig.Components[1].Images, 1)
	suite.Len(differentialZarfConfig.Components[1].Repos, 3)
	suite.Equal(differentialZarfConfig.Components[1].Images[0], "ghcr.io/stefanprodan/podinfo:latest")
	suite.Equal(differentialZarfConfig.Components[1].Repos[0], "https://github.com/stefanprodan/podinfo.git")

}

func TestOCIDifferentialSuite(t *testing.T) {
	e2e.SetupWithCluster(t)
	defer e2e.Teardown(t)
	suite.Run(t, new(OCIDifferentialSuite))
}
