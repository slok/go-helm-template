package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/slok/go-helm-template/helm"
)

var (
	chartPath = flag.String("chart-path", "./ingress-nginx", "chart path")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	chartFS := os.DirFS(*chartPath)

	chart, err := helm.LoadChart(ctx, chartFS)
	if err != nil {
		panic(err)
	}

	result, err := helm.Template(ctx, helm.TemplateConfig{
		Chart:       chart,
		ReleaseName: "test",
		Values: map[string]interface{}{
			"commonLabels": map[string]string{
				"example-from": "go-helm-template",
			},
			"controller": map[string]interface{}{
				"autoscaling": map[string]interface{}{
					"enabled":     true,
					"maxReplicas": 42,
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
