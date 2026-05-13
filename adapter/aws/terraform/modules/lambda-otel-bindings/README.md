# lambda-otel-bindings

Reusable Terraform module from the platform product to materialize OpenTelemetry bindings for AWS Lambda workloads.

Main inputs:

- a contract already validated and translated into Terraform values by the platform pipeline
- the `collector` or `direct` export strategy
- OTLP endpoints from the platform-managed data plane
- the `code` or `adot_layer` instrumentation mode

Outputs:

- `environment`
- `layers`
- `managed_policies`
- effective OTLP endpoints

Expected usage:

1. the application team submits the contract to the platform
2. the planner generates neutral intent
3. the AWS adapter resolves endpoints and policies
4. this module returns bindings ready for workload Terraform
