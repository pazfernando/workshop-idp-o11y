# AWS Lambda Consumer Example

This example shows how a client repository can consume the platform product through the reusable workflow interface instead of invoking `o11yctl` directly.

## Files

- workflow example:
  `.github/workflows/consume-observability-platform.yml`
- contract example:
  `../../contracts/aws-lambda-order-processing.yaml`
- Terraform example for local binding consumption:
  `main.tf`

## What The Workflow Demonstrates

The example workflow:

1. calls the platform reusable workflow
2. validates the contract
3. builds the neutral plan
4. deploys the managed suite
5. generates AWS Lambda bindings
6. passes the resulting JSON outputs to a downstream deployment job

## Important Adaptations For A Real Client Repository

Before using this workflow in a real repository, replace:

- `your-org/workshop-iidp-o11y` with the actual platform repository
- `contract_path` with the path of the contract inside the client repository
- CIDR inputs with the ranges appropriate for the client environment
- the placeholder deployment step with the real deployment logic that consumes `bindings.json`

## Migration Intent

This example is the concrete switch target for clients that previously ran:

- `o11yctl validate`
- `o11yctl plan`
- `o11yctl bindings aws-lambda`

directly in their own workflow.
