# Operating Model

## Platform Contract Flow

1. An application team authors and versions an `ObservabilityContract`.
2. The platform validates the contract against the embedded schema and semantic rules.
3. The planner translates the contract into a backend-neutral `ProvisioningPlan`.
4. A target-specific adapter materializes runtime bindings and platform asset references.
5. The application deployment consumes those bindings instead of embedding observability implementation details directly.

## Operational Components

- Contract input:
  `examples/contracts` and application repositories that consume the platform.
- Validation:
  `internal/validation`
- Planner:
  `internal/planner`
- AWS runtime adapter:
  `internal/adapters/aws`
- Terraform runtime bindings module:
  `adapter/aws/terraform`
- Catalog classes:
  `catalog/classes`
- Platform assets:
  `assets/collector`, `assets/dashboards`, `assets/alerts`

## Current Execution Interface

Today, the operational interface is split across a stateless control plane and a productized infrastructure path:

- `o11yctl validate`
- `o11yctl plan`
- `o11yctl bindings aws-lambda`
- `o11yd` HTTP endpoints for remote `validate`, `plan`, and `bindings aws-lambda`
- `.github/workflows/contract-consumer.yml` for GitHub-based client consumption
- `infra/terraform/managed-suite`
- `.github/workflows/managed-suite-apply.yml`
- `.github/workflows/managed-suite-destroy.yml`

The CLI and HTTP surfaces are suitable for CI/CD pipeline execution because they are non-interactive and return deterministic validation results or JSON payloads.

## Current Idempotency Model

The contract validation and planning steps are stateless. Idempotent infrastructure convergence is currently the responsibility of the execution backend that consumes the generated plan or bindings, such as Terraform with remote state.

This repository already includes the reusable building blocks for that model:

- contract validation
- backend-neutral planning
- adapter-driven runtime binding generation
- reusable Terraform module inputs for AWS Lambda
- a root Terraform path for platform-managed collector and visualization infrastructure

## AWS Lambda Runtime Flow

1. The contract declares `target: aws` and `runtime: aws-lambda`.
2. The planner selects the `aws-lambda-otel-bindings` class.
3. The AWS adapter resolves the runtime export strategy, such as `collector` or `direct`.
4. The adapter returns OpenTelemetry environment variables, optional Lambda layer settings, effective OTLP routing, and platform asset references.
5. The application deployment injects those outputs into the Lambda runtime.

## Product Boundary

This repository should be treated as the platform product source for:

- the observability contract model
- the validation logic
- the planner
- target adapters
- reusable classes and assets

Application-specific deployment stacks may consume these outputs, but they are not the product boundary of the platform itself.
