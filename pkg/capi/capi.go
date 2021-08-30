// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package capi manages CAPI installation, provides default client for CAPI CRDs.
package capi

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/talos-systems/capi-utils/pkg/capi/infrastructure"
)

// Manager installs and controls cluster API installation.
type Manager struct {
	kubeconfig    client.Kubeconfig
	client        client.Client
	clientset     *kubernetes.Clientset
	config        *rest.Config
	runtimeClient runtimeclient.Client

	options Options
}

// Options for the CAPI installer.
type Options struct {
	Kubeconfig              client.Kubeconfig
	ClusterctlConfigPath    string
	CoreProvider            string
	ContextName             string
	InfrastructureProviders []infrastructure.Provider
	BootstrapProviders      []string
	ControlPlaneProviders   []string
}

// NewManager creates new Manager object.
func NewManager(ctx context.Context, options Options) (*Manager, error) {
	clusterAPI := &Manager{
		options: options,
	}

	var err error

	clusterAPI.client, err = client.New(options.ClusterctlConfigPath)
	if err != nil {
		return nil, err
	}

	clusterConfig, err := clusterAPI.GetKubeconfig(ctx)
	if err != nil {
		return nil, err
	}

	clusterAPI.config, err = clientcmd.BuildConfigFromKubeconfigGetter("", func() (*clientcmdapi.Config, error) {
		c, e := clientcmd.LoadFromFile(clusterConfig.Path)
		if e != nil {
			return nil, e
		}

		if clusterAPI.options.ContextName == "" {
			clusterAPI.options.ContextName = c.CurrentContext
		}

		return c, nil
	})
	if err != nil {
		return nil, err
	}

	clusterAPI.clientset, err = kubernetes.NewForConfig(clusterAPI.config)
	if err != nil {
		return nil, err
	}

	_, err = clusterAPI.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	return clusterAPI, nil
}

// GetKubeconfig returns kubeconfig in clusterctl expected format.
func (clusterAPI *Manager) GetKubeconfig(ctx context.Context) (client.Kubeconfig, error) {
	if clusterAPI.kubeconfig.Path != "" {
		return clusterAPI.kubeconfig, nil
	}

	var path string

	if v := os.Getenv(clientcmd.RecommendedConfigPathEnvVar); v != "" {
		path = v
	} else {
		usr, err := user.Current()
		if err != nil {
			return client.Kubeconfig{}, err
		}

		path = filepath.Join(usr.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName)
	}

	clusterAPI.kubeconfig.Path = path
	clusterAPI.kubeconfig.Context = clusterAPI.options.ContextName

	return clusterAPI.kubeconfig, nil
}

// GetManagerClient client returns instance of cluster API client.
func (clusterAPI *Manager) GetManagerClient() client.Client {
	return clusterAPI.client
}

// GetClient returns k8s client stuffed with CAPI CRDs.
func (clusterAPI *Manager) GetClient(ctx context.Context) (client runtimeclient.Client, err error) {
	if clusterAPI.runtimeClient != nil {
		return clusterAPI.runtimeClient, nil
	}

	clusterAPI.runtimeClient, err = GetMetalClient(clusterAPI.config)

	return clusterAPI.runtimeClient, err
}

// Install the Manager components and wait for them to be ready.
func (clusterAPI *Manager) Install(ctx context.Context) error {
	kubeconfig, err := clusterAPI.GetKubeconfig(ctx)
	if err != nil {
		return err
	}

	providers := make([]string, len(clusterAPI.options.InfrastructureProviders))
	for i, provider := range clusterAPI.options.InfrastructureProviders {
		providers[i] = provider.Name()

		if provider.Version() != "" {
			providers[i] += ":" + provider.Version()
		}

		if err = provider.PreInstall(); err != nil {
			return err
		}
	}

	opts := client.InitOptions{
		Kubeconfig:              kubeconfig,
		CoreProvider:            clusterAPI.options.CoreProvider,
		BootstrapProviders:      clusterAPI.options.BootstrapProviders,
		ControlPlaneProviders:   clusterAPI.options.ControlPlaneProviders,
		InfrastructureProviders: providers,
		TargetNamespace:         "",
		WatchingNamespace:       "",
		LogUsageInstructions:    false,
	}

	for _, provider := range clusterAPI.options.InfrastructureProviders {
		var installed bool

		if installed, err = provider.IsInstalled(ctx, clusterAPI.clientset); err != nil {
			return err
		}

		if !installed {
			if _, err = clusterAPI.client.Init(opts); err != nil {
				return err
			}
		}

		if err = provider.WaitReady(ctx, clusterAPI.clientset); err != nil {
			return err
		}
	}

	return nil
}

type ref struct {
	types.NamespacedName
	gvk schema.GroupVersionKind
}

func getRef(in map[string]interface{}, keys ...string) (*ref, error) {
	res := &ref{}

	refInterface, found, err := unstructured.NestedMap(in, keys...)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fieldNotFound(keys...)
	}

	res.Name, found, err = unstructured.NestedString(refInterface, "name")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fieldNotFound(append(keys, "name")...)
	}

	res.Namespace, found, err = unstructured.NestedString(refInterface, "namespace")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fieldNotFound(append(keys, "namespace")...)
	}

	groupVersion, found, err := unstructured.NestedString(refInterface, "apiVersion")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fieldNotFound(append(keys, "apiVersion")...)
	}

	kind, found, err := unstructured.NestedString(refInterface, "kind")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fieldNotFound(append(keys, "kind")...)
	}

	res.gvk = schema.FromAPIVersionAndKind(groupVersion, kind)

	return res, nil
}

func fieldNotFound(fields ...string) error {
	return fmt.Errorf("failed to find field %s", strings.Join(fields, "."))
}
