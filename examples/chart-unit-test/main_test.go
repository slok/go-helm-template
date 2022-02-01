package chartunittest_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/go-helm-template/helm"
)

func TestSomeChart(t *testing.T) {
	chart, err := helm.LoadChart(context.TODO(), os.DirFS("some-chart"))
	require.NoError(t, err)

	tests := map[string]struct {
		name        string
		namespace   string
		template    string
		values      map[string]interface{}
		expErr      bool
		expDataFile string
	}{
		"A chart with default values should render correctly (configmap).": {
			name:        "test-svc",
			namespace:   "test",
			template:    "templates/configmap.yaml",
			expDataFile: "testdata/configmap-default.yaml",
		},

		"A chart with custom values should render correctly (configmap).": {
			name:      "test-svc",
			namespace: "test",
			template:  "templates/configmap.yaml",
			values: map[string]interface{}{
				"labels": map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
			},
			expDataFile: "testdata/configmap-custom.yaml",
		},

		"A chart with default values should render correctly (secret).": {
			name:        "test-svc",
			namespace:   "test",
			template:    "templates/secret.yaml",
			expDataFile: "testdata/secret-default.yaml",
		},

		"A chart with custom values should render correctly (secret).": {
			name:      "test-svc",
			namespace: "test",
			template:  "templates/secret.yaml",
			values: map[string]interface{}{
				"labels": map[string]string{
					"k3": "v3",
					"k4": "v4",
				},
			},
			expDataFile: "testdata/secret-custom.yaml",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotData, err := helm.Template(context.TODO(), helm.TemplateConfig{
				Chart:       chart,
				ReleaseName: test.name,
				Values:      test.values,
				Namespace:   test.namespace,
				ShowFiles:   []string{test.template},
			})

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				expData, err := os.ReadFile(test.expDataFile)
				require.NoError(t, err)
				expDataS := strings.TrimSpace(string(expData))

				assert.Equal(expDataS, gotData)
			}
		})
	}

}
