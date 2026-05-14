# Support Matrix

This document defines what the platform product supports today.

## Contract Interface

Supported today:

- contract schema validation
- semantic validation
- backend-neutral planning
- AWS Lambda runtime binding generation
- HTTP API for `validate`, `plan`, and `bindings aws-lambda`
- reusable GitHub workflow for client CI/CD consumption
- root Terraform deployment path for a platform-managed observability suite

Not yet provided as a platform interface:

- gRPC control plane
- reconciler-driven persistent platform state inside this repository

## Command Surface

| Interface | Supported | Notes |
| :--- | :--- | :--- |
| `o11yctl validate` | Yes | Validates schema and semantic rules |
| `o11yctl plan` | Yes | Produces a backend-neutral JSON plan |
| `o11yctl bindings aws-lambda` | Yes | Produces AWS Lambda runtime bindings |
| `o11yd` HTTP API | Yes | Exposes `/v1/validate`, `/v1/plan`, and `/v1/bindings/aws-lambda` |
| reusable workflow | Yes | `.github/workflows/contract-consumer.yml` |
| gRPC API | No | Not implemented |
| Managed suite Terraform root | Yes | `infra/terraform/managed-suite` |
| Managed suite GitHub workflows | Yes | Apply and destroy workflows at the product root |

## Target And Runtime Coverage

| Target | Runtime | Validation | Planning | Adapter Output | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `aws` | `aws-lambda` | Yes | Yes | Yes | Primary implemented runtime path today |
| `kubernetes` | `kubernetes` | Yes | Yes | No target adapter yet | Contract and planning path exist |
| `hybrid` | `vm` or mixed | Yes | Yes | No target adapter yet | Contract and planning path exist |

## Ingestion Mode Coverage

| Ingestion | Validation | Planning | AWS Lambda Bindings | Notes |
| :--- | :--- | :--- | :--- | :--- |
| `collector` | Yes | Yes | Yes | Preferred mode for platform-managed routing |
| `direct` | Yes | Yes | Yes | Supported for AWS Lambda bindings with target-specific constraints |

## AWS Lambda Binding Combination Coverage

| Instrumentation Mode | Export Strategy | Supported | Notes |
| :--- | :--- | :--- | :--- |
| `code` | `collector` | Yes | Recommended default for application teams that already own in-process OTel instrumentation |
| `code` | `direct` | Yes | Supports explicit direct OTLP endpoints and AWS-inferred traces and metrics endpoints |
| `adot_layer` | `collector` | Yes | Supported when the client deployment attaches the returned ADOT layer, wrapper, and policy requirements |
| `adot_layer` | `direct` | Yes | Supported with the same client-side ADOT attachment responsibility as above |

Additional AWS Lambda notes:

- EMF compatibility mode defaults to `true`
- when `direct` is used against inferred AWS endpoints, traces and metrics can be inferred but logs still require CloudWatch Logs or a collector-based route
- when `collector` is used and the client does not provide `collector_endpoint`, the reusable workflow uses the managed-suite OTLP endpoint if it deployed the managed suite in the same run or reused a pre-existing one

## Capability Coverage

| Capability | Contract | Plan | AWS Adapter Materialization | Notes |
| :--- | :--- | :--- | :--- | :--- |
| Dashboards | Yes | Yes | Partial | Planner resolves preset classes and asset references; backend rendering is not fully generalized |
| Alerts | Yes | Yes | Partial | Planner carries rules; generalized backend rendering is not complete |
| SLOs | Yes | Yes | Partial | Planner carries objectives; execution backend is not generalized |
| Data retention | Yes | Yes | Partial | Planner carries policy intent; enforcement depends on execution backend |
| Runtime bindings | Yes | Yes | Yes for AWS Lambda | Concrete adapter output exists |
| Governed preset metrics | Yes | Yes | Partial | Only metrics that belong to supported dashboard presets are accepted today |

## CI/CD Consumption Model

Supported today:

- local CLI execution inside a CI/CD runner
- remote HTTP execution through `o11yd`
- remote GitHub-native execution through the reusable workflow interface
- deterministic validation and plan generation
- JSON output that can be consumed by downstream steps
- adapter output generation for AWS Lambda runtime configuration
- product-owned suite provisioning through Terraform state

Not yet productized:

- first-class execution state managed by this repository
- platform-owned asynchronous reconciliation beyond Terraform state

## Guidance For Application Teams

Before authoring a contract, check:

1. whether your target and runtime combination is in the supported adapter path
2. whether the capability you want is fully materialized or only represented at planning level
3. whether your requested metrics belong to a supported dashboard preset
4. whether your delivery pipeline consumes CLI output directly, the reusable workflow interface, or a remote platform API

## Companion References

- [observability-contract.md](observability-contract.md)
- [contract-authoring-guide.md](contract-authoring-guide.md)
- [client-consumption.md](client-consumption.md)
- [operating-model.md](operating-model.md)
