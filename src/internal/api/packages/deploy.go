package packages

import (
	"encoding/json"
	"net/http"
	"path"
	"path/filepath"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/internal/api/common"
	"github.com/defenseunicorns/zarf/src/internal/message"
	"github.com/defenseunicorns/zarf/src/internal/packager"
	"github.com/defenseunicorns/zarf/src/internal/utils"
	"github.com/defenseunicorns/zarf/src/types"
)

// DeployPackage deploys a package to the Zarf cluster.
func DeployPackage(w http.ResponseWriter, r *http.Request) {
	isInitPkg := r.URL.Query().Get("isInitPkg") == "true"

	if isInitPkg {
		var body = types.ZarfInitOptions{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			message.ErrorWebf(err, w, "Unable to decode the request to deploy the cluster")
			return
		}
		config.InitOptions = body
		initPackageName := config.GetInitPackageName()
		config.DeployOptions.PackagePath = initPackageName

		// Try to use an init-package in the executable directory if none exist in current working directory
		if utils.InvalidPath(config.DeployOptions.PackagePath) {
			// Get the path to the executable
			if executablePath, err := utils.GetFinalExecutablePath(); err != nil {
				message.Errorf(err, "Unable to get the path to the executable")
			} else {
				executableDir := path.Dir(executablePath)
				config.DeployOptions.PackagePath = filepath.Join(executableDir, initPackageName)
			}

			// If the init-package doesn't exist in the executable directory, try the cache directory
			if err != nil || utils.InvalidPath(config.DeployOptions.PackagePath) {
				config.DeployOptions.PackagePath = filepath.Join(config.GetAbsCachePath(), initPackageName)

				// If the init-package doesn't exist in the cache directory, return an error
				if utils.InvalidPath(config.DeployOptions.PackagePath) {
					common.WriteJSONResponse(w, false, http.StatusBadRequest)
					return
				}
			}
		}
	} else {
		var body = types.ZarfDeployOptions{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			message.ErrorWebf(err, w, "Unable to decode the request to deploy the cluster")
			return
		}
		config.DeployOptions = body
	}

	config.CommonOptions.Confirm = true
	packager.Deploy()

	common.WriteJSONResponse(w, true, http.StatusCreated)
}
