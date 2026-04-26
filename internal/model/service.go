package model

import "time"

type Service struct {
	Name           string
	ClusterARN     string
	Status         string
	DesiredCount   int32
	RunningCount   int32
	PendingCount   int32
	TaskDefinition string
	LaunchType     string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	Events         []ServiceEvent
	LoadBalancers  []LoadBalancer
	Deployments    []Deployment
}

type ServiceEvent struct {
	CreatedAt time.Time
	Message   string
}

type LoadBalancer struct {
	TargetGroupARN string
	ContainerName  string
	ContainerPort  int32
}

type Deployment struct {
	ID           string
	Status       string
	DesiredCount int32
	RunningCount int32
	PendingCount int32
	LaunchType   string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}
