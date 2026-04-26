package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "List ECS resources",
}

var getClustersCmd = &cobra.Command{
	Use:               "clusters",
	Short:             "List ECS clusters",
	Aliases:           []string{"cluster"},
	PersistentPreRunE: initClients,
	RunE: func(cmd *cobra.Command, args []string) error {
		clusters, err := globalClients.ListClusters(context.Background())
		if err != nil {
			return err
		}
		if len(clusters) == 0 {
			fmt.Println("No clusters found.")
			return nil
		}
		return globalPrinter.Clusters(clusters)
	},
}

var getServicesCmd = &cobra.Command{
	Use:               "services",
	Short:             "List ECS services",
	Aliases:           []string{"service", "svc"},
	PersistentPreRunE: initClients,
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster, err := mustCluster()
		if err != nil {
			return err
		}
		services, err := globalClients.ListServices(context.Background(), cluster)
		if err != nil {
			return err
		}
		if len(services) == 0 {
			fmt.Println("No services found.")
			return nil
		}
		return globalPrinter.Services(services)
	},
}

var (
	getTasksService string
	getTasksStatus  string
)

var getTasksCmd = &cobra.Command{
	Use:               "tasks",
	Short:             "List ECS tasks",
	Aliases:           []string{"task"},
	PersistentPreRunE: initClients,
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster, err := mustCluster()
		if err != nil {
			return err
		}
		tasks, err := globalClients.ListTasks(context.Background(), cluster, getTasksService, getTasksStatus)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}
		return globalPrinter.Tasks(tasks)
	},
}

func init() {
	getTasksCmd.Flags().StringVarP(&getTasksService, "service", "s", "", "filter by service name")
	getTasksCmd.Flags().StringVar(&getTasksStatus, "status", "RUNNING", "desired status filter: RUNNING|STOPPED")

	getCmd.AddCommand(getClustersCmd, getServicesCmd, getTasksCmd)
	rootCmd.AddCommand(getCmd)
}
