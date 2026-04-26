package aws

import (
	"context"
	"fmt"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	ecsSDK "github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/bikramdhoju/ecsctl/internal/model"
)

func (c *Clients) ListClusters(ctx context.Context) ([]model.Cluster, error) {
	var arns []string
	pager := ecsSDK.NewListClustersPaginator(c.ECS, &ecsSDK.ListClustersInput{})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list clusters: %w", err)
		}
		arns = append(arns, page.ClusterArns...)
	}
	if len(arns) == 0 {
		return nil, nil
	}
	out, err := c.ECS.DescribeClusters(ctx, &ecsSDK.DescribeClustersInput{Clusters: arns})
	if err != nil {
		return nil, fmt.Errorf("describe clusters: %w", err)
	}
	result := make([]model.Cluster, 0, len(out.Clusters))
	for _, cl := range out.Clusters {
		result = append(result, clusterFromSDK(cl))
	}
	return result, nil
}

func (c *Clients) ListServices(ctx context.Context, cluster string) ([]model.Service, error) {
	cluster = c.resolveCluster(cluster)
	var arns []string
	pager := ecsSDK.NewListServicesPaginator(c.ECS, &ecsSDK.ListServicesInput{
		Cluster: &cluster,
	})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list services in %s: %w", cluster, err)
		}
		arns = append(arns, page.ServiceArns...)
	}
	if len(arns) == 0 {
		return nil, nil
	}
	return c.describeServicesByARN(ctx, cluster, arns)
}

func (c *Clients) DescribeService(ctx context.Context, cluster, service string) (*model.Service, error) {
	cluster = c.resolveCluster(cluster)
	out, err := c.ECS.DescribeServices(ctx, &ecsSDK.DescribeServicesInput{
		Cluster:  &cluster,
		Services: []string{service},
	})
	if err != nil {
		return nil, fmt.Errorf("describe service %s: %w", service, err)
	}
	if len(out.Services) == 0 {
		return nil, fmt.Errorf("service %q not found in cluster %q", service, cluster)
	}
	svc := serviceFromSDK(out.Services[0])
	return &svc, nil
}

func (c *Clients) ListTasks(ctx context.Context, cluster, service, status string) ([]model.Task, error) {
	cluster = c.resolveCluster(cluster)
	input := &ecsSDK.ListTasksInput{Cluster: &cluster}
	if service != "" {
		input.ServiceName = &service
	}
	if status != "" {
		input.DesiredStatus = types.DesiredStatus(strings.ToUpper(status))
	}

	var arns []string
	pager := ecsSDK.NewListTasksPaginator(c.ECS, input)
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list tasks: %w", err)
		}
		arns = append(arns, page.TaskArns...)
	}
	if len(arns) == 0 {
		return nil, nil
	}
	return c.describeTasks(ctx, cluster, arns)
}

func (c *Clients) DescribeTask(ctx context.Context, cluster, taskID string) (*model.Task, error) {
	cluster = c.resolveCluster(cluster)
	tasks, err := c.describeTasks(ctx, cluster, []string{taskID})
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("task %q not found in cluster %q", taskID, cluster)
	}
	return &tasks[0], nil
}

// ContainerLogConfig holds log configuration derived from a task definition.
type ContainerLogConfig struct {
	Group        string
	Region       string
	StreamPrefix string
}

// GetContainerLogConfigs fetches awslogs configuration for all containers
// in the given task definition ARN. Returns map[containerName]config.
func (c *Clients) GetContainerLogConfigs(ctx context.Context, taskDefARN string) (map[string]ContainerLogConfig, error) {
	out, err := c.ECS.DescribeTaskDefinition(ctx, &ecsSDK.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefARN,
	})
	if err != nil {
		return nil, fmt.Errorf("describe task definition %s: %w", taskDefARN, err)
	}
	configs := make(map[string]ContainerLogConfig)
	for _, cd := range out.TaskDefinition.ContainerDefinitions {
		if cd.LogConfiguration == nil || cd.LogConfiguration.LogDriver != types.LogDriverAwslogs {
			continue
		}
		opts := cd.LogConfiguration.Options
		region := opts["awslogs-region"]
		if region == "" {
			region = c.Region
		}
		configs[sdkaws.ToString(cd.Name)] = ContainerLogConfig{
			Group:        opts["awslogs-group"],
			Region:       region,
			StreamPrefix: opts["awslogs-stream-prefix"],
		}
	}
	return configs, nil
}

// --- internal ---

func (c *Clients) resolveCluster(cluster string) string {
	if cluster != "" {
		return cluster
	}
	return c.Cluster
}

func (c *Clients) describeServicesByARN(ctx context.Context, cluster string, arns []string) ([]model.Service, error) {
	var services []model.Service
	// DescribeServices accepts max 10
	for i := 0; i < len(arns); i += 10 {
		end := i + 10
		if end > len(arns) {
			end = len(arns)
		}
		out, err := c.ECS.DescribeServices(ctx, &ecsSDK.DescribeServicesInput{
			Cluster:  &cluster,
			Services: arns[i:end],
		})
		if err != nil {
			return nil, fmt.Errorf("describe services: %w", err)
		}
		for _, s := range out.Services {
			services = append(services, serviceFromSDK(s))
		}
	}
	return services, nil
}

