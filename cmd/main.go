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
	if len(os.Args[1:]) == 0 {
		fmt.Println("k8s version param is missing")
		os.Exit(1)
	}
	// parse deprecate and removed versions from k8s swagger api
	_, err := swagger.NewVersionCollector().ParseSwagger(os.Args[1:][0])
	if err != nil {
		fmt.Println("failed to Parse swagger")
		os.Exit(1)
	}
	// parse removed version from k8s deprecation mark down docs
	objs, err := markdown.NewRemovedVersion().ParseMarkDown()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// merge swagger and markdown results
	apis := collector.MergeMdSwaggerVersions(objs, map[string]collector.K8sObject{})
	// print result in a table
	tableprinter.Print(os.Stdout, apis)
}
