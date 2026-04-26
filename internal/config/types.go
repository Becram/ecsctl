package config

const DefaultConfigPath = "~/.ecsctl/config.yaml"

type EcsConfig struct {
	APIVersion     string         `yaml:"apiVersion"`
	Kind           string         `yaml:"kind"`
	CurrentContext string         `yaml:"current-context"`
	Contexts       []ContextEntry `yaml:"contexts"`
}

type ContextEntry struct {
	Name    string  `yaml:"name"`
	Context Context `yaml:"context"`
}

// Context maps to a named entry in the config file.
// AWSProfile maps to ~/.aws/credentials profile. If empty, the default
// credential chain (env vars, instance metadata, etc.) is used.
type Context struct {
	Cluster    string `yaml:"cluster"`
	Region     string `yaml:"region"`
	Namespace  string `yaml:"namespace,omitempty"`
	RoleARN    string `yaml:"role-arn,omitempty"`
	AWSProfile string `yaml:"aws-profile,omitempty"`
	Output     string `yaml:"output,omitempty"`
}
