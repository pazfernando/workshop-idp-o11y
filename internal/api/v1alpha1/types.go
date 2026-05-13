package v1alpha1

type ObservabilityContract struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	Spec       Spec     `json:"spec" yaml:"spec"`
	Status     *Status  `json:"status,omitempty" yaml:"status,omitempty"`
}

type Metadata struct {
	Name        string            `json:"name" yaml:"name"`
	Owner       string            `json:"owner" yaml:"owner"`
	System      string            `json:"system" yaml:"system"`
	Environment string            `json:"environment" yaml:"environment"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type Spec struct {
	Service      Service      `json:"service" yaml:"service"`
	Delivery     Delivery     `json:"delivery" yaml:"delivery"`
	Telemetry    Telemetry    `json:"telemetry" yaml:"telemetry"`
	Capabilities Capabilities `json:"capabilities" yaml:"capabilities"`
}

type Service struct {
	Name      string `json:"name" yaml:"name"`
	Runtime   string `json:"runtime" yaml:"runtime"`
	Language  string `json:"language" yaml:"language"`
	Framework string `json:"framework" yaml:"framework"`
	Tier      string `json:"tier,omitempty" yaml:"tier,omitempty"`
	Lifecycle string `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
}

type Delivery struct {
	Mode      string `json:"mode" yaml:"mode"`
	Target    string `json:"target" yaml:"target"`
	AccountID string `json:"accountId,omitempty" yaml:"accountId,omitempty"`
	Region    string `json:"region,omitempty" yaml:"region,omitempty"`
	Cluster   string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type Telemetry struct {
	OpenTelemetry      OpenTelemetryConfig `json:"openTelemetry" yaml:"openTelemetry"`
	ResourceAttributes map[string]string   `json:"resourceAttributes" yaml:"resourceAttributes"`
	Signals            Signals             `json:"signals" yaml:"signals"`
}

type OpenTelemetryConfig struct {
	Enabled             bool      `json:"enabled" yaml:"enabled"`
	SemanticConventions string    `json:"semanticConventions" yaml:"semanticConventions"`
	Propagators         []string  `json:"propagators" yaml:"propagators"`
	Sampling            *Sampling `json:"sampling,omitempty" yaml:"sampling,omitempty"`
}

type Sampling struct {
	Strategy string   `json:"strategy,omitempty" yaml:"strategy,omitempty"`
	Ratio    *float64 `json:"ratio,omitempty" yaml:"ratio,omitempty"`
}

type Signals struct {
	Traces  SignalSpec        `json:"traces" yaml:"traces"`
	Metrics MetricsSignalSpec `json:"metrics" yaml:"metrics"`
	Logs    LogsSignalSpec    `json:"logs" yaml:"logs"`
}

type SignalSpec struct {
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	BackendClass string `json:"backendClass" yaml:"backendClass"`
	Ingestion    string `json:"ingestion,omitempty" yaml:"ingestion,omitempty"`
}

type MetricsSignalSpec struct {
	SignalSpec `json:",inline" yaml:",inline"`
	Catalog    []MetricSpec `json:"catalog,omitempty" yaml:"catalog,omitempty"`
}

type LogsSignalSpec struct {
	SignalSpec `json:",inline" yaml:",inline"`
	Format     string `json:"format,omitempty" yaml:"format,omitempty"`
}

type MetricSpec struct {
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"`
	Unit        string   `json:"unit" yaml:"unit"`
	Description string   `json:"description" yaml:"description"`
	Dimensions  []string `json:"dimensions,omitempty" yaml:"dimensions,omitempty"`
}

type Capabilities struct {
	Dashboards     *DashboardCapability `json:"dashboards,omitempty" yaml:"dashboards,omitempty"`
	Alerts         *AlertCapability     `json:"alerts,omitempty" yaml:"alerts,omitempty"`
	SLOs           *SLOCapability       `json:"slos,omitempty" yaml:"slos,omitempty"`
	DataManagement *DataManagement      `json:"dataManagement,omitempty" yaml:"dataManagement,omitempty"`
}

type DashboardCapability struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Preset  string `json:"preset,omitempty" yaml:"preset,omitempty"`
}

type AlertCapability struct {
	Enabled bool        `json:"enabled" yaml:"enabled"`
	Rules   []AlertRule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

type AlertRule struct {
	Name       string `json:"name" yaml:"name"`
	Summary    string `json:"summary" yaml:"summary"`
	Severity   string `json:"severity" yaml:"severity"`
	Condition  string `json:"condition" yaml:"condition"`
	For        string `json:"for,omitempty" yaml:"for,omitempty"`
	RunbookURL string `json:"runbookUrl,omitempty" yaml:"runbookUrl,omitempty"`
}

type SLOCapability struct {
	Enabled    bool           `json:"enabled" yaml:"enabled"`
	Objectives []SLOObjective `json:"objectives,omitempty" yaml:"objectives,omitempty"`
}

type SLOObjective struct {
	Name        string  `json:"name" yaml:"name"`
	Description string  `json:"description" yaml:"description"`
	Indicator   string  `json:"indicator" yaml:"indicator"`
	Target      float64 `json:"target" yaml:"target"`
	Window      string  `json:"window" yaml:"window"`
}

type DataManagement struct {
	Retention          *Retention `json:"retention,omitempty" yaml:"retention,omitempty"`
	DataClassification string     `json:"dataClassification,omitempty" yaml:"dataClassification,omitempty"`
}

type Retention struct {
	TracesDays  int `json:"tracesDays,omitempty" yaml:"tracesDays,omitempty"`
	MetricsDays int `json:"metricsDays,omitempty" yaml:"metricsDays,omitempty"`
	LogsDays    int `json:"logsDays,omitempty" yaml:"logsDays,omitempty"`
}

type Status struct {
	Phase      string            `json:"phase,omitempty" yaml:"phase,omitempty"`
	Conditions []Condition       `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Outputs    map[string]string `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

type Condition struct {
	Type    string `json:"type" yaml:"type"`
	Status  string `json:"status" yaml:"status"`
	Reason  string `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}
