# Observability Deployment Guide

Short guide for choosing `code` vs `adot_layer` and `direct` vs `collector` in this repository.

## Decisions

There are two axes:

1. Who initializes OpenTelemetry
2. Where telemetry is exported

### 1. Who Initializes OpenTelemetry

| Variable | Values | What it means | When to use it |
| :--- | :--- | :--- | :--- |
| `OTEL_MODE` | `code` | The Lambda itself initializes the OTel SDK | Repository default |
| `OTEL_MODE` | `adot_layer` | An ADOT Lambda Layer initializes OTel before the handler | Useful for direct CloudWatch export |
| `ADOT_LAMBDA_LAYER_ARN` | ARN or empty | ARN of the ADOT layer | Required when `OTEL_MODE=adot_layer` |

Notes for this repository:

- with the AWS-managed Node.js layer used here, the effective wrapper is `AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler`
- when `OTEL_MODE=adot_layer`, Terraform attaches `CloudWatchLambdaApplicationSignalsExecutionRolePolicy` to the Lambda execution roles

### 2. Where Telemetry Is Exported

| Variable | Values | What it means | When to use it |
| :--- | :--- | :--- | :--- |
| `OTEL_EXPORT_STRATEGY` | `direct` | The Lambda exports OTLP directly to the final backend | Current operational default |
| `OTEL_EXPORT_STRATEGY` | `collector` | The Lambda exports OTLP to a Collector first | In this repository it provisions the EC2 suite with Alloy |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | URL or empty | Base OTLP endpoint of the final backend | Only with `direct` |
| `OTEL_COLLECTOR_ENDPOINT` | URL or empty | Base OTLP endpoint of the Collector | Only with `collector`; optional when Terraform infers Alloy |

Important OTLP/HTTP rule:

- if you use a base endpoint such as `http://host:4318`, the SDK derives `.../v1/traces` and `.../v1/metrics`
- if you use per-signal variables, you must pass the full URL, for example `http://host:4318/v1/traces` and `http://host:4318/v1/metrics`

## Combinations

| Mode | What it does | Main advantage | Cost / tradeoff |
| :--- | :--- | :--- | :--- |
| `code + direct` | The app starts OTel and exports directly | Simpler | Not suitable for direct CloudWatch OTLP |
| `code + collector` | The app starts OTel and exports to a Collector | Supported path for custom business metrics in Grafana | Requires a Collector |
| `adot_layer + direct` | ADOT starts OTel and exports directly | Enables direct CloudWatch export | Requires SigV4 |
| `adot_layer + collector` | ADOT starts OTel and exports to a Collector | Do not use in this repository | The deployment blocks it to avoid a false positive for metrics |

### Signals Supported By Combination

| Combination | Traces | Custom business metrics | State in this repository |
| :--- | :--- | :--- | :--- |
| `code + direct` | Yes | Yes, to generic non-AWS OTLP | Supported |
| `code + collector` | Yes | Yes, to Alloy/Prometheus/Grafana | Supported; this repository runs `forceFlush` on each invocation |
| `adot_layer + direct` | Yes | Yes, for direct CloudWatch OTLP | Supported |
| `adot_layer + collector` | Yes, potentially | Not guaranteed for this repository's custom metrics | Unsupported and blocked |

## EC2 Suite For `collector`

When `OTEL_EXPORT_STRATEGY=collector`, this repository uses a single EC2 instance for:

- `Alloy` as the OTLP collector
- `Prometheus` as the metrics backend
- `Tempo` as the traces backend
- `Grafana` as the UI
- `Loki` as a backend ready for future OTLP logs

Current scope:

- OTLP metrics: supported and visualized in Grafana through Prometheus
- in Lambda, custom business metrics are forced to export at the end of each invocation so they do not rely only on the SDK timer
- OTLP traces: supported and explorable in Grafana through Tempo
- the provisioned Grafana dashboard includes a `Trace View` guide, but span exploration happens in `Explore` with the `Tempo` data source
- OTLP logs: the collector and Loki are ready when the app starts emitting them
- networking: it tries to use a public subnet in the region first and falls back to the first available subnet otherwise
- Grafana access: it uses `admin` and reads the password from GitHub `Secrets` via `OBSERVABILITY_SUITE_GRAFANA_ADMIN_PASSWORD` when available; otherwise it falls back to `Variables` and then to a random Terraform-generated password

