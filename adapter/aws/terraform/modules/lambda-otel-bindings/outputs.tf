output "environment" {
  description = "Environment variables that should be injected into the Lambda workload."
  value       = local.environment
}

output "layers" {
  description = "Lambda layers required by the selected instrumentation mode."
  value       = local.layers
}

output "managed_policies" {
  description = "Managed IAM policies recommended for the Lambda execution role."
  value       = local.managed_policies
}

output "otlp_authentication_mode" {
  description = "Authentication mode required by the effective OTLP route."
  value       = local.otlp_authentication_mode
}

output "otlp_base_endpoint" {
  description = "Effective OTLP base endpoint."
  value       = local.effective_otlp_base_endpoint
}

output "otlp_traces_endpoint" {
  description = "Effective OTLP traces endpoint."
  value       = local.effective_otlp_traces_endpoint
}

output "otlp_metrics_endpoint" {
  description = "Effective OTLP metrics endpoint."
  value       = local.effective_otlp_metrics_endpoint
}
