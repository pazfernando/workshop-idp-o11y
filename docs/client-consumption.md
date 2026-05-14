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

The client does not need to pass platform repository or ref inputs. The reusable workflow checks out its own pinned source and the caller repository separately so platform logic stays aligned with the exact workflow revision being executed.

That reusable workflow can:

- validate a contract
- build a backend-neutral plan
- build AWS Lambda bindings
- optionally deploy the platform-managed suite first and then use its OTLP endpoint for bindings generation
- fail early on invalid caller paths, unsupported targets, invalid JSON-array CIDR inputs, and incompatible instrumentation values

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

If both are set, the explicit `collector_endpoint` wins and the workflow emits that precedence in the run logs.

When `deploy_managed_suite: true`, the workflow also applies the platform standard `Ensure Terraform state bucket` strategy so the suite never converges against local Terraform state.

For AWS Lambda bindings, `emf_compatibility_mode` defaults to `true`. Clients should only turn it off deliberately after confirming that their metrics path no longer depends on EMF compatibility behavior.

## What The Client Receives

Workflow outputs:

- `validation_message`
- `plan_json`
- `bindings_json`
- `managed_suite_enabled`
- `managed_suite_grafana_url`
- `managed_suite_otlp_http_endpoint`
- `effective_collector_endpoint`
- `effective_collector_source`

The workflow also writes a short GitHub Actions job summary with the selected contract path, whether bindings were generated, whether the managed suite was deployed, the effective collector endpoint, and whether that endpoint came from an explicit client input or from the platform-managed suite.

Artifacts:

- `generated/validation.txt`
- `generated/plan.json`
- `generated/bindings.json` when bindings are enabled
- `generated/client-handoff/implementation-summary.md`
- `generated/client-handoff/implementation-manifest.json`
- uploaded artifact `o11y-client-handoff`

The `o11y-client-handoff` artifact is the client-facing delivery package after the platform workflow finishes. It is intended to be the easiest thing for the client team to download, inspect, and pass to the deployment owner.

The handoff package is also the authoritative record of the effective implementation chosen by the platform pipeline. In particular, `implementation-summary.md` and `implementation-manifest.json` now capture:

- the effective instrumentation mode and export strategy
- the OTLP authentication mode
- the current metric support policy and selected dashboard preset
- the requested and supported preset metric names
- the effective collector endpoint and whether it came from an explicit client input or the managed suite
- the effective OTLP endpoints returned by the adapter
- whether EMF compatibility remained enabled
- whether ADOT wrapper, layers, and managed policies were required
- any unresolved required inputs still left for the client deployment step
- the exact files produced at the end of the workflow

## Responsibility Split

For AWS Lambda, the platform product generates bindings and deployment guidance; the client deployment still materializes the workload.

- `instrumentation_mode: code` means the client application owns in-process OTel instrumentation and the platform returns runtime configuration.
- `instrumentation_mode: adot_layer` means the platform returns the wrapper, layer, and policy requirements, but the client still has to attach those to the Lambda definition in its own deployment stack.

The platform does not deploy the client Lambda itself.

## Required Secrets For Managed Suite Deployment

When `deploy_managed_suite` is enabled, the caller should pass `secrets: inherit` and provide:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_SESSION_TOKEN` when temporary credentials are used
- optionally `OBSERVABILITY_SUITE_GRAFANA_ADMIN_PASSWORD`

## Input Guardrails

The reusable workflow now enforces a few caller-facing rules before doing any expensive work:

- `contract_path` must be a relative path inside the caller repository
- the contract file must exist
- `bindings_target` must be `aws-lambda` when bindings are enabled
- `instrumentation_mode` must be `code` or `adot_layer`
- managed-suite CIDR inputs must be valid JSON arrays
- `managed_suite_name` must be non-empty when managed-suite deployment is enabled

## AWS Lambda Binding Defaults And Notes

- `instrumentation_mode` defaults to `code`
- `emf_compatibility_mode` defaults to `true`
- when the contract resolves to `collector` export strategy and `collector_endpoint` is empty, the workflow uses the managed-suite OTLP endpoint if the managed suite was deployed
- when the contract resolves to `direct` export strategy and no direct endpoints are supplied, the adapter can infer traces and metrics endpoints for AWS backends
- in direct AWS inference mode, logs are not inferred as OTLP endpoints; application logs still land in CloudWatch Logs unless the client routes logs through a collector or another backend-specific path

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
