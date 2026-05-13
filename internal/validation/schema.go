package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
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

	return nil
}
