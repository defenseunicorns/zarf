// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package cluster contains Zarf-specific cluster management functions.
package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/types"
	autoscalingV2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployedZarfPackages gets metadata information about packages that have been deployed to the cluster.
// We determine what packages have been deployed to the cluster by looking for specific secrets in the Zarf namespace.
// Returns a list of DeployedPackage structs and a list of errors.
func (c *Cluster) GetDeployedZarfPackages() ([]types.DeployedPackage, []error) {
	var deployedPackages = []types.DeployedPackage{}
	var errorList []error
	// Get the secrets that describe the deployed packages
	secrets, err := c.Kube.GetSecretsWithLabel(ZarfNamespaceName, ZarfPackageInfoLabel)
	if err != nil {
		return deployedPackages, append(errorList, err)
	}

	// Process the k8s secret into our internal structs
	for _, secret := range secrets.Items {
		var deployedPackage types.DeployedPackage
		err := json.Unmarshal(secret.Data["data"], &deployedPackage)
		// add the error to the error list
		if err != nil {
			errorList = append(errorList, fmt.Errorf("unable to unmarshal the secret %s/%s", secret.Namespace, secret.Name))
		} else {
			deployedPackages = append(deployedPackages, deployedPackage)
		}

	}

	// TODO: If we move this function out of `internal` we should return a more standard singular error.
	return deployedPackages, errorList
}

// GetDeployedPackage gets the metadata information about the package name provided (if it exists in the cluster).
// We determine what packages have been deployed to the cluster by looking for specific secrets in the Zarf namespace.
func (c *Cluster) GetDeployedPackage(packageName string) (types.DeployedPackage, error) {
	var deployedPackage = types.DeployedPackage{}

	// Get the secret that describes the deployed init package
	secret, err := c.Kube.GetSecret(ZarfNamespaceName, config.ZarfPackagePrefix+packageName)
	if err != nil {
		return deployedPackage, err
	}

	err = json.Unmarshal(secret.Data["data"], &deployedPackage)
	return deployedPackage, err
}

// StripZarfLabelsAndSecretsFromNamespaces removes metadata and secrets from existing namespaces no longer manged by Zarf.
func (c *Cluster) StripZarfLabelsAndSecretsFromNamespaces() {
	spinner := message.NewProgressSpinner("Removing zarf metadata & secrets from existing namespaces not managed by Zarf")
	defer spinner.Stop()

	deleteOptions := metav1.DeleteOptions{}
	listOptions := metav1.ListOptions{
		LabelSelector: config.ZarfManagedByLabel + "=zarf",
	}

	if namespaces, err := c.Kube.GetNamespaces(); err != nil {
		spinner.Errorf(err, "Unable to get k8s namespaces")
	} else {
		for _, namespace := range namespaces.Items {
			if _, ok := namespace.Labels[agentLabel]; ok {
				spinner.Updatef("Removing Zarf Agent label for namespace %s", namespace.Name)
				delete(namespace.Labels, agentLabel)
				if _, err = c.Kube.UpdateNamespace(&namespace); err != nil {
					// This is not a hard failure, but we should log it
					spinner.Errorf(err, "Unable to update the namespace labels for %s", namespace.Name)
				}
			}

			for _, namespace := range namespaces.Items {
				spinner.Updatef("Removing Zarf secrets for namespace %s", namespace.Name)
				err := c.Kube.Clientset.CoreV1().
					Secrets(namespace.Name).
					DeleteCollection(context.TODO(), deleteOptions, listOptions)
				if err != nil {
					spinner.Errorf(err, "Unable to delete secrets from namespace %s", namespace.Name)
				}
			}
		}
	}

	spinner.Success()
}

// PackageSecretNeedsWait checks if the the package status or any of the component statuses are being mutated by a webhook and need to be waited on.
func (c *Cluster) PackageSecretNeedsWait(secretName string) (bool, int, error) {
	// Get the secret that describes the deployed package
	packageSecret, err := c.Kube.GetSecret(ZarfNamespaceName, secretName)
	if err != nil {
		return false, 0, err
	}

	// Parse the secret
	var deployedPackage types.DeployedPackage
	err = json.Unmarshal(packageSecret.Data["data"], &deployedPackage)
	if err != nil {
		return false, 0, err
	}

	// Check if there are any component level statuses that we need to wait for
	for componentName, hookMap := range deployedPackage.ComponentWebhooks {
		for hookName, webhook := range hookMap {
			if webhook.Status == string(types.WebhookStatusRunning) {
				message.Debugf("The component %s is still running the webhook %s", componentName, hookName)
				return true, webhook.WaitDurationSeconds, nil
			}
		}
	}

	// If we get here, none of the components need to wait for a webhook to run
	return false, 0, nil
}

