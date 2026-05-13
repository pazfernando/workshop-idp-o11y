# Architecture

## Layers

1. Contract
   Versioned document owned by the application that expresses declarative intent.

2. Validation
   JSON Schema plus target-specific semantic rules.

3. Planner
   Translates intent into a backend-independent `ProvisioningPlan`, including references to classes and assets.

4. Catalog
   Defines reusable classes and official paths to platform assets.

5. Assets
   Groups collector configs, managed suites, and dashboard or alert templates that are decoupled from the consumer.

6. Adapters
   Materialize the plan into concrete technologies and deliver bindings or outputs to the consumer.

## Decisions

- `ObservabilityContract` is the stable input.
- `ProvisioningPlan` is the stable intermediate output.
- catalog classes describe reusable presets that can be selected per target.
- adapters are replaceable.

This separation prevents the application from needing to know details such as:

- exact OTLP URLs
- dashboard names
- collector wiring
- cloud-specific resources

## Model-Supported Targets

- `aws`
- `kubernetes`
- `hybrid`

## Covered Workloads

- distributed applications
- monoliths
- workloads running on Kubernetes
- serverless

## Extension Path

Additional adapters can be added without changing the contract model:

- `adapter/aws`
- `adapter/kubernetes`
- `adapter/grafana`
- `adapter/cloudwatch`
- `adapter/openslo`
