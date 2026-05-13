locals {
  inferred_cloudwatch_traces_endpoint  = "https://xray.${var.aws_region}.amazonaws.com/v1/traces"
  inferred_cloudwatch_metrics_endpoint = "https://monitoring.${var.aws_region}.amazonaws.com/v1/metrics"

  resource_attributes = merge({
    "cloud.provider"         = "aws"
    "deployment.environment" = var.environment_name
    "faas.name"              = var.workload_name
    "faas.platform"          = "aws_lambda"
  }, var.resource_attributes)

  sorted_resource_attribute_keys = sort(keys(local.resource_attributes))
  rendered_resource_attributes = join(",", [
    for key in local.sorted_resource_attribute_keys : "${key}=${local.resource_attributes[key]}"
  ])

  effective_otlp_base_endpoint = var.otel_export_strategy == "collector" ? trim(var.collector_endpoint, " ") : trim(var.direct_endpoint, " ")

  effective_otlp_traces_endpoint = var.otel_export_strategy == "collector" ? (
    trim(var.collector_traces_endpoint, " ")
    ) : (
    trim(var.direct_traces_endpoint, " ") != "" ? trim(var.direct_traces_endpoint, " ") : (
      trim(var.direct_endpoint, " ") == "" ? local.inferred_cloudwatch_traces_endpoint : ""
    )
  )

  effective_otlp_metrics_endpoint = var.otel_export_strategy == "collector" ? (
    trim(var.collector_metrics_endpoint, " ")
    ) : (
    trim(var.direct_metrics_endpoint, " ") != "" ? trim(var.direct_metrics_endpoint, " ") : (
      trim(var.direct_endpoint, " ") == "" ? local.inferred_cloudwatch_metrics_endpoint : ""
    )
  )

  otlp_authentication_mode = var.otel_export_strategy == "collector" ? "collector-managed" : (
    local.effective_otlp_base_endpoint == "" && local.effective_otlp_traces_endpoint == local.inferred_cloudwatch_traces_endpoint && local.effective_otlp_metrics_endpoint == local.inferred_cloudwatch_metrics_endpoint ? "sigv4" : "backend-defined"
  )

  base_environment = {
    OBSERVABILITY_OTEL_ENABLED           = tostring(var.otel_enabled)
    OBSERVABILITY_EMF_COMPATIBILITY_MODE = tostring(var.enable_emf_compatibility_mode)
    OTEL_EXPORTER_OTLP_ENDPOINT          = local.effective_otlp_base_endpoint
    OTEL_EXPORTER_OTLP_TRACES_ENDPOINT   = local.effective_otlp_traces_endpoint
    OTEL_EXPORTER_OTLP_METRICS_ENDPOINT  = local.effective_otlp_metrics_endpoint
    OTEL_EXPORTER_OTLP_PROTOCOL          = "http/protobuf"
    OTEL_EXPORT_STRATEGY                 = var.otel_export_strategy
    OTEL_LOGS_EXPORTER                   = var.signals.logs ? "otlp" : "none"
    OTEL_METRICS_EXPORTER                = var.signals.metrics ? "otlp" : "none"
    OTEL_METRIC_EXPORT_INTERVAL_MS       = tostring(var.metric_export_interval_ms)
    OTEL_PROPAGATORS                     = join(",", var.propagators)
    OTEL_RESOURCE_ATTRIBUTES             = local.rendered_resource_attributes
    OTEL_TRACES_EXPORTER                 = var.signals.traces ? "otlp" : "none"
  }

  wrapper_environment = var.instrumentation_mode == "adot_layer" ? {
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-handler"
  } : {}

  environment = merge(local.base_environment, local.wrapper_environment, var.extra_environment)

  layers = var.instrumentation_mode == "adot_layer" && trim(var.adot_lambda_layer_arn, " ") != "" ? [trim(var.adot_lambda_layer_arn, " ")] : []

  managed_policies = var.instrumentation_mode == "adot_layer" ? [
    "arn:aws:iam::aws:policy/CloudWatchLambdaApplicationSignalsExecutionRolePolicy"
  ] : []
}
