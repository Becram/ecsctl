package model

import "time"

type Task struct {
	TaskID            string
	TaskARN           string
	ClusterARN        string
	ServiceName       string
	TaskDefinition    string
	Status            string
	DesiredStatus     string
	HealthStatus      string
	LaunchType        string
	CPU               string
	Memory            string
	StartedAt         *time.Time
	StoppedAt         *time.Time
	StoppedReason     string
	EnableExecCommand bool
	Containers        []Container
}
