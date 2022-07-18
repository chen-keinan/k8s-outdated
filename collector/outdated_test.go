package collector

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapResources(t *testing.T) {
	tests := []struct {
		name       string
		mdAPI      []*K8sObject
		swaggerAPI map[string]*K8sObject
		want       []K8sAPI
	}{
		{name: "override removed version",
			mdAPI:      []*K8sObject{{Removed: "1.25", Gav: Gav{Group: "storage.k8s.io", Version: "v1beta1", Kind: "CSIStorageCapacity"}}},
			swaggerAPI: map[string]*K8sObject{"io.k8s.api.storage.k8s.io.v1beta1.CSIStorageCapacity": {Removed: "1.23", Deprecated: "1.21", Gav: Gav{Group: "storage.k8s.io", Version: "v1beta1", Kind: "CSIStorageCapacity"}}},
			want:       []K8sAPI{{DeprecatedVersion: "1.21", RemovedVersion: "1.25", API: "storage.k8s.io.v1beta1.CSIStorageCapacity"}},
		},
		{name: "append api",
			mdAPI:      []*K8sObject{{Removed: "1.25", Gav: Gav{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta1", Kind: "FlowSchema"}}},
			swaggerAPI: map[string]*K8sObject{"io.k8s.api.storage.k8s.io.v1beta1.CSIStorageCapacity": {Removed: "1.23", Deprecated: "1.21", Gav: Gav{Group: "storage.k8s.io", Version: "v1beta1", Kind: "CSIStorageCapacity"}}},
			want: []K8sAPI{{DeprecatedVersion: "", RemovedVersion: "1.25", API: "flowcontrol.apiserver.k8s.io.v1beta1.FlowSchema"},
				{DeprecatedVersion: "1.21", RemovedVersion: "1.23", API: "storage.k8s.io.v1beta1.CSIStorageCapacity"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeMdSwaggerVersions(tt.mdAPI, tt.swaggerAPI)
			for index, r := range got {
				assert.Equal(t, r.API, tt.want[index].API)
				assert.Equal(t, r.DeprecatedVersion, tt.want[index].DeprecatedVersion)
				assert.Equal(t, r.RemovedVersion, tt.want[index].RemovedVersion)
			}
		})
	}
}
