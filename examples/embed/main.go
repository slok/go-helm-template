package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"

	"github.com/slok/go-helm-template/helm"
)

//go:embed all:charts/*
// Raw embedded charts.
//
// Make sure you use Go >=1.18, so `_` prefixed files are included on embed.
// https://github.com/golang/go/issues/43854.
var embeddedCharts embed.FS
var chartName = flag.String("chart", "", "Chart name.")

func main() {
	ctx := context.Background()
	flag.Parse()

	if *chartName == "" {
		panic("chart is required")
	}

	chartFS, err := fs.Sub(embeddedCharts, fmt.Sprintf("charts/%s", *chartName))
	if err != nil {
		panic(err)
	}

	chart, err := helm.LoadChart(ctx, chartFS)
	if err != nil {
		panic(err)
	}

	result, err := helm.Template(ctx, helm.TemplateConfig{
		Chart:       chart,
		ReleaseName: "test",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
