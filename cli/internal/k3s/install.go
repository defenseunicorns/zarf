package k3s

import (
	"os"

	"github.com/defenseunicorns/zarf/cli/config"
	"github.com/defenseunicorns/zarf/cli/internal/git"
	"github.com/defenseunicorns/zarf/cli/internal/packager"
	"github.com/defenseunicorns/zarf/cli/internal/utils"
	"github.com/sirupsen/logrus"
)

type InstallOptions struct {
	PKI        utils.PKIConfig
	Confirmed  bool
	Components string
}

func Install(options InstallOptions) {
	utils.RunPreflightChecks()

	logrus.Info("Installing K3s")

	packager.Deploy(config.PackageInitName, options.Confirmed, options.Components)

	// Install RHEL RPMs if applicable
	if utils.IsRHEL() {
		configureRHEL()
	}

	// Create the K3s systemd service
	createService()

	createK3sSymlinks()

	utils.HandlePKI(options.PKI)

	gitSecret := git.GetOrCreateZarfSecret()

	logrus.Info("Installation complete.  You can run \"/usr/local/bin/k9s\" to monitor the status of the deployment.")
	logrus.WithFields(logrus.Fields{
		"Gitea Username (if installed)": config.ZarfGitUser,
		"Grafana Username":              "zarf-admin",
		"Password (all)":                gitSecret,
	}).Warn("Credentials stored in ~/.git-credentials")
}

func createK3sSymlinks() {
	logrus.Info("Creating kube config symlink")

	// Make the k3s kubeconfig available to other standard K8s tools that bind to the default ~/.kube/config
	err := utils.CreateDirectory("/root/.kube", 0700)
	if err != nil {
		logrus.Debug(err)
		logrus.Warn("Unable to create the root kube config directory")
	} else {
		// Dont log an error for now since re-runs throw an invalid error
		_ = os.Symlink("/etc/rancher/k3s/k3s.yaml", "/root/.kube/config")
	}

	// Add aliases for k3s
	_ = os.Symlink(config.K3sBinary, "/usr/local/bin/kubectl")
	_ = os.Symlink(config.K3sBinary, "/usr/local/bin/ctr")
	_ = os.Symlink(config.K3sBinary, "/usr/local/bin/crictl")
}

func createService() {
	servicePath := "/etc/systemd/system/k3s.service"

	_ = os.Symlink(servicePath, "/etc/systemd/system/multi-user.target.wants/k3s.service")

	_, err := utils.ExecCommand(nil, "systemctl", "daemon-reload")
	if err != nil {
		logrus.Debug(err)
		logrus.Warn("Unable to reload systemd")
	}

	_, err = utils.ExecCommand(nil, "systemctl", "enable", "--now", "k3s")
	if err != nil {
		logrus.Debug(err)
		logrus.Warn("Unable to enable or start k3s via systemd")
	}
}
