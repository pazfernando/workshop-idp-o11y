# Contract Authoring Guide

This guide explains how an application team should create an `ObservabilityContract`.

## Authoring Workflow

1. Start from the example that matches your workload shape.
2. Fill in `metadata` and `spec.service`.
3. Define the `spec.delivery` target and context.
4. Declare traces, metrics, and logs intent in `spec.telemetry`.
5. Add the platform capabilities you want in `spec.capabilities`.
6. Validate the contract locally or in CI with `o11yctl validate`.
7. Generate a plan with `o11yctl plan` and review the resulting classes, assets, and bindings.
8. For AWS Lambda consumers, review the final bindings and handoff artifact to confirm the effective runtime implementation details that are intentionally outside the contract itself.

## Minimal Contract Skeleton

```yaml
apiVersion: observability.platform/v1alpha1
kind: ObservabilityContract
metadata:
  name: my-service
  owner: my-team
  system: customer-platform
  environment: prod
spec:
  service:
    name: my-service
    runtime: aws-lambda
    language: nodejs
    framework: aws-serverless
  delivery:
    mode: provision
    target: aws
    region: us-east-1
  telemetry:
    openTelemetry:
      enabled: true
      semanticConventions: opentelemetry
      propagators:
        - tracecontext
        - baggage
    resourceAttributes:
      service.name: my-service
      service.namespace: customer-platform
      deployment.environment: prod
    signals:
      traces:
        enabled: true
        ingestion: collector
        backendClass: traces-standard
      metrics:
        enabled: true
        ingestion: collector
        backendClass: metrics-standard
        catalog:
          - name: HttpRequests
            type: counter
            unit: "{request}"
            description: Total HTTP requests.
      logs:
        enabled: true
        ingestion: collector
        backendClass: logs-standard
        format: json
  capabilities:
    dashboards:
      enabled: true
      preset: serverless-api
```

## Section-By-Section Guidance

### `metadata`

Required fields:

- `name`
- `owner`
- `system`
- `environment`

Guidance:

- `name` should be workload-scoped and DNS-safe.
- `owner` should identify the responsible engineering team.
- `system` should represent the broader product or business domain.
- `environment` must be one of `dev`, `test`, `staging`, or `prod`.

### `spec.service`

Required fields:

- `name`
- `runtime`
- `language`
- `framework`

Useful optional fields:

- `tier`
- `lifecycle`

Guidance:

- Use `runtime` to reflect the actual execution environment, such as `aws-lambda`, `kubernetes`, or `vm`.
- Use `framework` as a platform-facing hint for selecting presets and integration conventions.

### `spec.delivery`

Required fields:

- `mode`
- `target`

Target-specific requirements:

- if `target: aws`, set `region`
- if `target: kubernetes`, set both `cluster` and `namespace`

Guidance:

- `mode: provision` means the platform is expected to materialize platform-facing resources.
- `mode: adopt` means the platform binds to an existing operating environment instead of treating the workload as greenfield.

### `spec.telemetry.openTelemetry`

Required fields:

- `enabled`
- `semanticConventions`
- `propagators`

Guidance:

- Use `semanticConventions: opentelemetry` unless your platform introduces a more specific profile later.
- Include only propagators that the workload actually supports end to end.
- Use `sampling` only when the workload needs to declare non-default behavior.

### `spec.telemetry.resourceAttributes`

Required fields:

- `service.name`
- `service.namespace`
- `deployment.environment`

Guidance:

- Keep these values stable across environments where possible.
- Add target-specific resource attributes only when they help routing, querying, or asset generation.

### `spec.telemetry.signals`

Each signal section defines:

- whether the signal is enabled
- how the signal reaches the platform
- which backend class the platform should resolve

Guidance:

- Use `ingestion: collector` when the platform should route through a managed observability data plane.
- Use `ingestion: direct` only when the supported target adapter and backend path are designed for direct export.
- `backendClass` is a platform-facing abstraction, not a vendor SKU.
- When a signal is enabled, both `ingestion` and `backendClass` are required.
- For AWS Lambda, all enabled signals must currently use the same ingestion mode because bindings materialize one effective export strategy for the runtime.
- If a signal is disabled, clients may omit `ingestion` and `backendClass` for that signal.

