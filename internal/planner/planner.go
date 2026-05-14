package planner

import (
	"fmt"
	"sort"

	metriccatalog "github.com/example/workshop-iidp-o11y/catalog/metrics"
	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
)

type ProvisioningPlan struct {
	ContractRef  ContractRef      `json:"contractRef"`
	Platform     PlatformContext  `json:"platform"`
	Capabilities []CapabilityPlan `json:"capabilities"`
	Classes      []ClassPlan      `json:"classes,omitempty"`
	Artifacts    []ArtifactPlan   `json:"artifacts,omitempty"`
	Bindings     RuntimeBindings  `json:"bindings"`
}

type ContractRef struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	System      string `json:"system"`
	Environment string `json:"environment"`
}

type PlatformContext struct {
	DeliveryMode string `json:"deliveryMode"`
	Target       string `json:"target"`
	Region       string `json:"region,omitempty"`
	Cluster      string `json:"cluster,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
}

type CapabilityPlan struct {
	Name         string         `json:"name"`
	Kind         string         `json:"kind"`
	BackendClass string         `json:"backendClass,omitempty"`
	Intent       string         `json:"intent"`
	Properties   map[string]any `json:"properties,omitempty"`
	Dependencies []string       `json:"dependencies,omitempty"`
}

type RuntimeBindings struct {
	ResourceAttributes map[string]string `json:"resourceAttributes"`
	Propagators        []string          `json:"propagators"`
	IngestionMode      string            `json:"ingestionMode"`
}

type ClassPlan struct {
	Name    string `json:"name"`
	Layer   string `json:"layer"`
	Path    string `json:"path"`
	Purpose string `json:"purpose"`
}

type ArtifactPlan struct {
	Name    string `json:"name"`
	Layer   string `json:"layer"`
	Path    string `json:"path"`
	Purpose string `json:"purpose"`
}

func Build(contract *v1alpha1.ObservabilityContract) (*ProvisioningPlan, error) {
	if contract == nil {
		return nil, fmt.Errorf("contract is nil")
	}

	plan := &ProvisioningPlan{
		ContractRef: ContractRef{
			Name:        contract.Metadata.Name,
			Owner:       contract.Metadata.Owner,
			System:      contract.Metadata.System,
			Environment: contract.Metadata.Environment,
		},
		Platform: PlatformContext{
			DeliveryMode: contract.Spec.Delivery.Mode,
			Target:       contract.Spec.Delivery.Target,
			Region:       contract.Spec.Delivery.Region,
			Cluster:      contract.Spec.Delivery.Cluster,
			Namespace:    contract.Spec.Delivery.Namespace,
		},
		Bindings: RuntimeBindings{
			ResourceAttributes: contract.Spec.Telemetry.ResourceAttributes,
			Propagators:        contract.Spec.Telemetry.OpenTelemetry.Propagators,
			IngestionMode:      resolveIngestion(contract),
		},
	}

	plan.Capabilities = append(plan.Capabilities, signalCapabilities(contract)...)
	plan.Capabilities = append(plan.Capabilities, dashboardCapabilities(contract)...)
	plan.Capabilities = append(plan.Capabilities, alertCapabilities(contract)...)
	plan.Capabilities = append(plan.Capabilities, sloCapabilities(contract)...)
	plan.Capabilities = append(plan.Capabilities, retentionCapabilities(contract)...)
	plan.Classes = append(plan.Classes, classPlans(contract)...)
	plan.Artifacts = append(plan.Artifacts, artifactPlans(contract)...)

	sort.Slice(plan.Capabilities, func(i, j int) bool {
		return plan.Capabilities[i].Name < plan.Capabilities[j].Name
	})
	sort.Slice(plan.Classes, func(i, j int) bool {
		return plan.Classes[i].Name < plan.Classes[j].Name
	})
	sort.Slice(plan.Artifacts, func(i, j int) bool {
		return plan.Artifacts[i].Name < plan.Artifacts[j].Name
	})

	return plan, nil
}

func signalCapabilities(contract *v1alpha1.ObservabilityContract) []CapabilityPlan {
	var capabilities []CapabilityPlan

	if contract.Spec.Telemetry.Signals.Traces.Enabled {
		capabilities = append(capabilities, CapabilityPlan{
			Name:         "traces-pipeline",
			Kind:         "signal",
			BackendClass: contract.Spec.Telemetry.Signals.Traces.BackendClass,
			Intent:       describeSignalIntent(contract.Spec.Delivery.Target, "traces"),
			Properties: map[string]any{
				"signal":    "traces",
				"ingestion": contract.Spec.Telemetry.Signals.Traces.Ingestion,
			},
		})
	}

	if contract.Spec.Telemetry.Signals.Metrics.Enabled {
		dashboardPreset := ""
		supportedMetrics := []string{}
		recommendedRuntimeMetrics := []string{}
		if contract.Spec.Capabilities.Dashboards != nil {
			dashboardPreset = contract.Spec.Capabilities.Dashboards.Preset
			if metrics, ok := metriccatalog.SupportedMetrics(dashboardPreset); ok {
				supportedMetrics = metrics
			}
			if metrics, ok := metriccatalog.RecommendedRuntimeMetrics(dashboardPreset); ok {
				recommendedRuntimeMetrics = metrics
			}
		}
		capabilities = append(capabilities, CapabilityPlan{
			Name:         "metrics-pipeline",
			Kind:         "signal",
			BackendClass: contract.Spec.Telemetry.Signals.Metrics.BackendClass,
			Intent:       describeSignalIntent(contract.Spec.Delivery.Target, "metrics"),
			Properties: map[string]any{
				"signal":                "metrics",
				"ingestion":             contract.Spec.Telemetry.Signals.Metrics.Ingestion,
				"metricCount":           len(contract.Spec.Telemetry.Signals.Metrics.Catalog),
				"metricNames":           metricNames(contract.Spec.Telemetry.Signals.Metrics.Catalog),
				"dashboardPreset":       dashboardPreset,
				"supportedMetricNames":  supportedMetrics,
				"recommendedRuntimeMetrics": recommendedRuntimeMetrics,
				"metricSupportPolicy":   "preset-only",
				"monitoredSLO":          contract.Spec.Capabilities.SLOs != nil && contract.Spec.Capabilities.SLOs.Enabled,
			},
		})
	}

	if contract.Spec.Telemetry.Signals.Logs.Enabled {
		capabilities = append(capabilities, CapabilityPlan{
			Name:         "logs-pipeline",
			Kind:         "signal",
			BackendClass: contract.Spec.Telemetry.Signals.Logs.BackendClass,
			Intent:       describeSignalIntent(contract.Spec.Delivery.Target, "logs"),
			Properties: map[string]any{
				"signal":    "logs",
				"ingestion": contract.Spec.Telemetry.Signals.Logs.Ingestion,
				"format":    contract.Spec.Telemetry.Signals.Logs.Format,
			},
		})
	}

	return capabilities
}

func dashboardCapabilities(contract *v1alpha1.ObservabilityContract) []CapabilityPlan {
	if contract.Spec.Capabilities.Dashboards == nil || !contract.Spec.Capabilities.Dashboards.Enabled {
		return nil
	}

	return []CapabilityPlan{
		{
			Name:   "dashboards",
			Kind:   "capability",
			Intent: "Provision or bind a standard observability dashboard set for the workload.",
			Dependencies: []string{
				"metrics-pipeline",
				"traces-pipeline",
			},
			Properties: map[string]any{
				"preset":                    contract.Spec.Capabilities.Dashboards.Preset,
				"metricSupportPolicy":       "preset-only",
				"supportedMetrics":          supportedMetricsForPreset(contract.Spec.Capabilities.Dashboards.Preset),
				"recommendedRuntimeMetrics": recommendedRuntimeMetricsForPreset(contract.Spec.Capabilities.Dashboards.Preset),
			},
		},
	}
}

func alertCapabilities(contract *v1alpha1.ObservabilityContract) []CapabilityPlan {
	if contract.Spec.Capabilities.Alerts == nil || !contract.Spec.Capabilities.Alerts.Enabled {
		return nil
	}

	return []CapabilityPlan{
		{
			Name:   "alerts",
			Kind:   "capability",
			Intent: "Provision alert rules and route them according to platform policy.",
			Dependencies: []string{
				"metrics-pipeline",
			},
			Properties: map[string]any{
				"rules": contract.Spec.Capabilities.Alerts.Rules,
			},
		},
	}
}

func sloCapabilities(contract *v1alpha1.ObservabilityContract) []CapabilityPlan {
	if contract.Spec.Capabilities.SLOs == nil || !contract.Spec.Capabilities.SLOs.Enabled {
		return nil
	}

	return []CapabilityPlan{
		{
			Name:   "slos",
			Kind:   "capability",
			Intent: "Provision SLO definitions and wire them to the appropriate signal sources.",
			Dependencies: []string{
				"metrics-pipeline",
				"traces-pipeline",
			},
			Properties: map[string]any{
				"objectives": contract.Spec.Capabilities.SLOs.Objectives,
			},
		},
	}
}

func retentionCapabilities(contract *v1alpha1.ObservabilityContract) []CapabilityPlan {
	if contract.Spec.Capabilities.DataManagement == nil || contract.Spec.Capabilities.DataManagement.Retention == nil {
		return nil
	}

	retention := contract.Spec.Capabilities.DataManagement.Retention

	return []CapabilityPlan{
		{
			Name:   "data-retention",
			Kind:   "governance",
			Intent: "Apply retention and data handling policy to provisioned observability storage.",
			Properties: map[string]any{
				"tracesDays":         retention.TracesDays,
				"metricsDays":        retention.MetricsDays,
				"logsDays":           retention.LogsDays,
				"dataClassification": contract.Spec.Capabilities.DataManagement.DataClassification,
			},
		},
	}
}

func resolveIngestion(contract *v1alpha1.ObservabilityContract) string {
	signals := []string{
		contract.Spec.Telemetry.Signals.Traces.Ingestion,
		contract.Spec.Telemetry.Signals.Metrics.Ingestion,
		contract.Spec.Telemetry.Signals.Logs.Ingestion,
	}

	for _, signal := range signals {
		if signal == "collector" {
			return "collector"
		}
	}

	return "direct"
}

func describeSignalIntent(target, signal string) string {
	switch target {
	case "kubernetes":
		return fmt.Sprintf("Bind %s emission to a cluster-level observability data plane and policy-managed backend.", signal)
	case "aws":
		return fmt.Sprintf("Bind %s emission to an AWS-aligned observability pipeline through platform-managed backends.", signal)
	default:
		return fmt.Sprintf("Bind %s emission to a platform-managed observability pipeline.", signal)
	}
}

func metricNames(metrics []v1alpha1.MetricSpec) []string {
	names := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		names = append(names, metric.Name)
	}
	sort.Strings(names)
	return names
}

func supportedMetricsForPreset(preset string) []string {
	metrics, ok := metriccatalog.SupportedMetrics(preset)
	if !ok {
		return nil
	}
	return metrics
}

func recommendedRuntimeMetricsForPreset(preset string) []string {
	metrics, ok := metriccatalog.RecommendedRuntimeMetrics(preset)
	if !ok {
		return nil
	}
	return metrics
}

func classPlans(contract *v1alpha1.ObservabilityContract) []ClassPlan {
	var classes []ClassPlan

	if contract.Spec.Delivery.Target == "aws" && contract.Spec.Service.Runtime == "aws-lambda" {
		classes = append(classes, ClassPlan{
			Name:    "aws-lambda-otel-bindings",
			Layer:   "adapter",
			Path:    "catalog/classes/aws/lambda-otel-bindings.yaml",
			Purpose: "Reusable class for materializing OpenTelemetry runtime bindings for AWS Lambda workloads.",
		})
	}

	if resolveIngestion(contract) == "collector" && contract.Spec.Delivery.Target == "aws" {
		classes = append(classes, ClassPlan{
			Name:    "aws-collector-gateway-cloudwatch",
			Layer:   "catalog",
			Path:    "catalog/classes/aws/collector-gateway-cloudwatch.yaml",
			Purpose: "AWS collector class that fulfills neutral traces, metrics, and logs classes through platform-managed pipelines.",
		})
	}

	if contract.Spec.Capabilities.Dashboards != nil && contract.Spec.Capabilities.Dashboards.Enabled && contract.Spec.Delivery.Target == "aws" {
		classes = append(classes, ClassPlan{
			Name:    "aws-serverless-api-observability",
			Layer:   "catalog",
			Path:    "catalog/classes/aws/serverless-api-observability.yaml",
			Purpose: "AWS class that fulfills the neutral serverless-api dashboard preset for Lambda and API Gateway workloads.",
		})
	}

	return classes
}

func artifactPlans(contract *v1alpha1.ObservabilityContract) []ArtifactPlan {
	var artifacts []ArtifactPlan

	if resolveIngestion(contract) == "collector" && contract.Spec.Delivery.Target == "aws" {
		artifacts = append(artifacts, ArtifactPlan{
			Name:    "collector-cloudwatch-config",
			Layer:   "assets",
			Path:    "assets/collector/aws/collector-cloudwatch.yaml",
			Purpose: "Collector pipeline for AWS-native metrics and traces export.",
		})
		artifacts = append(artifacts, ArtifactPlan{
			Name:    "collector-cloudwatch-third-party-config",
			Layer:   "assets",
			Path:    "assets/collector/aws/collector-cloudwatch-third-party.yaml",
			Purpose: "Collector pipeline for AWS-native export plus third-party OTLP fan-out.",
		})
	}

	if contract.Spec.Capabilities.Dashboards != nil && contract.Spec.Capabilities.Dashboards.Enabled && contract.Spec.Delivery.Target == "aws" {
		artifacts = append(artifacts, ArtifactPlan{
			Name:    "cloudwatch-dashboard-template",
			Layer:   "assets",
			Path:    "assets/dashboards/cloudwatch/aws-lambda-api-workload.json.tftpl",
			Purpose: "Reusable CloudWatch dashboard template for API Gateway and Lambda runtime metrics.",
		})
		artifacts = append(artifacts, ArtifactPlan{
			Name:    "managed-suite-assets",
			Layer:   "assets",
			Path:    "assets/collector/managed-suite",
			Purpose: "Platform-managed observability suite assets for collector-managed OTLP endpoints and Grafana.",
		})
	}

	if contract.Spec.Capabilities.Alerts != nil && contract.Spec.Capabilities.Alerts.Enabled && contract.Spec.Delivery.Target == "aws" {
		artifacts = append(artifacts, ArtifactPlan{
			Name:    "aws-alert-template-conventions",
			Layer:   "assets",
			Path:    "assets/alerts/aws",
			Purpose: "Reserved location for AWS alert templates that translate contract rules into backend-specific alarms.",
		})
	}

	return artifacts
}
