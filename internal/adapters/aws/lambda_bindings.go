package aws

import (
	"fmt"
	"sort"
	"strings"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
	"github.com/example/workshop-iidp-o11y/internal/planner"
)

const (
	defaultMetricExportIntervalMs = 10000
	adotExecWrapper               = "/opt/otel-handler"
	adotManagedPolicy             = "arn:aws:iam::aws:policy/CloudWatchLambdaApplicationSignalsExecutionRolePolicy"
)

type LambdaBindingOptions struct {
	InstrumentationMode        string
	ADOTLambdaLayerARN         string
	CollectorEndpoint          string
	CollectorTracesEndpoint    string
	CollectorMetricsEndpoint   string
	DirectEndpoint             string
	DirectTracesEndpoint       string
	DirectMetricsEndpoint      string
	MetricExportIntervalMs     int
	EnableEMFCompatibilityMode bool
}

type LambdaBindings struct {
	ContractRef     planner.ContractRef     `json:"contractRef"`
	Platform        planner.PlatformContext `json:"platform"`
	Instrumentation InstrumentationSpec     `json:"instrumentation"`
	Environment     map[string]string       `json:"environment"`
	Layers          []string                `json:"layers,omitempty"`
	ManagedPolicies []string                `json:"managedPolicies,omitempty"`
	RequiredInputs  []RequiredInput         `json:"requiredInputs,omitempty"`
	Assets          AssetReferences         `json:"assets"`
	Outputs         BindingOutputs          `json:"outputs"`
	Notes           []string                `json:"notes,omitempty"`
}

type InstrumentationSpec struct {
	Mode                   string `json:"mode"`
	ExportStrategy         string `json:"exportStrategy"`
	OTLPAuthenticationMode string `json:"otlpAuthenticationMode"`
	MetricExportIntervalMs int    `json:"metricExportIntervalMs"`
	EMFCompatibilityMode   bool   `json:"emfCompatibilityMode"`
}

type RequiredInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssetReferences struct {
	TerraformModule           string `json:"terraformModule"`
	CollectorConfig           string `json:"collectorConfig,omitempty"`
	CollectorThirdPartyConfig string `json:"collectorThirdPartyConfig,omitempty"`
	ManagedSuitePath          string `json:"managedSuitePath,omitempty"`
	CloudWatchDashboard       string `json:"cloudWatchDashboard,omitempty"`
}

type BindingOutputs struct {
	OTLPBaseEndpoint    string `json:"otlpBaseEndpoint,omitempty"`
	OTLPTracesEndpoint  string `json:"otlpTracesEndpoint,omitempty"`
	OTLPMetricsEndpoint string `json:"otlpMetricsEndpoint,omitempty"`
	LambdaExecWrapper   string `json:"lambdaExecWrapper,omitempty"`
	ADOTLambdaLayerARN  string `json:"adotLambdaLayerArn,omitempty"`
}

