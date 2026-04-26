package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show detailed information about an ECS resource",
}

var describeClusterCmd = &cobra.Command{
	Use:               "cluster <name>",
	Short:             "Describe an ECS cluster",
	Args:              cobra.ExactArgs(1),
	PersistentPreRunE: initClients,
	RunE: func(cmd *cobra.Command, args []string) error {
		clusters, err := globalClients.ListClusters(context.Background())
		if err != nil {
			return err
		}
		for _, c := range clusters {
			if c.Name == args[0] || c.ARN == args[0] {
				return globalPrinter.DescribeCluster(&c)
			}
		}
		return fmt.Errorf("cluster %q not found", args[0])
	},
}

var (
	describeServiceCluster string
)

var describeServiceCmd = &cobra.Command{
	Use:               "service <name>",
	Short:             "Describe an ECS service",
	Aliases:           []string{"svc"},
	Args:              cobra.ExactArgs(1),
	PersistentPreRunE: initClients,
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster := describeServiceCluster
		if cluster == "" {
			var err error
			cluster, err = mustCluster()
			if err != nil {
				return err
			}
		}
		svc, err := globalClients.DescribeService(context.Background(), cluster, args[0])
		if err != nil {
			return err
		}
		return globalPrinter.DescribeService(svc)
	},
}

var (
	describeTaskCluster string
)

var describeTaskCmd = &cobra.Command{
	Use:               "task <task-id>",
	Short:             "Describe an ECS task",
	Args:              cobra.ExactArgs(1),
	PersistentPreRunE: initClients,
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster := describeTaskCluster
		if cluster == "" {
			var err error
			cluster, err = mustCluster()
			if err != nil {
				return err
			}
		}
		task, err := globalClients.DescribeTask(context.Background(), cluster, args[0])
		if err != nil {
			return err
		}
		return globalPrinter.DescribeTask(task)
	},
}

func init() {
	describeServiceCmd.Flags().StringVar(&describeServiceCluster, "cluster", "", "cluster name or ARN (overrides context)")
	describeTaskCmd.Flags().StringVar(&describeTaskCluster, "cluster", "", "cluster name or ARN (overrides context)")

	describeCmd.AddCommand(describeClusterCmd, describeServiceCmd, describeTaskCmd)
	rootCmd.AddCommand(describeCmd)
}
