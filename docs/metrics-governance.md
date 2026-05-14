# Metrics Governance

This document defines the intended product model for workload metrics, dashboard coverage, and metric approval in the observability platform.

## Problem Statement

If every client defines arbitrary metrics with arbitrary dimensions, the platform quickly becomes hard to operate:

- metric names drift across teams
- semantic meaning becomes inconsistent
- dashboards cannot be reused safely
- alert and SLO semantics become non-comparable
- dimensions can explode cardinality and cost
- target adapters cannot provide predictable coverage

For that reason, workload metrics should be governed platform API objects, not free-form backend-specific definitions.

## Product Direction

The platform should converge on a governed metric catalog model:

- clients select from a platform-owned catalog of approved metrics
- dashboard presets consume approved metrics from that catalog
- adapters declare which catalog metrics they support and how they materialize them
- clients can propose new metrics, but new definitions should not bypass governance in the normal production flow

## Current Enforced Policy

In the current repository state, the standard product path is stricter than the long-term vision:

- only metrics that belong to supported dashboard presets are accepted
- free-form client-defined production metrics outside preset coverage are rejected
- plan outputs, bindings notes, and handoff artifacts report this as a `preset-only` metric support policy

## Roles And Responsibilities

### Client Team

The client team should:

- select approved metrics that match the workload
- request dashboard presets
- propose new metrics only when an approved metric does not satisfy a legitimate workload need
- instrument the application to emit the approved metric set

The client team should not:

- create arbitrary production metrics with uncontrolled names or dimensions
- treat dashboards as backend-specific handwritten resources in the standard platform flow

### Platform / Observability Team

The platform or observability team should:

- own the governed metric catalog
- define semantic rules for names, units, dimensions, and cardinality
- define preset-to-metric mappings
- define adapter support coverage for each approved metric
- review and approve new metric proposals

## Governed Metric Catalog

Each governed metric should have metadata similar to the following conceptual model:

```yaml
id: business.order.created.count
displayName: Orders Created
signal: metrics
otelName: business.order.created.count
type: counter
unit: "{order}"
description: Successful order creation events.
allowedDimensions:
  - tenant
  - region
  - result
cardinalityPolicy:
  maxRecommendedDimensions: 3
  forbiddenDimensions:
    - user_id
    - request_id
support:
  aws-lambda:
    status: supported
    dashboardPresets:
      - serverless-api
  kubernetes:
    status: planned
lifecycle:
  status: approved
  owner: observability-platform
```

The key design point is that the metric identifier is stable and governed, while adapter-specific materialization stays behind the platform boundary.

## Dashboard Model

Dashboard ownership should also be split cleanly:

- the client requests dashboard intent through a preset
- the platform owns the preset definition
- the preset maps to a known set of governed metrics
- the adapter renders the best backend-specific dashboard implementation available for that target

That means the client does not define the final CloudWatch or Grafana dashboard layout in the standard platform path.

## Adapter Coverage Model

Each adapter should declare per governed metric:

- `supported`
- `partially_supported`
- `not_supported`
- backend materialization notes
- supported dimensions
- dashboard preset coverage
- alert or SLO eligibility where relevant

This is what prevents the metric contract from becoming ambiguous across runtimes and targets.

## Contract Direction

The current repository still models `spec.telemetry.signals.metrics.catalog` as free-form metric definitions.

The intended evolution is:

1. keep the current field for transition compatibility
2. introduce governed metric references as the primary path
3. restrict free-form custom metric definitions to a proposal workflow

Conceptually, the contract should move toward something like:

```yaml
metrics:
  enabled: true
  ingestion: collector
  backendClass: metrics-standard
  catalogRefs:
    - business.order.created.count
    - http.server.request.duration
```

An optional extension path could remain available:

```yaml
customMetrics:
  - id: business.order.retry.count
    type: counter
    unit: "{retry}"
    description: Order retry attempts.
    governanceStatus: proposed
```

But custom metrics should not be treated as standard preset-ready metrics until governance approval completes.

## Approval Workflow

Metric approval should not be a blocker for normal delivery. The workflow should be lightweight and mostly asynchronous.

### Fast Path

If a client needs observability now:

- they consume approved metrics from the catalog
- they select a supported dashboard preset
- the platform path proceeds immediately with no approval dependency

This should be the default and should cover most workloads.

### Proposal Path

If a client needs a new metric that does not exist:

1. the client submits a metric proposal
2. the proposal includes semantic purpose, type, unit, candidate dimensions, expected usage, and expected dashboard or alert use
3. the platform team reviews naming, overlap, cardinality, and cross-workload reuse potential
4. the proposal is either:
   - approved into the governed catalog
   - approved with changes
   - rejected in favor of an existing metric
   - accepted as workload-local experimental metric

### Experimental Path

To avoid blocking delivery, the platform can allow temporary experimental metrics under clear limits:

- marked as `experimental`
- not eligible for standard dashboard presets by default
- not guaranteed across adapters
- dimension policy enforced
- time-boxed for later catalog review

This prevents governance from becoming a delivery stopper while still protecting the platform surface.

## Approval Criteria

The review should answer a small set of concrete questions:

- Is the metric semantically distinct from an existing approved metric?
- Is the name stable and backend-neutral?
- Is the unit correct and unambiguous?
- Are the dimensions operationally useful and cardinality-safe?
- Is this metric likely reusable by more than one workload or class of workload?
- Does at least one current or planned adapter have a credible materialization path?
- Does the metric belong in a standard dashboard preset, a workload-specific extension, or both?

## Proposed Lifecycle States

The metric catalog should support simple lifecycle states:

- `proposed`
- `experimental`
- `approved`
- `deprecated`

Clients should freely consume `approved` metrics.

`experimental` metrics may be used with explicit caveats and reduced platform guarantees.

`deprecated` metrics should remain readable for transition periods but should no longer be recommended in new contracts.

## Handoff Expectations

The client handoff artifact should eventually report:

- requested governed metrics
- approved metrics resolved for the workload
- dashboard presets requested
- dashboard presets materialized
- metrics supported by the selected adapter
- metrics partially supported or unsupported
- any experimental metrics accepted with caveats

This gives the deployment owner an exact record of what observability coverage was actually delivered.

## Implementation Roadmap

Suggested sequence:

1. define a platform-owned governed metric catalog artifact in the repository
2. add adapter coverage metadata for the first supported target
3. introduce `catalogRefs` as a preferred contract path
4. keep free-form `metrics.catalog` only as a transitional compatibility field
5. add validation warnings or failures for ungoverned production metric definitions
6. connect dashboard presets to explicit metric coverage
7. expose coverage in plan outputs and client handoff artifacts

## Current Repository Status

Today, the repository is only partially at this destination:

- the contract already treats the metrics catalog as central intent
- dashboard presets already exist
- adapter dashboard materialization is still partial
- governed metric references are not yet enforced
- approval workflow is not yet implemented in product interfaces

This document describes the intended hardening direction.
