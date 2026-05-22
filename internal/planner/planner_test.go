package planner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	if len(supportedMetrics) != 3 {
		t.Fatalf("expected three supported metrics, got %v", supportedMetrics)
	}

	runtimeMetrics, ok := metricsCapability.Properties["recommendedRuntimeMetrics"].([]string)
	if !ok {
		t.Fatalf("expected recommendedRuntimeMetrics to be []string, got %T", metricsCapability.Properties["recommendedRuntimeMetrics"])
	}
	if len(runtimeMetrics) == 0 {
		t.Fatalf("expected recommended runtime metrics, got %v", runtimeMetrics)
	}
}

func TestBuildAssignsMonolithDashboardClass(t *testing.T) {
	t.Parallel()

	plan, err := Build(validMonolithContract())
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	if classByName(plan, "aws-monolith-business-app-observability") == nil {
		t.Fatalf("expected monolith dashboard class, got %v", plan.Classes)
	}
	if classByName(plan, "aws-serverless-api-observability") != nil {
		t.Fatalf("did not expect serverless dashboard class for monolith preset")
	}
}

func TestMonolithDashboardTemplateRendersServiceScopedQueries(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "infra", "observability-suite", "grafana-dashboard-monolith-business-app.json.tftpl")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read dashboard template: %v", err)
	}

	rendered := strings.NewReplacer(
		"${dashboard_uid}", "test-monolith",
		"${dashboard_title}", "Test Monolith",
		"${service_name}", "order-satellite-service",
		"${service_namespace}", "observability-demo",
		"${deployment_environment}", "dev",
	).Replace(string(raw))

	if strings.Contains(rendered, "aws_lambda_function_name") {
		t.Fatalf("monolith dashboard must not contain Lambda function filters")
	}

	var dashboard map[string]any
	if err := json.Unmarshal([]byte(rendered), &dashboard); err != nil {
		t.Fatalf("dashboard template did not render valid JSON: %v", err)
	}

	for _, query := range dashboardQueries(dashboard) {
		if strings.Contains(query, "{__name__=~") {
			for _, label := range []string{`service_name="order-satellite-service"`, `service_namespace="observability-demo"`, `deployment_environment="dev"`} {
				if !strings.Contains(query, label) {
					t.Fatalf("query missing label %s: %s", label, query)
				}
			}
		}
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

func classByName(plan *ProvisioningPlan, name string) *ClassPlan {
	for i := range plan.Classes {
		if plan.Classes[i].Name == name {
			return &plan.Classes[i]
		}
	}
	return nil
}

func dashboardQueries(value any) []string {
	var queries []string
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if (key == "expr" || key == "query") && nested != nil {
				if query, ok := nested.(string); ok {
					queries = append(queries, query)
					continue
				}
			}
			queries = append(queries, dashboardQueries(nested)...)
		}
	case []any:
		for _, nested := range typed {
			queries = append(queries, dashboardQueries(nested)...)
		}
	}
	return queries
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
							{Name: "HttpServerRequestCount", Type: "counter", Unit: "{request}", Description: "Request count"},
							{Name: "HttpServerRequestDuration", Type: "histogram", Unit: "ms", Description: "Request latency"},
							{Name: "HttpServerRequestErrors", Type: "counter", Unit: "{error}", Description: "Request errors"},
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

func validMonolithContract() *v1alpha1.ObservabilityContract {
	contract := validContract()
	contract.Metadata.Name = "order-satellite-demo"
	contract.Metadata.System = "order-processing-system"
	contract.Spec.Service.Name = "order-satellite-service"
	contract.Spec.Service.Runtime = "jvm"
	contract.Spec.Service.Language = "java"
	contract.Spec.Service.Framework = "micronaut"
	contract.Spec.Telemetry.ResourceAttributes = map[string]string{
		"service.name":           "order-satellite-service",
		"service.namespace":      "observability-demo",
		"deployment.environment": "dev",
	}
	contract.Spec.Telemetry.Signals.Metrics.Catalog = []v1alpha1.MetricSpec{
		{Name: "http.server.request.duration", Type: "histogram", Unit: "s", Description: "HTTP server request duration."},
		{Name: "http.client.request.duration", Type: "histogram", Unit: "s", Description: "HTTP client request duration."},
	}
	contract.Spec.Capabilities.Dashboards.Preset = "monolith-business-app"
	return contract
}
