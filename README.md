# go-helm-template

[![CI](https://github.com/slok/go-helm-template/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/slok/go-helm-template/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/slok/go-helm-template)](https://goreportcard.com/report/github.com/slok/go-helm-template)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/slok/go-helm-template/master/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/slok/go-helm-template)](https://github.com/slok/go-helm-template/releases/latest)

Simple, fast and easy to use Go library to run [helm template][helm-template] without the need of a [Helm] binary nor its execution as an external command.

## Features

- Simple
- Fast
- Compatible with go [`fs.FS`](https://pkg.go.dev/io/fs#FS) (Template charts from FS, embedded, memory...)
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

// Chart data in memory.
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

## Examples

- [Chart unit test](./examples/chart-unit-test): An example that shows how to use the library for chart unit testing.
- [Custom](examples/custom): An example that templates a chart with custom options (e.g CRDs).
- [Embed](examples/embed): An example that renders charts embedded in the binary using [`embed.FS`][embed-fs].
- [Memory](examples/memory): An example that templates a chart from memory.
- [simple](examples/simple): A simple way of templating a chart in the FS.


## Tradeoffs

This library doesn't support anything apart from simple `helm template`. dependencies, hooks... and similar _fancy_ features, are not supported.

## Why

One of the [Helm]'s most powerful feature (if not the most) is its template system, lots of users only use [Helm] for this usage.

Not depending on helm as a system dependency, nor requiring to execute an external command, improves the portability and performance of applications that use Helm internally.


## Some use cases

- Remove process execution for simple helm template calls.
- Control better the execution flow of rendering multiple charts.
- Embed charts in compiled binaries with [`embed.FS`][embed-fs].
- Increase Helm template speed.
- Chart unit testing.

[helm]: https://helm.sh
[helm-template]: https://helm.sh/docs/helm/helm_template/
[embed-fs]: https://pkg.go.dev/embed#FS
