package main

import "fmt"

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	fmt.Println("Build version:", buildValue(buildVersion))
	fmt.Println("Build date:", buildValue(buildDate))
	fmt.Println("Build commit:", buildValue(buildCommit))
}

func buildValue(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
