provider "aws" {
  region = var.aws_region
}

data "aws_subnets" "public" {
  filter {
    name   = "map-public-ip-on-launch"
    values = ["true"]
  }
}

data "aws_subnets" "all" {}

locals {
  selected_subnet_id = trim(var.subnet_id, " ") != "" ? trim(var.subnet_id, " ") : (
    length(data.aws_subnets.public.ids) > 0 ? sort(data.aws_subnets.public.ids)[0] : (
      length(data.aws_subnets.all.ids) > 0 ? sort(data.aws_subnets.all.ids)[0] : ""
    )
  )
}

data "aws_subnet" "selected" {
  count = local.selected_subnet_id != "" ? 1 : 0
  id    = local.selected_subnet_id
}

data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "state"
    values = ["available"]
  }
}

resource "random_password" "grafana_admin_password" {
  count   = trim(var.grafana_admin_password, " ") == "" ? 1 : 0
  length  = 20
  special = false
}

locals {
  grafana_admin_password = trim(var.grafana_admin_password, " ") != "" ? trim(var.grafana_admin_password, " ") : random_password.grafana_admin_password[0].result
  common_tags = merge(var.tags, {
    Name    = var.name
    Product = "o11y-platform"
  })
  grafana_datasources_yaml        = templatefile("${path.module}/../../observability-suite/grafana-datasources.yml.tftpl", {})
  grafana_dashboard_provider_yaml = templatefile("${path.module}/../../observability-suite/grafana-dashboard-provider.yml.tftpl", {})
  prometheus_config_yaml          = templatefile("${path.module}/../../observability-suite/prometheus.yml.tftpl", {})
  loki_config_yaml                = templatefile("${path.module}/../../observability-suite/loki-config.yml.tftpl", {})
  tempo_config_yaml               = templatefile("${path.module}/../../observability-suite/tempo-config.yml.tftpl", {})
  alloy_config = templatefile("${path.module}/../../observability-suite/alloy-config.alloy.tftpl", {
    loki_otlp_endpoint       = "http://loki:3100/otlp"
    prometheus_remote_write  = "http://prometheus:9090/api/v1/write"
    tempo_otlp_grpc_endpoint = "tempo:4319"
    deployment_environment   = "platform"
  })
  grafana_dashboard_json = templatefile("${path.module}/../../observability-suite/grafana-dashboard.json.tftpl", {
    dashboard_uid   = var.dashboard_uid
    dashboard_title = var.dashboard_title
  })
  docker_compose = templatefile("${path.module}/../../observability-suite/docker-compose.yml.tftpl", {
    grafana_admin_user_json     = jsonencode("admin")
    grafana_admin_password_json = jsonencode(local.grafana_admin_password)
    grafana_image               = "grafana/grafana:latest"
    alloy_image                 = "grafana/alloy:latest"
    loki_image                  = "grafana/loki:latest"
    tempo_image                 = "grafana/tempo:2.10.4"
    prometheus_image            = "prom/prometheus:latest"
  })
  user_data = templatefile("${path.module}/../../observability-suite/user-data.sh.tftpl", {
    compose_b64                    = base64encode(local.docker_compose)
    alloy_config_b64               = base64encode(local.alloy_config)
    prometheus_config_b64          = base64encode(local.prometheus_config_yaml)
    loki_config_b64                = base64encode(local.loki_config_yaml)
    tempo_config_b64               = base64encode(local.tempo_config_yaml)
    grafana_datasources_b64        = base64encode(local.grafana_datasources_yaml)
    grafana_dashboard_provider_b64 = base64encode(local.grafana_dashboard_provider_yaml)
    grafana_dashboard_b64          = base64encode(local.grafana_dashboard_json)
    grafana_dashboard_file_name    = "${var.dashboard_uid}.json"
  })
}

resource "aws_security_group" "managed_suite" {
  name        = "${var.name}-sg"
  description = "Security group for the platform-managed observability suite."
  vpc_id      = data.aws_subnet.selected[0].vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

resource "aws_vpc_security_group_ingress_rule" "grafana" {
  for_each = toset(var.grafana_allowed_cidrs)

  security_group_id = aws_security_group.managed_suite.id
  cidr_ipv4         = each.value
  from_port         = 3000
  to_port           = 3000
  ip_protocol       = "tcp"
  description       = "Grafana access"
}

resource "aws_vpc_security_group_ingress_rule" "otlp_http" {
  for_each = toset(var.otlp_allowed_cidrs)

  security_group_id = aws_security_group.managed_suite.id
  cidr_ipv4         = each.value
  from_port         = 4318
  to_port           = 4318
  ip_protocol       = "tcp"
  description       = "Alloy OTLP HTTP ingest"
}

resource "aws_vpc_security_group_ingress_rule" "otlp_grpc" {
  for_each = toset(var.otlp_allowed_cidrs)

  security_group_id = aws_security_group.managed_suite.id
  cidr_ipv4         = each.value
  from_port         = 4317
  to_port           = 4317
  ip_protocol       = "tcp"
  description       = "Alloy OTLP gRPC ingest"
}

resource "aws_vpc_security_group_ingress_rule" "ssh" {
  for_each = toset(var.ssh_allowed_cidrs)

  security_group_id = aws_security_group.managed_suite.id
  cidr_ipv4         = each.value
  from_port         = 22
  to_port           = 22
  ip_protocol       = "tcp"
  description       = "SSH access"
}

data "aws_iam_policy_document" "ec2_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "managed_suite" {
  name               = "${var.name}-role"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
  tags               = local.common_tags
}

resource "aws_iam_role_policy_attachment" "managed_suite_ssm" {
  role       = aws_iam_role.managed_suite.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "managed_suite" {
  name = "${var.name}-profile"
  role = aws_iam_role.managed_suite.name
}

resource "aws_instance" "managed_suite" {
  ami                         = data.aws_ami.amazon_linux_2023.id
  instance_type               = var.instance_type
  subnet_id                   = data.aws_subnet.selected[0].id
  vpc_security_group_ids      = [aws_security_group.managed_suite.id]
  iam_instance_profile        = aws_iam_instance_profile.managed_suite.name
  associate_public_ip_address = true
  user_data_base64            = base64gzip(local.user_data)
  user_data_replace_on_change = true

  lifecycle {
    precondition {
      condition     = local.selected_subnet_id != ""
      error_message = "The managed suite requires at least one usable subnet in the target account and region."
    }
    ignore_changes = [ami]
  }

  root_block_device {
    volume_size = var.root_volume_size_gb
    volume_type = "gp3"
  }

  tags = local.common_tags
}

resource "aws_eip" "managed_suite" {
  domain   = "vpc"
  instance = aws_instance.managed_suite.id
  tags     = local.common_tags
}
