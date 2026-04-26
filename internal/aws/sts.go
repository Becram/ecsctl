package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// ValidateCredentials calls GetCallerIdentity to verify credentials work.
// Called early in PersistentPreRunE so failures surface before any ECS call.
func (c *Clients) ValidateCredentials(ctx context.Context) error {
	_, err := c.STS.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("AWS credentials invalid or not configured: %w", err)
	}
	return nil
}