### `spec.telemetry.signals.metrics.catalog`

This is one of the most important sections for self-service adoption.

Each metric should define:

- `name`
- `type`
- `unit`
- `description`

Recommended practice:

- include dimensions only when they are operationally useful
- avoid uncontrolled high-cardinality dimensions
- name the metric after a meaningful domain event or user-visible operation

### `spec.capabilities`

Use this section to request platform products:

- `dashboards`
- `alerts`
- `slos`
- `dataManagement`

Guidance:

- dashboards should align with the workload shape through a known `preset`
- alerts should capture business or operational conditions, not implementation-specific alarm resources
- SLOs should describe service objectives in user-facing terms
- data management should reflect retention and classification requirements

## Supported Values And Validation Rules

Use the schema for the exact machine-validated contract:

- [schemas/observability-contract-v1alpha1.schema.json](../schemas/observability-contract-v1alpha1.schema.json)

Important validation rules enforced today:

- `kind` must be `ObservabilityContract`
- `target: aws` requires `delivery.region`
- `target: kubernetes` requires both `delivery.cluster` and `delivery.namespace`
- when metrics are enabled, `metrics.catalog` must not be empty
- when a signal is enabled, `ingestion` and `backendClass` must both be set
- AWS Lambda contracts cannot mix `collector` and `direct` ingestion across enabled signals

## Contract Versus Runtime Bindings

The contract expresses observability intent. The client-facing runtime integration is completed later through target-specific bindings.

For AWS Lambda in particular, the contract does not directly encode:

- `instrumentation_mode`
- ADOT layer attachment details
- EMF compatibility mode
- the final collector endpoint chosen by the workflow or managed suite

Those values are resolved by the bindings path and returned to the client through `bindings.json` and the final handoff artifact.

## Effective Fields Today

Clients should treat these contract fields as effective today in validation, plan generation, or current AWS Lambda bindings:

- `metadata`
- `spec.delivery`
- `spec.telemetry.resourceAttributes`
- `spec.telemetry.openTelemetry.propagators`
- `spec.telemetry.signals.*.enabled`
- `spec.telemetry.signals.*.ingestion`
- `spec.telemetry.signals.*.backendClass`
- `spec.telemetry.signals.metrics.catalog`
- `spec.capabilities.*` as planning intent

Clients should treat these as valid model fields that are currently more declarative than operational in this repository:

- `spec.telemetry.openTelemetry.semanticConventions`
- `spec.telemetry.openTelemetry.sampling`
- `spec.service.tier`
- `spec.service.lifecycle`
- `status`

## Supported Starting Examples

- AWS Lambda:
  [examples/contracts/aws-lambda-order-processing.yaml](../examples/contracts/aws-lambda-order-processing.yaml)
- Kubernetes:
  [examples/contracts/kubernetes-payments.yaml](../examples/contracts/kubernetes-payments.yaml)
- Hybrid:
  [examples/contracts/monolith-crm.yaml](../examples/contracts/monolith-crm.yaml)

## Common Authoring Mistakes

- Omitting `delivery.region` for AWS workloads.
- Enabling metrics without defining a catalog.
- Enabling a signal without setting both `ingestion` and `backendClass`.
- Mixing `collector` and `direct` ingestion in AWS Lambda contracts and expecting separate per-signal runtime export paths.
- Treating `backendClass` as a vendor-specific implementation field.
- Using unstable runtime values as resource attributes.
- Requesting capabilities that are not yet supported by the target adapter in production flow.

## Validation Commands

Validate a contract:

```bash
go run ./cmd/o11yctl validate -f examples/contracts/aws-lambda-order-processing.yaml
```

Generate a plan:

```bash
go run ./cmd/o11yctl plan -f examples/contracts/aws-lambda-order-processing.yaml
```

Generate AWS Lambda bindings:

```bash
go run ./cmd/o11yctl bindings aws-lambda -f examples/contracts/aws-lambda-order-processing.yaml
```

## Companion References

- [observability-contract.md](observability-contract.md)
- [support-matrix.md](support-matrix.md)
- [consumer-interaction-flow.md](consumer-interaction-flow.md)
