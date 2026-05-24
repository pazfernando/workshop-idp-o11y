variable "aws_region" {
  description = "AWS region where the managed observability suite will be deployed."
  type        = string
  default     = "us-east-1"
}

variable "name" {
  description = "Base name used for managed-suite resources."
  type        = string
  default     = "o11y-platform"
}

variable "vpc_id" {
  description = "VPC ID where the managed observability suite will be deployed."
  type        = string
  default     = ""
}

variable "subnet_id" {
  description = "Subnet ID where the managed suite will be deployed. If empty, vpc_id must be provided so Terraform can auto-select a public subnet."
  type        = string
  default     = ""
}

variable "instance_type" {
  description = "EC2 instance type for the managed suite."
  type        = string
  default     = "t3.small"
}

variable "root_volume_size_gb" {
  description = "Root volume size in GiB for the managed-suite instance."
  type        = number
  default     = 20
}

variable "grafana_admin_password" {
  description = "Optional fixed Grafana admin password. If empty, Terraform generates one."
  type        = string
  default     = ""
  sensitive   = true
}

variable "grafana_allowed_cidrs" {
  description = "CIDR ranges allowed to access Grafana on port 3000."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "otlp_allowed_cidrs" {
  description = "CIDR ranges allowed to send OTLP traffic to Alloy on ports 4317 and 4318."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "ssh_allowed_cidrs" {
  description = "CIDR ranges allowed to access the managed suite over SSH on port 22."
  type        = list(string)
  default     = []
}

variable "dashboard_uid" {
  description = "Grafana dashboard UID used in the default managed-suite dashboard URL."
  type        = string
  default     = "platform-observability"
}

variable "dashboard_title" {
  description = "Grafana dashboard title rendered into the default managed-suite dashboard."
  type        = string
  default     = "Platform Observability"
}

variable "tags" {
  description = "Additional tags applied to managed-suite resources."
  type        = map(string)
  default     = {}
}
