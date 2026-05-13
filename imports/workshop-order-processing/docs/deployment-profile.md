# Deployment Profile

This document describes **how the solution is deployed by default** when the repository workflow is used without additional overrides.

It does not describe the ideal target architecture. It describes the **actual operational state** that SRE / DevOps should assume if they only run the standard deployment.

## Default Workflow Profile

### OpenTelemetry Initialization

| Variable | Default value | Resulting state |
| :--- | :--- | :--- |
| `OTEL_MODE` | `code` | The Lambda tries to initialize OTel from `src/shared/otel-bootstrap.js` |
| `ADOT_LAMBDA_LAYER_ARN` | empty | No ADOT Lambda Layer is attached |

Result:

- the default deployment **does not use `adot_layer`**
- the OTel bootstrap remains in application code

### Export Path

| Variable | Default value | Resulting state |
| :--- | :--- | :--- |
| `OTEL_EXPORT_STRATEGY` | `direct` | No Collector is expected as a mandatory dependency |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | empty | No direct OTLP endpoint is configured |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | empty | No trace override |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | empty | No metrics override |
| `OTEL_COLLECTOR_ENDPOINT` | empty | Collector not used |

Result:

- the default deployment **does not use a Collector**
- the default deployment **does not export OTLP to any backend**
- with `OTEL_MODE=code`, leaving `direct` without endpoints does not infer CloudWatch and keeps OTLP inactive

## Which Observability Is Actually Active

Although the code is already prepared with `otel-first` abstractions, under the default profile the effective observability remains:

| Signal | Default active mechanism |
| :--- | :--- |
| Logs | CloudWatch Logs |
| Current business metrics | EMF to CloudWatch Metrics |
| AWS Lambda traces | X-Ray |
| OTLP export | Inactive because no endpoint is configured |
| Collector | Inactive / not required |
| ADOT Layer | Inactive / not attached |

## Effective Default Variables

| Variable | Default value |
| :--- | :--- |
| `STACK_NAME` | `observability-business-case` |
| `RESOURCE_PREFIX` | `aws-dev` en GitHub Actions |
| `AWS_REGION` | `us-east-1` |
| `PAYMENT_FAILURE_MODE` | `random_fail` |
| `LOG_RETENTION_IN_DAYS` | `7` |
| `METRICS_NAMESPACE` | `Workshop/OrderProcessing` |
| `OTEL_MODE` | `code` |
| `OTEL_EXPORT_STRATEGY` | `direct` |
| `ADOT_LAMBDA_LAYER_ARN` | empty |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | empty |
| `OTEL_COLLECTOR_ENDPOINT` | empty |
| `OBSERVABILITY_EMF_COMPATIBILITY_MODE` | `false` |
| `CREATE_OBSERVABILITY_DASHBOARD` | `true` |
| `CREATE_OBSERVABILITY_ALARMS` | `true` |

## What Changes If SRE / DevOps Applies Overrides

### Case 1: Enable Direct OTLP Export

You must define:

```text
OTEL_MODE=code
OTEL_EXPORT_STRATEGY=direct
OTEL_EXPORTER_OTLP_ENDPOINT=https://...
```

Result:

- the application still initializes OTel from code
- the Lambdas export OTLP directly to the configured backend
- this case is for OTLP backends that do not depend on SigV4

### Case 1b: Enable Direct CloudWatch Export With ADOT Layer

You must define:

```text
OTEL_MODE=adot_layer
OTEL_EXPORT_STRATEGY=direct
ADOT_LAMBDA_LAYER_ARN=arn:aws:lambda:...
OTEL_EXPORTER_OTLP_ENDPOINT=
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=
```

Result:

- the ADOT layer initializes OTel before the handler
- Terraform infers `https://xray.<region>.amazonaws.com/v1/traces`
- Terraform infers `https://monitoring.<region>.amazonaws.com/v1/metrics`
- the effective direct OTLP backend requires `SigV4` authentication
- the expected effective wrapper for this Node.js repository is `/opt/otel-handler`
- Terraform attaches `CloudWatchLambdaApplicationSignalsExecutionRolePolicy` to the execution roles

### Case 2: Enable a Collector

You must define:

```text
OTEL_MODE=code
OTEL_EXPORT_STRATEGY=collector
OTEL_COLLECTOR_ENDPOINT=
```

Result:

- the application still initializes OTel from code
- the Lambdas export OTLP to the Collector
- each handler runs `forceFlush` before invocation end to avoid leaving short-lived OTLP metrics in memory
- when `OTEL_COLLECTOR_ENDPOINT` is empty, Terraform provisions the EC2 suite and infers Alloy for traces and metrics
- if you define `OTEL_COLLECTOR_ENDPOINT`, you point to an external Collector instead of the inferred Alloy
- Grafana receives spans in `Tempo`; the provisioned dashboard adds a `Trace View` guide, but detailed trace exploration happens in `Explore`

### Case 3: Try ADOT Layer + Collector

You must define:

```text
OTEL_MODE=adot_layer
ADOT_LAMBDA_LAYER_ARN=arn:aws:lambda:...
OTEL_EXPORT_STRATEGY=collector
OTEL_COLLECTOR_ENDPOINT=
```

Result:

- the deployment fails explicitly
- in this repository, that combination is not considered supported for custom business metrics
- use `code + collector` for Grafana/Alloy/Prometheus or `adot_layer + direct` for direct CloudWatch OTLP

## Current Operational Recommendation

### To Operate The Repository Today Without Adding Extra Infrastructure

Assume this profile:

```text
OTEL_MODE=code
OTEL_EXPORT_STRATEGY=direct
OTEL_EXPORTER_OTLP_ENDPOINT=
OBSERVABILITY_EMF_COMPATIBILITY_MODE=false
```

Interpretation:

- the design convention is `otel-first`
- but the effective operation still relies on CloudWatch Logs, EMF, and X-Ray
- direct OTLP to CloudWatch is not inferred in this profile because bootstrap remains in code
- the EC2 suite with Grafana/Alloy/Prometheus/Tempo/Loki is not part of the default profile; it appears when you switch to `OTEL_EXPORT_STRATEGY=collector`

### To Move Toward A More Mature Grafana-Based Operation

The next recommended transition is:

```text
OTEL_MODE=code
OTEL_EXPORT_STRATEGY=collector
OTEL_COLLECTOR_ENDPOINT=
```

Interpretation:

- bootstrap remains in code because that is where this repository's custom business metrics live
- the Collector becomes the central routing point
- Grafana/Alloy/Prometheus receive the workshop OTLP metrics
- this path also depends on `forceFlush` per invocation so Lambda does not wait for the SDK export interval
- if you do not provide an external endpoint, Terraform uses the workshop EC2 suite
