package helm

import (
	"context"
	"fmt"
	"io/fs"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// Chart represents a loaded Helm chart.
type Chart struct {
	c *chart.Chart
}

// TemplateConfig is the configuration for Helm Template rendering.
type TemplateConfig struct {
	ReleaseName string
	Chart       *Chart
	Values      map[string]interface{}
	IncludeCRDs bool
	Namespace   string
}

func (c *TemplateConfig) defaults() error {
	if c.Chart == nil {
		return fmt.Errorf("chart is required")
	}

	if c.ReleaseName == "" {
		return fmt.Errorf("release name is required")
	}

	if c.Values == nil {
		c.Values = map[string]interface{}{}
	}

	return nil
}

// Template will execute run helm template in the provided chart and values.
func Template(ctx context.Context, config TemplateConfig) (string, error) {
	err := config.defaults()
	if err != nil {
		return "", fmt.Errorf("invalid configuration: %w", err)
	}

	// Create chart renderer.
	client := action.NewInstall(&action.Configuration{})
	client.ClientOnly = true
	client.DryRun = true
	client.ReleaseName = config.ReleaseName
	client.IncludeCRDs = config.IncludeCRDs
	client.Namespace = config.Namespace

	// Render chart.
	rel, err := client.Run(config.Chart.c, config.Values)
	if err != nil {
		return "", fmt.Errorf("could not render helm chart correctly: %w", err)
	}

	return rel.Manifest, nil
}

// LoadChart loads a chart from a fs.FS system.
// There chart files must be at the root of the provided fs.FS.
// e.g: ./Chart.yaml, ./values.yaml ./templates/deployment.yaml...
//
// You can use `fs.Sub` as a helper tool to get the root chart.
func LoadChart(ctx context.Context, f fs.FS) (*Chart, error) {
	files := []*loader.BufferedFile{}

	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(f, path)
		if err != nil {
			return fmt.Errorf("could not read manifest %s: %w", path, err)
		}

		files = append(files, &loader.BufferedFile{
			Name: path,
			Data: data,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk chart directory: %w", err)
	}

	chart, err := loader.LoadFiles(files)
	if err != nil {
		return nil, fmt.Errorf("could not load chart from files: %w", err)
	}

	return &Chart{c: chart}, nil
}
