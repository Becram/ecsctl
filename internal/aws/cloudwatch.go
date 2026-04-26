package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// FetchLogs retrieves log events from CloudWatch Logs.
// If follow is true, it polls continuously until ctx is cancelled.
func (c *Clients) FetchLogs(ctx context.Context, group, stream string, tail int32, since time.Time, follow bool, out io.Writer) error {
	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:   &group,
		LogStreamNames: []string{stream},
	}
	if !since.IsZero() {
		ms := since.UnixMilli()
		input.StartTime = &ms
	}
	if !follow && tail > 0 {
		input.Limit = &tail
	}

	var lastTimestamp int64

	for {
		if lastTimestamp > 0 {
			next := lastTimestamp + 1
			input.StartTime = &next
			input.Limit = nil
		}

		resp, err := c.CWL.FilterLogEvents(ctx, input)
		if err != nil {
			return fmt.Errorf("fetch logs from %s/%s: %w", group, stream, err)
		}

		for _, e := range resp.Events {
			ts := time.UnixMilli(aws.ToInt64(e.Timestamp)).Local().Format("2006-01-02T15:04:05")
			fmt.Fprintf(out, "%s  %s\n", ts, aws.ToString(e.Message))
			if aws.ToInt64(e.Timestamp) > lastTimestamp {
				lastTimestamp = aws.ToInt64(e.Timestamp)
			}
		}

		if !follow {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
}

// LogStreamName constructs the CloudWatch log stream name for an ECS container.
// Standard ECS awslogs format: {prefix}/{containerName}/{shortTaskID}
func LogStreamName(prefix, containerName, taskID string) string {
	return fmt.Sprintf("%s/%s/%s", prefix, containerName, taskID)
}
