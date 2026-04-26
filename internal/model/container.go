package model

type Container struct {
	Name         string
	Image        string
	Status       string
	ExitCode     *int32
	Reason       string
	HealthStatus string
	// Log config — populated from task definition when needed
	LogGroup     string
	LogStream    string
	LogRegion    string
}
