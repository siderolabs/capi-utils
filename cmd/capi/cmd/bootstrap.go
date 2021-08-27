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
	"github.com/talos-systems/talos/pkg/cli"

	"github.com/talos-systems/capi-utils/pkg/capi"
	"github.com/talos-systems/capi-utils/pkg/capi/infrastructure"
)

var awsOptions struct {
	b64EncodedCredentials string
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Install and patch CAPI.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.WithContext(context.Background(), func(ctx context.Context) error {
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

			clusterAPI, err := capi.NewManager(ctx, capi.Options{
				ClusterctlConfigPath:    options.ClusterctlConfigPath,
				CoreProvider:            options.CoreProvider,
				BootstrapProviders:      options.BootstrapProviders,
				InfrastructureProviders: providers,
				ControlPlaneProviders:   options.ControlPlaneProviders,
			})
			if err != nil {
				return err
			}

			return clusterAPI.Install(ctx)
		})
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.Flags().StringVar(&options.BootstrapClusterName, "bootstrap-cluster-name", options.BootstrapClusterName, "bootstrap cluster name")
	bootstrapCmd.Flags().StringSliceVar(&options.InfrastructureProviders, "infrastructure-providers", options.InfrastructureProviders, "infrastructure providers to set up")
	bootstrapCmd.Flags().StringVar(&options.ClusterctlConfigPath, "clusterctl-config", options.ClusterctlConfigPath, "path to the clusterctl config file")
	// AWS provider flags
	bootstrapCmd.Flags().StringVar(&awsOptions.b64EncodedCredentials, "aws-base64-encoded-credentials", awsOptions.b64EncodedCredentials, "AWS_B64ENCODED_CREDENTIALS")
}
