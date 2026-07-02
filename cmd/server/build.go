package main

import buildinfo "github.com/Viva-Fidel/metrics-and-alerting/internal/build"

func printBuildInfo() {
	buildinfo.PrintInfo(buildVersion, buildDate, buildCommit)
}
