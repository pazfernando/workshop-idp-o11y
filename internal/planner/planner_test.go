package planner

import (
	"testing"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
)

func TestBuildIncludesPresetMetricSupportMetadata(t *testing.T) {
	t.Parallel()

	plan, err := Build(validContract())
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	metricsCapability := capabilityByName(plan, "metrics-pipeline")
	if metricsCapability == nil {
		t.Fatalf("expected metrics-pipeline capability")
	}

	if got := metricsCapability.Properties["metricSupportPolicy"]; got != "preset-only" {
		t.Fatalf("expected preset-only support policy, got %v", got)
	}
	if got := metricsCapability.Properties["dashboardPreset"]; got != "serverless-api" {
		t.Fatalf("expected serverless-api preset, got %v", got)
	}

	supportedMetrics, ok := metricsCapability.Properties["supportedMetricNames"].([]string)
	if !ok {
		t.Fatalf("expected supportedMetricNames to be []string, got %T", metricsCapability.Properties["supportedMetricNames"])
	}
	if len(supportedMetrics) != 2 {
		t.Fatalf("expected two supported metrics, got %v", supportedMetrics)
	}
}

func capabilityByName(plan *ProvisioningPlan, name string) *CapabilityPlan {
	for i := range plan.Capabilities {
		if plan.Capabilities[i].Name == name {
			return &plan.Capabilities[i]
		}
	}
	return nil
}

func validContract() *v1alpha1.ObservabilityContract {
	return &v1alpha1.ObservabilityContract{
		APIVersion: "observability.platform/v1alpha1",
		Kind:       "ObservabilityContract",
		Metadata: v1alpha1.Metadata{
			Name:        "order-processing",
			Owner:       "payments-platform",
			System:      "commerce",
			Environment: "dev",
		},
		Spec: v1alpha1.Spec{
			Service: v1alpha1.Service{
				Name:      "order-processing",
				Runtime:   "aws-lambda",
				Language:  "nodejs",
				Framework: "lambda",
			},
			Delivery: v1alpha1.Delivery{
				Mode:   "provision",
				Target: "aws",
				Region: "us-east-1",
			},
			Telemetry: v1alpha1.Telemetry{
				OpenTelemetry: v1alpha1.OpenTelemetryConfig{
					Enabled:             true,
					SemanticConventions: "opentelemetry",
					Propagators:         []string{"tracecontext", "baggage", "xray"},
				},
				ResourceAttributes: map[string]string{
					"service.name":           "order-processing",
					"service.namespace":      "commerce",
					"deployment.environment": "dev",
				},
				Signals: v1alpha1.Signals{
					Traces: v1alpha1.SignalSpec{
						Enabled:      true,
						BackendClass: "aws-xray",
						Ingestion:    "collector",
					},
					Metrics: v1alpha1.MetricsSignalSpec{
						SignalSpec: v1alpha1.SignalSpec{
							Enabled:      true,
							BackendClass: "aws-cloudwatch-metrics",
							Ingestion:    "collector",
						},
						Catalog: []v1alpha1.MetricSpec{
							{Name: "OrdersCreated", Type: "counter", Unit: "{order}", Description: "Orders created"},
							{Name: "CreateOrderLatencyMs", Type: "histogram", Unit: "ms", Description: "Order latency"},
						},
					},
					Logs: v1alpha1.LogsSignalSpec{
						SignalSpec: v1alpha1.SignalSpec{
							Enabled:      true,
							BackendClass: "aws-cloudwatch-logs",
							Ingestion:    "collector",
						},
						Format: "json",
					},
				},
			},
			Capabilities: v1alpha1.Capabilities{
				Dashboards: &v1alpha1.DashboardCapability{
					Enabled: true,
					Preset:  "serverless-api",
				},
			},
		},
	}
}