// RecordPackageDeploymentAndWait records the deployment of a package to the cluster and waits for any webhooks to complete.
func (c *Cluster) RecordPackageDeploymentAndWait(pkg types.ZarfPackage, components []types.DeployedComponent, connectStrings types.ConnectStrings, generation int) (*corev1.Secret, error) {

	packageSecret, err := c.RecordPackageDeployment(pkg, components, connectStrings, generation)
	if err != nil {
		return packageSecret, err
	}

	packageNeedsWait, waitSeconds, err := c.PackageSecretNeedsWait(packageSecret.Name)
	if err != nil {
		return packageSecret, err
	}
	if !packageNeedsWait {
		return packageSecret, nil
	}

	// Timebox the amount of time we wait for a mutation to finish before erroring
	waitDuration := types.DefaultWebhookWaitDuration
	if waitSeconds > 0 {
		waitDuration = time.Duration(waitSeconds) * time.Second
	}
	timeout := time.After(waitDuration)

	// We need to wait for this package to finish having webhooks run, create a spinner and keep checking until it's ready
	spinner := message.NewProgressSpinner("Waiting for component webhooks to complete")
	defer spinner.Stop()
	for packageNeedsWait {
		select {
		// On timeout, abort and return an error.
		case <-timeout:
			return nil, errors.New("timed out waiting for package deployment to complete")
		default:
			// Wait for 3 seconds before checking the secret again
			time.Sleep(3 * time.Second)
			packageNeedsWait, _, err = c.PackageSecretNeedsWait(packageSecret.Name)
		}
	}

	spinner.Success()
	return packageSecret, err
}

// RecordPackageDeployment saves metadata about a package that has been deployed to the cluster.
func (c *Cluster) RecordPackageDeployment(pkg types.ZarfPackage, components []types.DeployedComponent, connectStrings types.ConnectStrings, generation int) (*corev1.Secret, error) {

	// Attempt to load information about webhooks for the package
	componentWebhooks := map[string]map[string]types.Webhook{}
	existingPackageSecret, err := c.GetDeployedPackage(pkg.Metadata.Name)
	if err == nil {
		componentWebhooks = existingPackageSecret.ComponentWebhooks
	}

	// Generate a secret that describes the package that is being deployed
	packageName := pkg.Metadata.Name
	deployedPackageSecret := c.Kube.GenerateSecret(ZarfNamespaceName, config.ZarfPackagePrefix+packageName, corev1.SecretTypeOpaque)
	deployedPackageSecret.Labels[ZarfPackageInfoLabel] = packageName

	stateData, err := json.Marshal(types.DeployedPackage{
		Name:               packageName,
		CLIVersion:         config.CLIVersion,
		Data:               pkg,
		DeployedComponents: components,
		ConnectStrings:     connectStrings,
		Generation:         generation,
		ComponentWebhooks:  componentWebhooks,
	})
	if err != nil {
		return nil, err
	}

	deployedPackageSecret.Data = map[string][]byte{"data": stateData}

	return c.Kube.CreateOrUpdateSecret(deployedPackageSecret)
}

// EnableRegHPAScaleDown enables the HPA scale down for the Zarf Registry.
func (c *Cluster) EnableRegHPAScaleDown() error {
	hpa, err := c.Kube.GetHPA(ZarfNamespaceName, "zarf-docker-registry")
	if err != nil {
		return err
	}

	// Enable HPA scale down.
	policy := autoscalingV2.MinChangePolicySelect
	hpa.Spec.Behavior.ScaleDown.SelectPolicy = &policy

	// Save the HPA changes.
	if _, err = c.Kube.UpdateHPA(hpa); err != nil {
		return err
	}

	return nil
}

// DisableRegHPAScaleDown disables the HPA scale down for the Zarf Registry.
func (c *Cluster) DisableRegHPAScaleDown() error {
	hpa, err := c.Kube.GetHPA(ZarfNamespaceName, "zarf-docker-registry")
	if err != nil {
		return err
	}

	// Disable HPA scale down.
	policy := autoscalingV2.DisabledPolicySelect
	hpa.Spec.Behavior.ScaleDown.SelectPolicy = &policy

	// Save the HPA changes.
	if _, err = c.Kube.UpdateHPA(hpa); err != nil {
		return err
	}

	return nil
}
