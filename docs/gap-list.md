# Gap List

This document captures the main product gaps identified from the current repository state.

## 1. Platform-Managed Suite Provisioning

- The repository contains collector and managed-suite assets, but the operational deployment flow is not yet packaged as a reusable platform workflow at the root product boundary.
- Dashboard and alert execution are still stronger at the planning or asset-reference layer than at a fully generalized adapter execution layer.
- The platform needs a productized execution path for platform-managed observability suite provisioning that is independent from any single sample workload.

## 2. Remote Platform Execution Surface

- `o11yctl` supports `validate`, `plan`, and `bindings aws-lambda`, but it does not yet expose those capabilities through a remote platform interface.
- The repository does not yet expose a remote HTTP or gRPC control-plane interface.
- Idempotent convergence currently depends on the downstream execution backend rather than on a first-class platform execution service in this repository.

## 3. Self-Service Contract Experience

- The contract schema is strong, but clients need clearer authoring guidance than schema-only validation.
- Supported values and supported combinations need to be explicit from a product perspective.
- Clients need a field-by-field authoring guide, support matrix, and example-driven onboarding path.

## Recommended Execution Order

1. finalize the documentation set for a product-facing platform experience
2. productize the platform-managed suite deployment path
3. define and implement the remote `validate`, `plan`, and `bindings` control-plane surface
4. harden idempotent execution semantics and persistent state ownership
