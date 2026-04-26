# ecsctl

A `kubectl`-like CLI for AWS ECS. Manage clusters, services, tasks, and container logs from the terminal with familiar commands.

## Installation

```bash
go install github.com/bikramdhoju/ecsctl@latest
```

Or build from source:

```bash
git clone https://github.com/bikramdhoju/ecsctl
cd ecsctl
make install
```

## Configuration

```bash
# Add a context (one per environment/cluster)
ecsctl config set-context prod \
  --cluster arn:aws:ecs:eu-west-1:123456789:cluster/main \
  --region eu-west-1 \
  --profile prod-sso

# Switch context
ecsctl config use-context prod

# List all contexts
ecsctl config get-contexts
```

Config is stored at `~/.ecsctl/config.yaml`.

### Context fields

| Field | Description |
|---|---|
| `--cluster` | Cluster name or full ARN |
| `--region` | AWS region |
| `--profile` | AWS credentials profile (`~/.aws/credentials`) |
| `--role-arn` | IAM role to assume before calling ECS |
| `--output` | Default output format for this context |

### Global overrides

Any context field can be overridden per-command:

```bash
ecsctl get services --profile staging --region eu-west-1 --cluster my-cluster
```

## Commands

### get

```bash
ecsctl get clusters
ecsctl get services
ecsctl get services -c my-cluster
ecsctl get tasks
ecsctl get tasks -s my-service --status RUNNING
ecsctl get tasks --status STOPPED

# Wide output — extra columns
ecsctl get services -o wide

# JSON / YAML
ecsctl get clusters -o json
ecsctl get tasks -o yaml
```

### describe

```bash
ecsctl describe cluster my-cluster
ecsctl describe service my-service
ecsctl describe service my-service -c my-cluster
ecsctl describe task a1b2c3d4e5f6
```

### logs

```bash
# Last 100 lines
ecsctl logs a1b2c3d4e5f6

# Stream (follow)
ecsctl logs a1b2c3d4e5f6 -f

# Specific container (required when task has multiple)
ecsctl logs a1b2c3d4e5f6 -c web

# Last N lines
ecsctl logs a1b2c3d4e5f6 --tail 200

# Since a duration
ecsctl logs a1b2c3d4e5f6 --since 1h
ecsctl logs a1b2c3d4e5f6 --since 30m -f
```

Logs are fetched from CloudWatch using the `awslogs` configuration in the task definition.

### config

```bash
ecsctl config view                        # Print all contexts
ecsctl config get-contexts                # Table of contexts
ecsctl config current-context            # Show active context name
ecsctl config use-context staging        # Switch active context
ecsctl config set-context prod \         # Create or update a context
  --cluster my-cluster \
  --region eu-west-1
ecsctl config delete-context old-ctx     # Remove a context
```

## AWS Permissions Required

```json
{
  "Effect": "Allow",
  "Action": [
    "ecs:DescribeClusters",
    "ecs:ListClusters",
    "ecs:DescribeServices",
    "ecs:ListServices",
    "ecs:DescribeTasks",
    "ecs:ListTasks",
    "ecs:DescribeTaskDefinition",
    "logs:FilterLogEvents",
    "sts:GetCallerIdentity"
  ],
  "Resource": "*"
}
```

## Roadmap

- [ ] `ecsctl exec` — ECS Exec into a running container
- [ ] `ecsctl scale` — update service desired count
- [ ] `ecsctl restart` — force new deployment
- [ ] `ecsctl events` — service and task events
- [ ] `ecsctl port-forward` — SSM port forwarding
- [ ] `--watch` mode for `get` commands
- [ ] Shell completion (bash/zsh/fish)
