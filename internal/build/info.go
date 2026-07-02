package build

import "fmt"

func PrintInfo(version, date, commit string) {
	fmt.Println("Build version:", value(version))
	fmt.Println("Build date:", value(date))
	fmt.Println("Build commit:", value(commit))
}

func value(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
