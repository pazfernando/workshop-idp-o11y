package validation

import (
	"strings"
	"testing"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
)

func TestValidateContractRejectsMissingEnabledSignalIngestion(t *testing.T) {
	t.Parallel()

	contract := validContract()
	contract.Spec.Telemetry.Signals.Traces.Ingestion = ""

	err := ValidateContract(contract)
	if err == nil || !strings.Contains(err.Error(), "missing properties: 'ingestion'") {
		t.Fatalf("expected missing ingestion error, got %v", err)
	}
}

func TestValidateContractAllowsDisabledSignalWithoutBackendClass(t *testing.T) {
	t.Parallel()

	contract := validContract()
	contract.Spec.Telemetry.Signals.Logs.Enabled = false
	contract.Spec.Telemetry.Signals.Logs.BackendClass = ""
	contract.Spec.Telemetry.Signals.Logs.Ingestion = ""

	if err := ValidateContract(contract); err != nil {
		t.Fatalf("expected disabled signal without backendClass to be valid, got %v", err)
	}
}

func TestValidateContractRejectsMixedIngestionForAWSLambda(t *testing.T) {
	t.Parallel()

	contract := validContract()
	contract.Spec.Telemetry.Signals.Logs.Ingestion = "direct"

	err := ValidateContract(contract)
	if err == nil || !strings.Contains(err.Error(), "mixed direct and collector signal ingestion is not supported") {
		t.Fatalf("expected mixed ingestion error, got %v", err)
	}
}

func TestValidateContractRejectsMetricsOutsidePresetSupport(t *testing.T) {
	t.Parallel()

	contract := validContract()
	contract.Spec.Telemetry.Signals.Metrics.Catalog = append(contract.Spec.Telemetry.Signals.Metrics.Catalog, v1alpha1.MetricSpec{
		Name:        "UnexpectedMetric",
		Type:        "counter",
		Unit:        "1",
		Description: "Unsupported metric",
	})

	err := ValidateContract(contract)
	if err == nil || !strings.Contains(err.Error(), "outside preset") {
		t.Fatalf("expected preset metric support error, got %v", err)
	}
}

func TestValidateContractRejectsMetricsWithoutDashboardPreset(t *testing.T) {
	t.Parallel()

	contract := validContract()
	contract.Spec.Capabilities.Dashboards = nil

	err := ValidateContract(contract)
	if err == nil || !strings.Contains(err.Error(), "requires capabilities.dashboards.enabled=true with a supported preset") {
		t.Fatalf("expected dashboard preset requirement error, got %v", err)
	}
}

func TestValidateContractRejectsUnknownDashboardPreset(t *testing.T) {
	t.Parallel()

	contract := validContract()
	contract.Spec.Capabilities.Dashboards.Preset = "unknown-preset"

	err := ValidateContract(contract)
	if err == nil || !strings.Contains(err.Error(), "dashboard preset") {
		t.Fatalf("expected dashboard preset error, got %v", err)
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
							{
								Name:        "HttpServerRequestCount",
								Type:        "counter",
								Unit:        "{request}",
								Description: "Total HTTP requests",
							},
							{
								Name:        "HttpServerRequestDuration",
								Type:        "histogram",
								Unit:        "ms",
								Description: "HTTP request latency",
							},
							{
								Name:        "HttpServerRequestErrors",
								Type:        "counter",
								Unit:        "{error}",
								Description: "HTTP request errors",
							},
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
