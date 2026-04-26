package aws

import (
	"context"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Clients holds all AWS service clients.
type Clients struct {
	ECS     *ecs.Client
	CWL     *cloudwatchlogs.Client
	SSM     *ssm.Client
	STS     *sts.Client
	Region  string
	Cluster string
}

// NewClients builds AWS service clients from explicit parameters.
// profile and roleARN are optional; region defaults to eu-west-1 if empty.
func NewClients(ctx context.Context, cluster, region, profile, roleARN string) (*Clients, error) {
	if region == "" {
		region = "eu-west-1"
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	if roleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
		cfg.Credentials = sdkaws.NewCredentialsCache(provider)
	}

	return &Clients{
		ECS:     ecs.NewFromConfig(cfg),
		CWL:     cloudwatchlogs.NewFromConfig(cfg),
		SSM:     ssm.NewFromConfig(cfg),
		STS:     sts.NewFromConfig(cfg),
		Region:  region,
		Cluster: cluster,
	}, nil
}
