package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	awsadapter "github.com/example/workshop-iidp-o11y/internal/adapters/aws"
	"github.com/example/workshop-iidp-o11y/internal/planner"
	"github.com/example/workshop-iidp-o11y/internal/validation"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "validate":
		if err := runValidate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "validate failed: %v\n", err)
			os.Exit(1)
		}
	case "plan":
		if err := runPlan(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "plan failed: %v\n", err)
			os.Exit(1)
		}
	case "bindings":
		if err := runBindings(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "bindings failed: %v\n", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	file := fs.String("f", "", "Path to the observability contract YAML file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required -f contract path")
	}

	contract, err := validation.LoadContract(*file)
	if err != nil {
		return err
	}

	if err := validation.ValidateContract(contract); err != nil {
		return err
	}

	fmt.Printf("contract %q is valid\n", contract.Metadata.Name)
	return nil
}

func runPlan(args []string) error {
	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	file := fs.String("f", "", "Path to the observability contract YAML file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required -f contract path")
	}

	contract, err := validation.LoadContract(*file)
	if err != nil {
		return err
	}

	if err := validation.ValidateContract(contract); err != nil {
		return err
	}

	plan, err := planner.Build(contract)
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(payload))
	return nil
}

func runBindings(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing bindings target; expected aws-lambda")
	}

	switch args[0] {
	case "aws-lambda":
		return runAWSLambdaBindings(args[1:])
	default:
		return fmt.Errorf("unsupported bindings target %q", args[0])
	}
}

func runAWSLambdaBindings(args []string) error {
	fs := flag.NewFlagSet("bindings aws-lambda", flag.ContinueOnError)
	file := fs.String("f", "", "Path to the observability contract YAML file")
	mode := fs.String("mode", "code", "Instrumentation mode: code or adot_layer")
	adotLayerARN := fs.String("adot-layer-arn", "", "ADOT Lambda layer ARN used when mode=adot_layer")
	collectorEndpoint := fs.String("collector-endpoint", "", "Collector OTLP base endpoint")
	collectorTracesEndpoint := fs.String("collector-traces-endpoint", "", "Collector OTLP traces endpoint override")
	collectorMetricsEndpoint := fs.String("collector-metrics-endpoint", "", "Collector OTLP metrics endpoint override")
	directEndpoint := fs.String("direct-endpoint", "", "Direct OTLP base endpoint")
	directTracesEndpoint := fs.String("direct-traces-endpoint", "", "Direct OTLP traces endpoint override")
	directMetricsEndpoint := fs.String("direct-metrics-endpoint", "", "Direct OTLP metrics endpoint override")
	metricExportIntervalMs := fs.Int("metric-export-interval-ms", 10000, "Metric export interval in milliseconds")
	emfCompatibilityMode := fs.Bool("emf-compatibility-mode", true, "Keep EMF compatibility mode enabled")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required -f contract path")
	}

	contract, err := validation.LoadContract(*file)
	if err != nil {
		return err
	}

	if err := validation.ValidateContract(contract); err != nil {
		return err
	}

	bindings, err := awsadapter.BuildLambdaBindings(contract, awsadapter.LambdaBindingOptions{
		InstrumentationMode:        *mode,
		ADOTLambdaLayerARN:         *adotLayerARN,
		CollectorEndpoint:          *collectorEndpoint,
		CollectorTracesEndpoint:    *collectorTracesEndpoint,
		CollectorMetricsEndpoint:   *collectorMetricsEndpoint,
		DirectEndpoint:             *directEndpoint,
		DirectTracesEndpoint:       *directTracesEndpoint,
		DirectMetricsEndpoint:      *directMetricsEndpoint,
		MetricExportIntervalMs:     *metricExportIntervalMs,
		EnableEMFCompatibilityMode: emfCompatibilityMode,
	})
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(bindings, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(payload))
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <validate|plan> -f <contract.yaml>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s bindings aws-lambda -f <contract.yaml> [options]\n", os.Args[0])
}
