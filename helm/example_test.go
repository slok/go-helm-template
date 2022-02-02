package helm_test

import (
	"context"
	"fmt"
	"testing/fstest"

	"github.com/slok/go-helm-template/helm"
)

// Template shows a basic example of how you would use helm template by using a fake chart
// created in memory.
// To load a chart from disk you could use `os.DirFS`.
func ExampleTemplate_memory() {
	// Chart data in memory.
	const (
		chartData = `
apiVersion: v2
name: example-memory
version: 0.1.0
`
		configmap = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ printf "%s-%s" .Chart.Name .Release.Name | trunc 63 | trimSuffix "-" }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- with .Values.labels -}}
    {{ toYaml . | nindent 4 }}
    {{- end }}
data:
  something: something
`
	)

	ctx := context.Background()

	// Create chart in memory.
	// We could use `os.DirFS("./some-chart")` if the chart is in the disk.
	chartFS := make(fstest.MapFS)
	chartFS["Chart.yaml"] = &fstest.MapFile{Data: []byte(chartData)}
	chartFS["templates/configmap.yaml"] = &fstest.MapFile{Data: []byte(configmap)}

	// Load chart.
	chart, err := helm.LoadChart(ctx, chartFS)
	if err != nil {
		panic(err)
	}

	// Execute helm template.
	result, err := helm.Template(ctx, helm.TemplateConfig{
		Chart:       chart,
		ReleaseName: "test",
		Namespace:   "no-kube-system",
		Values: map[string]interface{}{
			"labels": map[string]string{
				"example-from": "go-helm-template",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result)

	// Output:
	// ---
	// # Source: example-memory/templates/configmap.yaml
	// apiVersion: v1
	// kind: ConfigMap
	// metadata:
	//   name: example-memory-test
	//   namespace: no-kube-system
	//   labels:
	//     example-from: go-helm-template
	// data:
	//   something: something
}
