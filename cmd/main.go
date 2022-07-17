package main

import (
	"fmt"
	"github.com/lensesio/tableprinter"
	"k8s-outdated/collector"
	"k8s-outdated/collector/markdown"
	"k8s-outdated/collector/swagger"
	"os"
)

func main() {
	// parse removed version from k8s deprecation mark down docs
	objs, err := markdown.NewRemovedVersion().ParseMarkDown()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// parse deprecate and removed versions from k8s swagger api
	mDetails := swagger.NewVersionCollector().VersionToDetails(os.Args[1:][0])
	apis := collector.MergeMdSwaggerVersions(objs, mDetails)
	tableprinter.Print(os.Stdout, apis)
}
