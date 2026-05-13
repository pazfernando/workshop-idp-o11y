# Platform Consumer Flow

This document describes how an application team consumes the observability platform defined by this repository.

## Summary

The application team does not request dashboards, collectors, alarms, or cloud resources directly. Instead, it declares observability intent in an `ObservabilityContract`, validates that contract, generates a neutral plan, and consumes target-specific bindings produced by the platform.

## Main Actors

- Application team:
  authors the workload contract and applies the generated bindings.
- Platform product:
  validates the contract, builds the neutral plan, and maps intent to supported classes and assets.
- Target adapter:
  produces target-specific runtime bindings and platform integration outputs.

## Interaction Flow

1. The application team creates and versions an `ObservabilityContract` with its application code.
2. The platform validates the contract with `o11yctl validate`.
3. The platform generates a neutral `ProvisioningPlan` with `o11yctl plan`.
4. The target adapter materializes supported runtime bindings, currently through `o11yctl bindings aws-lambda`.
5. The application deployment injects the generated outputs into the workload runtime.
6. Platform-managed assets such as dashboards, alerts, collector configs, and managed-suite assets are selected from the catalog and asset layers.

## Concrete Example

The AWS Lambda example in [examples/contracts/aws-lambda-order-processing.yaml](../examples/contracts/aws-lambda-order-processing.yaml) follows this path:

1. The contract declares an AWS Lambda workload with OpenTelemetry enabled.
2. The contract requests collector-based ingestion for traces, metrics, and logs.
3. The contract requests dashboards, alerts, SLOs, and retention settings as platform capabilities.
4. The planner resolves the AWS Lambda bindings class, collector class, and relevant assets.
5. The AWS adapter produces Lambda-ready bindings so the workload can integrate with the platform without hardcoding observability infrastructure details.

## CLI Walkthrough

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

## Related Documents

- [README.md](../README.md)
- [architecture.md](architecture.md)
- [operating-model.md](operating-model.md)
- [observability-contract.md](observability-contract.md)
- [contract-authoring-guide.md](contract-authoring-guide.md)
- [support-matrix.md](support-matrix.md)
