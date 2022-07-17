package swagger

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"k8s-outdated/collector"
	"k8s-outdated/utils"
	"net/http"
	"os"
	"strings"
)

const (
	k8sTagsUrl = "https://api.github.com/repos/kubernetes/kubernetes/git/refs/tags"
	baseURL    = "https://raw.githubusercontent.com/kubernetes/kubernetes"
	fileURL    = "api/openapi-spec/swagger.json"

	servedIn     = "served in"
	removedIn    = "removal in"
	deprecatedIn = "deprecated in"
)

type Ref struct {
	Ref    string `json:"ref"`
	NodeId string `json:"node_id"`
	Url    string `json:"url"`
}

type VersionCollector struct {
}

func NewVersionCollector() *VersionCollector {
	return &VersionCollector{}
}

func (vc VersionCollector) ParseSwagger(k8sVer string) (map[string]*collector.K8sObject, error) {
	r, err := http.Get(k8sTagsUrl)
	if err != nil {
		return nil, err
	}
	var refs []Ref
	err = json.NewDecoder(r.Body).Decode(&refs)
	if err != nil {
		return nil, err
	}
	v1, err := version.NewVersion(k8sVer)
	kVer := make([]string, 0)
	for _, r := range refs {
		if strings.Contains(r.Ref, "-rc") ||
			strings.Contains(r.Ref, "-alpha") ||
			strings.Contains(r.Ref, "-beta") {
			continue
		}
		v := strings.Replace(r.Ref, "refs/tags/", "", -1)
		v2, err := version.NewVersion(strings.Replace(v, "v", "", -1))
		if err != nil {
			return nil, err
		}
		if v1.LessThanOrEqual(v2) {
			kVer = append(kVer, v)
		}
	}
	vList, err := vc.fetchSwaggerVersions(kVer)
	if err != nil {
		return nil, err
	}
	return vc.versionToDetails(vList), nil
}

func (vc VersionCollector) fetchSwaggerVersions(versions []string) ([]map[string]interface{}, error) {
	swaggerVersionsData := make([]map[string]interface{}, 0)
	for _, kv := range versions {
		url := fmt.Sprintf("%s/%s/%s", baseURL, kv, fileURL)
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		var apiMap map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&apiMap)
		if err != nil {
			return nil, err
		}
		swaggerVersionsData = append(swaggerVersionsData, apiMap)
	}
	return swaggerVersionsData, nil
}

func (vc VersionCollector) versionToDetails(swaggerData []map[string]interface{}) map[string]*collector.K8sObject {
	gavMap := make(map[string]*collector.K8sObject)
	for _, data := range swaggerData {
		p := data["definitions"]
		for key, val := range p.(map[string]interface{}) {
			mval := val.(map[string]interface{})
			gav, ok := mval["x-kubernetes-group-version-kind"].(interface{})
			b, err := json.Marshal(&gav)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if !ok {
				continue
			}
			var ga []collector.Gav
			err = json.Unmarshal(b, &ga)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if len(ga) == 0 {
				continue
			}
			desc, ok := mval["description"].(string)
			if !ok {
				continue
			}
			var dep string
			var rem string
			lower := strings.ToLower(desc)
			if strings.Contains(lower, deprecatedIn) {
				dep = utils.RemovedDeprecatedVersion(lower, deprecatedIn)
			}
			if strings.Contains(lower, removedIn) {
				rem = utils.RemovedDeprecatedVersion(lower, removedIn)
			}
			if strings.Contains(lower, servedIn) {
				rem = utils.RemovedDeprecatedVersion(lower, servedIn)
			}
			object := collector.K8sObject{Description: desc, Gav: ga[0], Deprecated: dep, Removed: rem}
			if len(object.Deprecated) == 0 && len(object.Removed) == 0 {
				continue
			}
			if len(object.Gav.Kind) == 0 || len(object.Gav.Version) == 0 || len(object.Gav.Group) == 0 {
				continue
			}
			gavMap[key] = &object
		}
	}
	return gavMap
}