func BuildLambdaBindings(contract *v1alpha1.ObservabilityContract, opts LambdaBindingOptions) (*LambdaBindings, error) {
	if contract == nil {
		return nil, fmt.Errorf("contract is nil")
	}
	if contract.Spec.Delivery.Target != "aws" {
		return nil, fmt.Errorf("aws lambda bindings require delivery.target=aws")
	}
	if contract.Spec.Service.Runtime != "aws-lambda" {
		return nil, fmt.Errorf("aws lambda bindings require service.runtime=aws-lambda")
	}

	mode := opts.InstrumentationMode
	if mode == "" {
		mode = "code"
	}
	if mode != "code" && mode != "adot_layer" {
		return nil, fmt.Errorf("instrumentation mode must be one of: code, adot_layer")
	}

	exportStrategy := resolveExportStrategy(contract)
	metricInterval := opts.MetricExportIntervalMs
	if metricInterval == 0 {
		metricInterval = defaultMetricExportIntervalMs
	}

	baseAttrs := copyMap(contract.Spec.Telemetry.ResourceAttributes)
	baseAttrs["cloud.provider"] = "aws"
	baseAttrs["faas.name"] = contract.Spec.Service.Name
	baseAttrs["faas.platform"] = "aws_lambda"
	baseAttrs["deployment.environment"] = contract.Metadata.Environment

	env := map[string]string{
		"OBSERVABILITY_OTEL_ENABLED":           boolString(contract.Spec.Telemetry.OpenTelemetry.Enabled),
		"OBSERVABILITY_EMF_COMPATIBILITY_MODE": boolString(opts.EnableEMFCompatibilityMode),
		"OTEL_EXPORTER_OTLP_PROTOCOL":          "http/protobuf",
		"OTEL_EXPORT_STRATEGY":                 exportStrategy,
		"OTEL_METRIC_EXPORT_INTERVAL_MS":       fmt.Sprintf("%d", metricInterval),
		"OTEL_PROPAGATORS":                     strings.Join(contract.Spec.Telemetry.OpenTelemetry.Propagators, ","),
		"OTEL_RESOURCE_ATTRIBUTES":             renderResourceAttributes(baseAttrs),
		"OTEL_TRACES_EXPORTER":                 exporterState(contract.Spec.Telemetry.Signals.Traces.Enabled),
		"OTEL_METRICS_EXPORTER":                exporterState(contract.Spec.Telemetry.Signals.Metrics.Enabled),
		"OTEL_LOGS_EXPORTER":                   exporterState(contract.Spec.Telemetry.Signals.Logs.Enabled),
	}

	if mode == "adot_layer" {
		env["AWS_LAMBDA_EXEC_WRAPPER"] = adotExecWrapper
	}

	outputs := BindingOutputs{
		LambdaExecWrapper:  env["AWS_LAMBDA_EXEC_WRAPPER"],
		ADOTLambdaLayerARN: strings.TrimSpace(opts.ADOTLambdaLayerARN),
	}

	requiredInputs := []RequiredInput{}
	notes := []string{}
	managedPolicies := []string{}
	layers := []string{}

	switch exportStrategy {
	case "collector":
		configureCollectorBindings(env, &outputs, opts, &requiredInputs)
		notes = append(notes,
			"Collector mode assumes a platform-managed OTLP gateway; the consumer only receives runtime endpoints and propagators.",
			"Use assets/collector/aws for collector configuration and assets/collector/managed-suite when the platform also manages a shared or dedicated gateway endpoint.",
		)
	default:
		authMode := configureDirectBindings(contract, env, &outputs, opts, &requiredInputs)
		if authMode == "sigv4" {
			notes = append(notes, "Direct AWS export uses SigV4-authenticated OTLP endpoints for X-Ray and CloudWatch metrics.")
		} else {
			notes = append(notes, "Direct mode without inferred AWS endpoints requires platform-supplied OTLP endpoints.")
		}
	}

	authMode := determineAuthMode(exportStrategy, outputs)
	if mode == "adot_layer" {
		if strings.TrimSpace(opts.ADOTLambdaLayerARN) == "" {
			requiredInputs = append(requiredInputs, RequiredInput{
				Name:        "adot_lambda_layer_arn",
				Description: "Provide the AWS Distro for OpenTelemetry Lambda layer ARN for the target region and architecture.",
			})
		} else {
			layers = append(layers, strings.TrimSpace(opts.ADOTLambdaLayerARN))
			managedPolicies = append(managedPolicies, adotManagedPolicy)
		}
		notes = append(notes, "ADOT layer mode delegates Lambda bootstrap to the platform-supplied Lambda layer.")
	}

	sort.Strings(layers)
	sort.Strings(managedPolicies)
	sort.Slice(requiredInputs, func(i, j int) bool {
		return requiredInputs[i].Name < requiredInputs[j].Name
	})

	result := &LambdaBindings{
		ContractRef: contractRef(contract),
		Platform: planner.PlatformContext{
			DeliveryMode: contract.Spec.Delivery.Mode,
			Target:       contract.Spec.Delivery.Target,
			Region:       contract.Spec.Delivery.Region,
		},
		Instrumentation: InstrumentationSpec{
			Mode:                   mode,
			ExportStrategy:         exportStrategy,
			OTLPAuthenticationMode: authMode,
			MetricExportIntervalMs: metricInterval,
			EMFCompatibilityMode:   opts.EnableEMFCompatibilityMode,
		},
		Environment:     env,
		Layers:          layers,
		ManagedPolicies: managedPolicies,
		RequiredInputs:  requiredInputs,
		Assets: AssetReferences{
			TerraformModule:           "adapter/aws/terraform/modules/lambda-otel-bindings",
			CollectorConfig:           "assets/collector/aws/collector-cloudwatch.yaml",
			CollectorThirdPartyConfig: "assets/collector/aws/collector-cloudwatch-third-party.yaml",
			ManagedSuitePath:          "assets/collector/managed-suite",
			CloudWatchDashboard:       "assets/dashboards/cloudwatch/aws-lambda-api-workload.json.tftpl",
		},
		Outputs: outputs,
		Notes:   notes,
	}

	return result, nil
}

