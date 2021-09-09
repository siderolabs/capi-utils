// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package capi

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/talos-systems/go-retry/retry"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/talos-systems/capi-utils/pkg/capi/infrastructure"
)

// DeployOption defines a single CAPI cluster creation option.
type DeployOption func(opts *DeployOptions) error

// DeployOptions cluster deployment options.
type DeployOptions struct {
	providerOptions interface{}

	Provider          string
	ProviderVersion   string
	ClusterName       string
	ClusterNamespace  string
	TalosVersion      string
	KubernetesVersion string
	Template          []byte
	ControlPlaneNodes int64
	WorkerNodes       int64
}

// DefaultDeployOptions default deployment settings.
func DefaultDeployOptions() *DeployOptions {
	return &DeployOptions{
		ControlPlaneNodes: 1,
		WorkerNodes:       1,
		ClusterNamespace:  "default",
		TalosVersion:      "v0.11",
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}
}

// WithControlPlaneNodes creates a cluster with N control plane nodes.
func WithControlPlaneNodes(count int64) DeployOption {
	return func(o *DeployOptions) error {
		o.ControlPlaneNodes = count

		return nil
	}
}

// WithWorkerNodes creates a cluster with N worker nodes.
func WithWorkerNodes(count int64) DeployOption {
	return func(o *DeployOptions) error {
		o.WorkerNodes = count

		return nil
	}
}

// WithProvider sets cluster provider.
func WithProvider(name string) DeployOption {
	return func(o *DeployOptions) error {
		o.Provider = name

		return nil
	}
}

// WithProviderVersion picks provider with a specific version.
func WithProviderVersion(version string) DeployOption {
	return func(o *DeployOptions) error {
		o.ProviderVersion = version

		return nil
	}
}

// WithTalosVersion sets Talos version.
func WithTalosVersion(version string) DeployOption {
	return func(o *DeployOptions) error {
		o.TalosVersion = version

		return nil
	}
}

// WithKubernetesVersion sets Kubernetes version.
func WithKubernetesVersion(version string) DeployOption {
	return func(o *DeployOptions) error {
		o.KubernetesVersion = version

		return nil
	}
}

// WithTemplateFile load cluster template from the file.
func WithTemplateFile(path string) DeployOption {
	return func(o *DeployOptions) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}

		defer f.Close() //nolint:errcheck

		o.Template, err = ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		return nil
	}
}

// WithTemplate loads cluster template from memory.
func WithTemplate(data []byte) DeployOption {
	return func(o *DeployOptions) error {
		o.Template = data

		return nil
	}
}

// WithDeployOptions sets deploy options as a struct.
func WithDeployOptions(val *DeployOptions) DeployOption {
	return func(o *DeployOptions) error {
		*o = *val

		return nil
	}
}

// WithProviderOptions sets provider specific deploy options.
func WithProviderOptions(val interface{}) DeployOption {
	return func(o *DeployOptions) error {
		o.providerOptions = val

		return nil
	}
}

// WithClusterNamespace sets cluster namespace.
func WithClusterNamespace(val string) DeployOption {
	return func(o *DeployOptions) error {
		o.ClusterNamespace = val

		return nil
	}
}

// DeployCluster creates a new cluster.
//nolint:gocognit
func (clusterAPI *Manager) DeployCluster(ctx context.Context, clusterName string, setters ...DeployOption) (*Cluster, error) {
	if len(clusterAPI.providers) == 0 {
		return nil, fmt.Errorf("no infrastructure providers are installed")
	}

	options := DefaultDeployOptions()

	for _, setter := range setters {
		if err := setter(options); err != nil {
			return nil, err
		}
	}

	options.ClusterName = clusterName

	var provider infrastructure.Provider

	if options.Provider != "" {
		for _, p := range clusterAPI.providers {
			if p.Name() == options.Provider {
				if options.ProviderVersion != "" && p.Version() != options.ProviderVersion {
					continue
				}

				provider = p

				break
			}
		}

		if provider == nil {
			return nil, fmt.Errorf("no provider with name %s is installed", options.Provider)
		}
	} else {
		provider = clusterAPI.providers[0]
	}

	// set up env variables common for all providers
	vars := map[string]string{
		"TALOS_VERSION":               options.TalosVersion,
		"KUBERNETES_VERSION":          options.KubernetesVersion,
		"CLUSTER_NAME":                options.ClusterName,
		"CONTROL_PLANE_MACHINE_COUNT": strconv.FormatInt(options.ControlPlaneNodes, 10),
		"WORKER_MACHINE_COUNT":        strconv.FormatInt(options.WorkerNodes, 10),
	}

	for key, value := range vars {
		if err := os.Setenv(key, value); err != nil {
			return nil, err
		}
	}

	templateOptions := client.GetClusterTemplateOptions{
		Kubeconfig:               clusterAPI.kubeconfig,
		ClusterName:              options.ClusterName,
		ControlPlaneMachineCount: &options.ControlPlaneNodes,
		WorkerMachineCount:       &options.WorkerNodes,
	}

	if options.Template != nil {
		file, err := ioutil.TempFile("", "clusterTemplate")
		if err != nil {
			log.Fatal(err)
		}

		templateOptions.URLSource = &client.URLSourceOptions{
			URL: file.Name(),
		}

		if _, err = file.Write(options.Template); err != nil {
			return nil, err
		}

		defer file.Close() //nolint:errcheck
	}

	template, err := provider.GetClusterTemplate(clusterAPI.client, templateOptions, options.providerOptions)
	if err != nil {
		return nil, err
	}

	for _, obj := range template.Objs() {
		if err = clusterAPI.runtimeClient.Create(ctx, &obj); err != nil {
			return nil, err
		}
	}

	deployedCluster, err := clusterAPI.NewCluster(ctx, options.ClusterName, options.ClusterNamespace)
	if err != nil {
		return nil, err
	}

	if err = retry.Constant(30*time.Minute, retry.WithUnits(10*time.Second), retry.WithErrorLogging(true)).Retry(func() error {
		return clusterAPI.CheckClusterReady(ctx, deployedCluster)
	}); err != nil {
		return nil, err
	}

	return deployedCluster, nil
}

// DestroyCluster deletes cluster.
func (clusterAPI *Manager) DestroyCluster(ctx context.Context, name, namespace string) error {
	cluster := &unstructured.Unstructured{}
	cluster.SetName(name)
	cluster.SetNamespace(namespace)
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Kind:    "Cluster",
		Version: clusterAPI.version,
	})

	if err := clusterAPI.runtimeClient.Delete(ctx, cluster); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	return retry.Constant(30*time.Minute, retry.WithUnits(10*time.Second), retry.WithErrorLogging(true)).Retry(func() error {
		err := clusterAPI.runtimeClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, cluster)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}

			return err
		}

		return retry.ExpectedError(fmt.Errorf("cluster is being deleted"))
	})
}
