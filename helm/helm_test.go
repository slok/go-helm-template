package helm_test

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/slok/go-helm-template/helm"
)

func newTestChartFS() fstest.MapFS {
	m := make(fstest.MapFS)
	m["Chart.yaml"] = &fstest.MapFile{Data: []byte("apiVersion: v2\nname: test-chart\nversion: 0.1.0")}

	return m
}

func mustLoadChart(f fs.FS) *helm.Chart {
	c, err := helm.LoadChart(context.TODO(), f)
	if err != nil {
		panic(err)
	}
	return c
}

func TestTemplate(t *testing.T) {
	tests := map[string]struct {
		chart        func() *helm.Chart
		config       helm.TemplateConfig
		expManifests string
		expErr       bool
	}{
		"Empty chart should not error.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				c := mustLoadChart(chartFS)
				return c

			},
			config:       helm.TemplateConfig{ReleaseName: "test"},
			expManifests: "",
		},

		"Simple chart should render default values.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: {{ .Values.someValue }}`)}
				c := mustLoadChart(chartFS)
				return c

			},
			config:       helm.TemplateConfig{ReleaseName: "test"},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: something\n",
		},

		"Having a regular chart, default values could be override.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: {{ .Values.someValue }}`)}
				c := mustLoadChart(chartFS)
				return c

			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				Values: map[string]interface{}{
					"someValue": "otherthing",
				},
			},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: otherthing\n",
		},

		"Having a regular chart, namespace should be set correctly.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: {{ .Release.Namespace }}`)}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				Namespace:   "somens",
			},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: somens\n",
		},

		"Having a regular chart, release name should be set correctly.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: {{ .Release.Name }}`)}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
			},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: test\n",
		},

		"Having a chart with CRDs and these disabled, it should not return the CRDs.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: something`)}
				chartFS["crds/something.yaml"] = &fstest.MapFile{Data: []byte(`this-is: a CRD`)}
				c := mustLoadChart(chartFS)
				return c
			},
			config:       helm.TemplateConfig{ReleaseName: "test"},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: something\n",
		},

		"Having a chart with CRDs and these enabled, it should return the CRDs.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: something`)}
				chartFS["crds/something.yaml"] = &fstest.MapFile{Data: []byte(`this-is: a CRD`)}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				IncludeCRDs: true,
			},
			expManifests: "---\n# Source: test-chart/crds/something.yaml\nthis-is: a CRD\n---\n# Source: test-chart/templates/something.yaml\nsomething: something\n",
		},

		"Having a chart with hooks and these disabled, it should not return the hooks.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: something`)}
				chartFS["templates/myhook.yaml"] = &fstest.MapFile{Data: []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  annotations:\n    helm.sh/hook: pre-install")}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				IncludeCRDs: true,
			},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: something\n",
		},

		"Having a chart with hooks and these enabled, it should return also the hooks.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: something`)}
				chartFS["templates/myhook.yaml"] = &fstest.MapFile{Data: []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  annotations:\n    helm.sh/hook: pre-install")}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				IncludeCRDs: true,
				EnableHooks: true,
			},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: something\n---\n# Source: test-chart/templates/myhook.yaml\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  annotations:\n    helm.sh/hook: pre-install",
		},

		"Filtering files should only return the files specified.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: something`)}
				chartFS["templates/something1.yaml"] = &fstest.MapFile{Data: []byte(`something1: something1`)}
				chartFS["templates/something2.yaml"] = &fstest.MapFile{Data: []byte(`something2: something2`)}
				chartFS["templates/something3.yaml"] = &fstest.MapFile{Data: []byte(`something3: something3`)}
				chartFS["templates/something4.yaml"] = &fstest.MapFile{Data: []byte(`something4: something4`)}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				ShowFiles: []string{
					"templates/something.yaml",
					"templates/something1.yaml",
					"templates/something3.yaml",
				},
			},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: something\n---\n# Source: test-chart/templates/something1.yaml\nsomething1: something1\n---\n# Source: test-chart/templates/something3.yaml\nsomething3: something3",
		},

		"Filtering files should only return the files specified (multiple yamls per file).": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte("something: something\n---\nsomething0: something0")}
				chartFS["templates/something1.yaml"] = &fstest.MapFile{Data: []byte(`something1: something1`)}
				chartFS["templates/something2.yaml"] = &fstest.MapFile{Data: []byte(`something2: something2`)}
				chartFS["templates/something3.yaml"] = &fstest.MapFile{Data: []byte("something3: something3\n---\nsomething31: something31\n---\nsomething32: something32")}
				chartFS["templates/something4.yaml"] = &fstest.MapFile{Data: []byte(`something4: something4`)}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				ShowFiles: []string{
					"templates/something.yaml",
					"templates/something1.yaml",
					"templates/something3.yaml",
				},
			},
			expManifests: "---\n# Source: test-chart/templates/something.yaml\nsomething: something\n---\n# Source: test-chart/templates/something.yaml\nsomething0: something0\n---\n# Source: test-chart/templates/something1.yaml\nsomething1: something1\n---\n# Source: test-chart/templates/something3.yaml\nsomething3: something3\n---\n# Source: test-chart/templates/something3.yaml\nsomething31: something31\n---\n# Source: test-chart/templates/something3.yaml\nsomething32: something32",
		},

		"Filtering missing files should fail.": {
			chart: func() *helm.Chart {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("someValue: something")}
				chartFS["templates/something.yaml"] = &fstest.MapFile{Data: []byte(`something: something`)}
				c := mustLoadChart(chartFS)
				return c
			},
			config: helm.TemplateConfig{
				ReleaseName: "test",
				ShowFiles: []string{
					"templates/something.yaml",
					"templates/something1.yaml",
				},
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			config := test.config
			config.Chart = test.chart()
			gotManifests, err := helm.Template(context.TODO(), config)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expManifests, gotManifests)
			}
		})
	}

}

func TestLoadChart(t *testing.T) {
	tests := map[string]struct {
		fs     func() fs.FS
		expErr bool
	}{
		"No chart should error.": {
			fs: func() fs.FS {
				chartFS := make(fstest.MapFS)
				chartFS["something.yaml"] = &fstest.MapFile{Data: []byte("")}

				return chartFS
			},
			expErr: true,
		},

		"Invalid chart should error.": {
			fs: func() fs.FS {
				chartFS := newTestChartFS()
				chartFS["values.yaml"] = &fstest.MapFile{Data: []byte("{[[]}}}")}

				return chartFS
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := helm.LoadChart(context.TODO(), test.fs())

			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