func configureCollectorBindings(env map[string]string, outputs *BindingOutputs, opts LambdaBindingOptions, requiredInputs *[]RequiredInput) {
	base := strings.TrimSpace(opts.CollectorEndpoint)
	traces := strings.TrimSpace(opts.CollectorTracesEndpoint)
	metrics := strings.TrimSpace(opts.CollectorMetricsEndpoint)

	if base == "" && traces == "" && metrics == "" {
		*requiredInputs = append(*requiredInputs, RequiredInput{
			Name:        "collector_endpoint",
			Description: "Provide the OTLP base endpoint or per-signal OTLP endpoints of the platform-managed collector gateway.",
		})
	} else {
		if base != "" {
			env["OTEL_EXPORTER_OTLP_ENDPOINT"] = base
			outputs.OTLPBaseEndpoint = base
		}
		if traces != "" {
			env["OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"] = traces
			outputs.OTLPTracesEndpoint = traces
		}
		if metrics != "" {
			env["OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"] = metrics
			outputs.OTLPMetricsEndpoint = metrics
		}
	}
}

func configureDirectBindings(contract *v1alpha1.ObservabilityContract, env map[string]string, outputs *BindingOutputs, opts LambdaBindingOptions, requiredInputs *[]RequiredInput) string {
	base := strings.TrimSpace(opts.DirectEndpoint)
	traces := strings.TrimSpace(opts.DirectTracesEndpoint)
	metrics := strings.TrimSpace(opts.DirectMetricsEndpoint)

	if base != "" {
		env["OTEL_EXPORTER_OTLP_ENDPOINT"] = base
		outputs.OTLPBaseEndpoint = base
		return "backend-defined"
	}
	if traces != "" {
		env["OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"] = traces
		outputs.OTLPTracesEndpoint = traces
	}
	if metrics != "" {
		env["OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"] = metrics
		outputs.OTLPMetricsEndpoint = metrics
	}
	if traces != "" || metrics != "" {
		return "backend-defined"
	}

	if strings.HasPrefix(contract.Spec.Telemetry.Signals.Traces.BackendClass, "aws-") || strings.HasPrefix(contract.Spec.Telemetry.Signals.Metrics.BackendClass, "aws-") {
		traces = fmt.Sprintf("https://xray.%s.amazonaws.com/v1/traces", contract.Spec.Delivery.Region)
		metrics = fmt.Sprintf("https://monitoring.%s.amazonaws.com/v1/metrics", contract.Spec.Delivery.Region)
		env["OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"] = traces
		env["OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"] = metrics
		outputs.OTLPTracesEndpoint = traces
		outputs.OTLPMetricsEndpoint = metrics
		if contract.Spec.Telemetry.Signals.Logs.Enabled {
			*requiredInputs = append(*requiredInputs, RequiredInput{
				Name:        "logs_export_path",
				Description: "Logs are enabled in the contract, but direct AWS OTLP log export is not inferred; route logs through a collector or provide a backend-specific path.",
			})
		}
		return "sigv4"
	}

	*requiredInputs = append(*requiredInputs, RequiredInput{
		Name:        "direct_otlp_endpoint",
		Description: "Provide a direct OTLP endpoint, or per-signal traces/metrics endpoints, for direct export mode.",
	})
	return "backend-defined"
}

func determineAuthMode(exportStrategy string, outputs BindingOutputs) string {
	if exportStrategy == "collector" {
		return "collector-managed"
	}
	if strings.Contains(outputs.OTLPTracesEndpoint, ".amazonaws.com/v1/traces") || strings.Contains(outputs.OTLPMetricsEndpoint, ".amazonaws.com/v1/metrics") {
		return "sigv4"
	}
	if outputs.OTLPBaseEndpoint != "" || outputs.OTLPTracesEndpoint != "" || outputs.OTLPMetricsEndpoint != "" {
		return "backend-defined"
	}
	return "inactive"
}

func resolveExportStrategy(contract *v1alpha1.ObservabilityContract) string {
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

func contractRef(contract *v1alpha1.ObservabilityContract) planner.ContractRef {
	return planner.ContractRef{
		Name:        contract.Metadata.Name,
		Owner:       contract.Metadata.Owner,
		System:      contract.Metadata.System,
		Environment: contract.Metadata.Environment,
	}
}

func exporterState(enabled bool) string {
	if enabled {
		return "otlp"
	}
	return "none"
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func renderResourceAttributes(attrs map[string]string) string {
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, attrs[key]))
	}
	return strings.Join(pairs, ",")
}

func copyMap(source map[string]string) map[string]string {
	target := make(map[string]string, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}
