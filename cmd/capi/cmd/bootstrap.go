// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package cmd contains all bootstrap CLI commands.
package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/talos-systems/capi-utils/pkg/capi"
	"github.com/talos-systems/capi-utils/pkg/capi/infrastructure"
)

var setupOptions = map[string]interface{}{}

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
			provider, err := infrastructure.NewProvider(name)
			if err != nil {
				return err
			}

			if opts, ok := setupOptions[provider.Name()]; ok {
				if err = provider.Configure(opts); err != nil {
					return err
				}
			}

			providers[i] = provider
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
	opts := infrastructure.NewAWSSetupOptions()
	bootstrapCmd.PersistentFlags().StringVar(&opts.AWSCredentials, "aws-base64-encoded-credentials", awsOptions.b64EncodedCredentials, "AWS_B64ENCODED_CREDENTIALS")
	setupOptions[infrastructure.AWSProviderName] = opts
}
