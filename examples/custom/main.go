package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/slok/go-helm-template/helm"
)

var (
	chartPath  = flag.String("chart-path", "", "chart path")
	name       = flag.String("name", "test", "release name")
	namespace  = flag.String("namespace", "", "namespace")
	enableCRDs = flag.Bool("crds", false, "enable CRDs")
)

// Example:
// git clone git@github.com:prometheus-community/helm-charts.git
// go run ./ --name test --crds --chart-path ./helm-charts/charts/kube-prometheus-stack
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
		ReleaseName: *name,
		Namespace:   *namespace,
		IncludeCRDs: *enableCRDs,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
