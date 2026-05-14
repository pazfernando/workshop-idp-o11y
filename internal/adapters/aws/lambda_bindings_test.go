package aws

import (
	"strings"
	"testing"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
)

func TestBuildLambdaBindingsIncludesPresetMetricSupportNote(t *testing.T) {
	t.Parallel()

	emf := true
	bindings, err := BuildLambdaBindings(validContract(), LambdaBindingOptions{
		InstrumentationMode:        "code",
		CollectorEndpoint:          "http://collector.internal:4318",
		EnableEMFCompatibilityMode: &emf,
	})
	if err != nil {
		t.Fatalf("build bindings: %v", err)
	}

	found := false
	for _, note := range bindings.Notes {
		if strings.Contains(note, "Current metric support is limited to dashboard preset") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected preset metric support note, got %v", bindings.Notes)
	}
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
				Framework: "aws-serverless",
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
					"service.namespace":      "workshop",
					"deployment.environment": "dev",
				},
				Signals: v1alpha1.Signals{
					Traces: v1alpha1.SignalSpec{
						Enabled:      true,
						BackendClass: "traces-standard",
						Ingestion:    "collector",
					},
					Metrics: v1alpha1.MetricsSignalSpec{
						SignalSpec: v1alpha1.SignalSpec{
							Enabled:      true,
							BackendClass: "metrics-standard",
							Ingestion:    "collector",
						},
						Catalog: []v1alpha1.MetricSpec{
							{Name: "OrdersCreated", Type: "counter", Unit: "{order}", Description: "Orders"},
							{Name: "CreateOrderLatencyMs", Type: "histogram", Unit: "ms", Description: "Latency"},
						},
					},
					Logs: v1alpha1.LogsSignalSpec{
						SignalSpec: v1alpha1.SignalSpec{
							Enabled:      true,
							BackendClass: "logs-standard",
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
