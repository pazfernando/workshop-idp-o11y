# Managed Suite Deployment

This repository now includes a root productized deployment path for a platform-managed observability suite.

## What It Deploys

The managed suite provisions:

- Grafana
- Grafana Alloy
- Prometheus
- Loki
- Tempo
- supporting AWS security group, IAM, EC2, and Elastic IP resources

The deployment path is meant to represent the platform product boundary for a shared or dedicated observability data plane.

## Product Interfaces

Infrastructure path:

- Terraform root: `infra/terraform/managed-suite`
- GitHub Actions apply workflow: `.github/workflows/managed-suite-apply.yml`
- GitHub Actions destroy workflow: `.github/workflows/managed-suite-destroy.yml`
- reusable client workflow path: `.github/workflows/contract-consumer.yml` with `deploy_managed_suite: true`

Asset layer:

- `infra/observability-suite`

## Terraform Usage

Initialize and apply:

```bash
terraform -chdir=infra/terraform/managed-suite init
terraform -chdir=infra/terraform/managed-suite apply
```

Key inputs:

- `aws_region`
- `name`
- `subnet_id`
- `instance_type`
- `root_volume_size_gb`
- `grafana_admin_password`
- `grafana_allowed_cidrs`
- `otlp_allowed_cidrs`
- `ssh_allowed_cidrs`
- `dashboard_uid`
- `dashboard_title`

Key outputs:

- `public_ip`
- `grafana_url`
- `grafana_dashboard_url`
- `otlp_http_endpoint`
- `otlp_grpc_endpoint`
- `grafana_admin_user`
- `grafana_admin_password`

## GitHub Actions Usage

The apply workflow is intended for product operations teams that want a standard, repository-owned deployment path.

For GitHub-based application teams, the same suite can also be provisioned indirectly through the reusable workflow documented in [client-consumption.md](client-consumption.md).

The workflow:

1. checks out the repository
2. configures AWS credentials
3. resolves or creates the Terraform state bucket
4. initializes the managed-suite Terraform root
5. applies the suite and prints outputs

The destroy workflow reuses the same state key and destroys that same Terraform root.

## Current Design Notes

- The suite is EC2-based and stateful at the infrastructure layer.
- The repository now owns the reusable deployment path, but persistent reconciliation still depends on Terraform state.
- The default dashboard is generic platform observability, not tied to a sample workload.
