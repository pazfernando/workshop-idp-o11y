variable "workload_name" {
  description = "Logical workload name used to derive resource attributes."
  type        = string
}

variable "environment_name" {
  description = "Platform environment name such as dev, test, staging, or prod."
  type        = string
}

variable "aws_region" {
  description = "AWS region where the workload runs."
  type        = string
}

variable "otel_enabled" {
  description = "Whether OpenTelemetry should be enabled for the workload."
  type        = bool
  default     = true
}

variable "instrumentation_mode" {
  description = "Use code for in-process SDK bootstrap or adot_layer to rely on the ADOT Lambda layer."
  type        = string
  default     = "code"

  validation {
    condition     = contains(["code", "adot_layer"], var.instrumentation_mode)
    error_message = "instrumentation_mode must be one of: code, adot_layer."
  }
}

variable "otel_export_strategy" {
  description = "Use direct to export to a backend endpoint or collector to target a platform-managed gateway."
  type        = string
  default     = "collector"

  validation {
    condition     = contains(["direct", "collector"], var.otel_export_strategy)
    error_message = "otel_export_strategy must be one of: direct, collector."
  }
}

variable "propagators" {
  description = "OpenTelemetry propagators exposed to the runtime."
  type        = list(string)
  default     = ["tracecontext", "baggage"]
}

variable "resource_attributes" {
  description = "Additional resource attributes merged into OTEL_RESOURCE_ATTRIBUTES."
  type        = map(string)
  default     = {}
}

variable "collector_endpoint" {
  description = "Collector OTLP base endpoint."
  type        = string
  default     = ""
}

variable "collector_traces_endpoint" {
  description = "Collector OTLP traces endpoint override."
  type        = string
  default     = ""
}

variable "collector_metrics_endpoint" {
  description = "Collector OTLP metrics endpoint override."
  type        = string
  default     = ""
}

variable "direct_endpoint" {
  description = "Direct OTLP base endpoint."
  type        = string
  default     = ""
}

variable "direct_traces_endpoint" {
  description = "Direct OTLP traces endpoint override."
  type        = string
  default     = ""
}

variable "direct_metrics_endpoint" {
  description = "Direct OTLP metrics endpoint override."
  type        = string
  default     = ""
}

variable "adot_lambda_layer_arn" {
  description = "ADOT Lambda layer ARN used when instrumentation_mode is adot_layer."
  type        = string
  default     = ""
}

variable "metric_export_interval_ms" {
  description = "Metric export interval in milliseconds."
  type        = number
  default     = 10000
}

variable "enable_emf_compatibility_mode" {
  description = "Whether to preserve EMF compatibility mode while OTel rollout progresses."
  type        = bool
  default     = true
}

variable "signals" {
  description = "Signal enablement flags used to set OTEL_*_EXPORTER variables."
  type = object({
    traces  = bool
    metrics = bool
    logs    = bool
  })
  default = {
    traces  = true
    metrics = true
    logs    = true
  }
}

variable "extra_environment" {
  description = "Additional environment variables appended to the generated runtime bindings."
  type        = map(string)
  default     = {}
}
