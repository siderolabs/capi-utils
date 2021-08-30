// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package cmd contains all bootstrap CLI commands.
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/talos-systems/capi-utils/pkg/capi"
	"github.com/talos-systems/capi-utils/pkg/capi/infrastructure"
)

var awsOptions struct {
	b64EncodedCredentials string
}

var manager *capi.Manager

var bootstrapCmd = &cobra.Command{
	Use: "bootstrap",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		providers := make([]infrastructure.Provider, len(options.InfrastructureProviders))
		for i, name := range options.InfrastructureProviders {
			parts := strings.Split(name, ":")
			var version string
			if len(parts) > 1 {
				version = parts[1]
			}

			switch parts[0] {
			case infrastructure.AWSProviderName:
				providers[i] = infrastructure.NewAWSProvider(awsOptions.b64EncodedCredentials, version)
			default:
				return fmt.Errorf("unknown infrastructure provider type %s", name)
			}
		}

		var err error

		manager, err = capi.NewManager(ctx, capi.Options{
			ClusterctlConfigPath:    options.ClusterctlConfigPath,
			CoreProvider:            options.CoreProvider,
			BootstrapProviders:      options.BootstrapProviders,
			InfrastructureProviders: providers,
			ControlPlaneProviders:   options.ControlPlaneProviders,
		})
		if err != nil {
			return err
		}

		return manager.Install(ctx)
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.PersistentFlags().StringVar(&options.BootstrapClusterName, "bootstrap-cluster-name", options.BootstrapClusterName, "bootstrap cluster name")
	bootstrapCmd.PersistentFlags().StringSliceVar(&options.InfrastructureProviders, "infrastructure-providers", options.InfrastructureProviders, "infrastructure providers to set up")
	bootstrapCmd.PersistentFlags().StringVar(&options.ClusterctlConfigPath, "clusterctl-config", options.ClusterctlConfigPath, "path to the clusterctl config file")
	// AWS provider flags
	bootstrapCmd.PersistentFlags().StringVar(&awsOptions.b64EncodedCredentials, "aws-base64-encoded-credentials", awsOptions.b64EncodedCredentials, "AWS_B64ENCODED_CREDENTIALS")
}
