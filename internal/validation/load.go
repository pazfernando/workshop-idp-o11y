package validation

import (
	"os"

	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
	"gopkg.in/yaml.v3"
)

func LoadContract(path string) (*v1alpha1.ObservabilityContract, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var contract v1alpha1.ObservabilityContract
	if err := yaml.Unmarshal(content, &contract); err != nil {
		return nil, err
	}

	return &contract, nil
}
