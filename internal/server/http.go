package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	awsadapter "github.com/example/workshop-iidp-o11y/internal/adapters/aws"
	v1alpha1 "github.com/example/workshop-iidp-o11y/internal/api/v1alpha1"
	"github.com/example/workshop-iidp-o11y/internal/platform"
)

type ValidateRequest struct {
	Contract *v1alpha1.ObservabilityContract `json:"contract"`
}

type ValidateResponse struct {
	Valid       bool              `json:"valid"`
	ContractRef map[string]string `json:"contractRef"`
}

type AWSLambdaBindingsRequest struct {
	Contract *v1alpha1.ObservabilityContract `json:"contract"`
	Options  AWSLambdaBindingOptionsPayload  `json:"options"`
}

type AWSLambdaBindingOptionsPayload struct {
	InstrumentationMode        string `json:"instrumentationMode"`
	ADOTLambdaLayerARN         string `json:"adotLambdaLayerArn"`
	CollectorEndpoint          string `json:"collectorEndpoint"`
	CollectorTracesEndpoint    string `json:"collectorTracesEndpoint"`
	CollectorMetricsEndpoint   string `json:"collectorMetricsEndpoint"`
	DirectEndpoint             string `json:"directEndpoint"`
	DirectTracesEndpoint       string `json:"directTracesEndpoint"`
	DirectMetricsEndpoint      string `json:"directMetricsEndpoint"`
	MetricExportIntervalMs     int    `json:"metricExportIntervalMs"`
	EnableEMFCompatibilityMode *bool  `json:"enableEmfCompatibilityMode"`
}

func NewHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/v1/validate", handleValidate)
	mux.HandleFunc("/v1/plan", handlePlan)
	mux.HandleFunc("/v1/bindings/aws-lambda", handleAWSLambdaBindings)
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method))
		return
	}

	var req ValidateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := platform.ValidateContract(req.Contract); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ref, err := platform.ContractRef(req.Contract)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, ValidateResponse{
		Valid: true,
		ContractRef: map[string]string{
			"name":        ref.Name,
			"owner":       ref.Owner,
			"system":      ref.System,
			"environment": ref.Environment,
		},
	})
}

func handlePlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method))
		return
	}

	var req ValidateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	plan, err := platform.BuildPlan(req.Contract)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

func handleAWSLambdaBindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method))
		return
	}

	var req AWSLambdaBindingsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	bindings, err := platform.BuildAWSLambdaBindings(req.Contract, awsadapter.LambdaBindingOptions{
		InstrumentationMode:        req.Options.InstrumentationMode,
		ADOTLambdaLayerARN:         req.Options.ADOTLambdaLayerARN,
		CollectorEndpoint:          req.Options.CollectorEndpoint,
		CollectorTracesEndpoint:    req.Options.CollectorTracesEndpoint,
		CollectorMetricsEndpoint:   req.Options.CollectorMetricsEndpoint,
		DirectEndpoint:             req.Options.DirectEndpoint,
		DirectTracesEndpoint:       req.Options.DirectTracesEndpoint,
		DirectMetricsEndpoint:      req.Options.DirectMetricsEndpoint,
		MetricExportIntervalMs:     req.Options.MetricExportIntervalMs,
		EnableEMFCompatibilityMode: req.Options.EnableEMFCompatibilityMode,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, bindings)
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON request: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain a single JSON object")
	}

	switch payload := target.(type) {
	case *ValidateRequest:
		if payload.Contract == nil {
			return fmt.Errorf("contract is required")
		}
	case *AWSLambdaBindingsRequest:
		if payload.Contract == nil {
			return fmt.Errorf("contract is required")
		}
	}

	return nil
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
