# go-helm-template

Simple, fast and easy to use Go library to run [helm template][helm-template] without the need of a helm binary or its execution as an external command.

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

This library doesn't support anything apart from simple `helm template`. dependencies, hooks... and similar _fancy_ features, are not supported.

## Examples

- [Chart unit test](./examples/chart-unit-test): An example that shows how to use the library for chart unit testing.
- [Custom](examples/custom): An example that templates a chart with custom options (e.g CRDs).
- [Embed](examples/embed): An example that renders charts embedded in the binary using [`embed.FS`][embed-fs].
- [Memory](examples/memory): An example that templates a chart from memory.
- [simple](examples/simple): A simple way of templating a chart in the FS.

## Why

[Helm]'s most powerful feature is its template system, lots of users only use [Helm] for this.

Having a library for this use, that doesn't depend on helm dependency on the system, nor executing an external command improves the portability and performance of applications.



## Some use cases

- Remove process execution for simple helm template calls.
- Control better the execution flow of rendering multiple charts.
- Embed charts in compiled binaries with [`embed.FS`][embed-fs].
- Increase Helm template speed.
- Chart unit testing.

[helm]: https://helm.sh
[helm-template]: https://helm.sh/docs/helm/helm_template/
[embed-fs]: https://pkg.go.dev/embed#FS
