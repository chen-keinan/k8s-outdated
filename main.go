package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/lensesio/tableprinter"
	"net/http"
	"os"
	"strings"
)

const (
	baseURL = "https://raw.githubusercontent.com/kubernetes/kubernetes"
	fileURL = "api/openapi-spec/swagger.json"
)

type K8sAPI struct {
	Kind              string `header:"k8s api"`
	DeprecatedVersion string `header:"deprecated Version"`
	RemovedVersion    string `header:"removed Version"`
}

func main() {
	k8sVer := os.Args[1:]
	kVer := getRelevantK8sVersions(k8sVer[0])
	kVer = append(kVer, "master")
	mapList := make(map[string]map[string]interface{}, 0)
	mDetails := versionToDetails(kVer, mapList)
	apis := make([]K8sAPI, 0)
	fmt.Println(fmt.Sprintf("Kubernetes Version: %s", k8sVer))
	for _, data := range mDetails {
		if len(data.Deprecated) == 0 && len(data.Removed) == 0 {
			continue
		}
		if len(data.Gav.Kind) == 0 || len(data.Gav.Version) == 0 || len(data.Gav.Group) == 0 {
			continue
		}
		apis = append(apis, K8sAPI{Kind: fmt.Sprintf("%s.%s.%s", data.Gav.Group, data.Gav.Version, data.Gav.Kind), DeprecatedVersion: data.Deprecated, RemovedVersion: data.Removed})
	}
	tableprinter.Print(os.Stdout, apis)
}

func versionToDetails(kVer []string, mapList map[string]map[string]interface{}) map[string]k8sObject {
	gavMap := make(map[string]k8sObject)
	for _, kv := range kVer {
		url := fmt.Sprintf("%s/%s/%s", baseURL, kv, fileURL)
		res, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var apiMap map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&apiMap)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		mapList[kv] = apiMap
		p := apiMap["definitions"]
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
			var ga []Gav
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
			if strings.Contains(lower, "deprecated in") {
				dIndex := strings.Index(lower, "deprecated in")
				ndes := lower[dIndex+13:]
				sndes := strings.Split(strings.TrimPrefix(ndes, " "), " ")
				dep = strings.TrimSuffix(strings.TrimSuffix(sndes[0], ","), ".")
			}
			if strings.Contains(lower, "removal in") {
				dIndex := strings.Index(lower, "removal in")
				ndes := lower[dIndex+11:]
				sndes := strings.Split(strings.TrimPrefix(ndes, " "), " ")
				rem = strings.TrimSuffix(strings.TrimSuffix(sndes[0], ","), ".")
			}
			if strings.Contains(lower, "served in") {
				dIndex := strings.Index(lower, "served in")
				ndes := lower[dIndex+11:]
				sndes := strings.Split(strings.TrimPrefix(ndes, " "), " ")
				rem = strings.TrimSuffix(strings.TrimSuffix(sndes[0], ","), ".")
			}

			object := k8sObject{Description: desc, Gav: ga[0], Deprecated: dep, Removed: rem}
			gavMap[key] = object
		}
	}
	return gavMap
}

type k8sObject struct {
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

func getRelevantK8sVersions(k8sVer string) []string {
	r, err := http.Get("https://api.github.com/repos/kubernetes/kubernetes/git/refs/tags")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var refs []Ref
	err = json.NewDecoder(r.Body).Decode(&refs)
	if err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
			os.Exit(1)
		}
		if v1.LessThanOrEqual(v2) {
			kVer = append(kVer, v)
		}
	}
	return kVer
}

type Ref struct {
	Ref    string `json:"ref"`
	NodeId string `json:"node_id"`
	Url    string `json:"url"`
}
