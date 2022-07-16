package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/lensesio/tableprinter"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	baseURL  = "https://raw.githubusercontent.com/kubernetes/kubernetes"
	fileURL  = "api/openapi-spec/swagger.json"
	depGuide = "https://raw.githubusercontent.com/kubernetes/website/main/content/en/docs/reference/using-api/deprecation-guide.md"

	// verbs for detecting removal and deprecation
	servedIn             = "served in"
	removedIn            = "removal in"
	deprecatedIn         = "deprecated in"
	willNoLongerBeServed = "will no longer be served in"
	isNoLongerServedAsOf = "is no longer served as of"
	apiVersionOf         = "API version of"
	apiVersionsOf        = "API versions of"
	apiVersions          = "API version"
	the                  = "The"
)

type K8sAPI struct {
	API               string `header:"k8s api"`
	DeprecatedVersion string `header:"deprecated Version"`
	RemovedVersion    string `header:"removed Version"`
}

func main() {
	res, err := http.Get(depGuide)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	objs := MarkdownToJson(string(body))
	k8sVer := os.Args[1:]
	kVer := getRelevantK8sVersions(k8sVer[0])
	kVer = append(kVer, "master")
	mapList := make(map[string]map[string]interface{}, 0)
	mDetails := versionToDetails(kVer, mapList)
	for _, obj := range objs {
		definition := fmt.Sprintf("%s.%s.%s", obj.Gav.Group, obj.Gav.Version, obj.Gav.Kind)
		if val, ok := mDetails[definition]; ok {
			val.Removed = obj.Removed
			continue
		}
		mDetails[definition] = obj
	}
	apis := make([]K8sAPI, 0)
	fmt.Println(fmt.Sprintf("Kubernetes Version: %s", k8sVer))
	for _, data := range mDetails {
		if len(data.Deprecated) == 0 && len(data.Removed) == 0 {
			continue
		}
		if len(data.Gav.Kind) == 0 || len(data.Gav.Version) == 0 || len(data.Gav.Group) == 0 {
			continue
		}
		apis = append(apis, K8sAPI{API: fmt.Sprintf("%s.%s.%s", data.Gav.Group, data.Gav.Version, data.Gav.Kind), DeprecatedVersion: data.Deprecated, RemovedVersion: data.Removed})
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
			if strings.Contains(lower, deprecatedIn) {
				dep = removedDeprecatedVersion(lower, deprecatedIn)
			}
			if strings.Contains(lower, removedIn) {
				rem = removedDeprecatedVersion(lower, removedIn)
			}
			if strings.Contains(lower, servedIn) {
				rem = removedDeprecatedVersion(lower, servedIn)
			}
			object := k8sObject{Description: desc, Gav: ga[0], Deprecated: dep, Removed: rem}
			gavMap[key] = object
		}
	}
	return gavMap
}

func removedDeprecatedVersion(lower string, verb string) string {
	dIndex := strings.Index(lower, verb)
	ndes := lower[dIndex+len(verb):]
	sndes := strings.Split(strings.TrimPrefix(ndes, " "), " ")
	rem := strings.TrimSuffix(strings.TrimSuffix(sndes[0], ","), ".")
	return rem
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

func MarkdownToJson(markdown string) []k8sObject {
	k8sObjects := make([]k8sObject, 0)
	scanner := bufio.NewScanner(strings.NewReader(markdown))
	scanner.Split(bufio.ScanLines)
	var currentVersion string
	k8sAPIs := make(map[string][]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		lineWithoutSpace := strings.TrimSpace(line)
		if len(lineWithoutSpace) == 0 {
			continue
		}
		if strings.Contains(line, "### v1.") {
			currentVersion = strings.Replace(lineWithoutSpace, "###", "", -1)
			if _, ok := k8sAPIs[currentVersion]; !ok {
				k8sAPIs[currentVersion] = []string{}
			}
			continue
		}
		if _, ok := k8sAPIs[currentVersion]; ok {
			if strings.Contains(line, "VolumeAttachment") {
				fmt.Println("here")
			}
			if strings.Contains(line, willNoLongerBeServed) || strings.Contains(line, isNoLongerServedAsOf) {
				removedVersion := findVersion(line, []string{willNoLongerBeServed, isNoLongerServedAsOf})
				if len(removedVersion) == 0 {
					continue
				}
				apis := findResourcesAPI([]string{the}, []string{apiVersionOf, apiVersionsOf, apiVersions}, line, []string{"**"})
				resources := findResourcesAPI([]string{apiVersionOf, apiVersionsOf}, []string{willNoLongerBeServed, isNoLongerServedAsOf}, line, []string{",", "and"})
				for _, api := range apis {
					apiParts := strings.Split(api, "/")
					if len(apiParts) == 2 {
						for _, res := range resources {
							k8sObjects = append(k8sObjects, k8sObject{Description: line, Removed: removedVersion, Gav: Gav{Group: apiParts[0], Version: apiParts[1], Kind: res}})
						}
					}
				}
				k8sAPIs[removedVersion] = append(k8sAPIs[removedVersion], line)
			}
		}
	}
	return k8sObjects
}

func findVersion(line string, keyWords []string) string {
	var partLine string
	for _, keyWord := range keyWords {
		partLine = removedDeprecatedVersion(strings.ToLower(line), keyWord)
		if strings.HasPrefix(partLine, "v1.") {
			return partLine
		}
	}
	return ""
}

func findResourcesAPI(beginWords []string, endWords []string, line string, removedSigns []string) []string {
	resources := make([]string, 0)
	var beginWord string
	var beginIndex, endIndex int
	for _, b := range beginWords {
		beginIndex = strings.Index(line, b)
		if beginIndex == -1 {
			continue
		} else {
			beginWord = b
			break
		}
	}
	for _, e := range endWords {
		endIndex = strings.Index(line, e)
		if endIndex == -1 {
			continue
		} else {
			break
		}
	}
	if beginIndex == -1 {
		beginIndex = 0
	}
	resourceLine := line[beginIndex+len(beginWord) : endIndex]
	splitResource := strings.Split(resourceLine, " ")
	for _, r := range splitResource {
		if len(strings.TrimSpace(r)) == 0 {
			continue
		}
		for _, sign := range removedSigns {
			resourceLine = strings.Replace(r, sign, " ", -1)
		}
		resWithoutSpace := strings.TrimSpace(resourceLine)
		if len(resWithoutSpace) == 0 {
			continue
		}
		resources = append(resources, strings.TrimSpace(resWithoutSpace))
	}
	return resources
}
