package collector

import (
	"testing"
)

func TestMapResources(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		markDownLine string
		values       []string
		mdAPI        []*K8sObject
		swaggerAPI   map[string]*K8sObject
	}{
		{name: "k8s api v1.20.1 apis", filePath: "./testdata/fixture/k8s_v1.20.1.api.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeMdSwaggerVersions(tt.mdAPI, tt.swaggerAPI)
		})
	}
}
