// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package cmd contains all bootstrap CLI commands.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/siderolabs/capi-utils/pkg/capi"
)

var setupOptions = map[string]interface{}{}

var awsOptions struct {
	b64EncodedCredentials string
}

var manager *capi.Manager

var bootstrapCmd = &cobra.Command{
	Use: "bootstrap",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.PersistentFlags().StringVar(&options.ClusterctlConfigPath, "clusterctl-config", options.ClusterctlConfigPath, "path to the clusterctl config file")
}
