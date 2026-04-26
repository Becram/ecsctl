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

	"github.com/bikramdhoju/ecsctl/internal/config"
)

// Clients holds all AWS service clients for a single context.
type Clients struct {
	ECS     *ecs.Client
	CWL     *cloudwatchlogs.Client
	SSM     *ssm.Client
	STS     *sts.Client
	Region  string
	Cluster string
}

// NewClients builds AWS service clients from the resolved context.
// profileOverride, if non-empty, takes precedence over ctx.AWSProfile.
func NewClients(ctx context.Context, activeCtx *config.Context, profileOverride string) (*Clients, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(activeCtx.Region),
	}

	profile := activeCtx.AWSProfile
	if profileOverride != "" {
		profile = profileOverride
	}
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	if activeCtx.RoleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, activeCtx.RoleARN)
		cfg.Credentials = sdkaws.NewCredentialsCache(provider)
	}

	return &Clients{
		ECS:     ecs.NewFromConfig(cfg),
		CWL:     cloudwatchlogs.NewFromConfig(cfg),
		SSM:     ssm.NewFromConfig(cfg),
		STS:     sts.NewFromConfig(cfg),
		Region:  activeCtx.Region,
		Cluster: activeCtx.Cluster,
	}, nil
}
