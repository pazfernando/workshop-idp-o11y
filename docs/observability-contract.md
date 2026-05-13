# Observability Contract

The `ObservabilityContract` is the application-facing API of this platform product.

Its purpose is to let an application team declare observability intent in a stable, backend-neutral format. The contract describes what the workload needs from the platform, while the platform decides how to implement that intent with the supported adapters, classes, and assets.

## What The Contract Must Express

The contract allows an application team to declare:

- workload identity and ownership
- delivery target and operating mode
- OpenTelemetry configuration and resource attributes
- traces, metrics, and logs intent
- platform capabilities such as dashboards, alerts, SLOs, and retention

The contract does not require the application team to know whether the platform uses Terraform, Alloy, Grafana, CloudWatch, or another backend behind the scenes.

## Design Principles

- Declare capabilities, not infrastructure components.
- Use OpenTelemetry as the shared language between application teams and the platform.
- Keep implementation-specific values out of the contract whenever possible.
- Treat the metrics catalog as part of the API contract between the workload and the platform.
- Reserve `status` for platform-produced outputs and lifecycle reporting.

## Structure

The contract has five major sections:

1. `metadata`
   Workload identity, owner, system, environment, labels, and annotations.
2. `spec.service`
   Runtime, language, framework, lifecycle, and service metadata.
3. `spec.delivery`
   Delivery mode and target context such as AWS region or Kubernetes namespace.
4. `spec.telemetry`
   OpenTelemetry intent, resource attributes, and per-signal configuration.
5. `spec.capabilities`
   Platform products derived from telemetry such as dashboards, alerts, SLOs, and retention.

## Validation Surface

The contract is validated in two layers:

- JSON Schema validation:
  [schemas/observability-contract-v1alpha1.schema.json](../schemas/observability-contract-v1alpha1.schema.json)
- semantic validation in code:
  `internal/validation`

Examples:

- AWS Lambda:
  [examples/contracts/aws-lambda-order-processing.yaml](../examples/contracts/aws-lambda-order-processing.yaml)
- Kubernetes:
  [examples/contracts/kubernetes-payments.yaml](../examples/contracts/kubernetes-payments.yaml)
- Hybrid / VM-style workload:
  [examples/contracts/monolith-crm.yaml](../examples/contracts/monolith-crm.yaml)

## How The Platform Uses The Contract

1. The platform validates the contract.
2. The planner converts the contract into a neutral `ProvisioningPlan`.
3. The target adapter resolves supported classes, assets, and runtime outputs.
4. The execution backend consumes the adapter outputs to converge infrastructure or runtime configuration.

## What Application Teams Should Read Next

For self-service contract authoring, the contract definition alone is not enough. Use these companion documents:

- [contract-authoring-guide.md](contract-authoring-guide.md)
- [support-matrix.md](support-matrix.md)
- [consumer-interaction-flow.md](consumer-interaction-flow.md)
