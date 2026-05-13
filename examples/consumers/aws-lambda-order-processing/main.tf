module "otel_bindings" {
  source = "../../../adapter/aws/terraform/modules/lambda-otel-bindings"

  workload_name        = "order-processing"
  environment_name     = "dev"
  aws_region           = "us-east-1"
  otel_export_strategy = "collector"
  instrumentation_mode = "code"

  propagators = ["tracecontext", "baggage", "xray"]
  resource_attributes = {
    "service.name"      = "order-processing"
    "service.namespace" = "workshop"
    "service.version"   = "1.0.0"
  }

  collector_endpoint = "http://platform-collector.internal:4318"
  signals = {
    traces  = true
    metrics = true
    logs    = true
  }
}

output "lambda_environment" {
  value = module.otel_bindings.environment
}

output "lambda_layers" {
  value = module.otel_bindings.layers
}
