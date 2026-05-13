package platform

import (
	"fmt"

	awsadapter "github.com/example/workshop-iidp-o11y/internal/adapters/aws"
	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
	"github.com/example/workshop-iidp-o11y/internal/planner"
	"github.com/example/workshop-iidp-o11y/internal/validation"
)

func ValidateContract(contract *v1alpha1.ObservabilityContract) error {
	if err := validation.ValidateContract(contract); err != nil {
		return err
	}
	return nil
}

func BuildPlan(contract *v1alpha1.ObservabilityContract) (*planner.ProvisioningPlan, error) {
	if err := ValidateContract(contract); err != nil {
		return nil, err
	}

	return planner.Build(contract)
}

func BuildAWSLambdaBindings(contract *v1alpha1.ObservabilityContract, opts awsadapter.LambdaBindingOptions) (*awsadapter.LambdaBindings, error) {
	if err := ValidateContract(contract); err != nil {
		return nil, err
	}

	return awsadapter.BuildLambdaBindings(contract, opts)
}

func ContractRef(contract *v1alpha1.ObservabilityContract) (planner.ContractRef, error) {
	if contract == nil {
		return planner.ContractRef{}, fmt.Errorf("contract is nil")
	}

	return planner.ContractRef{
		Name:        contract.Metadata.Name,
		Owner:       contract.Metadata.Owner,
		System:      contract.Metadata.System,
		Environment: contract.Metadata.Environment,
	}, nil
}
