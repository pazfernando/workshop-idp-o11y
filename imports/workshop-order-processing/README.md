# Imported from workshop-order-processing

This folder contains the exact observability-related deployment and CI/CD assets imported from the `workshop-order-processing` consumer repository.

Purpose:
- preserve the original assets during the platform extraction
- provide source material for refactoring into reusable IDP adapters and modules
- keep a traceable boundary between inherited consumer-specific implementation and the future platform-native implementation

Imported areas:
- GitHub Actions workflows related to deploy, teardown, and CI reference
- observability contracts used by the consumer
- observability deployment documentation
- observability-specific infrastructure assets, including collector configs, suite templates, and Terraform files that currently mix application and observability concerns

Status:
- imported as-is
- not yet normalized into reusable platform modules
- should be treated as migration input, not the final target design
