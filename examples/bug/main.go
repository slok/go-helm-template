package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/slok/go-helm-template/helm"
)

var (
	chartPath = flag.String("chart-path", "./bug", "chart path")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	chartFS := os.DirFS(*chartPath)

	chart, err := helm.LoadChart(ctx, chartFS)
	if err != nil {
		panic(err)
	}
	fmt.Println()

	result, err := helm.Template(ctx, helm.TemplateConfig{
		Chart:       chart,
		ReleaseName: "test",
		Values:      map[string]any{},
		EnableHooks: true,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
