# go-helm-template

A simple, fast and easy to use Go library to run [helm template][helm-template] helm charts without the need of a helm binary or execution of helm as external command.

## Features

- Simple
- Fast
- Compatible with go [`fs.FS`](https://pkg.go.dev/io/fs#FS)
- Testable.
- No Helm binary required.
- No external command execution from Go.
- Template specific files option.

## Getting started

```golang
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
```

## Tradeoffs

The most powerful feature of Helm is the template system. This library doesn't support dependencies, hooks... just plain charts.

## Why

This gives the ability to Go applications to use the most powerful feature of helm, without the need to have to install and executing a process. This improves the portability and speed.

Apart from this, we gain the ability to embed charts in binaries.

## Use cases

- Remove process execution for simple helm template calls.
- Control better the execution flow of rendering multiple charts.
- Embed charts in compiled binaries with [`embed.FS`](https://pkg.go.dev/embed#FS).
- Increase Helm template speed.
- Chart unit testing.

[helm]: https://helm.sh
[helm-template]: https://helm.sh/docs/helm/helm_template/
