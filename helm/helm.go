package helm

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

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
	// ReleaseName is the name of the release.
	ReleaseName string
	// Chart is the loaded chart. Use `LoadChart`.
	Chart *Chart
	// Values are the custom values to be used on the chart template..
	Values map[string]interface{}
	// IncludeCRDs when enabled will template/render the CRDs.
	IncludeCRDs bool
	// Namespace is the namespace used to render the chart.
	Namespace string
	// ShowFiles is a list of files that can be used to only template the provided files,
	// by default it will render all.
	// This can be handy on specific use cases like unit tests for charts.
	ShowFiles []string
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

// Template will runhelm template in the provided chart and values without the need of the Helm binary
// and without executing an external command.
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

	manifests := rel.Manifest
	if len(config.ShowFiles) > 0 {
		manifests, err = filterFiles(manifests, config.ShowFiles)
		if err != nil {
			return "", fmt.Errorf("could not filter manifest files: %w", err)
		}
	}

	return manifests, nil
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

// MustLoadChart is the same as LoadChart but panics if there is
// any error while loading the chart.
func MustLoadChart(ctx context.Context, f fs.FS) *Chart {
	chart, err := LoadChart(ctx, f)
	if err != nil {
		panic(err)
	}

	return chart
}

var (
	splitMarkRe    = regexp.MustCompile("(?m)^---")
	fileMatchReFmt = `(?m)^# Source:.*%s$`
)

func filterFiles(rendered string, files []string) (string, error) {
	renderedSplit := splitMarkRe.Split(rendered, -1)
	filteredRendered := []string{}
	for _, f := range files {
		found := false
		for _, text := range renderedSplit {
			regex := fmt.Sprintf(fileMatchReFmt, f)
			match, err := regexp.MatchString(regex, text)
			if err != nil {
				return "", fmt.Errorf("could not match file: %w", err)
			}
			if match {
				found = true
				filteredRendered = append(filteredRendered, text)
				break
			}
		}
		if !found {
			return "", fmt.Errorf("no match for file: %q ", f)
		}
	}

	var b bytes.Buffer
	for _, m := range filteredRendered {
		_, _ = fmt.Fprintf(&b, "---\n%s", strings.TrimSpace(m))
	}

	return b.String(), nil
}
