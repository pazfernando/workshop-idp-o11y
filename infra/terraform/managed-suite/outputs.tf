output "instance_id" {
  description = "EC2 instance ID of the managed observability suite."
  value       = aws_instance.managed_suite.id
}

output "subnet_id" {
  description = "Subnet ID used by the managed suite."
  value       = data.aws_subnet.selected[0].id
}

output "security_group_id" {
  description = "Security group ID attached to the managed suite."
  value       = aws_security_group.managed_suite.id
}

output "public_ip" {
  description = "Public IP address of the managed suite."
  value       = aws_eip.managed_suite.public_ip
}

output "grafana_url" {
  description = "Grafana base URL for the managed suite."
  value       = "http://${aws_eip.managed_suite.public_ip}:3000"
}

output "grafana_dashboard_url" {
  description = "Grafana URL for the default managed-suite dashboard."
  value       = "http://${aws_eip.managed_suite.public_ip}:3000/d/${var.dashboard_uid}/${var.dashboard_uid}"
}

output "otlp_http_endpoint" {
  description = "OTLP HTTP endpoint exposed by Alloy."
  value       = "http://${aws_eip.managed_suite.public_ip}:4318"
}

output "otlp_grpc_endpoint" {
  description = "OTLP gRPC endpoint exposed by Alloy."
  value       = "${aws_eip.managed_suite.public_ip}:4317"
}

output "grafana_admin_user" {
  description = "Grafana admin username."
  value       = "admin"
}

output "grafana_admin_password" {
  description = "Grafana admin password."
  value       = local.grafana_admin_password
  sensitive   = true
}
