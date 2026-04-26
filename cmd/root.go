package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	awsclient "github.com/bikramdhoju/ecsctl/internal/aws"
	"github.com/bikramdhoju/ecsctl/internal/output"
)

var (
	clusterFlag string
	regionFlag  string
	profileFlag string
	roleARNFlag string
	outputFlag  string

	globalClients *awsclient.Clients
	globalPrinter *output.Printer
)

var rootCmd = &cobra.Command{
	Use:   "ecsctl",
	Short: "kubectl-like CLI for AWS ECS",
	Long:  `ecsctl — manage AWS ECS clusters, services, tasks, and containers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterFlag, "cluster", "", "ECS cluster name or ARN")
	rootCmd.PersistentFlags().StringVar(&regionFlag, "region", "eu-west-1", "AWS region")
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "AWS credentials profile (~/.aws/credentials)")
	rootCmd.PersistentFlags().StringVar(&roleARNFlag, "role-arn", "", "IAM role ARN to assume")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "table", "output format: table|wide|json|yaml")
}

// initClients builds AWS clients from flags and validates credentials.
// Used as PersistentPreRunE on commands that need AWS access.
func initClients(cmd *cobra.Command, args []string) error {
	clients, err := awsclient.NewClients(context.Background(), clusterFlag, regionFlag, profileFlag, roleARNFlag)
	if err != nil {
		return err
	}
	if err := clients.ValidateCredentials(context.Background()); err != nil {
		return err
	}
	globalClients = clients

	fmt, err := output.ParseFormat(outputFlag)
	if err != nil {
		return err
	}
	globalPrinter = output.New(fmt)
	return nil
}

func mustCluster() (string, error) {
	if globalClients.Cluster == "" {
		return "", fmt.Errorf("--cluster is required")
	}
	return globalClients.Cluster, nil
}
