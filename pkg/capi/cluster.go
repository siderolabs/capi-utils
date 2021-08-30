// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package capi

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/talos-systems/go-retry/retry"
	talosclusterapi "github.com/talos-systems/talos/pkg/machinery/api/cluster"
	talosclient "github.com/talos-systems/talos/pkg/machinery/client"
	clientconfig "github.com/talos-systems/talos/pkg/machinery/client/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Cluster attaches to the provisioned CAPI cluster and provides talos.Cluster.
type Cluster struct {
	client            *talosclient.Client
	name              string
	controlPlaneNodes []string
	workerNodes       []string
}

// NewCluster fetches cluster info from the CAPI state.
//nolint:gocyclo,cyclop
func (clusterAPI *Manager) NewCluster(ctx context.Context, clusterName, version string) (*Cluster, error) {
	var (
		cluster      unstructured.Unstructured
		controlPlane unstructured.Unstructured
		machines     unstructured.UnstructuredList
		talosConfig  unstructured.Unstructured

		controlPlaneNodes = []string{}
		workerNodes       = []string{}
		configEndpoints   = []string{}
	)

	cluster.SetGroupVersionKind(
		schema.GroupVersionKind{
			Version: version,
			Group:   "cluster.x-k8s.io",
			Kind:    "Cluster",
		},
	)

	machines.SetGroupVersionKind(
		schema.GroupVersionKind{
			Version: version,
			Group:   "cluster.x-k8s.io",
			Kind:    "Machine",
		},
	)

	clusterRef := types.NamespacedName{Namespace: "default", Name: clusterName}

	if err := clusterAPI.runtimeClient.Get(ctx, clusterRef, &cluster); err != nil {
		return nil, err
	}

	var (
		controlPlaneSelector string
		found                bool
		err                  error
	)

	controlPlaneRef, err := getRef(cluster.Object, "spec", "controlPlaneRef")
	if err != nil {
		return nil, err
	}

	controlPlane.SetGroupVersionKind(controlPlaneRef.gvk)

	if err = clusterAPI.runtimeClient.Get(ctx, controlPlaneRef.NamespacedName, &controlPlane); err != nil {
		return nil, err
	}

	if controlPlaneSelector, found, err = unstructured.NestedString(controlPlane.Object, "status", "selector"); err != nil {
		return nil, err
	} else if !found {
		return nil, fieldNotFound("status", "selector")
	}

	labelSelector, err := labels.Parse(controlPlaneSelector)
	if err != nil {
		return nil, err
	}

	if err = clusterAPI.runtimeClient.List(ctx, &machines, runtimeclient.MatchingLabelsSelector{Selector: labelSelector}); err != nil {
		return nil, err
	}

	if len(machines.Items) < 1 {
		return nil, fmt.Errorf("not enough machines found")
	}

	configRef, err := getRef(machines.Items[0].Object, "spec", "bootstrap", "configRef")
	if err != nil {
		return nil, err
	}

	talosConfig.SetGroupVersionKind(configRef.gvk)

	if err = clusterAPI.runtimeClient.Get(ctx, configRef.NamespacedName, &talosConfig); err != nil {
		return nil, err
	}

	var (
		clientConfig      *clientconfig.Config
		talosConfigString string
	)

	if talosConfigString, found, err = unstructured.NestedString(talosConfig.Object, "status", "talosConfig"); err != nil {
		return nil, err
	} else if !found {
		return nil, fieldNotFound("status", "talosConfig")
	}

	clientConfig, err = clientconfig.FromString(talosConfigString)

	if err != nil {
		return nil, err
	}

	kubeconfig, err := clusterAPI.GetKubeconfig(ctx)
	if err != nil {
		return nil, err
	}

	options := client.GetKubeconfigOptions{
		Kubeconfig:          kubeconfig,
		WorkloadClusterName: clusterRef.Name,
		Namespace:           clusterRef.Namespace,
	}

	raw, err := clusterAPI.client.GetKubeconfig(options)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(raw))
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		_, controlPlane := node.Labels["node-role.kubernetes.io/control-plane"]

		for _, address := range node.Status.Addresses {
			switch {
			case address.Type == v1.NodeExternalIP && controlPlane:
				configEndpoints = append(configEndpoints, address.Address)
			case address.Type == v1.NodeInternalIP && controlPlane:
				controlPlaneNodes = append(controlPlaneNodes, address.Address)
			case address.Type == v1.NodeInternalIP:
				workerNodes = append(workerNodes, address.Address)
			}
		}
	}

	if len(configEndpoints) < 1 {
		return nil, fmt.Errorf("failed to find control plane nodes")
	}

	clientConfig.Contexts[clientConfig.Context].Endpoints = configEndpoints

	var talosClient *talosclient.Client

	talosClient, err = talosclient.New(ctx, talosclient.WithConfig(clientConfig))
	if err != nil {
		return nil, err
	}

	return &Cluster{
		name:              clusterName,
		controlPlaneNodes: controlPlaneNodes,
		workerNodes:       workerNodes,
		client:            talosClient,
	}, nil
}

// Health runs the healthcheck for the cluster.
func (cluster *Cluster) Health(ctx context.Context) error {
	return retry.Constant(5*time.Minute, retry.WithUnits(10*time.Second)).Retry(func() error {
		// retry health checks as sometimes bootstrap bootkube issues break the check
		return retry.ExpectedError(cluster.health(ctx))
	})
}

func (cluster *Cluster) health(ctx context.Context) error {
	resp, err := cluster.client.ClusterHealthCheck(talosclient.WithNodes(ctx, cluster.controlPlaneNodes[0]), 3*time.Minute, &talosclusterapi.ClusterInfo{
		ControlPlaneNodes: cluster.controlPlaneNodes,
		WorkerNodes:       cluster.workerNodes,
	})
	if err != nil {
		return err
	}

	if err := resp.CloseSend(); err != nil {
		return err
	}

	for {
		msg, err := resp.Recv()
		if err != nil {
			if err == io.EOF || status.Code(err) == codes.Canceled { //nolint:errorlint
				return nil
			}

			return err
		}

		if msg.GetMetadata().GetError() != "" {
			return fmt.Errorf("healthcheck error: %s", msg.GetMetadata().GetError())
		}

		fmt.Fprintln(os.Stderr, msg.GetMessage())
	}
}

// Name of the cluster.
func (cluster *Cluster) Name() string {
	return cluster.name
}
