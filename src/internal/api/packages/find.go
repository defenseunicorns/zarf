package packages

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/internal/api/common"
	"github.com/defenseunicorns/zarf/src/internal/message"
	"github.com/defenseunicorns/zarf/src/internal/utils"
	"github.com/defenseunicorns/zarf/src/types"
)

var packagePattern = regexp.MustCompile(`zarf-package-.*\.tar`)
var initPattern = regexp.MustCompile(config.GetInitPackageName())

// Find returns all packages anywhere down the directory tree of the working directory.
func Find(w http.ResponseWriter, r *http.Request) {
	message.Debug("packages.Find()")
	findPackage(packagePattern, w, os.Getwd)
}

// FindInHome returns all packages in the user's home directory.
func FindInHome(w http.ResponseWriter, r *http.Request) {
	message.Debug("packages.FindInHome()")
	findPackage(packagePattern, w, os.UserHomeDir)
}

// FindInitPackage returns all init packages anywhere down the directory tree of the working directory.
func FindInitPackage(w http.ResponseWriter, r *http.Request) {
	message.Debug("packages.FindInitPackage()")
	findPackage(initPattern, w, os.Getwd)
}

func findPackage(pattern *regexp.Regexp, w http.ResponseWriter, setDir func() (string, error)) {
	targetDir, err := setDir()
	if err != nil {
		message.ErrorWebf(err, w, "Error getting directory")
	}

	// Intentionally ignore errors
	files, err := utils.RecursiveFileList(targetDir, pattern)
	if err != nil {
		pkgNotFoundMsg := fmt.Sprintf("Package not found: %s", pattern.String())
		message.Errorf(err, pkgNotFoundMsg)
		common.WriteJSONResponse(w, types.APIError{
			Error:   "PackageNotFound",
			Message: pkgNotFoundMsg,
		}, http.StatusNotFound)
		return
	}
	common.WriteJSONResponse(w, files, http.StatusOK)
}