## Recommendation

### Today

Use:

```text
OTEL_MODE=code
OTEL_EXPORT_STRATEGY=direct
OBSERVABILITY_EMF_COMPATIBILITY_MODE=true
```

This keeps the repository functional without requiring a Collector.

### Recommended Next Step

Use:

```text
OTEL_MODE=code
OTEL_EXPORT_STRATEGY=collector
OTEL_COLLECTOR_ENDPOINT=
OBSERVABILITY_EMF_COMPATIBILITY_MODE=true
```

This decouples instrumentation from the final backend.

### Direct CloudWatch

Use:

```text
OTEL_MODE=adot_layer
ADOT_LAMBDA_LAYER_ARN=arn:aws:lambda:...
OTEL_EXPORT_STRATEGY=direct
```

Use the ADOT layer for direct CloudWatch OTLP when that is the operational goal.

## Relevant Variables

### Base Stack

| Variable | Allowed values | Required | Recommended |
| :--- | :--- | :--- | :--- |
| `STACK_NAME` | string | Sí | `observability-business-case` |
| `RESOURCE_PREFIX` | string | Sí | `aws-dev-1` |
| `AWS_REGION` | valid AWS region | Sí | `us-east-1` |
| `PAYMENT_FAILURE_MODE` | `none`, `always_fail`, `random_fail`, `slow_response`, `random_reject` | Sí | `random_fail` |
| `LOG_RETENTION_IN_DAYS` | positive integer | No | `7` |
| `METRICS_NAMESPACE` | string | No | `Workshop/OrderProcessing` |

### OTel and ADOT

| Variable | Allowed values | Required | Recommended |
| :--- | :--- | :--- | :--- |
| `OTEL_MODE` | `code`, `adot_layer` | Sí | `code` |
| `ADOT_LAMBDA_LAYER_ARN` | ARN or empty | Only with `adot_layer` | empty |
| `OTEL_EXPORT_STRATEGY` | `direct`, `collector` | Sí | `direct` today |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | URL or empty | Only with `direct` | empty |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | URL or empty | No | empty; for OTLP/HTTP it must end with `/v1/traces` |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | URL or empty | No | empty; for OTLP/HTTP it must end with `/v1/metrics` |
| `OTEL_COLLECTOR_ENDPOINT` | URL or empty | Only with `collector` | empty to infer Alloy |
| `OTEL_COLLECTOR_TRACES_ENDPOINT` | URL or empty | No | empty; for OTLP/HTTP it must end with `/v1/traces` |
| `OTEL_COLLECTOR_METRICS_ENDPOINT` | URL or empty | No | empty; for OTLP/HTTP it must end with `/v1/metrics` |
| `OTEL_METRIC_EXPORT_INTERVAL_MS` | positive integer | No | `10000` |
| `OBSERVABILITY_EMF_COMPATIBILITY_MODE` | `true`, `false` | No | `false` |

### Dashboards and Alarms

| Variable | Allowed values | Required | Recommended |
| :--- | :--- | :--- | :--- |
| `CREATE_OBSERVABILITY_DASHBOARD` | `true`, `false` | No | `true` |
| `CREATE_OBSERVABILITY_ALARMS` | `true`, `false` | No | `true` |
| `API_5XX_ALARM_THRESHOLD` | non-negative integer | No | `1` |
| `ORDER_PROCESSOR_ERROR_ALARM_THRESHOLD` | non-negative integer | No | `1` |
| `PAYMENT_LATENCY_ALARM_THRESHOLD_MS` | non-negative integer | No | `3000` |

### Remote State

