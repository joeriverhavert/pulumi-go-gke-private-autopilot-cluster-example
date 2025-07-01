package main

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/secretmanager"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type cluster struct {
	name           string
	network        string
	subnetwork     string
	location       string
	releaseChannel string
	networkPolicy  networkPolicy
	project        string
}

type networkPolicy struct {
	master_ipv4_cidr_block    string
	secondairy_pod_range_name string
	secondairy_svc_range_name string
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Input details
		cluster := cluster{
			name:           "autopilot-mgmt-sbx",
			network:        "conro-sbx",
			subnetwork:     "cnr-sbx-sub",
			location:       "europe-west1",
			releaseChannel: "STABLE",
			networkPolicy: networkPolicy{
				master_ipv4_cidr_block:    "10.4.0.0/28",
				secondairy_pod_range_name: "cnr-sbx1-pod-sub-c2",
				secondairy_svc_range_name: "cnr-sbx1-svc-sub-c2",
			},
			project: "conro-sbx",
		}

		// Create service account
		svc_obj, err := serviceaccount.NewAccount(
			ctx,
			"gke-autopilot-serviceaccount",
			&serviceaccount.AccountArgs{
				AccountId:   pulumi.String(fmt.Sprintf("sa-gke-%s", cluster.name)),
				DisplayName: pulumi.String(fmt.Sprintf("sa-gke-%s", cluster.name)),
				Description: pulumi.String(
					fmt.Sprintf("Service account for %s cluster", cluster.name),
				),
				Disabled: pulumi.Bool(false),
			},
		)
		if err != nil {
			panic(err)
		}

		cluster_obj, err := container.NewCluster(
			ctx, cluster.name,
			&container.ClusterArgs{
				Name: pulumi.String(cluster.name),

				// Network
				Network: pulumi.String(
					fmt.Sprintf(
						"https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s",
						cluster.project,
						cluster.network,
					),
				),
				Subnetwork: pulumi.String(
					fmt.Sprintf(
						"https://www.googleapis.com/compute/v1/projects/%s/regions/europe-west1/subnetworks/%s",
						cluster.project,
						cluster.subnetwork,
					),
				),
				Location: pulumi.String(cluster.location),

				// Release Channel
				ReleaseChannel: &container.ClusterReleaseChannelArgs{
					Channel: pulumi.String(cluster.releaseChannel),
				},

				// Extra configs
				EnableAutopilot:                      pulumi.Bool(true),
				EnableCiliumClusterwideNetworkPolicy: pulumi.Bool(true),
				EnableFqdnNetworkPolicy:              pulumi.Bool(true),
				EnableL4IlbSubsetting:                pulumi.Bool(true),
				EnableMultiNetworking:                pulumi.Bool(true),
				DeletionProtection:                   pulumi.Bool(false),

				// Private Cluster Config
				PrivateClusterConfig: &container.ClusterPrivateClusterConfigArgs{
					EnablePrivateNodes:    pulumi.Bool(true),
					EnablePrivateEndpoint: pulumi.Bool(true),
					MasterIpv4CidrBlock: pulumi.String(
						cluster.networkPolicy.master_ipv4_cidr_block,
					),
				},

				// Authorized allowed networks
				MasterAuthorizedNetworksConfig: &container.ClusterMasterAuthorizedNetworksConfigArgs{
					CidrBlocks: container.ClusterMasterAuthorizedNetworksConfigCidrBlockArray{
						&container.ClusterMasterAuthorizedNetworksConfigCidrBlockArgs{
							CidrBlock:   pulumi.String("10.0.0.0/8"),
							DisplayName: pulumi.String("RFC1918"),
						},
						&container.ClusterMasterAuthorizedNetworksConfigCidrBlockArgs{
							CidrBlock:   pulumi.String("172.16.0.0/12"),
							DisplayName: pulumi.String("RFC1918"),
						},
						&container.ClusterMasterAuthorizedNetworksConfigCidrBlockArgs{
							CidrBlock:   pulumi.String("192.168.0.0/16"),
							DisplayName: pulumi.String("RFC1918"),
						},
					},
				},

				// IpAllocationPolicy
				IpAllocationPolicy: &container.ClusterIpAllocationPolicyArgs{
					StackType: pulumi.String("IPV4"),
					ClusterSecondaryRangeName: pulumi.String(
						cluster.networkPolicy.secondairy_pod_range_name,
					),
					ServicesSecondaryRangeName: pulumi.String(
						cluster.networkPolicy.secondairy_svc_range_name,
					),
				},

				// Addons
				AddonsConfig: &container.ClusterAddonsConfigArgs{
					HttpLoadBalancing: container.ClusterAddonsConfigHttpLoadBalancingArgs{
						Disabled: pulumi.Bool(false),
					},
				},
			},
		)
		if err != nil {
			panic(err)
		}

		// Create kubeconfig secret
		secretName := fmt.Sprintf("kubeconfig-%s", cluster.name)
		secret_obj, err := secretmanager.NewSecret(ctx, secretName, &secretmanager.SecretArgs{
			SecretId: pulumi.String(secretName),
			Replication: &secretmanager.SecretReplicationArgs{
				Auto: &secretmanager.SecretReplicationAutoArgs{},
			},
		})
		if err != nil {
			panic(err)
		}

		// Create secret version
		secretData := pulumi.All(
			cluster_obj.MasterAuth.ClusterCaCertificate(),
			cluster_obj.Endpoint,
			cluster_obj.Name,
		).ApplyT(func(args []interface{}) string {
			ca := *args[0].(*string)
			endpoint := args[1]
			name := args[2]

			return fmt.Sprintf(`
---
 apiVersion: v1
 clusters:
 - cluster:
     certificate-authority-data: %s
     server: https://%s
   name: %s
 contexts:
 - context:
     cluster: %s
     user: %s
   name: %s
 current-context: %s
 kind: Config
 preferences: {}
 users:
 - name: %s
   user:
     exec:
       apiVersion: client.authentication.k8s.io/v1beta1
       command: gke-gcloud-auth-plugin
       installHint: Install gke-gcloud-auth-plugin for use with kubectl by following
         https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin
       provideClusterInfo: true

 `, ca, endpoint, name, name, name, name, name, name)
		}).(pulumi.StringOutput)

		secretVersion_obj, err := secretmanager.NewSecretVersion(
			ctx,
			secretName+"-version",
			&secretmanager.SecretVersionArgs{
				Enabled:    pulumi.Bool(true),
				Secret:     secret_obj.ID(),
				SecretData: secretData,
			},
		)
		if err != nil {
			panic(err)
		}

		ctx.Export("Service Account", svc_obj.AccountId)
		ctx.Export("Cluster:", cluster_obj.Name)
		ctx.Export("Secret:", secret_obj.Name)
		ctx.Export("SecretVersion:", secretVersion_obj.Name)
		return nil
	})
}
