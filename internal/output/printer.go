package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bikramdhoju/ecsctl/internal/model"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatWide  Format = "wide"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "table", "":
		return FormatTable, nil
	case "wide":
		return FormatWide, nil
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	default:
		return FormatTable, fmt.Errorf("unknown output format %q; valid: table, wide, json, yaml", s)
	}
}

type Printer struct {
	format Format
	out    io.Writer
}

func New(format Format) *Printer {
	return &Printer{format: format, out: os.Stdout}
}

func (p *Printer) Clusters(clusters []model.Cluster) error {
	switch p.format {
	case FormatJSON:
		return jsonOut(p.out, clusters)
	case FormatYAML:
		return yamlOut(p.out, clusters)
	}
	w := newTab(p.out)
	fmt.Fprintln(w, "NAME\tSTATUS\tRUNNING\tPENDING\tSERVICES\t")
	for _, c := range clusters {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t\n",
			c.Name, c.Status, c.RunningTasksCount, c.PendingTasksCount, c.ActiveServicesCount)
	}
	return w.Flush()
}

func (p *Printer) Services(services []model.Service) error {
	switch p.format {
	case FormatJSON:
		return jsonOut(p.out, services)
	case FormatYAML:
		return yamlOut(p.out, services)
	}
	w := newTab(p.out)
	if p.format == FormatWide {
		fmt.Fprintln(w, "NAME\tDESIRED\tRUNNING\tPENDING\tTASK DEFINITION\tLAUNCH TYPE\tSTATUS\tCLUSTER\t")
	} else {
		fmt.Fprintln(w, "NAME\tDESIRED\tRUNNING\tPENDING\tTASK DEFINITION\tLAUNCH TYPE\tSTATUS\t")
	}
	for _, s := range services {
		row := fmt.Sprintf("%s\t%d\t%d\t%d\t%s\t%s\t%s\t",
			s.Name, s.DesiredCount, s.RunningCount, s.PendingCount,
			s.TaskDefinition, s.LaunchType, s.Status)
		if p.format == FormatWide {
			row += shortARN(s.ClusterARN) + "\t"
		}
		fmt.Fprintln(w, row)
	}
	return w.Flush()
}

func (p *Printer) Tasks(tasks []model.Task) error {
	switch p.format {
	case FormatJSON:
		return jsonOut(p.out, tasks)
	case FormatYAML:
		return yamlOut(p.out, tasks)
	}
	w := newTab(p.out)
	if p.format == FormatWide {
		fmt.Fprintln(w, "TASK ID\tSTATUS\tDESIRED\tSERVICE\tTASK DEFINITION\tLAUNCH TYPE\tSTARTED\tHEALTH\t")
	} else {
		fmt.Fprintln(w, "TASK ID\tSTATUS\tDESIRED\tSERVICE\tTASK DEFINITION\tLAUNCH TYPE\t")
	}
	for _, t := range tasks {
		row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t",
			shortID(t.TaskID), t.Status, t.DesiredStatus,
			t.ServiceName, t.TaskDefinition, t.LaunchType)
		if p.format == FormatWide {
			row += relativeTime(t.StartedAt) + "\t" + t.HealthStatus + "\t"
		}
		fmt.Fprintln(w, row)
	}
	return w.Flush()
}

func (p *Printer) DescribeCluster(c *model.Cluster) error {
	switch p.format {
	case FormatJSON:
		return jsonOut(p.out, c)
	case FormatYAML:
		return yamlOut(p.out, c)
	}
	pf(p.out, "Name", c.Name)
	pf(p.out, "ARN", c.ARN)
	pf(p.out, "Status", c.Status)
	pf(p.out, "Running Tasks", fmt.Sprintf("%d", c.RunningTasksCount))
	pf(p.out, "Pending Tasks", fmt.Sprintf("%d", c.PendingTasksCount))
	pf(p.out, "Active Services", fmt.Sprintf("%d", c.ActiveServicesCount))
	if len(c.CapacityProviders) > 0 {
		pf(p.out, "Capacity Providers", strings.Join(c.CapacityProviders, ", "))
	}
	return nil
}