| Variable | Allowed values | Required | Recommended |
| :--- | :--- | :--- | :--- |
| `TF_STATE_BUCKET` | S3 bucket or empty | No | empty if the workflow creates it |
| `TF_STATE_KEY` | S3 key or empty | No | `${environment}/${RESOURCE_PREFIX}-${STACK_NAME}.tfstate` |

### EC2 Suite

| Variable | Allowed values | Required | Recommended |
| :--- | :--- | :--- | :--- |
| `OBSERVABILITY_SUITE_INSTANCE_TYPE` | valid EC2 type | No | `t3.small` |
| `OBSERVABILITY_SUITE_GRAFANA_ADMIN_PASSWORD` | string or empty | No | empty |
| `OBSERVABILITY_SUITE_ROOT_VOLUME_SIZE_GB` | positive integer | No | `20` |
| `OBSERVABILITY_SUITE_GRAFANA_ALLOWED_CIDRS` | JSON list of CIDRs | No | `["0.0.0.0/0"]` |
| `OBSERVABILITY_SUITE_OTLP_ALLOWED_CIDRS` | JSON list of CIDRs | No | `["0.0.0.0/0"]` |
| `OBSERVABILITY_SUITE_SSH_ALLOWED_CIDRS` | JSON list of CIDRs | No | `[]` |

## Examples

### Example 1: Simple Repository Deployment

```bash
export OTEL_MODE="code"
export OTEL_EXPORT_STRATEGY="direct"
export OTEL_EXPORTER_OTLP_ENDPOINT=""
export OBSERVABILITY_EMF_COMPATIBILITY_MODE="false"
```

### Example 2: Direct To A Generic OTLP Backend

```bash
export OTEL_MODE="code"
export OTEL_EXPORT_STRATEGY="direct"
export OTEL_EXPORTER_OTLP_ENDPOINT="https://your-otlp-backend.example.com"
export OBSERVABILITY_EMF_COMPATIBILITY_MODE="false"
```

Note:

- CloudWatch does not use a single shared base endpoint for traces and metrics.
- For direct CloudWatch export, use `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=https://xray.<region>.amazonaws.com/v1/traces` and `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=https://monitoring.<region>.amazonaws.com/v1/metrics`.
- CloudWatch OTLP endpoints require `SigV4` authentication.
- In this repository, the supported path for direct CloudWatch export is `OTEL_MODE=adot_layer`.
- In this repository, the effective wrapper for that Node.js layer is `/opt/otel-handler`.

### Example 3: Deployment With A Collector

```bash
export OTEL_MODE="code"
export OTEL_EXPORT_STRATEGY="collector"
export OTEL_COLLECTOR_ENDPOINT="http://collector.internal:4318"
export OBSERVABILITY_EMF_COMPATIBILITY_MODE="false"
```

Expected result:

- custom business metrics reach Alloy and Prometheus
- each handler runs `forceFlush` before invocation end so Lambda does not leave OTLP metrics unexported
- if `OTEL_COLLECTOR_ENDPOINT` is an OTLP/HTTP base endpoint, the SDK derives `.../v1/traces` and `.../v1/metrics` from that base
- Grafana can visualize them through the `Prometheus` data source
- OTLP traces still reach Alloy and Tempo
- to inspect spans, use `Grafana > Explore > Tempo`; the dashboard `Trace View` panel only documents the recommended navigation

## Local Test

Without deploying, you can validate the instrumentation path like this:

```bash
npm run test:otel-local
```

The test checks:

- `code` mode: OTLP HTTP export to `/v1/metrics`
- `adot_layer` mode: explicit `bootstrap skipped` log

### Example 4: Unsupported Combination In This Repository

```bash
export OTEL_MODE="adot_layer"
export ADOT_LAMBDA_LAYER_ARN="arn:aws:lambda:<region>:<account-or-publisher>:layer:<adot-layer-name>:<version>"
export OTEL_EXPORT_STRATEGY="collector"
export OTEL_COLLECTOR_ENDPOINT="http://collector.internal:4318"
```

Expected result:

