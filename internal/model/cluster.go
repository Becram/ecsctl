package model

type Cluster struct {
	Name                              string
	ARN                               string
	Status                            string
	RunningTasksCount                 int32
	PendingTasksCount                 int32
	ActiveServicesCount               int32
	RegisteredContainerInstancesCount int32
	CapacityProviders                 []string
}
