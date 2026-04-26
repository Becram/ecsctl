package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	awsclient "github.com/bikramdhoju/ecsctl/internal/aws"
)

var (
	logsContainer string
	logsFollow    bool
	logsTail      int32
	logsSince     string
)

var logsCmd = &cobra.Command{
	Use:               "logs <task-id>",
	Short:             "Fetch logs for an ECS task container",
	Args:              cobra.ExactArgs(1),
	PersistentPreRunE: initClients,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		cluster, err := mustCluster()
		if err != nil {
			return err
		}

		ctx := context.Background()

		task, err := globalClients.DescribeTask(ctx, cluster, taskID)
		if err != nil {
			return err
		}

		containerName := logsContainer
		if containerName == "" {
			if len(task.Containers) == 0 {
				return fmt.Errorf("task has no containers")
			}
			if len(task.Containers) > 1 {
				names := make([]string, 0, len(task.Containers))
				for _, c := range task.Containers {
					names = append(names, c.Name)
				}
				return fmt.Errorf("task has multiple containers; specify one with -c: %v", names)
			}
			containerName = task.Containers[0].Name
		}

		logConfigs, err := globalClients.GetContainerLogConfigs(ctx, task.TaskDefinition)
		if err != nil {
			return err
		}

		cfg, ok := logConfigs[containerName]
		if !ok {
			return fmt.Errorf("container %q does not use awslogs driver (check task definition logConfiguration)", containerName)
		}
		if cfg.Group == "" {
			return fmt.Errorf("awslogs-group not set for container %q", containerName)
		}

		stream := awsclient.LogStreamName(cfg.StreamPrefix, containerName, task.TaskID)

		var since time.Time
		if logsSince != "" {
			d, err := time.ParseDuration(logsSince)
			if err != nil {
				return fmt.Errorf("invalid --since value %q: use Go duration format e.g. 1h, 30m", logsSince)
			}
			since = time.Now().Add(-d)
		}

		fmt.Fprintf(os.Stderr, "Fetching logs from %s/%s\n", cfg.Group, stream)

		if logsFollow {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			return globalClients.FetchLogs(ctx, cfg.Group, stream, logsTail, since, true, os.Stdout)
		}
		return globalClients.FetchLogs(context.Background(), cfg.Group, stream, logsTail, since, false, os.Stdout)
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "container name (required if task has multiple containers)")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "stream logs")
	logsCmd.Flags().Int32Var(&logsTail, "tail", 100, "number of lines to show from the end")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "show logs since duration (e.g. 1h, 30m)")

	rootCmd.AddCommand(logsCmd)
}
