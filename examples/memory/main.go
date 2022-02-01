package main

import (
	"context"
	"fmt"
	"testing/fstest"

	"github.com/slok/go-helm-template/helm"
)

// Chart data.
const (
	chart = `
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

func main() {
	ctx := context.Background()

	// Create chart in memory.
	chartFS := make(fstest.MapFS)
	chartFS["Chart.yaml"] = &fstest.MapFile{Data: []byte(chart)}
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
}
