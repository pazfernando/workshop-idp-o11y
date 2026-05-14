package metricspreset

import "sort"

type Preset struct {
	Name    string
	Metrics []string
}

var presets = map[string]Preset{
	"serverless-api": {
		Name: "serverless-api",
		Metrics: []string{
			"OrdersCreated",
			"CreateOrderLatencyMs",
		},
	},
	"kubernetes-http-service": {
		Name: "kubernetes-http-service",
		Metrics: []string{
			"HttpServerRequests",
			"PaymentAuthorizations",
		},
	},
	"monolith-business-app": {
		Name: "monolith-business-app",
		Metrics: []string{
			"CustomerSearchLatency",
			"CustomerSearchErrors",
		},
	},
}

func Lookup(name string) (Preset, bool) {
	preset, ok := presets[name]
	if !ok {
		return Preset{}, false
	}
	return preset, true
}

func SupportedMetrics(name string) ([]string, bool) {
	preset, ok := Lookup(name)
	if !ok {
		return nil, false
	}

	metrics := append([]string(nil), preset.Metrics...)
	sort.Strings(metrics)
	return metrics, true
}

func KnownPresets() []string {
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