- the deployment fails explicitly
- the reason is that, in this repository, `adot_layer + collector` does not guarantee that custom business metrics reach the Collector
- use `code + collector` for Grafana/Alloy/Prometheus or `adot_layer + direct` for direct CloudWatch OTLP

### Example 5: Deployment With ADOT Layer + Direct CloudWatch

```bash
export OTEL_MODE="adot_layer"
export ADOT_LAMBDA_LAYER_ARN="arn:aws:lambda:<region>:<account-or-publisher>:layer:<adot-layer-name>:<version>"
export OTEL_EXPORT_STRATEGY="direct"
export OTEL_EXPORTER_OTLP_ENDPOINT=""
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=""
export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=""
```

Expected result:

- Terraform infers `https://xray.<region>.amazonaws.com/v1/traces`
- Terraform infers `https://monitoring.<region>.amazonaws.com/v1/metrics`
- the `effective_otlp_authentication_mode` output is `sigv4`, which indicates a requirement of the effective OTLP backend, not a manual workflow input
- the `effective_lambda_exec_wrapper` output is `/opt/otel-handler`
- the `application_signals_execution_role_policy_enabled` output is `true`

### Example 6: EC2 Suite With Alloy + Grafana + Prometheus + Tempo + Loki

```bash
export OBSERVABILITY_SUITE_INSTANCE_TYPE="t3.small"
export OTEL_EXPORT_STRATEGY="collector"
export OTEL_COLLECTOR_ENDPOINT=""
export OTEL_COLLECTOR_TRACES_ENDPOINT=""
export OTEL_COLLECTOR_METRICS_ENDPOINT=""
```

Expected result:

- Terraform creates an EC2 instance with Grafana, Alloy, Prometheus, Tempo, and Loki
- Terraform infers the Alloy HTTP endpoint for traces and metrics if `collector` is active and you did not provide explicit endpoints
- Grafana is provisioned with a dashboard for the current business metrics

## Reference Collector

The repository includes these configurations:

- [collector-cloudwatch.yaml](/Users/pazfernando/Documents/projects/windsurf/workshop-order-processing/infra/otel-collector/collector-cloudwatch.yaml)
- [collector-cloudwatch-third-party.yaml](/Users/pazfernando/Documents/projects/windsurf/workshop-order-processing/infra/otel-collector/collector-cloudwatch-third-party.yaml)

### What `collector-cloudwatch.yaml` Does

- receives OTLP on `4317` and `4318`
- applies `memory_limiter`
- applies `batch`
- filters health checks
- removes sensitive or high-cardinality attributes
- performs `tail_sampling`
- exports to CloudWatch OTLP metrics and X-Ray traces

`collector-cloudwatch-third-party.yaml` does the same, but also forwards to an additional OTLP backend.

## Workflows

### `deploy.yml`

`deploy.yml` installs dependencies, packages Lambdas, validates observability, and runs `terraform apply`.

Observability validation rules:

- if `OTEL_MODE=adot_layer`, `ADOT_LAMBDA_LAYER_ARN` must exist
- if `OTEL_EXPORT_STRATEGY=collector`, you can leave `OTEL_COLLECTOR_ENDPOINT` empty so Terraform infers Alloy, or define it if you want an external Collector
- if `OTEL_EXPORT_STRATEGY=direct` and `OTEL_MODE=adot_layer`, leaving the direct endpoints empty makes Terraform infer per-signal CloudWatch OTLP endpoints
- if `OTEL_EXPORT_STRATEGY=direct` and `OTEL_MODE=code`, do not use CloudWatch OTLP endpoints because this repository does not sign SigV4 requests from the code bootstrap
- if `OTEL_EXPORT_STRATEGY=collector`, you can leave `OTEL_COLLECTOR_TRACES_ENDPOINT` and `OTEL_COLLECTOR_METRICS_ENDPOINT` empty; Terraform infers them toward Alloy

`teardown.yml` reuses the same variables to destroy the stack and now confirms against the effective name, including `RESOURCE_PREFIX` when applicable.
