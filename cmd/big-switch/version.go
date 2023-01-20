package main

import "fmt"

var buildTime, buildVersion string

func showVersion() {
	if buildTime != "" && buildVersion != "" {
		fmt.Printf("%s (built: %s)\n", buildVersion, buildTime)
	} else {
		fmt.Println("big-switch: dev")
	}
}
