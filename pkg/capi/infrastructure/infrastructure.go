// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package infrastructure contains infrastructure providers setup hooks and ready conditions.
package infrastructure

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
)

// Provider defines an interface for the infrastructure provider.
type Provider interface {
	Name() string
	Version() string
	PreInstall() error
	IsInstalled(ctx context.Context, clientset *kubernetes.Clientset) (bool, error)
	GetClusterTemplate(client.Client, client.GetClusterTemplateOptions, interface{}) (client.Template, error)
	WaitReady(context.Context, *kubernetes.Clientset) error
}