func (p *Printer) DescribeService(s *model.Service) error {
	switch p.format {
	case FormatJSON:
		return jsonOut(p.out, s)
	case FormatYAML:
		return yamlOut(p.out, s)
	}
	pf(p.out, "Name", s.Name)
	pf(p.out, "Cluster", shortARN(s.ClusterARN))
	pf(p.out, "Status", s.Status)
	pf(p.out, "Task Definition", s.TaskDefinition)
	pf(p.out, "Launch Type", s.LaunchType)
	pf(p.out, "Desired", fmt.Sprintf("%d", s.DesiredCount))
	pf(p.out, "Running", fmt.Sprintf("%d", s.RunningCount))
	pf(p.out, "Pending", fmt.Sprintf("%d", s.PendingCount))
	if s.CreatedAt != nil {
		pf(p.out, "Created", s.CreatedAt.Local().Format(time.RFC3339))
	}
	if s.UpdatedAt != nil {
		pf(p.out, "Updated", fmt.Sprintf("%s (%s)", relativeTime(s.UpdatedAt), s.UpdatedAt.Local().Format(time.RFC3339)))
	}

	if len(s.Deployments) > 0 {
		fmt.Fprintln(p.out, "\nDeployments:")
		w := newTab(p.out)
		fmt.Fprintln(w, "  ID\tSTATUS\tDESIRED\tRUNNING\tPENDING\t")
		for _, d := range s.Deployments {
			fmt.Fprintf(w, "  %s\t%s\t%d\t%d\t%d\t\n", d.ID, d.Status, d.DesiredCount, d.RunningCount, d.PendingCount)
		}
		w.Flush() //nolint:errcheck
	}

	if len(s.LoadBalancers) > 0 {
		fmt.Fprintln(p.out, "\nLoad Balancers:")
		w := newTab(p.out)
		fmt.Fprintln(w, "  TARGET GROUP\tCONTAINER\tPORT\t")
		for _, lb := range s.LoadBalancers {
			fmt.Fprintf(w, "  %s\t%s\t%d\t\n", shortARN(lb.TargetGroupARN), lb.ContainerName, lb.ContainerPort)
		}
		w.Flush() //nolint:errcheck
	}

	if len(s.Events) > 0 {
		fmt.Fprintln(p.out, "\nEvents (last 10):")
		w := newTab(p.out)
		fmt.Fprintln(w, "  TIME\tMESSAGE\t")
		limit := 10
		if len(s.Events) < limit {
			limit = len(s.Events)
		}
		for _, e := range s.Events[:limit] {
			fmt.Fprintf(w, "  %s\t%s\t\n", relativeTime(&e.CreatedAt), e.Message)
		}
		w.Flush() //nolint:errcheck
	}
	return nil
}

func (p *Printer) DescribeTask(task *model.Task) error {
	switch p.format {
	case FormatJSON:
		return jsonOut(p.out, task)
	case FormatYAML:
		return yamlOut(p.out, task)
	}
	pf(p.out, "Task ID", shortID(task.TaskID))
	pf(p.out, "ARN", task.TaskARN)
	pf(p.out, "Cluster", shortARN(task.ClusterARN))
	if task.ServiceName != "" {
		pf(p.out, "Service", task.ServiceName)
	}
	pf(p.out, "Task Definition", task.TaskDefinition)
	pf(p.out, "Status", task.Status)
	pf(p.out, "Desired Status", task.DesiredStatus)
	pf(p.out, "Health", task.HealthStatus)
	pf(p.out, "Launch Type", task.LaunchType)
	pf(p.out, "CPU", task.CPU)
	pf(p.out, "Memory", task.Memory)
	pf(p.out, "ECS Exec Enabled", fmt.Sprintf("%v", task.EnableExecCommand))
	if task.StartedAt != nil {
		pf(p.out, "Started", task.StartedAt.Local().Format(time.RFC3339))
	}
	if task.StoppedAt != nil {
		pf(p.out, "Stopped", task.StoppedAt.Local().Format(time.RFC3339))
	}
	if task.StoppedReason != "" {
		pf(p.out, "Stopped Reason", task.StoppedReason)
	}

	if len(task.Containers) > 0 {
		fmt.Fprintln(p.out, "\nContainers:")
		w := newTab(p.out)
		fmt.Fprintln(w, "  NAME\tSTATUS\tEXIT CODE\tHEALTH\tIMAGE\t")
		for _, c := range task.Containers {
			exitCode := "-"
			if c.ExitCode != nil {
				exitCode = fmt.Sprintf("%d", *c.ExitCode)
			}
			name := c.Name
			if c.Reason != "" {
				name = c.Name + " (" + c.Reason + ")"
			}
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t\n", name, c.Status, exitCode, c.HealthStatus, c.Image)
		}
		w.Flush() //nolint:errcheck
	}
	return nil
}

// --- helpers ---

func newTab(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

func pf(w io.Writer, key, value string) {
	fmt.Fprintf(w, "%-25s %s\n", key+":", value)
}

func shortARN(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return arn
}

func shortID(id string) string {
	parts := strings.Split(id, "/")
	return parts[len(parts)-1]
}

func relativeTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	d := time.Since(*t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func jsonOut(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func yamlOut(w io.Writer, v any) error {
	return yaml.NewEncoder(w).Encode(v)
}
