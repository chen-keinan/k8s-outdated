package markdown

import (
	"bufio"
	"io/ioutil"
	"k8s-outdated/collector"
	"k8s-outdated/utils"
	"net/http"
	"strings"
)

const (
	willNoLongerBeServed = "will no longer be served in"
	isNoLongerServedAsOf = "is no longer served as of"
	apiVersionOf         = "API version of"
	apiVersionsOf        = "API versions of"
	apiVersions          = "API version"
	the                  = "The"

	depGuide = "https://raw.githubusercontent.com/kubernetes/website/main/content/en/docs/reference/using-api/deprecation-guide.md"
)

type RemovedVersion struct {
}

func NewRemovedVersion() *RemovedVersion {
	return &RemovedVersion{}
}

func (vz RemovedVersion) MarkdownToJson() ([]collector.K8sObject, error) {
	res, err := http.Get(depGuide)
	if err != nil {
		return nil, err
	}
	md, err := ioutil.ReadAll(res.Body)
	k8sObjects := make([]collector.K8sObject, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(md)))
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
							k8sObjects = append(k8sObjects, collector.K8sObject{Description: line, Removed: removedVersion, Gav: collector.Gav{Group: apiParts[0], Version: apiParts[1], Kind: res}})
						}
					}
				}
				k8sAPIs[removedVersion] = append(k8sAPIs[removedVersion], line)
			}
		}
	}
	return k8sObjects, nil
}

func findVersion(line string, keyWords []string) string {
	var partLine string
	for _, keyWord := range keyWords {
		partLine = utils.RemovedDeprecatedVersion(strings.ToLower(line), keyWord)
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
