package test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTemplating(t *testing.T) {
	t.Log("E2E: Temporary directory deploy")

	e2e.setup(t)
	defer e2e.teardown(t)

	// run `zarf package create` with a specified image cache location
	imageCachePath := "/tmp/.image_cache-location"
	decompressPath := "/tmp/.package-decompressed"

	e2e.cleanFiles(imageCachePath, decompressPath)

	pkgName := fmt.Sprintf("zarf-package-package-variables-%s.tar.zst", e2e.arch)

	// Test a simple package variable example
	stdOut, stdErr, err := e2e.execZarfCommand("package", "create", "examples/package-variables", "--set", "CAT=meow", "--set", "FOX=bark", "--confirm", "--zarf-cache", imageCachePath)
	require.NoError(t, err, stdOut, stdErr)

	stdOut, stdErr, err = e2e.execZarfCommand("t", "archiver", "decompress", pkgName, decompressPath)
	require.NoError(t, err, stdOut, stdErr)

	// Check that the configmap exists and is readable
	_, err = ioutil.ReadFile(decompressPath + "/components/variable-example/manifests/simple-configmap.yaml")
	require.NoError(t, err)

	// Check variables in zarf.yaml are replaced correctly
	builtConfig, err := ioutil.ReadFile(decompressPath + "/zarf.yaml")
	require.NoError(t, err)
	require.Contains(t, string(builtConfig), "###ZARF_VAR_WOLF### is the ancestor of woof but not of a meow or a bark")

	e2e.cleanFiles(imageCachePath, decompressPath, pkgName)

	pkgName = fmt.Sprintf("zarf-package-composable-package-variables-%s.tar.zst", e2e.arch)

	// Test a composable package variable example
	stdOut, stdErr, err = e2e.execZarfCommand("package", "create", "examples/composable-package-variables", "--set", "CAT=meow", "--confirm", "--zarf-cache", imageCachePath, "--log-level=debug")
	require.NoError(t, err, stdOut, stdErr)

	stdOut, stdErr, err = e2e.execZarfCommand("t", "archiver", "decompress", pkgName, decompressPath)
	require.NoError(t, err, stdOut, stdErr)

	// Check that the configmaps exist and are readable
	_, err = ioutil.ReadFile(decompressPath + "/components/nested-example/manifests/sub-package/package-variables-2/two-configmap.yaml")
	require.NoError(t, err)
	_, err = ioutil.ReadFile(decompressPath + "/components/single-example/manifests/sub-package/package-variables-1/one-configmap.yaml")
	require.NoError(t, err)

	// Check variables in zarf.yaml are replaced correctly
	builtConfig, err = ioutil.ReadFile(decompressPath + "/zarf.yaml")
	require.NoError(t, err)
	require.Contains(t, string(builtConfig), "Who let the woof's out? meow")

	e2e.cleanFiles(imageCachePath, decompressPath, pkgName)
}
