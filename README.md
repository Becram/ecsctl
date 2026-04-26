# ecsctl

A `kubectl`-like CLI for AWS ECS. Manage clusters, services, tasks, and container logs from the terminal with familiar commands.

## Installation

```bash
go install github.com/bikramdhoju/ecsctl@latest
```

Or build from source:

```bash
git clone https://github.com/Becram/ecsctl
cd ecsctl
make install
```

## Usage

```
ecsctl [command] --cluster <name> --profile <aws-profile> [flags]
```

### Global flags

| Flag | Default | Description |
|---|---|---|
| `-c, --cluster` | | ECS cluster name or ARN |
| `--profile` | | AWS credentials profile (`~/.aws/credentials`) |
| `--region` | `eu-west-1` | AWS region |
| `--role-arn` | | IAM role ARN to assume |
| `-o, --output` | `table` | Output format: `table`, `wide`, `json`, `yaml` |

## Commands

### get

```bash
ecsctl get clusters --profile prod
ecsctl get services -c my-cluster --profile prod
ecsctl get tasks   -c my-cluster --profile prod
ecsctl get tasks   -c my-cluster --profile prod -s my-service
ecsctl get tasks   -c my-cluster --profile prod --status STOPPED

# Wide output — extra columns
ecsctl get services -c my-cluster --profile prod -o wide

# JSON / YAML
ecsctl get clusters --profile prod -o json
ecsctl get tasks    -c my-cluster --profile prod -o yaml
```

### describe

```bash
ecsctl describe cluster my-cluster  --profile prod
ecsctl describe service my-service  -c my-cluster --profile prod
ecsctl describe task    a1b2c3d4    -c my-cluster --profile prod
```

### logs

```bash
# Last 100 lines
ecsctl logs a1b2c3d4 -c my-cluster --profile prod

# Stream (follow)
ecsctl logs a1b2c3d4 -c my-cluster --profile prod -f

# Specific container (required when task has multiple)
ecsctl logs a1b2c3d4 -c my-cluster --profile prod --container web

# Since a duration
ecsctl logs a1b2c3d4 -c my-cluster --profile prod --since 1h
ecsctl logs a1b2c3d4 -c my-cluster --profile prod --since 30m -f
```

Logs are fetched from CloudWatch using the `awslogs` configuration in the task definition.

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
