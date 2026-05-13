# Client Consumption Guide

This document explains how an application team consumes the platform product from its own CI/CD pipeline.

## Recommended Consumption Mode

The recommended path for GitHub-based clients is the reusable workflow interface:

- it avoids keeping a control-plane service running all the time
- it gives the client a stable product interface
- it keeps the platform logic centralized in this repository
- it lets the platform evolve validation, planning, bindings, and managed-suite logic without copying scripts into each client repository
- when managed-suite deployment is enabled, it also guarantees remote Terraform state through the workflow-managed S3 backend strategy

## What The Client Calls

The client workflow calls:

`<platform-repo>/.github/workflows/contract-consumer.yml`

That reusable workflow can:

- validate a contract
- build a backend-neutral plan
- build AWS Lambda bindings
- optionally deploy the platform-managed suite first and then use its OTLP endpoint for bindings generation

A concrete end-to-end example is included in [examples/consumers/aws-lambda-order-processing/README.md](../examples/consumers/aws-lambda-order-processing/README.md).

## Minimal Caller Example

```yaml
name: Consume Observability Platform

on:
  workflow_dispatch:
  push:
    branches: [main]

jobs:
  observability:
    uses: your-org/workshop-iidp-o11y/.github/workflows/contract-consumer.yml@main
    with:
      contract_path: contracts/observability/order-processing.yaml
      generate_bindings: true
      bindings_target: aws-lambda
      instrumentation_mode: code
      collector_endpoint: http://collector.internal:4318
    secrets: inherit
```

## Caller Example With Managed Suite Provisioning

If the client wants the platform workflow to stand up the managed suite and then reuse that endpoint:

```yaml
name: Consume Observability Platform

on:
  workflow_dispatch:

jobs:
  observability:
    uses: your-org/workshop-iidp-o11y/.github/workflows/contract-consumer.yml@main
    with:
      contract_path: contracts/observability/order-processing.yaml
      generate_bindings: true
      bindings_target: aws-lambda
      deploy_managed_suite: true
      aws_region: us-east-1
      managed_suite_name: team-a-o11y
      managed_suite_grafana_allowed_cidrs: '["10.0.0.0/8"]'
      managed_suite_otlp_allowed_cidrs: '["10.0.0.0/8"]'
    secrets: inherit
```

When `deploy_managed_suite: true` and `collector_endpoint` is empty, the workflow uses the managed-suite OTLP HTTP endpoint as the effective collector endpoint for bindings generation.

When `deploy_managed_suite: true`, the workflow also applies the platform standard `Ensure Terraform state bucket` strategy so the suite never converges against local Terraform state.

## What The Client Receives

Workflow outputs:

- `validation_message`
- `plan_json`
- `bindings_json`
- `managed_suite_enabled`
- `managed_suite_grafana_url`
- `managed_suite_otlp_http_endpoint`
- `effective_collector_endpoint`

Artifacts:

- `generated/validation.txt`
- `generated/plan.json`
- `generated/bindings.json` when bindings are enabled

## Required Secrets For Managed Suite Deployment

When `deploy_managed_suite` is enabled, the caller should pass `secrets: inherit` and provide:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_SESSION_TOKEN` when temporary credentials are used
- optionally `OBSERVABILITY_SUITE_GRAFANA_ADMIN_PASSWORD`

## How To Switch From Direct CLI Usage

### Before

Many clients start with local steps in their own workflow:

```yaml
- name: Validate contract
  run: go run ./cmd/o11yctl validate -f contracts/observability/order-processing.yaml

- name: Build plan
  run: go run ./cmd/o11yctl plan -f contracts/observability/order-processing.yaml > build/plan.json

- name: Build bindings
  run: go run ./cmd/o11yctl bindings aws-lambda -f contracts/observability/order-processing.yaml --collector-endpoint http://collector.internal:4318 > build/bindings.json
```

### After

Replace those local steps with one reusable job:

```yaml
jobs:
  observability:
    uses: your-org/workshop-iidp-o11y/.github/workflows/contract-consumer.yml@main
    with:
      contract_path: contracts/observability/order-processing.yaml
      generate_bindings: true
      bindings_target: aws-lambda
      collector_endpoint: http://collector.internal:4318
    secrets: inherit
```

This is the clearest switch point for clients:

1. move the contract file into the client repository
2. remove copied local platform scripts
3. replace local `o11yctl` commands with the reusable workflow call
4. consume workflow outputs or artifacts in downstream deployment jobs

For a concrete reference, see [examples/consumers/aws-lambda-order-processing/.github/workflows/consume-observability-platform.yml](../examples/consumers/aws-lambda-order-processing/.github/workflows/consume-observability-platform.yml).

## How The Client Uses The Outputs

A downstream deployment job can read the JSON outputs and inject them into deployment steps.

Example:

```yaml
jobs:
  observability:
    uses: your-org/workshop-iidp-o11y/.github/workflows/contract-consumer.yml@main
    with:
      contract_path: contracts/observability/order-processing.yaml
      generate_bindings: true
    secrets: inherit

  deploy:
    needs: observability
    runs-on: ubuntu-latest
    steps:
      - name: Use bindings output
        run: |
          printf '%s\n' '${{ needs.observability.outputs.bindings_json }}' > bindings.json
```

## Current Boundary

This reusable workflow is the recommended client-facing interface for GitHub consumers today.

If the platform later needs:

- non-GitHub consumers
- long-running API integrations
- asynchronous orchestration
- platform-owned tenancy and auth

then `o11yd` can become the next externally hosted interface.
