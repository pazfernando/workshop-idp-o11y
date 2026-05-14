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

Do not use local Terraform state for this root. The managed suite is intended to converge through the S3 backend declared in `versions.tf`.

For ad hoc operator-driven execution, first resolve or create the state bucket and key, then initialize Terraform against that remote backend.

Example:

```bash
AWS_REGION=us-east-1
SUITE_NAME=o11y-platform
ACCOUNT_ID="$(aws sts get-caller-identity --query Account --output text)"
TF_STATE_BUCKET="${SUITE_NAME}-${ACCOUNT_ID}-${AWS_REGION}-tfstate"
TF_STATE_KEY="platform/${SUITE_NAME}.tfstate"

if aws s3api head-bucket --bucket "$TF_STATE_BUCKET" 2>/dev/null; then
  echo "Using existing Terraform state bucket: $TF_STATE_BUCKET"
else
  if [ "$AWS_REGION" = "us-east-1" ]; then
    aws s3api create-bucket --bucket "$TF_STATE_BUCKET"
  else
    aws s3api create-bucket \
      --bucket "$TF_STATE_BUCKET" \
      --create-bucket-configuration "LocationConstraint=${AWS_REGION}"
  fi

  aws s3api put-bucket-versioning \
    --bucket "$TF_STATE_BUCKET" \
    --versioning-configuration Status=Enabled

  aws s3api put-bucket-encryption \
    --bucket "$TF_STATE_BUCKET" \
    --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
fi

terraform -chdir=infra/terraform/managed-suite init -reconfigure \
  -backend-config="bucket=${TF_STATE_BUCKET}" \
  -backend-config="key=${TF_STATE_KEY}" \
  -backend-config="region=${AWS_REGION}"

terraform -chdir=infra/terraform/managed-suite apply -auto-approve \
  -var="aws_region=${AWS_REGION}" \
  -var="name=${SUITE_NAME}"
```

This is the same strategy used by the product workflows through the `Ensure Terraform state bucket` step.

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
- The repository now owns the reusable deployment path, but persistent reconciliation still depends on Terraform state stored in S3 rather than local state files.
- The suite still provisions a default platform dashboard, and the contract-consumer path can now publish workload-specific dashboards into the same Grafana instance.
