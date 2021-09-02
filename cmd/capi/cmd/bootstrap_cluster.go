// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/talos-systems/capi-utils/pkg/capi"
	"github.com/talos-systems/capi-utils/pkg/capi/infrastructure"
)

var clusterCmdFlags struct {
	clusterName  string
	templatePath string
}

var deployOptions = capi.DefaultDeployOptions()

var awsDeployOptions = infrastructure.NewAWSDeployOptions()

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Deploy a cluster using CAPI.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		opts := []capi.DeployOption{
			capi.WithDeployOptions(deployOptions),
		}

		if deployOptions.Provider == "" || deployOptions.Provider == infrastructure.AWSProviderName {
			if clusterCmdFlags.templatePath == "" {
				deployOptions.Template = infrastructure.AWSTalosTemplate
			}

			opts = append(
				opts,
				capi.WithProviderOptions(awsDeployOptions),
			)
		}

		if clusterCmdFlags.templatePath != "" {
			opts = append(opts, capi.WithTemplateFile(clusterCmdFlags.templatePath))
		}

		cluster, err := manager.DeployCluster(ctx, clusterCmdFlags.clusterName, opts...)
		if err != nil {
			return err
		}

		return cluster.Health(ctx)
	},
}

func init() {
	bootstrapCmd.AddCommand(clusterCmd)

	clusterCmd.Flags().StringVarP(&clusterCmdFlags.clusterName, "name", "n", "talos-default", "Created CAPI cluster name")
	clusterCmd.Flags().StringVarP(&clusterCmdFlags.templatePath, "from", "f", "", "Custom path for the cluster template")
	clusterCmd.Flags().Int64Var(&deployOptions.ControlPlaneNodes, "control-plane-nodes", deployOptions.ControlPlaneNodes, "Number of control plane nodes to deploy")
	clusterCmd.Flags().Int64Var(&deployOptions.WorkerNodes, "worker-nodes", deployOptions.WorkerNodes, "Number of worker nodes to deploy")
	clusterCmd.Flags().StringVarP(&deployOptions.Provider, "provider", "p", deployOptions.Provider, "Infrastructure provider to use for the deployment")
	clusterCmd.Flags().StringVar(&deployOptions.ProviderVersion, "provider-version", deployOptions.ProviderVersion, "Provider version to use")
	clusterCmd.Flags().StringVar(&deployOptions.KubernetesVersion, "kubernetes-version", deployOptions.KubernetesVersion, "Kubernetes version to use")
	clusterCmd.Flags().StringVar(&deployOptions.TalosVersion, "talos-version", deployOptions.TalosVersion, "Talos version to use")
	// AWS provider flags
	clusterCmd.Flags().StringVar(&awsDeployOptions.CloudProviderVersion, "aws-cloud-provider-version", awsDeployOptions.CloudProviderVersion, "AWS cloud provider version")

	clusterCmd.Flags().StringVar(&awsDeployOptions.ControlPlaneAMIID, "aws-cp-ami-id", awsDeployOptions.ControlPlaneAMIID, "AWS AMI ID for control plane nodes")
	clusterCmd.Flags().StringVar(&awsDeployOptions.ControlPlaneADDLSecGroups, "aws-cp-addl-sec-groups", awsDeployOptions.ControlPlaneADDLSecGroups, "AWS control plane ADDL sec groups")
	clusterCmd.Flags().StringVar(&awsDeployOptions.ControlPlaneIAMProfile, "aws-cp-iam-profile", awsDeployOptions.ControlPlaneIAMProfile, "AWS control plane IAM profile")
	clusterCmd.Flags().StringVar(&awsDeployOptions.ControlPlaneMachineType, "aws-cp-machine-type", awsDeployOptions.ControlPlaneMachineType, "AWS control plane machine type")
	clusterCmd.Flags().Int64Var(&awsDeployOptions.ControlPlaneVolSize, "aws-cp-vol-size", awsDeployOptions.ControlPlaneVolSize, "AWS control plane vol size")

	clusterCmd.Flags().StringVar(&awsDeployOptions.NodeAMIID, "aws-worker-ami-id", awsDeployOptions.NodeAMIID, "AWS AMI ID for worker nodes")
	clusterCmd.Flags().StringVar(&awsDeployOptions.NodeADDLSecGroups, "aws-worker-addl-sec-groups", awsDeployOptions.NodeADDLSecGroups, "AWS worker ADDL sec groups")
	clusterCmd.Flags().StringVar(&awsDeployOptions.NodeIAMProfile, "aws-worker-iam-profile", awsDeployOptions.NodeIAMProfile, "AWS worker IAM profile")
	clusterCmd.Flags().StringVar(&awsDeployOptions.NodeMachineType, "aws-worker-machine-type", awsDeployOptions.NodeMachineType, "AWS worker machine type")
	clusterCmd.Flags().Int64Var(&awsDeployOptions.NodeVolSize, "aws-worker-vol-size", awsDeployOptions.NodeVolSize, "AWS worker vol size")

	clusterCmd.Flags().StringVar(&awsDeployOptions.Region, "aws-region", awsDeployOptions.Region, "AWS region")
	clusterCmd.Flags().StringVar(&awsDeployOptions.Subnet, "aws-subnet", awsDeployOptions.NodeADDLSecGroups, "AWS subnet")
	clusterCmd.Flags().StringVar(&awsDeployOptions.SSHKeyName, "aws-ssh-key-name", awsDeployOptions.SSHKeyName, "AWS ssh key name")
	clusterCmd.Flags().StringVar(&awsDeployOptions.VPCID, "aws-vpc-id", awsDeployOptions.VPCID, "AWS VPC ID")
}