func (c *Clients) describeTasks(ctx context.Context, cluster string, arns []string) ([]model.Task, error) {
	var tasks []model.Task
	// DescribeTasks accepts max 100
	for i := 0; i < len(arns); i += 100 {
		end := i + 100
		if end > len(arns) {
			end = len(arns)
		}
		out, err := c.ECS.DescribeTasks(ctx, &ecsSDK.DescribeTasksInput{
			Cluster: &cluster,
			Tasks:   arns[i:end],
		})
		if err != nil {
			return nil, fmt.Errorf("describe tasks: %w", err)
		}
		for _, t := range out.Tasks {
			tasks = append(tasks, taskFromSDK(t))
		}
	}
	return tasks, nil
}

// --- SDK type converters ---

func clusterFromSDK(c types.Cluster) model.Cluster {
	return model.Cluster{
		Name:                              sdkaws.ToString(c.ClusterName),
		ARN:                               sdkaws.ToString(c.ClusterArn),
		Status:                            sdkaws.ToString(c.Status),
		RunningTasksCount:                 c.RunningTasksCount,
		PendingTasksCount:                 c.PendingTasksCount,
		ActiveServicesCount:               c.ActiveServicesCount,
		RegisteredContainerInstancesCount: c.RegisteredContainerInstancesCount,
		CapacityProviders:                 c.CapacityProviders,
	}
}

func serviceFromSDK(s types.Service) model.Service {
	svc := model.Service{
		Name:           sdkaws.ToString(s.ServiceName),
		ClusterARN:     sdkaws.ToString(s.ClusterArn),
		Status:         sdkaws.ToString(s.Status),
		DesiredCount:   s.DesiredCount,
		RunningCount:   s.RunningCount,
		PendingCount:   s.PendingCount,
		TaskDefinition: shortName(sdkaws.ToString(s.TaskDefinition)),
		LaunchType:     string(s.LaunchType),
		CreatedAt:      s.CreatedAt,
	}
	for _, d := range s.Deployments {
		dep := model.Deployment{
			ID:           sdkaws.ToString(d.Id),
			Status:       sdkaws.ToString(d.Status),
			DesiredCount: d.DesiredCount,
			RunningCount: d.RunningCount,
			PendingCount: d.PendingCount,
			LaunchType:   string(d.LaunchType),
			CreatedAt:    d.CreatedAt,
			UpdatedAt:    d.UpdatedAt,
		}
		svc.Deployments = append(svc.Deployments, dep)
		if sdkaws.ToString(d.Status) == "PRIMARY" {
			svc.UpdatedAt = d.UpdatedAt
		}
	}
	for _, lb := range s.LoadBalancers {
		port := int32(0)
		if lb.ContainerPort != nil {
			port = *lb.ContainerPort
		}
		svc.LoadBalancers = append(svc.LoadBalancers, model.LoadBalancer{
			TargetGroupARN: sdkaws.ToString(lb.TargetGroupArn),
			ContainerName:  sdkaws.ToString(lb.ContainerName),
			ContainerPort:  port,
		})
	}
	for _, e := range s.Events {
		if e.CreatedAt == nil {
			continue
		}
		svc.Events = append(svc.Events, model.ServiceEvent{
			CreatedAt: *e.CreatedAt,
			Message:   sdkaws.ToString(e.Message),
		})
	}
	return svc
}

func taskFromSDK(t types.Task) model.Task {
	task := model.Task{
		TaskID:            shortTaskID(sdkaws.ToString(t.TaskArn)),
		TaskARN:           sdkaws.ToString(t.TaskArn),
		ClusterARN:        sdkaws.ToString(t.ClusterArn),
		TaskDefinition:    shortName(sdkaws.ToString(t.TaskDefinitionArn)),
		Status:            sdkaws.ToString(t.LastStatus),
		DesiredStatus:     sdkaws.ToString(t.DesiredStatus),
		HealthStatus:      string(t.HealthStatus),
		LaunchType:        string(t.LaunchType),
		CPU:               sdkaws.ToString(t.Cpu),
		Memory:            sdkaws.ToString(t.Memory),
		StartedAt:         t.StartedAt,
		StoppedAt:         t.StoppedAt,
		StoppedReason:     sdkaws.ToString(t.StoppedReason),
		EnableExecCommand: t.EnableExecuteCommand,
	}
	if t.Group != nil {
		parts := strings.SplitN(*t.Group, ":", 2)
		if len(parts) == 2 && parts[0] == "service" {
			task.ServiceName = parts[1]
		}
	}
	for _, ctr := range t.Containers {
		c := model.Container{
			Name:         sdkaws.ToString(ctr.Name),
			Image:        sdkaws.ToString(ctr.Image),
			Status:       sdkaws.ToString(ctr.LastStatus),
			Reason:       sdkaws.ToString(ctr.Reason),
			HealthStatus: string(ctr.HealthStatus),
			ExitCode:     ctr.ExitCode,
		}
		task.Containers = append(task.Containers, c)
	}
	return task
}

func shortName(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return arn
}

func shortTaskID(arn string) string {
	parts := strings.Split(arn, "/")
	return parts[len(parts)-1]
}
