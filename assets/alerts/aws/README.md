# AWS Alert Templates

The alert rules declared by the application team live in the contract.

This directory is reserved for AWS adapter templates and mappings that translate those rules into:

- CloudWatch alarms
- Grafana alerting
- Prometheus-compatible rule groups

The planner already separates the `alerts` capability. This directory is the adapter-facing location for rendering those rules into backend-specific alert resources.
