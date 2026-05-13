# Gap List

This document captures the remaining product gaps after the current repository implementation.

## Closed In The Current Repository State

- A root Terraform path now exists for the platform-managed observability suite.
- Root GitHub Actions workflows now exist to apply or destroy that suite.
- A remote HTTP surface now exists for `validate`, `plan`, and `bindings aws-lambda`.
- A reusable GitHub workflow now exists as the recommended CI/CD consumption interface for GitHub clients.
- The self-service documentation set now includes the contract guide, support matrix, consumer flow, managed-suite guide, and control-plane API guide.

## 1. Persistent Execution Ownership

- The HTTP control plane is stateless.
- Idempotent convergence still depends on Terraform state and the execution backend rather than on repository-owned reconciliation state.
- The platform does not yet persist contract submissions, plans, or binding history.

## 2. Generalized Capability Execution

- Dashboards, alerts, SLOs, and retention are well represented in the contract and plan layers.
- Their backend execution is still less generalized than AWS Lambda runtime bindings.
- The repository still needs a broader adapter execution story beyond asset references and planning metadata.

## 3. Expanded Target Coverage

- AWS Lambda is still the only target with concrete runtime adapter output.
- Kubernetes and hybrid targets still stop at validation and planning.
- The platform still needs additional target adapters if it wants uniform multi-runtime execution.

## 4. Control-Plane Hardening

- There is no gRPC surface yet.
- There is no authentication, authorization, or tenancy layer in `o11yd`.
- There is no asynchronous execution model, job tracking, or long-running operation API.

## Recommended Execution Order

1. harden the reusable workflow interface, client output contract, and workflow-based platform semantics
2. harden persistent execution semantics and platform-owned state
3. generalize capability execution beyond planning metadata
4. expand target adapter coverage beyond AWS Lambda
5. harden the remote control plane with auth, tenancy, and asynchronous operations only if that interface becomes a primary product path
