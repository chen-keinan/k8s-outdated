package collector

import (
	"fmt"
	"strings"
)

type K8sObject struct {
	Description string
	Deprecated  string
	Removed     string
	Gav         Gav
}

type Gav struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type K8sAPI struct {
	API               string `header:"k8s api"`
	DeprecatedVersion string `header:"deprecated Version"`
	RemovedVersion    string `header:"removed Version"`
}

func MergeMdSwaggerVersions(objs []K8sObject, mDetails map[string]K8sObject) []K8sAPI {
	for _, obj := range objs {
		definition := strings.TrimSpace(fmt.Sprintf("%s.%s.%s", obj.Gav.Group, obj.Gav.Version, obj.Gav.Kind))
		if val, ok := mDetails[definition]; ok {
			val.Removed = obj.Removed
			continue
		}
		mDetails[definition] = obj
	}
	apis := make([]K8sAPI, 0)
	for _, data := range mDetails {
		if len(data.Deprecated) == 0 && len(data.Removed) == 0 {
			continue
		}
		if len(data.Gav.Kind) == 0 || len(data.Gav.Version) == 0 || len(data.Gav.Group) == 0 {
			continue
		}
		apis = append(apis, K8sAPI{API: fmt.Sprintf("%s.%s.%s", data.Gav.Group, data.Gav.Version, data.Gav.Kind), DeprecatedVersion: data.Deprecated, RemovedVersion: data.Removed})
	}
	return apis
}
