# Metrics Governance

This document defines the current supported model for workload metrics in the observability platform.

## Current Boundary

The current product supports only metrics that belong to the repository-visible preset catalog under `catalog/metrics/presets/`.

Supported today:

- clients select metrics indirectly by choosing a supported dashboard preset
- clients may include only metrics that belong to that preset in `spec.telemetry.signals.metrics.catalog`
- validation rejects metrics outside the selected preset metric set
- plan outputs, bindings notes, and handoff artifacts report this as a `preset-only` metric support policy

Not supported today:

- client-defined production metrics outside the preset catalog
- productized approval or proposal workflow for new metrics
- experimental metric onboarding path
- `catalogRefs` or other governed metric reference fields in the contract
- rich per-metric adapter coverage metadata in the catalog

## Why The Boundary Exists

If every client defines arbitrary metrics with arbitrary dimensions, the platform becomes hard to operate:

- metric names drift across teams
- semantic meaning becomes inconsistent
- dashboards cannot be reused safely
- alert and SLO semantics become non-comparable
- dimensions can explode cardinality and cost
- target adapters cannot provide predictable coverage

For that reason, the current repository intentionally limits standard product usage to a governed preset metric set.

## Current Catalog Location

The current client-visible source of truth is:

- `catalog/metrics/presets/serverless-api.yaml`
- `catalog/metrics/presets/distributed-service.yaml`
- `catalog/metrics/presets/kubernetes-http-service.yaml`
- `catalog/metrics/presets/monolith-business-app.yaml`

These files define:

- `contractMetrics`: the standard workload metrics accepted in the contract for that preset
- `recommendedRuntimeMetrics`: the target or platform metrics that should also be observed for that workload shape

## Why Runtime Metrics Vary By Preset

The catalog intentionally separates workload metrics from runtime metrics.

- `serverless-api` focuses on request throughput, latency, and errors in the contract, while recommending AWS Lambda runtime metrics such as invocations, duration, throttles, and concurrency
- `kubernetes-http-service` keeps contract metrics centered on request throughput, latency, and errors, while recommending CPU, memory, and pressure metrics at pod or container level
- `distributed-service` extends the contract metric set to include outbound dependency call metrics because dependency behavior is usually part of the service reliability surface
- `monolith-business-app` keeps contract metrics centered on application operations while recommending process-level CPU, memory, and thread metrics for runtime health

This is deliberate:

- ephemeral serverless runtimes are usually better understood through latency, errors, concurrency, and throttling than through baseline CPU charts in the product contract
- long-running workloads such as Kubernetes services, distributed services, and monoliths benefit more directly from CPU, memory, and pressure visibility
- platform or target-native metrics remain important, but they are not all promoted to contract-authored workload metrics

## Client Responsibilities

The client team should:

- choose a supported dashboard preset
- instrument the workload to emit the metrics required by that preset
- declare only preset-supported metrics in the contract

The client team should not:

- define arbitrary new production metrics in the standard platform flow
- expect a proposal, approval, or exception workflow to exist in the current product path
- treat dashboard definitions as backend-specific handwritten resources in the standard path

## Platform Responsibilities

The platform currently owns:

- the preset metric catalog
- validation of metric-to-preset alignment
- plan metadata describing preset-only support
- target-specific notes and handoff reporting about supported metrics

## Contract Semantics Today

The contract still uses `spec.telemetry.signals.metrics.catalog`, but that field is not a free-form extension surface.

Today it means:

- the client declares the subset of preset metrics used by the workload
- the selected dashboard preset determines the allowed metric set
- any metric outside that preset is rejected
- runtime or platform-native metrics such as Lambda concurrency or Kubernetes CPU and memory remain adapter concerns, not contract metrics

## Plan, Bindings, And Handoff

The current implementation reflects the same policy across layers:

- validation enforces preset-only metric support
- plan outputs include the selected preset, requested metric names, supported metric names, recommended runtime metrics, and `metricSupportPolicy: preset-only`
- bindings emit notes that metric support is limited to the selected preset
- the client handoff artifact records requested and supported preset metric names together with recommended runtime metrics

## Future Work

Possible future enhancements, not supported today:

- richer metadata in `catalog/metrics/presets/`
- stable governed metric IDs independent of display names
- explicit adapter coverage metadata per metric
- governed metric reference fields such as `catalogRefs`
- a productized workflow for proposing or approving new metrics

Until those exist, the supported answer is simple: only catalog-backed preset metrics are allowed.
