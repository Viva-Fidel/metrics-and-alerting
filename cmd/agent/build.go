package main

import "fmt"

func printBuildInfo() {
	fmt.Println("Build version:", buildValue(buildVersion))
	fmt.Println("Build date:", buildValue(buildDate))
	fmt.Println("Build commit:", buildValue(buildCommit))
}

func buildValue(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
