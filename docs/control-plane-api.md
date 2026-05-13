# Control-Plane API

This repository now exposes a minimal HTTP control plane through `o11yd`.

## Purpose

The HTTP service exposes the same core product operations that already exist in the CLI:

- contract validation
- backend-neutral planning
- AWS Lambda binding generation

The service is intentionally stateless. It validates the submitted contract and returns the requested output immediately. Persistent convergence and long-lived execution state still belong to the downstream execution backend.

## Run The Service

```bash
go run ./cmd/o11yd -listen :8080
```

Available flags:

- `-listen`
- `-read-timeout`
- `-write-timeout`
- `-idle-timeout`

## Endpoints

| Endpoint | Method | Purpose |
| :--- | :--- | :--- |
| `/healthz` | `GET` | Liveness probe |
| `/v1/validate` | `POST` | Validate an `ObservabilityContract` |
| `/v1/plan` | `POST` | Build a backend-neutral `ProvisioningPlan` |
| `/v1/bindings/aws-lambda` | `POST` | Build AWS Lambda runtime bindings |

## Request Model

Validation and planning accept the same request shape:

```json
{
  "contract": {
    "apiVersion": "observability.platform/v1alpha1",
    "kind": "ObservabilityContract"
  }
}
```

AWS Lambda bindings accept:

```json
{
  "contract": {
    "apiVersion": "observability.platform/v1alpha1",
    "kind": "ObservabilityContract"
  },
  "options": {
    "instrumentationMode": "code",
    "collectorEndpoint": "http://collector.internal:4318"
  }
}
```

The request contract is submitted as JSON and must conform to the same schema and semantic rules enforced by `o11yctl`.

## Example Calls

Validate:

```bash
curl -X POST http://127.0.0.1:8080/v1/validate \
  -H 'Content-Type: application/json' \
  -d @examples/generated/validate-request.json
```

Plan:

```bash
curl -X POST http://127.0.0.1:8080/v1/plan \
  -H 'Content-Type: application/json' \
  -d @examples/generated/plan-request.json
```

AWS Lambda bindings:

```bash
curl -X POST http://127.0.0.1:8080/v1/bindings/aws-lambda \
  -H 'Content-Type: application/json' \
  -d @examples/generated/aws-lambda-bindings-request.json
```

## Response Model

- Successful validation returns `200` with `{ "valid": true, "contractRef": ... }`.
- Successful planning returns `200` with the JSON `ProvisioningPlan`.
- Successful bindings generation returns `200` with the JSON AWS Lambda bindings payload.
- Invalid input returns `400` with `{ "error": "..." }`.

## Current Boundary

The HTTP control plane is suitable for remote CI/CD usage and thin platform API wrappers.

It does not yet provide:

- gRPC
- asynchronous execution
- persistent reconciliation state
- long-running platform ownership of infrastructure convergence
