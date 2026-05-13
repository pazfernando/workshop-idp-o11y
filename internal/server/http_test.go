package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
)

func TestValidateEndpoint(t *testing.T) {
	t.Parallel()

	payload := ValidateRequest{Contract: validContract()}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	NewHTTPHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestPlanEndpointRejectsMissingContract(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/plan", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	NewHTTPHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
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
					SemanticConventions: "1.26.0",
					Propagators:         []string{"tracecontext", "baggage", "xray"},
				},
				ResourceAttributes: map[string]string{
					"service.name":           "order-processing",
					"service.namespace":      "commerce",
					"service.version":        "1.0.0",
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
								Name:        "orders_created_total",
								Type:        "counter",
								Unit:        "1",
								Description: "Number of created orders",
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
					Preset:  "api-workload",
				},
			},
		},
	}
}
