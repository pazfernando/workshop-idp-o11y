package metriccatalog

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed presets/*.yaml
var presetFS embed.FS

type Preset struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"displayName"`
	Description string   `yaml:"description"`
	Metrics     []string `yaml:"metrics"`
}

var (
	loadOnce sync.Once
	loadErr  error
	presets  map[string]Preset
)

func Lookup(name string) (Preset, bool) {
	if err := ensureLoaded(); err != nil {
		return Preset{}, false
	}

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
	if err := ensureLoaded(); err != nil {
		return nil
	}

	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func ensureLoaded() error {
	loadOnce.Do(func() {
		entries, err := presetFS.ReadDir("presets")
		if err != nil {
			loadErr = err
			return
		}

		loaded := make(map[string]Preset, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
				continue
			}

			raw, err := presetFS.ReadFile(filepath.Join("presets", entry.Name()))
			if err != nil {
				loadErr = err
				return
			}

			var preset Preset
			if err := yaml.Unmarshal(raw, &preset); err != nil {
				loadErr = fmt.Errorf("decode metric preset %s: %w", entry.Name(), err)
				return
			}

			preset.Name = strings.TrimSpace(preset.Name)
			if preset.Name == "" {
				loadErr = fmt.Errorf("metric preset file %s is missing name", entry.Name())
				return
			}

			sanitizedMetrics := make([]string, 0, len(preset.Metrics))
			for _, metric := range preset.Metrics {
				metric = strings.TrimSpace(metric)
				if metric != "" {
					sanitizedMetrics = append(sanitizedMetrics, metric)
				}
			}
			sort.Strings(sanitizedMetrics)
			preset.Metrics = sanitizedMetrics

			loaded[preset.Name] = preset
		}

		presets = loaded
	})

	return loadErr
}
