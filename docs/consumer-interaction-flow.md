# Platform Consumer Flow

This document describes how an application team consumes the observability platform defined by this repository.

## Summary

The application team does not request dashboards, collectors, alarms, or cloud resources directly. Instead, it declares observability intent in an `ObservabilityContract` and consumes the platform through the reusable workflow interface. The platform validates the contract, generates a neutral plan, optionally provisions the managed suite, and returns target-specific bindings and a client handoff package.

## Main Actors

- Application team:
  authors the workload contract, calls the platform workflow, and applies the generated bindings.
- Platform product:
  validates the contract, builds the neutral plan, maps intent to supported classes and assets, and packages the resulting outputs.
- Target adapter:
  produces target-specific runtime bindings and platform integration outputs.
- Managed suite path:
  optionally provisions the platform-managed observability data plane when requested by the client workflow.

## Interaction Flow

1. The application team creates and versions an `ObservabilityContract` with its application code.
2. The application team calls the reusable workflow documented in [client-consumption.md](client-consumption.md).
3. The platform validates the contract.
4. The platform generates a neutral `ProvisioningPlan`.
5. If requested, the platform provisions the managed suite and resolves its OTLP endpoint through the remote S3-backed Terraform state path.
6. The target adapter materializes supported runtime bindings, currently through the AWS Lambda path.
7. The platform publishes technical artifacts and a client-facing handoff package.
8. The application deployment injects the generated outputs into the workload runtime.
9. Platform-managed assets such as dashboards, alerts, collector configs, and managed-suite assets are selected from the catalog and asset layers.

## Concrete Example

The AWS Lambda example in [examples/contracts/aws-lambda-order-processing.yaml](../examples/contracts/aws-lambda-order-processing.yaml) follows this path:

1. The client repository stores the contract next to its application code.
2. The client workflow calls the platform reusable workflow.
3. The contract requests collector-based ingestion for traces, metrics, and logs.
4. The contract metric catalog is validated against the selected supported dashboard preset.
5. The platform can provision the managed suite if the client wants platform-owned collector routing.
6. The planner resolves the AWS Lambda bindings class, collector class, relevant assets, and preset metric support metadata.
7. The AWS adapter produces Lambda-ready bindings.
8. The client receives `plan.json`, `bindings.json`, and the final `o11y-client-handoff` artifact.
9. The handoff package records the effective export strategy, endpoint source, metric support policy, requested preset metrics, adapter notes, required follow-up inputs, and produced artifacts so the client deployer knows exactly how observability was materialized.

## Workflow-First Consumption

The recommended client entrypoint is the reusable workflow in [client-consumption.md](client-consumption.md).

That workflow wraps:

- contract validation
- neutral plan generation
- optional managed-suite provisioning
- AWS Lambda bindings generation
- handoff artifact generation for the client team

## Advanced CLI Walkthrough

The CLI remains useful for local inspection, debugging, or internal platform development.

Validate a contract:

```bash
go run ./cmd/o11yctl validate -f examples/contracts/aws-lambda-order-processing.yaml
```

Generate a neutral provisioning plan:

```bash
go run ./cmd/o11yctl plan -f examples/contracts/aws-lambda-order-processing.yaml
```

Generate AWS Lambda bindings:

```bash
go run ./cmd/o11yctl bindings aws-lambda \
  -f examples/contracts/aws-lambda-order-processing.yaml \
  --collector-endpoint http://collector.internal:4318
```

## Outputs Received By The Application Team

The application team receives stable outputs rather than platform internals. Depending on target and adapter mode, those outputs can include:

- OpenTelemetry environment variables
- OTLP collector or direct-export endpoints
- Lambda layers or execution policy requirements
- references to dashboards and platform-managed observability assets
- the final `o11y-client-handoff` artifact with an implementation summary and manifest that describe the effective runtime integration, not only the requested inputs

## Related Documents

- [README.md](../README.md)
- [architecture.md](architecture.md)
- [client-consumption.md](client-consumption.md)
- [operating-model.md](operating-model.md)
- [observability-contract.md](observability-contract.md)
- [contract-authoring-guide.md](contract-authoring-guide.md)
- [support-matrix.md](support-matrix.md)
