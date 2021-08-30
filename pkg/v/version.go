package v

import "fmt"

var (
	gitCommit  = ""
	buildTime  = ""
	goVersion  = ""
	appVersion = "1.0.0"
)

// Version returns the full version information for the application.
func Version() string {
	return fmt.Sprintf("version: %s\n", appVersion) +
		fmt.Sprintf("build:\t%s\n", buildTime) +
		fmt.Sprintf("git:\t%s\n", gitCommit) +
		fmt.Sprintf("go:\t%s\n", goVersion)
}
