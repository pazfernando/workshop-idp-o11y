package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
	"github.com/example/workshop-iidp-o11y/internal/metricspreset"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed observability-contract-v1alpha1.schema.json
var schemaFS embed.FS

func ValidateContract(contract *v1alpha1.ObservabilityContract) error {
	if contract == nil {
		return fmt.Errorf("contract is nil")
	}

	schema, err := compileSchema()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	var candidate any
	if err := json.Unmarshal(payload, &candidate); err != nil {
		return err
	}

	if err := schema.Validate(candidate); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return validateSemantics(contract)
}

func compileSchema() (*jsonschema.Schema, error) {
	raw, err := schemaFS.ReadFile("observability-contract-v1alpha1.schema.json")
	if err != nil {
		return nil, err
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("observability-contract-v1alpha1.schema.json", strings.NewReader(string(raw))); err != nil {
		return nil, err
	}

	return compiler.Compile("observability-contract-v1alpha1.schema.json")
}

func validateSemantics(contract *v1alpha1.ObservabilityContract) error {
	if contract.Kind != "ObservabilityContract" {
		return fmt.Errorf("kind must be ObservabilityContract")
	}

	target := contract.Spec.Delivery.Target
	switch target {
	case "aws":
		if contract.Spec.Delivery.Region == "" {
			return fmt.Errorf("delivery.region is required when target is aws")
		}
	case "kubernetes":
		if contract.Spec.Delivery.Cluster == "" || contract.Spec.Delivery.Namespace == "" {
			return fmt.Errorf("delivery.cluster and delivery.namespace are required when target is kubernetes")
		}
	}

	if contract.Spec.Telemetry.Signals.Metrics.Enabled && len(contract.Spec.Telemetry.Signals.Metrics.Catalog) == 0 {
		return fmt.Errorf("metrics catalog must not be empty when metrics are enabled")
	}

	if err := validateDashboardPreset(contract); err != nil {
		return err
	}

	if err := validateMetricsAgainstPreset(contract); err != nil {
		return err
	}

	if err := validateSignalSpec("traces", contract.Spec.Telemetry.Signals.Traces.Enabled, contract.Spec.Telemetry.Signals.Traces.BackendClass, contract.Spec.Telemetry.Signals.Traces.Ingestion); err != nil {
		return err
	}
	if err := validateSignalSpec("metrics", contract.Spec.Telemetry.Signals.Metrics.Enabled, contract.Spec.Telemetry.Signals.Metrics.BackendClass, contract.Spec.Telemetry.Signals.Metrics.Ingestion); err != nil {
		return err
	}
	if err := validateSignalSpec("logs", contract.Spec.Telemetry.Signals.Logs.Enabled, contract.Spec.Telemetry.Signals.Logs.BackendClass, contract.Spec.Telemetry.Signals.Logs.Ingestion); err != nil {
		return err
	}

	if contract.Spec.Delivery.Target == "aws" && contract.Spec.Service.Runtime == "aws-lambda" {
		if err := validateUniformIngestionForAWSLambda(contract); err != nil {
			return err
		}
	}

	return nil
}

func validateSignalSpec(name string, enabled bool, backendClass, ingestion string) error {
	backendClass = strings.TrimSpace(backendClass)
	ingestion = strings.TrimSpace(ingestion)

	if !enabled {
		return nil
	}
	if backendClass == "" {
		return fmt.Errorf("%s.backendClass is required when %s.enabled is true", name, name)
	}
	if ingestion == "" {
		return fmt.Errorf("%s.ingestion is required when %s.enabled is true", name, name)
	}

	return nil
}

func validateUniformIngestionForAWSLambda(contract *v1alpha1.ObservabilityContract) error {
	ingestions := []string{}
	signals := []struct {
		name      string
		enabled   bool
		ingestion string
	}{
		{name: "traces", enabled: contract.Spec.Telemetry.Signals.Traces.Enabled, ingestion: strings.TrimSpace(contract.Spec.Telemetry.Signals.Traces.Ingestion)},
		{name: "metrics", enabled: contract.Spec.Telemetry.Signals.Metrics.Enabled, ingestion: strings.TrimSpace(contract.Spec.Telemetry.Signals.Metrics.Ingestion)},
		{name: "logs", enabled: contract.Spec.Telemetry.Signals.Logs.Enabled, ingestion: strings.TrimSpace(contract.Spec.Telemetry.Signals.Logs.Ingestion)},
	}

	for _, signal := range signals {
		if signal.enabled {
			ingestions = append(ingestions, signal.ingestion)
		}
	}

	if len(ingestions) <= 1 {
		return nil
	}

	expected := ingestions[0]
	for _, ingestion := range ingestions[1:] {
		if ingestion != expected {
			return fmt.Errorf("aws-lambda bindings require one ingestion mode across all enabled signals; mixed direct and collector signal ingestion is not supported")
		}
	}

	return nil
}

func validateMetricsAgainstPreset(contract *v1alpha1.ObservabilityContract) error {
	if !contract.Spec.Telemetry.Signals.Metrics.Enabled {
		return nil
	}

	if contract.Spec.Capabilities.Dashboards == nil || !contract.Spec.Capabilities.Dashboards.Enabled || strings.TrimSpace(contract.Spec.Capabilities.Dashboards.Preset) == "" {
		return fmt.Errorf("metrics support currently requires capabilities.dashboards.enabled=true with a supported preset")
	}

	preset := strings.TrimSpace(contract.Spec.Capabilities.Dashboards.Preset)
	supportedMetrics, _ := metricspreset.SupportedMetrics(preset)

	supported := map[string]struct{}{}
	for _, metric := range supportedMetrics {
		supported[metric] = struct{}{}
	}

	unsupported := []string{}
	for _, metric := range contract.Spec.Telemetry.Signals.Metrics.Catalog {
		if _, ok := supported[metric.Name]; !ok {
			unsupported = append(unsupported, metric.Name)
		}
	}

	if len(unsupported) > 0 {
		return fmt.Errorf("metrics catalog contains metrics outside preset %q support: %s", preset, strings.Join(unsupported, ", "))
	}

	return nil
}

func validateDashboardPreset(contract *v1alpha1.ObservabilityContract) error {
	if contract.Spec.Capabilities.Dashboards == nil || !contract.Spec.Capabilities.Dashboards.Enabled {
		return nil
	}

	preset := strings.TrimSpace(contract.Spec.Capabilities.Dashboards.Preset)
	if preset == "" {
		return fmt.Errorf("capabilities.dashboards.preset is required when dashboards are enabled")
	}
	if _, ok := metricspreset.Lookup(preset); !ok {
		return fmt.Errorf("dashboard preset %q is not supported; supported presets: %s", preset, strings.Join(metricspreset.KnownPresets(), ", "))
	}

	return nil
}
